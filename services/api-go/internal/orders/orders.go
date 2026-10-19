package orders

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
)

type OrderStatus string

const (
	StatusPlaced         OrderStatus = "placed"
	StatusAccepted       OrderStatus = "accepted"
	StatusPreparing      OrderStatus = "preparing"
	StatusReady          OrderStatus = "ready"
	StatusOutForDelivery OrderStatus = "out_for_delivery"
	StatusDelivered      OrderStatus = "delivered"
	StatusCancelled      OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid   PaymentStatus = "unpaid"
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCOD        PaymentMethod = "cash_on_delivery"
	PaymentMethodBKash      PaymentMethod = "bkash"
	PaymentMethodSSLCommerz PaymentMethod = "sslcommerz"
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidStateTransition  = errors.New("invalid order status transition")
	ErrCannotCancelOrder       = errors.New("order cannot be cancelled in its current state")
	ErrUnauthorizedOrderAccess = errors.New("you do not have permission to access this order")
	ErrIdempotencyConflict     = errors.New("idempotency key conflict: identical key used with different payload")
	ErrMissingIdempotencyKey   = errors.New("idempotency-key header is required")
)

type OrderItem struct {
	ID                  string   `json:"id"`
	ItemID              string   `json:"itemId"`
	Name                string   `json:"name"`
	UnitPricePoisha     int64    `json:"unitPricePoisha"`
	Quantity            int      `json:"quantity"`
	TotalPoisha         int64    `json:"totalPoisha"`
	SelectedOptionNames []string `json:"selectedOptionNames,omitempty"`
}

