package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Role defines authorization role boundaries.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleMerchant Role = "merchant"
	RoleRider    Role = "rider"
	RoleAdmin    Role = "admin"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrSessionRevoked      = errors.New("session has been revoked")
	ErrTokenReplayDetected = errors.New("token reuse/replay detected")
	ErrWeakPassword        = errors.New("password must be at least 8 characters long")
	ErrInvalidEmail        = errors.New("invalid email address")
)

// User represents an identity in the Aupa system.
type User struct {
	ID           string    `json:"id"`
	FullName     string    `json:"fullName"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Phone        string    `json:"phone,omitempty"`
	Role         Role      `json:"role"`
	IsVerified   bool      `json:"isVerified"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TokenPair contains bearer credentials issued to client.
type TokenPair struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	ExpiresInSeconds int64  `json:"expiresInSeconds"`
}

// CustomClaims encapsulates verified JWT claims.
type CustomClaims struct {
	jwt.RegisteredClaims
	Role      string `json:"role"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
}

// Session tracks revocable refresh token families.
type Session struct {
	ID        string
	UserID    string
	TokenHash string
	FamilyID  string
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

// UserRepository defines storage contracts for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateVerification(ctx context.Context, id string, verified bool) error
}

// SessionRepository defines storage contracts for refresh sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	Revoke(ctx context.Context, sessionID string) error
	RevokeFamily(ctx context.Context, familyID string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// Service provides custom authentication, token issuance, and session management.
type Service struct {
	userRepo        UserRepository
	sessionRepo     SessionRepository
	jwtSecret       []byte
	jwtIssuer       string
	jwtAudience     string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewService instantiates an AuthService.
func NewService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	jwtSecret string,
	jwtIssuer string,
	jwtAudience string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		jwtSecret:       []byte(jwtSecret),
		jwtIssuer:       jwtIssuer,
		jwtAudience:     jwtAudience,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Register validates inputs, hashes password, and creates user identity.
func (s *Service) Register(ctx context.Context, fullName, email, password, phone string, role Role) (*User, *TokenPair, error) {
	trimmedName := strings.TrimSpace(fullName)
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	if trimmedName == "" {
		return nil, nil, errors.New("full name cannot be empty")
	}
	if !strings.Contains(normalizedEmail, "@") || len(normalizedEmail) < 5 {
		return nil, nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, nil, ErrWeakPassword
	}

	if role == "" {
		role = RoleCustomer
	}

	existing, _ := s.userRepo.GetByEmail(ctx, normalizedEmail)
	if existing != nil {
		return nil, nil, ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &User{
		ID:           generateID("usr_"),
		FullName:     trimmedName,
		Email:        normalizedEmail,
		PasswordHash: string(hash),
		Phone:        strings.TrimSpace(phone),
		Role:         role,
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Login validates user credentials and issues token pair.
func (s *Service) Login(ctx context.Context, email, password string) (*User, *TokenPair, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.GetByEmail(ctx, normalizedEmail)
	if err != nil || user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// RefreshToken rotates refresh token and issues a new token pair.
// If an already revoked token is submitted, all sessions in the family are revoked (anti-replay).
func (s *Service) RefreshToken(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(rawRefreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return nil, ErrInvalidToken
	}

	if session.IsRevoked {
		// Possible token theft / replay attack: revoke entire session family!
		_ = s.sessionRepo.RevokeFamily(ctx, session.FamilyID)
		return nil, ErrTokenReplayDetected
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	// Invalidate current refresh token
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return nil, err
	}

	// Issue new token pair preserving family ID
	return s.issueRotatedTokenPair(ctx, user, session.FamilyID)
}

// Logout revokes the given refresh token session.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}
	tokenHash := hashToken(rawRefreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return nil
	}
	return s.sessionRepo.Revoke(ctx, session.ID)
}

// VerifyAccessToken verifies signature, claims, audience, issuer, and expiration.
func (s *Service) VerifyAccessToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	}, jwt.WithIssuer(s.jwtIssuer), jwt.WithAudience(s.jwtAudience))

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || claims.TokenType != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) issueTokenPair(ctx context.Context, user *User) (*TokenPair, error) {
	familyID := generateID("fam_")
	return s.issueRotatedTokenPair(ctx, user, familyID)
}

func (s *Service) issueRotatedTokenPair(ctx context.Context, user *User, familyID string) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExpiry := now.Add(s.accessTokenTTL)

	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    s.jwtIssuer,
			Audience:  jwt.ClaimStrings{s.jwtAudience},
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Role:      string(user.Role),
		Email:     user.Email,
		TokenType: "access",
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := accessTokenObj.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	rawRefreshToken := generateSecureToken()
	refreshExpiry := now.Add(s.refreshTokenTTL)

	session := &Session{
		ID:        generateID("ses_"),
		UserID:    user.ID,
		TokenHash: hashToken(rawRefreshToken),
		FamilyID:  familyID,
		IsRevoked: false,
		ExpiresAt: refreshExpiry,
		CreatedAt: now,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save refresh session: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     rawRefreshToken,
		TokenType:        "Bearer",
		ExpiresInSeconds: int64(s.accessTokenTTL.Seconds()),
	}, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateID(prefix string) string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func generateSecureToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// MemoryUserRepository provides in-memory user repository test double.
type MemoryUserRepository struct {
	mu      sync.RWMutex
	users   map[string]*User
	byEmail map[string]*User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:   make(map[string]*User),
		byEmail: make(map[string]*User),
	}
}

func (m *MemoryUserRepository) Create(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.byEmail[user.Email]; exists {
		return ErrUserAlreadyExists
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *MemoryUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *MemoryUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, exists := m.byEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *MemoryUserRepository) UpdateVerification(ctx context.Context, id string, verified bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, exists := m.users[id]
	if !exists {
		return ErrUserNotFound
	}
	u.IsVerified = verified
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// MemorySessionRepository provides in-memory session repository test double.
type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*Session // tokenHash -> Session
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[string]*Session),
	}
}

func (m *MemorySessionRepository) Create(ctx context.Context, s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *MemorySessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, exists := m.sessions[tokenHash]
	if !exists {
		return nil, nil
	}
	return s, nil
}

func (m *MemorySessionRepository) Revoke(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sessions {
		if s.ID == sessionID {
			s.IsRevoked = true
			return nil
		}
	}
	return nil
}

func (m *MemorySessionRepository) RevokeFamily(ctx context.Context, familyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sessions {
		if s.FamilyID == familyID {
			s.IsRevoked = true
		}
	}
	return nil
}

func (m *MemorySessionRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sessions {
		if s.UserID == userID {
			s.IsRevoked = true
		}
	}
	return nil
}
