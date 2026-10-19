package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func setupAuthService() *Service {
	userRepo := NewMemoryUserRepository()
	sessionRepo := NewMemorySessionRepository()

	return NewService(
		userRepo,
		sessionRepo,
		"super-secret-test-key-must-be-long-enough-32-chars",
		"aupa.com.bd",
		"aupa-client",
		15*time.Minute,
		24*time.Hour,
	)
}

func TestRegister(t *testing.T) {
	svc := setupAuthService()
	ctx := context.Background()

	t.Run("successful customer registration", func(t *testing.T) {
		user, tokens, err := svc.Register(ctx, "Rahim Ahmed", "rahim@example.com", "SecureP@ss123", "+8801712345678", RoleCustomer)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if user.FullName != "Rahim Ahmed" || user.Email != "rahim@example.com" {
			t.Errorf("unexpected user details: %v", user)
		}
		if user.Role != RoleCustomer {
			t.Errorf("expected role customer, got %s", user.Role)
		}
		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Errorf("expected access and refresh tokens")
		}

		// Verify issued access token
		claims, err := svc.VerifyAccessToken(tokens.AccessToken)
		if err != nil {
			t.Fatalf("expected valid token, got %v", err)
		}
		if claims.Subject != user.ID || claims.Role != "customer" {
			t.Errorf("claims mismatch: %v", claims)
		}
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		_, _, err := svc.Register(ctx, "Duplicate", "rahim@example.com", "SecureP@ss123", "", RoleCustomer)
		if err != ErrUserAlreadyExists {
			t.Errorf("expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("weak password fails", func(t *testing.T) {
		_, _, err := svc.Register(ctx, "Short Pass", "short@example.com", "pass", "", RoleCustomer)
		if err != ErrWeakPassword {
			t.Errorf("expected ErrWeakPassword, got %v", err)
		}
	})

	t.Run("invalid email fails", func(t *testing.T) {
		_, _, err := svc.Register(ctx, "Bad Email", "notanemail", "SecureP@ss123", "", RoleCustomer)
		if err != ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})
}

func TestLogin(t *testing.T) {
	svc := setupAuthService()
	ctx := context.Background()

	_, _, err := svc.Register(ctx, "Karim Ullah", "karim@example.com", "ValidP@ss123", "", RoleCustomer)
	if err != nil {
		t.Fatalf("failed setup register: %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		user, tokens, err := svc.Login(ctx, "karim@example.com", "ValidP@ss123")
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if user.Email != "karim@example.com" {
			t.Errorf("expected user karim@example.com, got %s", user.Email)
		}
		if tokens.AccessToken == "" {
			t.Errorf("expected access token")
		}
	})

	t.Run("wrong password returns ErrInvalidCredentials", func(t *testing.T) {
		_, _, err := svc.Login(ctx, "karim@example.com", "WrongPassword!")
		if err != ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("non-existent email returns ErrInvalidCredentials", func(t *testing.T) {
		_, _, err := svc.Login(ctx, "nobody@example.com", "ValidP@ss123")
		if err != ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestTokenVerificationAndTampering(t *testing.T) {
	svc := setupAuthService()
	ctx := context.Background()

	_, tokens, err := svc.Register(ctx, "Tamper Test", "tamper@example.com", "ValidP@ss123", "", RoleCustomer)
	if err != nil {
		t.Fatalf("failed register: %v", err)
	}

	t.Run("tampered token fails verification", func(t *testing.T) {
		tampered := tokens.AccessToken + "tampered"
		_, err := svc.VerifyAccessToken(tampered)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken for tampered token, got %v", err)
		}
	})

	t.Run("expired token fails verification", func(t *testing.T) {
		shortLivedSvc := NewService(
			NewMemoryUserRepository(),
			NewMemorySessionRepository(),
			"super-secret-test-key-must-be-long-enough-32-chars",
			"aupa.com.bd",
			"aupa-client",
			1*time.Millisecond,
			1*time.Hour,
		)
		_, shortTokens, _ := shortLivedSvc.Register(ctx, "Fast Expiry", "fast@example.com", "ValidP@ss123", "", RoleCustomer)

		time.Sleep(10 * time.Millisecond)

		_, err := shortLivedSvc.VerifyAccessToken(shortTokens.AccessToken)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
		}
	})

	t.Run("token with mismatched secret fails verification", func(t *testing.T) {
		otherSvc := NewService(
			NewMemoryUserRepository(),
			NewMemorySessionRepository(),
			"completely-different-signing-secret-key-32-bytes",
			"aupa.com.bd",
			"aupa-client",
			15*time.Minute,
			24*time.Hour,
		)

		_, err := otherSvc.VerifyAccessToken(tokens.AccessToken)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken with different signing secret, got %v", err)
		}
	})
}

func TestRefreshTokenRotationAndAntiReplay(t *testing.T) {
	svc := setupAuthService()
	ctx := context.Background()

	_, initialTokens, err := svc.Register(ctx, "Rotation Test", "rotate@example.com", "ValidP@ss123", "", RoleCustomer)
	if err != nil {
		t.Fatalf("failed register: %v", err)
	}

	t.Run("successful refresh rotates token", func(t *testing.T) {
		newTokens, err := svc.RefreshToken(ctx, initialTokens.RefreshToken)
		if err != nil {
			t.Fatalf("expected successful refresh, got %v", err)
		}

		if newTokens.RefreshToken == initialTokens.RefreshToken {
			t.Errorf("expected new refresh token to be different from previous")
		}

		// Replay attack simulation: attempting to use the old refresh token again
		_, replayErr := svc.RefreshToken(ctx, initialTokens.RefreshToken)
		if replayErr != ErrTokenReplayDetected {
			t.Fatalf("expected ErrTokenReplayDetected on reused refresh token, got %v", replayErr)
		}

		// Entire family must now be revoked: the newly rotated token should also fail!
		_, familyRevokedErr := svc.RefreshToken(ctx, newTokens.RefreshToken)
		if familyRevokedErr != ErrTokenReplayDetected {
			t.Errorf("expected family revocation to reject new token, got %v", familyRevokedErr)
		}
	})
}

func TestLogout(t *testing.T) {
	svc := setupAuthService()
	ctx := context.Background()

	_, tokens, err := svc.Register(ctx, "Logout User", "logout@example.com", "ValidP@ss123", "", RoleCustomer)
	if err != nil {
		t.Fatalf("failed register: %v", err)
	}

	err = svc.Logout(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected clean logout, got %v", err)
	}

	// Attempting refresh after logout fails
	_, refreshErr := svc.RefreshToken(ctx, tokens.RefreshToken)
	if refreshErr != ErrTokenReplayDetected && refreshErr != ErrInvalidToken {
		t.Errorf("expected failure after logout, got %v", refreshErr)
	}
}

// Ensure jwt package is used
var _ = jwt.NewNumericDate
