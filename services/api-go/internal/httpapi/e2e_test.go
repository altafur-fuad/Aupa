package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/config"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/payments"
)

type apiResponseEnvelope struct {
	Status  string          `json:"status"`
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func parseEnvelope(t *testing.T, resp *http.Response) (int, apiResponseEnvelope) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	var env apiResponseEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("failed to parse JSON envelope (status %d, body %s): %v", resp.StatusCode, string(body), err)
	}
	return resp.StatusCode, env
}

func TestEndToEndCriticalCustomerFlow(t *testing.T) {
	cfg := &config.Config{
		Port:            "8080",
		Environment:     "test",
		Currency:        "BDT",
		JWTSecret:       "integration-test-secret-at-least-32-chars-long!",
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

	// ----------------------------------------------------
	// 1. Health & Readiness checks
	// ----------------------------------------------------
	t.Run("1. Health and Readiness probes", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/healthz")
		if err != nil {
			t.Fatalf("healthz request failed: %v", err)
		}
		status, env := parseEnvelope(t, res)
		if status != http.StatusOK || env.Status != "ok" {
			t.Errorf("healthz failed: status=%d, env=%+v", status, env)
		}

		resReady, err := client.Get(ts.URL + "/readyz")
		if err != nil {
			t.Fatalf("readyz request failed: %v", err)
		}
		statusReady, envReady := parseEnvelope(t, resReady)
		if statusReady != http.StatusOK || envReady.Status != "ready" {
			t.Errorf("readyz failed: status=%d, env=%+v", statusReady, envReady)
		}
	})

	// ----------------------------------------------------
	// 2. Auth: Register new user
	// ----------------------------------------------------
	var (
		customerAccessToken  string
		customerRefreshToken string
		customerID           string
	)

	t.Run("2. User Registration", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"fullName": "Tanvir Ahmed",
			"email":    "tanvir.ahmed@example.com",
			"phone":    "+8801712345678",
			"password": "Password123!",
			"role":     "customer",
		}
		b, _ := json.Marshal(reqBody)
		res, err := client.Post(ts.URL+"/v1/auth/register", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("register request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", status, string(env.Data))
		}

		var authData struct {
			Tokens struct {
				AccessToken  string `json:"accessToken"`
				RefreshToken string `json:"refreshToken"`
			} `json:"tokens"`
			User struct {
				ID       string `json:"id"`
				Email    string `json:"email"`
				FullName string `json:"fullName"`
			} `json:"user"`
		}
		if err := json.Unmarshal(env.Data, &authData); err != nil {
			t.Fatalf("failed to decode register data: %v", err)
		}

		if authData.Tokens.AccessToken == "" || authData.Tokens.RefreshToken == "" || authData.User.ID == "" {
			t.Fatalf("empty tokens or user id in response: %+v", authData)
		}

		customerAccessToken = authData.Tokens.AccessToken
		customerRefreshToken = authData.Tokens.RefreshToken
		customerID = authData.User.ID
	})

	// ----------------------------------------------------
	// 3. Auth: Fetch /v1/me profile with Bearer Token
	// ----------------------------------------------------
	t.Run("3. Fetch authenticated user profile", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+customerAccessToken)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("me request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", status, string(env.Data))
		}

		var me struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
			Email    string `json:"email"`
		}
		if err := json.Unmarshal(env.Data, &me); err != nil {
			t.Fatalf("failed to decode me data: %v", err)
		}
		if me.ID != customerID || me.Email != "tanvir.ahmed@example.com" {
			t.Errorf("unexpected profile data: %+v", me)
		}
	})

	// ----------------------------------------------------
	// 4. Auth: Token refresh & Replay attack detection
	// ----------------------------------------------------
	t.Run("4. Token Refresh and Replay Protection", func(t *testing.T) {
		// Valid refresh
		refreshPayload, _ := json.Marshal(map[string]string{
			"refreshToken": customerRefreshToken,
		})
		res, err := client.Post(ts.URL+"/v1/auth/refresh", "application/json", bytes.NewReader(refreshPayload))
		if err != nil {
			t.Fatalf("refresh request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusOK {
			t.Fatalf("expected 200 OK on token refresh, got %d: %s", status, string(env.Data))
		}

		var refreshed struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.Unmarshal(env.Data, &refreshed); err != nil {
			t.Fatalf("failed to decode refresh data: %v", err)
		}

		newRefreshToken := refreshed.RefreshToken
		customerAccessToken = refreshed.AccessToken

		// Replay attack: attempt to use the OLD already-rotated refresh token
		resReplay, err := client.Post(ts.URL+"/v1/auth/refresh", "application/json", bytes.NewReader(refreshPayload))
		if err != nil {
			t.Fatalf("replay refresh request failed: %v", err)
		}
		statusReplay, _ := parseEnvelope(t, resReplay)
		if statusReplay != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized on refresh token reuse replay, got %d", statusReplay)
		}

		// Re-login to get fresh tokens for remaining customer actions
		loginPayload, _ := json.Marshal(map[string]string{
			"email":    "tanvir.ahmed@example.com",
			"password": "Password123!",
		})
		resLogin, err := client.Post(ts.URL+"/v1/auth/login", "application/json", bytes.NewReader(loginPayload))
		if err != nil {
			t.Fatalf("re-login request failed: %v", err)
		}
		statusLogin, envLogin := parseEnvelope(t, resLogin)
		if statusLogin != http.StatusOK {
			t.Fatalf("expected 200 OK on login, got %d", statusLogin)
		}

		var freshTokens struct {
			Tokens struct {
				AccessToken string `json:"accessToken"`
			} `json:"tokens"`
		}
		_ = json.Unmarshal(envLogin.Data, &freshTokens)
		customerAccessToken = freshTokens.Tokens.AccessToken
		_ = newRefreshToken
	})

	// ----------------------------------------------------
	// 5. Catalog Discovery: Restaurants and Menus
	// ----------------------------------------------------
	var (
		restaurantID = "rest_01" // Burger Barn
		menuItemID   string
	)

	t.Run("5. Catalog Discovery", func(t *testing.T) {
		// List restaurants
		res, err := client.Get(ts.URL + "/v1/restaurants")
		if err != nil {
			t.Fatalf("list restaurants request failed: %v", err)
		}
		status, env := parseEnvelope(t, res)
		if status != http.StatusOK {
			t.Fatalf("expected 200 OK from /v1/restaurants, got %d", status)
		}

		var rests []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(env.Data, &rests); err != nil {
			t.Fatalf("failed to decode restaurants: %v", err)
		}
		if len(rests) == 0 {
			t.Fatalf("expected restaurants list to be non-empty")
		}

		// Get restaurant details
		resDetails, err := client.Get(fmt.Sprintf("%s/v1/restaurants/%s", ts.URL, restaurantID))
		if err != nil {
			t.Fatalf("get restaurant details request failed: %v", err)
		}
		statusDetails, envDetails := parseEnvelope(t, resDetails)
		if statusDetails != http.StatusOK {
			t.Fatalf("expected 200 OK from /v1/restaurants/%s, got %d", restaurantID, statusDetails)
		}

		var rest struct {
			ID                 string `json:"id"`
			Name               string `json:"name"`
			DeliveryFeePoisha  int64  `json:"deliveryFeePoisha"`
			MinimumOrderPoisha int64  `json:"minimumOrderPoisha"`
		}
		if err := json.Unmarshal(envDetails.Data, &rest); err != nil {
			t.Fatalf("failed to decode restaurant details: %v", err)
		}
		if rest.Name != "Burger Barn" || rest.DeliveryFeePoisha != 6000 {
			t.Errorf("unexpected restaurant details: %+v", rest)
		}

		// Get restaurant menu (returns []MenuItem)
		resMenu, err := client.Get(fmt.Sprintf("%s/v1/restaurants/%s/menu", ts.URL, restaurantID))
		if err != nil {
			t.Fatalf("get menu request failed: %v", err)
		}
		statusMenu, envMenu := parseEnvelope(t, resMenu)
		if statusMenu != http.StatusOK {
			t.Fatalf("expected 200 OK from /v1/restaurants/%s/menu, got %d", restaurantID, statusMenu)
		}

		var menuItems []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			PricePoisha int64  `json:"pricePoisha"`
			Category    string `json:"category"`
		}
		if err := json.Unmarshal(envMenu.Data, &menuItems); err != nil {
			t.Fatalf("failed to decode menu items: %v", err)
		}
		if len(menuItems) == 0 {
			t.Fatalf("expected at least one item in menu")
		}

		menuItemID = menuItems[0].ID
	})

	// ----------------------------------------------------
	// 6. Server-Authoritative Cart Quote Calculation
	// ----------------------------------------------------
	t.Run("6. Server Authoritative Cart Quote Calculation", func(t *testing.T) {
		quotePayload, _ := json.Marshal(map[string]interface{}{
			"restaurantId":      restaurantID,
			"deliveryAddressId": "addr_01",
			"items": []map[string]interface{}{
				{
					"itemId":   menuItemID,
					"quantity": 2,
				},
			},
			"promoCode": "AUPA50",
		})

		res, err := client.Post(ts.URL+"/v1/checkout/quote", "application/json", bytes.NewReader(quotePayload))
		if err != nil {
			t.Fatalf("quote request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusOK {
			t.Fatalf("expected 200 OK from quote, got %d: %s", status, string(env.Data))
		}

		var quote struct {
			SubtotalPoisha    int64  `json:"subtotalPoisha"`
			DeliveryFeePoisha int64  `json:"deliveryFeePoisha"`
			VATPoisha         int64  `json:"vatPoisha"`
			DiscountPoisha    int64  `json:"discountPoisha"`
			TotalPoisha       int64  `json:"totalPoisha"`
			Currency          string `json:"currency"`
		}
		if err := json.Unmarshal(env.Data, &quote); err != nil {
			t.Fatalf("failed to decode quote: %v", err)
		}

		if quote.SubtotalPoisha <= 0 {
			t.Errorf("expected positive subtotal, got %d", quote.SubtotalPoisha)
		}
		if quote.DeliveryFeePoisha != 6000 {
			t.Errorf("expected delivery fee 6000 poisha, got %d", quote.DeliveryFeePoisha)
		}
		if quote.DiscountPoisha != 20000 {
			t.Errorf("expected 20000 poisha discount for promo AUPA50, got %d", quote.DiscountPoisha)
		}
		expectedTotal := quote.SubtotalPoisha + quote.DeliveryFeePoisha + quote.VATPoisha - quote.DiscountPoisha
		if quote.TotalPoisha != expectedTotal {
			t.Errorf("expected total %d, got %d", expectedTotal, quote.TotalPoisha)
		}
	})

	// ----------------------------------------------------
	// 7. Idempotent Order Creation
	// ----------------------------------------------------
	var (
		orderID           string
		totalAmountPoisha int64
		idempotencyKey    = "idemp-customer-e2e-order-uuid-999"
	)

	t.Run("7. Idempotent Order Creation", func(t *testing.T) {
		orderPayload := map[string]interface{}{
			"restaurantId":        restaurantID,
			"deliveryAddressId":   "addr_01",
			"paymentMethod":       "bkash",
			"promoCode":           "AUPA50",
			"specialInstructions": "Extra napkins please",
			"items": []map[string]interface{}{
				{
					"itemId":   menuItemID,
					"quantity": 2,
				},
			},
		}
		b, _ := json.Marshal(orderPayload)

		// First order submission with idempotency key
		req, _ := http.NewRequest("POST", ts.URL+"/v1/orders", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+customerAccessToken)
		req.Header.Set("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("create order request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusCreated {
			t.Fatalf("expected 201 Created on order placement, got %d: %s", status, string(env.Data))
		}

		var createdOrder struct {
			ID                string `json:"id"`
			Status            string `json:"status"`
			PaymentStatus     string `json:"paymentStatus"`
			TotalAmountPoisha int64  `json:"totalAmountPoisha"`
		}
		if err := json.Unmarshal(env.Data, &createdOrder); err != nil {
			t.Fatalf("failed to decode created order: %v", err)
		}
		if createdOrder.ID == "" || createdOrder.Status != "placed" || createdOrder.PaymentStatus != "unpaid" {
			t.Errorf("unexpected created order fields: %+v", createdOrder)
		}

		orderID = createdOrder.ID
		totalAmountPoisha = createdOrder.TotalAmountPoisha

		// Replay identical submission with same Idempotency-Key -> returns 200 OK with SAME order
		reqReplay, _ := http.NewRequest("POST", ts.URL+"/v1/orders", bytes.NewReader(b))
		reqReplay.Header.Set("Authorization", "Bearer "+customerAccessToken)
		reqReplay.Header.Set("Idempotency-Key", idempotencyKey)
		reqReplay.Header.Set("Content-Type", "application/json")

		resReplay, err := client.Do(reqReplay)
		if err != nil {
			t.Fatalf("replay order request failed: %v", err)
		}
		statusReplay, envReplay := parseEnvelope(t, resReplay)
		if statusReplay != http.StatusOK {
			t.Fatalf("expected 200 OK on idempotent replay, got %d", statusReplay)
		}

		var replayedOrder struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(envReplay.Data, &replayedOrder)
		if replayedOrder.ID != orderID {
			t.Errorf("expected replayed order ID %s, got %s", orderID, replayedOrder.ID)
		}

		// Conflicting submission with same Idempotency-Key but different quantity -> returns 409 Conflict
		conflictingPayload := map[string]interface{}{
			"restaurantId":      restaurantID,
			"deliveryAddressId": "addr_01",
			"paymentMethod":     "bkash",
			"items": []map[string]interface{}{
				{
					"itemId":   menuItemID,
					"quantity": 5, // CHANGED QUANTITY
				},
			},
		}
		bConflict, _ := json.Marshal(conflictingPayload)
		reqConflict, _ := http.NewRequest("POST", ts.URL+"/v1/orders", bytes.NewReader(bConflict))
		reqConflict.Header.Set("Authorization", "Bearer "+customerAccessToken)
		reqConflict.Header.Set("Idempotency-Key", idempotencyKey)
		reqConflict.Header.Set("Content-Type", "application/json")

		resConflict, err := client.Do(reqConflict)
		if err != nil {
			t.Fatalf("conflict order request failed: %v", err)
		}
		statusConflict, _ := parseEnvelope(t, resConflict)
		if statusConflict != http.StatusConflict {
			t.Errorf("expected 409 Conflict on payload mismatch for same idempotency key, got %d", statusConflict)
		}
	})

	// ----------------------------------------------------
	// 8. Order List & Order Details
	// ----------------------------------------------------
	t.Run("8. Order List and Details retrieval", func(t *testing.T) {
		// List orders
		reqList, _ := http.NewRequest("GET", ts.URL+"/v1/orders", nil)
		reqList.Header.Set("Authorization", "Bearer "+customerAccessToken)
		resList, err := client.Do(reqList)
		if err != nil {
			t.Fatalf("list orders request failed: %v", err)
		}
		statusList, envList := parseEnvelope(t, resList)
		if statusList != http.StatusOK {
			t.Fatalf("expected 200 OK from /v1/orders, got %d", statusList)
		}

		var orderList []struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(envList.Data, &orderList)
		if len(orderList) != 1 || orderList[0].ID != orderID {
			t.Errorf("expected 1 order with ID %s, got %+v", orderID, orderList)
		}

		// Get single order details
		reqDetails, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/orders/%s", ts.URL, orderID), nil)
		reqDetails.Header.Set("Authorization", "Bearer "+customerAccessToken)
		resDetails, err := client.Do(reqDetails)
		if err != nil {
			t.Fatalf("get order details request failed: %v", err)
		}
		statusDetails, envDetails := parseEnvelope(t, resDetails)
		if statusDetails != http.StatusOK {
			t.Fatalf("expected 200 OK from /v1/orders/%s, got %d", orderID, statusDetails)
		}

		var details struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Items  []struct {
				ItemID   string `json:"itemId"`
				Quantity int    `json:"quantity"`
			} `json:"items"`
		}
		_ = json.Unmarshal(envDetails.Data, &details)
		if details.ID != orderID || len(details.Items) != 1 || details.Items[0].Quantity != 2 {
			t.Errorf("unexpected order details: %+v", details)
		}
	})

	// ----------------------------------------------------
	// 9. Payment Initiation
	// ----------------------------------------------------
	var paymentAttemptID string

	t.Run("9. Payment Checkout Initiation", func(t *testing.T) {
		payPayload, _ := json.Marshal(map[string]string{
			"provider": "mock",
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/orders/%s/pay", ts.URL, orderID), bytes.NewReader(payPayload))
		req.Header.Set("Authorization", "Bearer "+customerAccessToken)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("initiate payment request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusOK {
			t.Fatalf("expected 200 OK from payment initiation, got %d: %s", status, string(env.Data))
		}

		var payRes struct {
			CheckoutURL string `json:"checkoutUrl"`
			AttemptID   string `json:"attemptId"`
		}
		if err := json.Unmarshal(env.Data, &payRes); err != nil {
			t.Fatalf("failed to decode payment response: %v", err)
		}

		if payRes.CheckoutURL == "" || payRes.AttemptID == "" {
			t.Errorf("empty checkoutUrl or attemptId: %+v", payRes)
		}
		paymentAttemptID = payRes.AttemptID
	})

	// ----------------------------------------------------
	// 10. Webhook Reconciliation & State Transition
	// ----------------------------------------------------
	t.Run("10. Payment Webhook Verification and Transition", func(t *testing.T) {
		webhookEvent := payments.WebhookEvent{
			EventID:          "evt_e2e_wh_001",
			Provider:         "mock",
			OrderID:          orderID,
			PaymentAttemptID: paymentAttemptID,
			TransactionID:    "trx_e2e_live_888",
			Status:           payments.AttemptSucceeded,
			AmountPoisha:     totalAmountPoisha,
			Currency:         "BDT",
			Timestamp:        time.Now().UTC(),
		}
		payloadBytes, _ := json.Marshal(webhookEvent)

		// 10a. Forged / invalid signature rejection
		reqInvalid, _ := http.NewRequest("POST", ts.URL+"/v1/webhooks/payments/mock", bytes.NewReader(payloadBytes))
		reqInvalid.Header.Set("Content-Type", "application/json")
		reqInvalid.Header.Set("X-Aupa-Signature", "forged_invalid_signature_hex")

		resInvalid, err := client.Do(reqInvalid)
		if err != nil {
			t.Fatalf("invalid signature webhook request failed: %v", err)
		}
		statusInvalid, _ := parseEnvelope(t, resInvalid)
		if statusInvalid != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for forged signature, got %d", statusInvalid)
		}

		// 10b. Valid HMAC-SHA256 signature
		validSig := payments.GenerateSignature(payloadBytes, cfg.JWTSecret)

		reqValid, _ := http.NewRequest("POST", ts.URL+"/v1/webhooks/payments/mock", bytes.NewReader(payloadBytes))
		reqValid.Header.Set("Content-Type", "application/json")
		reqValid.Header.Set("X-Aupa-Signature", validSig)

		resValid, err := client.Do(reqValid)
		if err != nil {
			t.Fatalf("valid signature webhook request failed: %v", err)
		}

		statusValid, envValid := parseEnvelope(t, resValid)
		if statusValid != http.StatusOK {
			t.Fatalf("expected 200 OK for valid webhook, got %d: %s", statusValid, string(envValid.Data))
		}

		var whRes struct {
			Received    bool `json:"received"`
			IsDuplicate bool `json:"isDuplicate"`
		}
		_ = json.Unmarshal(envValid.Data, &whRes)
		if !whRes.Received || whRes.IsDuplicate {
			t.Errorf("expected received=true, isDuplicate=false, got %+v", whRes)
		}

		// Verify order was transitioned to 'accepted' and 'paid'
		reqCheck, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/orders/%s", ts.URL, orderID), nil)
		reqCheck.Header.Set("Authorization", "Bearer "+customerAccessToken)
		resCheck, _ := client.Do(reqCheck)
		_, envCheck := parseEnvelope(t, resCheck)

		var updatedOrder struct {
			Status        string `json:"status"`
			PaymentStatus string `json:"paymentStatus"`
		}
		_ = json.Unmarshal(envCheck.Data, &updatedOrder)
		if updatedOrder.Status != "accepted" {
			t.Errorf("expected order status 'accepted', got '%s'", updatedOrder.Status)
		}
		if updatedOrder.PaymentStatus != "paid" {
			t.Errorf("expected payment status 'paid', got '%s'", updatedOrder.PaymentStatus)
		}

		// 10c. Idempotent webhook replay
		reqReplay, _ := http.NewRequest("POST", ts.URL+"/v1/webhooks/payments/mock", bytes.NewReader(payloadBytes))
		reqReplay.Header.Set("Content-Type", "application/json")
		reqReplay.Header.Set("X-Aupa-Signature", validSig)

		resReplay, _ := client.Do(reqReplay)
		statusReplay, envReplay := parseEnvelope(t, resReplay)
		if statusReplay != http.StatusOK {
			t.Fatalf("expected 200 OK on webhook replay, got %d", statusReplay)
		}
		var whReplayRes struct {
			Received    bool `json:"received"`
			IsDuplicate bool `json:"isDuplicate"`
		}
		_ = json.Unmarshal(envReplay.Data, &whReplayRes)
		if !whReplayRes.Received || !whReplayRes.IsDuplicate {
			t.Errorf("expected received=true, isDuplicate=true on replay, got %+v", whReplayRes)
		}
	})

	// ----------------------------------------------------
	// 11. Delivery Tracking & Courier Coordinates
	// ----------------------------------------------------
	t.Run("11. Delivery Tracking Timeline", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/orders/%s/tracking", ts.URL, orderID), nil)
		req.Header.Set("Authorization", "Bearer "+customerAccessToken)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("delivery tracking request failed: %v", err)
		}

		status, env := parseEnvelope(t, res)
		if status != http.StatusOK {
			t.Fatalf("expected 200 OK from delivery tracking, got %d: %s", status, string(env.Data))
		}

		var tracking struct {
			OrderID                   string `json:"orderId"`
			Status                    string `json:"status"`
			StageIndex                int    `json:"stageIndex"`
			StageName                 string `json:"stageName"`
			EstimatedMinutesRemaining int    `json:"estimatedMinutesRemaining"`
			Rider                     *struct {
				Name            string `json:"name"`
				Phone           string `json:"phone"`
				CurrentLocation *struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"currentLocation"`
			} `json:"rider"`
		}
		if err := json.Unmarshal(env.Data, &tracking); err != nil {
			t.Fatalf("failed to decode tracking data: %v", err)
		}

		if tracking.OrderID != orderID {
			t.Errorf("expected orderId %s, got %s", orderID, tracking.OrderID)
		}
		if tracking.Status != "accepted" {
			t.Errorf("expected status 'accepted', got '%s'", tracking.Status)
		}
		if tracking.StageIndex < 1 {
			t.Errorf("expected stageIndex >= 1 for accepted order, got %d", tracking.StageIndex)
		}
		if tracking.StageName == "" {
			t.Errorf("expected non-empty stageName")
		}
		if tracking.EstimatedMinutesRemaining <= 0 {
			t.Errorf("expected positive estimated minutes, got %d", tracking.EstimatedMinutesRemaining)
		}
	})
}
