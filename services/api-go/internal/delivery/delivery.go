package delivery

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
)

var (
	ErrTrackingNotFound     = errors.New("tracking information not found")
	ErrUnauthorizedTracking = errors.New("you do not have permission to track this order")
)

type DeliveryLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

type RiderInfo struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Phone           string            `json:"phone"`
	VehicleInfo     string            `json:"vehicleInfo"`
	CurrentLocation *DeliveryLocation `json:"currentLocation,omitempty"`
}

type TrackingInfo struct {
	OrderID                   string             `json:"orderId"`
	Status                    orders.OrderStatus `json:"status"`
	StageIndex                int                `json:"stageIndex"`
	StageName                 string             `json:"stageName"`
	EstimatedMinutesRemaining int                `json:"estimatedMinutesRemaining"`
	Rider                     *RiderInfo         `json:"rider,omitempty"`
	UpdatedAt                 time.Time          `json:"updatedAt"`
}

type OutboxEvent struct {
	ID            string     `json:"id"`
	AggregateType string     `json:"aggregateType"`
	AggregateID   string     `json:"aggregateId"`
	EventType     string     `json:"eventType"`
	Payload       string     `json:"payload"`
	CreatedAt     time.Time  `json:"createdAt"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
	RetryCount    int        `json:"retryCount"`
}

type OutboxRepository interface {
	Save(ctx context.Context, event *OutboxEvent) error
	GetPending(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkProcessed(ctx context.Context, id string) error
}

type OrderProvider interface {
	GetOrder(ctx context.Context, orderID string, customerID string, isPrivileged bool) (*orders.Order, error)
}

type Service struct {
	orderProvider OrderProvider
	outboxRepo    OutboxRepository
	defaultRider  *RiderInfo
}

func NewService(orderProvider OrderProvider, outboxRepo OutboxRepository) *Service {
	// Standard simulated courier for Banani/Gulshan Dhaka zone
	rider := &RiderInfo{
		ID:          "rdr_001",
		Name:        "Rafiqul Islam",
		Phone:       "+880 1711-223344",
		VehicleInfo: "Honda CG125 • DHA-HA-11-2049",
		CurrentLocation: &DeliveryLocation{
			Latitude:  23.7937,
			Longitude: 90.4066,
			Address:   "Road 11, Banani, Dhaka",
		},
	}

	return &Service{
		orderProvider: orderProvider,
		outboxRepo:    outboxRepo,
		defaultRider:  rider,
	}
}

// GetTrackingInfo computes authoritative live stage and ETA for an order
func (s *Service) GetTrackingInfo(ctx context.Context, orderID string, customerID string, isPrivileged bool) (*TrackingInfo, error) {
	order, err := s.orderProvider.GetOrder(ctx, orderID, customerID, isPrivileged)
	if err != nil {
		if errors.Is(err, orders.ErrOrderNotFound) {
			return nil, ErrTrackingNotFound
		}
		if errors.Is(err, orders.ErrUnauthorizedOrderAccess) {
			return nil, ErrUnauthorizedTracking
		}
		return nil, err
	}

	stageIndex, stageName, estMinutes := computeStageAndETA(order.Status)

	info := &TrackingInfo{
		OrderID:                   order.ID,
		Status:                    order.Status,
		StageIndex:                stageIndex,
		StageName:                 stageName,
		EstimatedMinutesRemaining: estMinutes,
		UpdatedAt:                 order.UpdatedAt,
	}

	// Assign rider details once order is in preparing, ready, or out_for_delivery
	if stageIndex >= 2 && order.Status != orders.StatusCancelled {
		info.Rider = s.defaultRider
	}

	return info, nil
}

// PublishOutboxEvent records an event using the transactional outbox pattern
func (s *Service) PublishOutboxEvent(ctx context.Context, aggType, aggID, eventType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox event: %w", err)
	}

	evt := &OutboxEvent{
		ID:            generateEventID(),
		AggregateType: aggType,
		AggregateID:   aggID,
		EventType:     eventType,
		Payload:       string(raw),
		CreatedAt:     time.Now().UTC(),
		RetryCount:    0,
	}

	return s.outboxRepo.Save(ctx, evt)
}

func computeStageAndETA(status orders.OrderStatus) (int, string, int) {
	switch status {
	case orders.StatusPlaced:
		return 0, "Order Placed", 30
	case orders.StatusAccepted:
		return 1, "Order Accepted", 25
	case orders.StatusPreparing:
		return 2, "Food Preparing", 18
	case orders.StatusReady:
		return 3, "Food Ready for Pickup", 12
	case orders.StatusOutForDelivery:
		return 3, "Out for Delivery", 8
	case orders.StatusDelivered:
		return 4, "Delivered", 0
	case orders.StatusCancelled:
		return 0, "Order Cancelled", 0
	default:
		return 0, "Processing", 30
	}
}

func generateEventID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("evt_%s", hex.EncodeToString(b))
}

// MemoryOutboxRepository in-memory double
type MemoryOutboxRepository struct {
	mu     sync.RWMutex
	events []*OutboxEvent
}

func NewMemoryOutboxRepository() *MemoryOutboxRepository {
	return &MemoryOutboxRepository{
		events: make([]*OutboxEvent, 0),
	}
}

func (m *MemoryOutboxRepository) Save(ctx context.Context, event *OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *MemoryOutboxRepository) GetPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var pending []OutboxEvent
	for _, e := range m.events {
		if e.ProcessedAt == nil {
			pending = append(pending, *e)
			if limit > 0 && len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (m *MemoryOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	for _, e := range m.events {
		if e.ID == id {
			e.ProcessedAt = &now
			return nil
		}
	}
	return nil
}
