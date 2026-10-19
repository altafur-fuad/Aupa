package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/config"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi"
	"github.com/golang-jwt/jwt/v5"
)

func TestSecurityAndNegativeAuthorization(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		Environment:     "test",
		Currency:        "BDT",
		JWTSecret:       "super-secure-production-ready-jwt-secret-key-32!",
		JWTIssuer:       "aupa-test",
		JWTAudience:     "aupa-mobile",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	router := httpapi.NewRouter(cfg, httpapi.Repositories{}, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()

	// Helper to register a customer
	registerCustomer := func(name, email string) (string, string) {
		payload, _ := json.Marshal(map[string]interface{}{
			"fullName": name,
			"email":    email,
			"phone":    "+8801700000000",
			"password": "Password123!",
			"role":     "customer",
		})
		res, err := client.Post(ts.URL+"/v1/auth/register", "application/json", bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("failed to register customer %s: %v", email, err)
		}
		_, env := parseEnvelope(t, res)
		var authData struct {
			Tokens struct {
				AccessToken string `json:"accessToken"`
			} `json:"tokens"`
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		_ = json.Unmarshal(env.Data, &authData)
		return authData.User.ID, authData.Tokens.AccessToken
	}

	custAID, custAToken := registerCustomer("Customer A", "cust.a@example.com")
	custBID, custBToken := registerCustomer("Customer B", "cust.b@example.com")

	if custAID == "" || custAToken == "" || custBID == "" || custBToken == "" {
		t.Fatalf("failed to create test customers")
	}

	// Customer A creates an order
	orderPayload, _ := json.Marshal(map[string]interface{}{
		"restaurantId":      "rest_01",
		"deliveryAddressId": "addr_01",
		"paymentMethod":     "bkash",
		"items": []map[string]interface{}{
			{
				"itemId":   "item_01",
				"quantity": 1,
			},
		},
	})
	reqCreate, _ := http.NewRequest("POST", ts.URL+"/v1/orders", bytes.NewReader(orderPayload))
	reqCreate.Header.Set("Authorization", "Bearer "+custAToken)
	reqCreate.Header.Set("Idempotency-Key", "idemp-sec-test-ord-001")
	reqCreate.Header.Set("Content-Type", "application/json")

	resCreate, err := client.Do(reqCreate)
	if err != nil {
		t.Fatalf("failed to create order for customer A: %v", err)
	}
	statusCreate, envCreate := parseEnvelope(t, resCreate)
	if statusCreate != http.StatusCreated {
		t.Fatalf("expected 201 Created for order A, got %d", statusCreate)
	}
	var orderA struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(envCreate.Data, &orderA)
	orderAID := orderA.ID

	// -------------------------------------------------------------------
	// 1. Cross-Customer Resource Access Rejection (Horizontal Privilege Escalation)
	// -------------------------------------------------------------------
	t.Run("Customer B cannot view Customer A's order", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/orders/%s", ts.URL, orderAID), nil)
		req.Header.Set("Authorization", "Bearer "+custBToken)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-customer order view, got %d", status)
		}
	})

	t.Run("Customer B cannot initiate payment for Customer A's order", func(t *testing.T) {
		payBody, _ := json.Marshal(map[string]string{"provider": "mock"})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/orders/%s/pay", ts.URL, orderAID), bytes.NewReader(payBody))
		req.Header.Set("Authorization", "Bearer "+custBToken)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-customer payment initiation, got %d", status)
		}
	})

	t.Run("Customer B cannot cancel Customer A's order", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/orders/%s/cancel", ts.URL, orderAID), nil)
		req.Header.Set("Authorization", "Bearer "+custBToken)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-customer order cancellation, got %d", status)
		}
	})

	t.Run("Customer B cannot track delivery of Customer A's order", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/orders/%s/tracking", ts.URL, orderAID), nil)
		req.Header.Set("Authorization", "Bearer "+custBToken)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-customer tracking, got %d", status)
		}
	})

	// -------------------------------------------------------------------
	// 2. Token Security: Expired, Forged, and None-Algorithm JWT Rejections
	// -------------------------------------------------------------------
	t.Run("Unauthenticated request to protected endpoint rejected with 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/v1/orders", nil)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for missing token, got %d", status)
		}
	})

	t.Run("Forged secret JWT rejected with 401", func(t *testing.T) {
		forgedClaims := &auth.CustomClaims{
			Role:      string(auth.RoleCustomer),
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   custAID,
				Issuer:    cfg.JWTIssuer,
				Audience:  jwt.ClaimStrings{cfg.JWTAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, forgedClaims)
		signedForged, _ := token.SignedString([]byte("wrong-attacker-key-1234567890123456"))

		req, _ := http.NewRequest("GET", ts.URL+"/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+signedForged)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for forged secret signature, got %d", status)
		}
	})

	t.Run("Expired JWT rejected with 401", func(t *testing.T) {
		expiredClaims := &auth.CustomClaims{
			Role:      string(auth.RoleCustomer),
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   custAID,
				Issuer:    cfg.JWTIssuer,
				Audience:  jwt.ClaimStrings{cfg.JWTAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		signedExpired, _ := token.SignedString([]byte(cfg.JWTSecret))

		req, _ := http.NewRequest("GET", ts.URL+"/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+signedExpired)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for expired JWT, got %d", status)
		}
	})

	// -------------------------------------------------------------------
	// 3. Input Validation & Bounds Checking
	// -------------------------------------------------------------------
	t.Run("Cart quote with 0 quantity rejected", func(t *testing.T) {
		badPayload, _ := json.Marshal(map[string]interface{}{
			"restaurantId": "rest_01",
			"items": []map[string]interface{}{
				{"itemId": "item_01", "quantity": 0},
			},
		})
		res, err := client.Post(ts.URL+"/v1/checkout/quote", "application/json", bytes.NewReader(badPayload))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for zero quantity, got %d", status)
		}
	})

	t.Run("Cart quote with empty items list rejected", func(t *testing.T) {
		emptyPayload, _ := json.Marshal(map[string]interface{}{
			"restaurantId": "rest_01",
			"items":        []map[string]interface{}{},
		})
		res, err := client.Post(ts.URL+"/v1/checkout/quote", "application/json", bytes.NewReader(emptyPayload))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for empty items, got %d", status)
		}
	})

	t.Run("Order creation missing Idempotency-Key rejected", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/v1/orders", bytes.NewReader(orderPayload))
		req.Header.Set("Authorization", "Bearer "+custAToken)
		req.Header.Set("Content-Type", "application/json")
		// NO Idempotency-Key header!

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		status, _ := parseEnvelope(t, res)
		if status != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for missing idempotency key, got %d", status)
		}
	})
}