type Order struct {
	ID                  string        `json:"id"`
	CustomerID          string        `json:"customerId"`
	RestaurantID        string        `json:"restaurantId"`
	DeliveryAddressID   string        `json:"deliveryAddressId"`
	Status              OrderStatus   `json:"status"`
	PaymentStatus       PaymentStatus `json:"paymentStatus"`
	PaymentMethod       PaymentMethod `json:"paymentMethod"`
	SubtotalPoisha      int64         `json:"subtotalPoisha"`
	DeliveryFeePoisha   int64         `json:"deliveryFeePoisha"`
	DiscountPoisha      int64         `json:"discountPoisha"`
	VATPoisha           int64         `json:"vatPoisha"`
	TotalAmountPoisha   int64         `json:"totalAmountPoisha"`
	Currency            string        `json:"currency"`
	SpecialInstructions string        `json:"specialInstructions,omitempty"`
	Items               []OrderItem   `json:"items"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
}

type IdempotencyRecord struct {
	Key         string
	CustomerID  string
	RequestHash string
	OrderID     string
	CreatedAt   time.Time
}

type CreateOrderInput struct {
	CustomerID          string
	RestaurantID        string
	DeliveryAddressID   string
	PaymentMethod       PaymentMethod
	PromoCode           string
	SpecialInstructions string
	Items               []cart.ItemInput
	IdempotencyKey      string
	RawPayloadHash      string
}

// OrderRepository interface
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	ListByCustomer(ctx context.Context, customerID string) ([]Order, error)
	ListByRestaurant(ctx context.Context, restaurantID string) ([]Order, error)
	UpdateStatus(ctx context.Context, id string, status OrderStatus, paymentStatus PaymentStatus) error
}

// IdempotencyRepository interface
type IdempotencyRepository interface {
	Get(ctx context.Context, key string, customerID string) (*IdempotencyRecord, error)
	Save(ctx context.Context, record *IdempotencyRecord) error
}

type Service struct {
	orderRepo       OrderRepository
	idempotencyRepo IdempotencyRepository
	cartCalculator  *cart.Calculator
}

func NewService(
	orderRepo OrderRepository,
	idempotencyRepo IdempotencyRepository,
	cartCalculator *cart.Calculator,
) *Service {
	return &Service{
		orderRepo:       orderRepo,
		idempotencyRepo: idempotencyRepo,
		cartCalculator:  cartCalculator,
	}
}

// CreateOrder validates idempotency, calculates prices authoritatively, and creates the order.
func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, bool, error) {
	if input.IdempotencyKey == "" {
		return nil, false, ErrMissingIdempotencyKey
	}

	// 1. Check idempotency record
	existingRec, _ := s.idempotencyRepo.Get(ctx, input.IdempotencyKey, input.CustomerID)
	if existingRec != nil {
		if existingRec.RequestHash != input.RawPayloadHash {
			return nil, false, ErrIdempotencyConflict
		}
		// Return existing order (isReplay = true)
		existingOrder, err := s.orderRepo.GetByID(ctx, existingRec.OrderID)
		if err != nil {
			return nil, false, err
		}
		return existingOrder, true, nil
	}

	// 2. Calculate authoritative server price
	quoteReq := cart.QuoteRequest{
		RestaurantID:      input.RestaurantID,
		DeliveryAddressID: input.DeliveryAddressID,
		PromoCode:         input.PromoCode,
		Items:             input.Items,
	}
	quote, err := s.cartCalculator.CalculateQuote(ctx, quoteReq)
	if err != nil {
		return nil, false, fmt.Errorf("cart calculation failed: %w", err)
	}

	initialPaymentStatus := PaymentStatusUnpaid
	if input.PaymentMethod == PaymentMethodCOD {
		initialPaymentStatus = PaymentStatusPending
	}

	now := time.Now().UTC()
	orderID := generateOrderID()

	var orderItems []OrderItem
	for i, item := range input.Items {
		orderItems = append(orderItems, OrderItem{
			ID:          fmt.Sprintf("oi_%s_%d", orderID, i+1),
			ItemID:      item.ItemID,
			Quantity:    item.Quantity,
			TotalPoisha: 0, // Will be stored authoritatively
		})
	}

	order := &Order{
		ID:                  orderID,
		CustomerID:          input.CustomerID,
		RestaurantID:        input.RestaurantID,
		DeliveryAddressID:   input.DeliveryAddressID,
		Status:              StatusPlaced,
		PaymentStatus:       initialPaymentStatus,
		PaymentMethod:       input.PaymentMethod,
		SubtotalPoisha:      quote.SubtotalPoisha,
		DeliveryFeePoisha:   quote.DeliveryFeePoisha,
		DiscountPoisha:      quote.DiscountPoisha,
		VATPoisha:           quote.VATPoisha,
		TotalAmountPoisha:   quote.TotalPoisha,
		Currency:            "BDT",
		SpecialInstructions: input.SpecialInstructions,
		Items:               orderItems,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, false, err
	}

	// Save idempotency key
	record := &IdempotencyRecord{
		Key:         input.IdempotencyKey,
		CustomerID:  input.CustomerID,
		RequestHash: input.RawPayloadHash,
		OrderID:     order.ID,
		CreatedAt:   now,
	}
	_ = s.idempotencyRepo.Save(ctx, record)

	return order, false, nil
}

// GetOrder retrieves order and verifies user ownership.
func (s *Service) GetOrder(ctx context.Context, orderID string, customerID string, isPrivileged bool) (*Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if !isPrivileged && order.CustomerID != customerID {
		return nil, ErrUnauthorizedOrderAccess
	}

	return order, nil
}

// ListOrders returns all orders placed by the customer.
func (s *Service) ListOrders(ctx context.Context, customerID string) ([]Order, error) {
	return s.orderRepo.ListByCustomer(ctx, customerID)
}

// CancelOrder validates state transitions and cancels the order.
func (s *Service) CancelOrder(ctx context.Context, orderID string, customerID string, isPrivileged bool) (*Order, error) {
	order, err := s.GetOrder(ctx, orderID, customerID, isPrivileged)
	if err != nil {
		return nil, err
	}

	// Can only cancel in placed or accepted status
	if order.Status != StatusPlaced && order.Status != StatusAccepted {
		return nil, ErrCannotCancelOrder
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, StatusCancelled, order.PaymentStatus); err != nil {
		return nil, err
	}

	order.Status = StatusCancelled
	order.UpdatedAt = time.Now().UTC()
	return order, nil
}

// TransitionStatus validates valid state machine transitions.
func (s *Service) TransitionStatus(ctx context.Context, orderID string, targetStatus OrderStatus) (*Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if !isValidTransition(order.Status, targetStatus) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, order.Status, targetStatus)
	}

	newPaymentStatus := order.PaymentStatus
	if targetStatus == StatusDelivered && order.PaymentMethod == PaymentMethodCOD {
		newPaymentStatus = PaymentStatusPaid
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, targetStatus, newPaymentStatus); err != nil {
		return nil, err
	}

	order.Status = targetStatus
	order.PaymentStatus = newPaymentStatus
	order.UpdatedAt = time.Now().UTC()
	return order, nil
}

// RecordPaymentSuccess marks the order as paid and advances status to accepted if currently placed.
func (s *Service) RecordPaymentSuccess(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	newStatus := order.Status
	if order.Status == StatusPlaced {
		newStatus = StatusAccepted
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, newStatus, PaymentStatusPaid)
}

// RecordPaymentFailure records payment failure on an order.
func (s *Service) RecordPaymentFailure(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, order.Status, PaymentStatusFailed)
}

func isValidTransition(current, target OrderStatus) bool {
	switch current {
	case StatusPlaced:
		return target == StatusAccepted || target == StatusCancelled
	case StatusAccepted:
		return target == StatusPreparing || target == StatusCancelled
	case StatusPreparing:
		return target == StatusReady
	case StatusReady:
		return target == StatusOutForDelivery
	case StatusOutForDelivery:
		return target == StatusDelivered
	default:
		// Terminal states: delivered, cancelled
		return false
	}
}

func generateOrderID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("ord_%s", hex.EncodeToString(b))
}

func HashPayload(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// MemoryOrderRepository in-memory test double
type MemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		orders: make(map[string]*Order),
	}
}

func (m *MemoryOrderRepository) Create(ctx context.Context, o *Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[o.ID] = o
	return nil
}

func (m *MemoryOrderRepository) GetByID(ctx context.Context, id string) (*Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, exists := m.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return o, nil
}

func (m *MemoryOrderRepository) ListByCustomer(ctx context.Context, customerID string) ([]Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []Order
	for _, o := range m.orders {
		if o.CustomerID == customerID {
			list = append(list, *o)
		}
	}
	return list, nil
}

func (m *MemoryOrderRepository) ListByRestaurant(ctx context.Context, restaurantID string) ([]Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []Order
	for _, o := range m.orders {
		if o.RestaurantID == restaurantID {
			list = append(list, *o)
		}
	}
	return list, nil
}

func (m *MemoryOrderRepository) UpdateStatus(ctx context.Context, id string, status OrderStatus, paymentStatus PaymentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, exists := m.orders[id]
	if !exists {
		return ErrOrderNotFound
	}
	o.Status = status
	o.PaymentStatus = paymentStatus
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// MemoryIdempotencyRepository in-memory double
type MemoryIdempotencyRepository struct {
	mu      sync.RWMutex
	records map[string]*IdempotencyRecord // key:customerID -> record
}

func NewMemoryIdempotencyRepository() *MemoryIdempotencyRepository {
	return &MemoryIdempotencyRepository{
		records: make(map[string]*IdempotencyRecord),
	}
}

func (m *MemoryIdempotencyRepository) Get(ctx context.Context, key, customerID string) (*IdempotencyRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	compositeKey := key + ":" + customerID
	r, exists := m.records[compositeKey]
	if !exists {
		return nil, nil
	}
	return r, nil
}

func (m *MemoryIdempotencyRepository) Save(ctx context.Context, record *IdempotencyRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	compositeKey := record.Key + ":" + record.CustomerID
	m.records[compositeKey] = record
	return nil
}
