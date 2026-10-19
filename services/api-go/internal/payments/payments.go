package payments

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
)

type PaymentAttemptStatus string

const (
	AttemptPending   PaymentAttemptStatus = "pending"
	AttemptSucceeded PaymentAttemptStatus = "succeeded"
	AttemptFailed    PaymentAttemptStatus = "failed"
	AttemptCancelled PaymentAttemptStatus = "cancelled"
)

var (
	ErrPaymentAlreadyCompleted = errors.New("order is already paid")
	ErrInvalidOrderState       = errors.New("cannot initiate payment for cancelled or delivered order")
	ErrProviderNotFound        = errors.New("payment provider not supported")
	ErrInvalidSignature        = errors.New("invalid webhook signature")
	ErrAmountMismatch          = errors.New("webhook amount does not match order total")
	ErrAttemptNotFound         = errors.New("payment attempt not found")
	ErrMissingSignatureHeader  = errors.New("missing webhook signature header")
)

type PaymentAttempt struct {
	ID            string               `json:"id"`
	OrderID       string               `json:"orderId"`
	CustomerID    string               `json:"customerId"`
	Provider      string               `json:"provider"`
	AmountPoisha  int64                `json:"amountPoisha"`
	Currency      string               `json:"currency"`
	Status        PaymentAttemptStatus `json:"status"`
	TransactionID string               `json:"transactionId,omitempty"`
	RedirectURL   string               `json:"redirectUrl,omitempty"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
}

type PaymentSession struct {
	PaymentAttemptID string    `json:"attemptId"`
	CheckoutURL      string    `json:"checkoutUrl"`
	Provider         string    `json:"provider"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

type WebhookEvent struct {
	EventID          string               `json:"eventId"`
	Provider         string               `json:"provider"`
	OrderID          string               `json:"orderId"`
	PaymentAttemptID string               `json:"paymentAttemptId"`
	TransactionID    string               `json:"transactionId"`
	Status           PaymentAttemptStatus `json:"status"`
	AmountPoisha     int64                `json:"amountPoisha"`
	Currency         string               `json:"currency"`
	Timestamp        time.Time            `json:"timestamp"`
}

type PaymentProvider interface {
	Name() string
	CreateSession(ctx context.Context, attempt *PaymentAttempt) (*PaymentSession, error)
	VerifyWebhook(r *http.Request, secret string, rawBody []byte) (*WebhookEvent, error)
}

type OrderProvider interface {
	GetOrder(ctx context.Context, orderID string, customerID string, isPrivileged bool) (*orders.Order, error)
	RecordPaymentSuccess(ctx context.Context, orderID string) error
	RecordPaymentFailure(ctx context.Context, orderID string) error
}

type PaymentAttemptRepository interface {
	Create(ctx context.Context, attempt *PaymentAttempt) error
	GetByID(ctx context.Context, id string) (*PaymentAttempt, error)
	GetByOrderID(ctx context.Context, orderID string) ([]PaymentAttempt, error)
	UpdateStatus(ctx context.Context, id string, status PaymentAttemptStatus, txID string) error
}

type WebhookAuditRepository interface {
	HasProcessed(ctx context.Context, provider string, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, provider string, eventID string, orderID string) error
}

type Service struct {
	providers     map[string]PaymentProvider
	orderProvider OrderProvider
	attemptRepo   PaymentAttemptRepository
	auditRepo     WebhookAuditRepository
	webhookSecret string
}

func NewService(
	orderProvider OrderProvider,
	attemptRepo PaymentAttemptRepository,
	auditRepo WebhookAuditRepository,
	webhookSecret string,
) *Service {
	s := &Service{
		providers:     make(map[string]PaymentProvider),
		orderProvider: orderProvider,
		attemptRepo:   attemptRepo,
		auditRepo:     auditRepo,
		webhookSecret: webhookSecret,
	}

	// Register default mock provider
	mock := NewMockPaymentProvider()
	s.RegisterProvider(mock)
	// Register sandbox aliases for bKash and SSLCommerz
	s.RegisterProvider(NewNamedSandboxProvider("bkash"))
	s.RegisterProvider(NewNamedSandboxProvider("sslcommerz"))

	return s
}

func (s *Service) RegisterProvider(provider PaymentProvider) {
	s.providers[provider.Name()] = provider
}

func (s *Service) InitiatePayment(ctx context.Context, customerID, orderID, providerName string) (*PaymentSession, error) {
	if providerName == "" {
		providerName = "mock"
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, ErrProviderNotFound
	}

	order, err := s.orderProvider.GetOrder(ctx, orderID, customerID, false)
	if err != nil {
		return nil, err
	}

	if order.PaymentStatus == orders.PaymentStatusPaid {
		return nil, ErrPaymentAlreadyCompleted
	}

	if order.Status == orders.StatusCancelled || order.Status == orders.StatusDelivered {
		return nil, ErrInvalidOrderState
	}

	attempt := &PaymentAttempt{
		ID:           generateAttemptID(),
		OrderID:      order.ID,
		CustomerID:   customerID,
		Provider:     providerName,
		AmountPoisha: order.TotalAmountPoisha,
		Currency:     order.Currency,
		Status:       AttemptPending,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	session, err := provider.CreateSession(ctx, attempt)
	if err != nil {
		return nil, err
	}

	attempt.RedirectURL = session.CheckoutURL
	if err := s.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) ProcessWebhook(ctx context.Context, providerName string, r *http.Request, rawBody []byte) (*WebhookEvent, bool, error) {
	provider, exists := s.providers[providerName]
	if !exists {
		return nil, false, ErrProviderNotFound
	}

	event, err := provider.VerifyWebhook(r, s.webhookSecret, rawBody)
	if err != nil {
		return nil, false, err
	}

	// Check idempotency (duplicate webhook delivery)
	alreadyProcessed, err := s.auditRepo.HasProcessed(ctx, providerName, event.EventID)
	if err != nil {
		return nil, false, err
	}
	if alreadyProcessed {
		return event, true, nil
	}

	order, err := s.orderProvider.GetOrder(ctx, event.OrderID, "", true)
	if err != nil {
		return nil, false, err
	}

	if event.AmountPoisha > 0 && event.AmountPoisha != order.TotalAmountPoisha {
		return nil, false, ErrAmountMismatch
	}

	if event.PaymentAttemptID != "" {
		_ = s.attemptRepo.UpdateStatus(ctx, event.PaymentAttemptID, event.Status, event.TransactionID)
	}

	if event.Status == AttemptSucceeded {
		if err := s.orderProvider.RecordPaymentSuccess(ctx, event.OrderID); err != nil {
			return nil, false, err
		}
	} else if event.Status == AttemptFailed {
		if err := s.orderProvider.RecordPaymentFailure(ctx, event.OrderID); err != nil {
			return nil, false, err
		}
	}

	_ = s.auditRepo.MarkProcessed(ctx, providerName, event.EventID, event.OrderID)

	return event, false, nil
}

func (s *Service) GetPaymentAttempts(ctx context.Context, orderID string) ([]PaymentAttempt, error) {
	return s.attemptRepo.GetByOrderID(ctx, orderID)
}

// MockPaymentProvider handles sandbox checkout simulation and HMAC webhook validation
type MockPaymentProvider struct {
	name string
}

func NewMockPaymentProvider() *MockPaymentProvider {
	return &MockPaymentProvider{name: "mock"}
}

func NewNamedSandboxProvider(name string) *MockPaymentProvider {
	return &MockPaymentProvider{name: name}
}

func (m *MockPaymentProvider) Name() string {
	return m.name
}

func (m *MockPaymentProvider) CreateSession(ctx context.Context, attempt *PaymentAttempt) (*PaymentSession, error) {
	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	checkoutURL := fmt.Sprintf("https://checkout.aupa.dev/%s/pay?attempt_id=%s&order_id=%s&amount=%d",
		m.name, attempt.ID, attempt.OrderID, attempt.AmountPoisha)

	return &PaymentSession{
		PaymentAttemptID: attempt.ID,
		CheckoutURL:      checkoutURL,
		Provider:         m.name,
		ExpiresAt:        expiresAt,
	}, nil
}

func (m *MockPaymentProvider) VerifyWebhook(r *http.Request, secret string, rawBody []byte) (*WebhookEvent, error) {
	sig := r.Header.Get("X-Aupa-Signature")
	if sig == "" {
		sig = r.Header.Get("X-Signature")
	}
	if sig == "" {
		return nil, ErrMissingSignatureHeader
	}

	expectedSig := GenerateSignature(rawBody, secret)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return nil, ErrInvalidSignature
	}

	var event WebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return nil, fmt.Errorf("invalid webhook json: %w", err)
	}

	if event.EventID == "" {
		return nil, errors.New("webhook payload missing eventId")
	}
	if event.OrderID == "" {
		return nil, errors.New("webhook payload missing orderId")
	}

	return &event, nil
}

