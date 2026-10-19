package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/config"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/delivery"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/handlers"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/payments"
)

// Repositories holds the storage dependencies for building the HTTP API router.
type Repositories struct {
	UserRepo    auth.UserRepository
	SessionRepo auth.SessionRepository
	RestRepo    catalog.RestaurantRepository
	MenuRepo    catalog.MenuRepository
	OrderRepo   orders.OrderRepository
	IdempRepo   orders.IdempotencyRepository
	AttemptRepo payments.PaymentAttemptRepository
	AuditRepo   payments.WebhookAuditRepository
	OutboxRepo  delivery.OutboxRepository
}

// NewRouter constructs the full HTTP handler and middleware pipeline for the Aupa API.
func NewRouter(cfg *config.Config, repos Repositories, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	// Fallback to in-memory test doubles if repositories are not provided
	if repos.UserRepo == nil {
		repos.UserRepo = auth.NewMemoryUserRepository()
	}
	if repos.SessionRepo == nil {
		repos.SessionRepo = auth.NewMemorySessionRepository()
	}
	if repos.RestRepo == nil || repos.MenuRepo == nil {
		memCat := catalog.NewMemoryCatalogRepository()
		if repos.RestRepo == nil {
			repos.RestRepo = memCat
		}
		if repos.MenuRepo == nil {
			repos.MenuRepo = memCat
		}
	}
	if repos.OrderRepo == nil {
		repos.OrderRepo = orders.NewMemoryOrderRepository()
	}
	if repos.IdempRepo == nil {
		repos.IdempRepo = orders.NewMemoryIdempotencyRepository()
	}
	if repos.AttemptRepo == nil {
		repos.AttemptRepo = payments.NewMemoryPaymentAttemptRepository()
	}
	if repos.AuditRepo == nil {
		repos.AuditRepo = payments.NewMemoryWebhookAuditRepository()
	}
	if repos.OutboxRepo == nil {
		repos.OutboxRepo = delivery.NewMemoryOutboxRepository()
	}

	mux := http.NewServeMux()

	// 1. Health and Readiness probes
	healthHandler := handlers.NewHealthHandler(func() bool { return true })
	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)

	// 2. Auth service & routes
	authService := auth.NewService(
		repos.UserRepo,
		repos.SessionRepo,
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)
	authHandler := handlers.NewAuthHandler(authService)
	authMw := middleware.Authenticate(authService)

	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /v1/auth/verify-contact", authHandler.VerifyContact)
	mux.Handle("GET /v1/me", authMw(http.HandlerFunc(authHandler.GetMe)))

	// 3. Catalog service & routes
	catalogService := catalog.NewService(repos.RestRepo, repos.MenuRepo)
	catalogHandler := handlers.NewCatalogHandler(catalogService)

	mux.HandleFunc("GET /v1/restaurants", catalogHandler.ListRestaurants)
	mux.HandleFunc("GET /v1/restaurants/{id}", catalogHandler.GetRestaurant)
	mux.HandleFunc("GET /v1/restaurants/{id}/menu", catalogHandler.GetMenu)

	// 4. Cart quote service & routes
	cartCalculator := cart.NewCalculator(catalogService)
	cartHandler := handlers.NewCartHandler(cartCalculator)
	mux.HandleFunc("POST /v1/checkout/quote", cartHandler.CalculateQuote)

	// 5. Orders service & routes
	orderService := orders.NewService(repos.OrderRepo, repos.IdempRepo, cartCalculator)
	ordersHandler := handlers.NewOrdersHandler(orderService)

	mux.Handle("POST /v1/orders", authMw(http.HandlerFunc(ordersHandler.CreateOrder)))
	mux.Handle("GET /v1/orders", authMw(http.HandlerFunc(ordersHandler.ListOrders)))
	mux.Handle("GET /v1/orders/{id}", authMw(http.HandlerFunc(ordersHandler.GetOrder)))
	mux.Handle("POST /v1/orders/{id}/cancel", authMw(http.HandlerFunc(ordersHandler.CancelOrder)))

	// 6. Payments service & routes
	paymentsService := payments.NewService(orderService, repos.AttemptRepo, repos.AuditRepo, cfg.JWTSecret)
	paymentsHandler := handlers.NewPaymentsHandler(paymentsService)

	mux.Handle("POST /v1/orders/{id}/pay", authMw(http.HandlerFunc(paymentsHandler.InitiatePayment)))
	mux.HandleFunc("POST /v1/webhooks/payments/{provider}", paymentsHandler.HandleWebhook)

	// 7. Delivery tracking service & routes
	deliveryService := delivery.NewService(orderService, repos.OutboxRepo)
	deliveryHandler := handlers.NewDeliveryHandler(deliveryService)

	mux.Handle("GET /v1/orders/{id}/tracking", authMw(http.HandlerFunc(deliveryHandler.GetTracking)))

	// Wrap mux with global middleware chain
	return middleware.Chain(mux,
		middleware.RequestID,
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.CORS("*"),
		middleware.Timeout(cfg.WriteTimeout),
	)
}