// GenerateSignature produces hex-encoded HMAC-SHA256 signature of body with secret
func GenerateSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func generateAttemptID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("pay_%s", hex.EncodeToString(b))
}

// MemoryPaymentAttemptRepository in-memory double
type MemoryPaymentAttemptRepository struct {
	mu       sync.RWMutex
	attempts map[string]*PaymentAttempt
}

func NewMemoryPaymentAttemptRepository() *MemoryPaymentAttemptRepository {
	return &MemoryPaymentAttemptRepository{
		attempts: make(map[string]*PaymentAttempt),
	}
}

func (m *MemoryPaymentAttemptRepository) Create(ctx context.Context, attempt *PaymentAttempt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attempts[attempt.ID] = attempt
	return nil
}

func (m *MemoryPaymentAttemptRepository) GetByID(ctx context.Context, id string) (*PaymentAttempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	att, exists := m.attempts[id]
	if !exists {
		return nil, ErrAttemptNotFound
	}
	return att, nil
}

func (m *MemoryPaymentAttemptRepository) GetByOrderID(ctx context.Context, orderID string) ([]PaymentAttempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []PaymentAttempt
	for _, att := range m.attempts {
		if att.OrderID == orderID {
			list = append(list, *att)
		}
	}
	return list, nil
}

func (m *MemoryPaymentAttemptRepository) UpdateStatus(ctx context.Context, id string, status PaymentAttemptStatus, txID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	att, exists := m.attempts[id]
	if !exists {
		return ErrAttemptNotFound
	}
	att.Status = status
	if txID != "" {
		att.TransactionID = txID
	}
	att.UpdatedAt = time.Now().UTC()
	return nil
}

// MemoryWebhookAuditRepository in-memory double
type MemoryWebhookAuditRepository struct {
	mu        sync.RWMutex
	processed map[string]string // provider:eventId -> orderId
}

func NewMemoryWebhookAuditRepository() *MemoryWebhookAuditRepository {
	return &MemoryWebhookAuditRepository{
		processed: make(map[string]string),
	}
}

func (m *MemoryWebhookAuditRepository) HasProcessed(ctx context.Context, provider string, eventID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", provider, eventID)
	_, exists := m.processed[key]
	return exists, nil
}

func (m *MemoryWebhookAuditRepository) MarkProcessed(ctx context.Context, provider string, eventID string, orderID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", provider, eventID)
	m.processed[key] = orderID
	return nil
}
