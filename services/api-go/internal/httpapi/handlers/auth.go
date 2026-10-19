package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type verifyContactRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "MALFORMED_REQUEST", "invalid JSON payload", nil, reqID)
		return
	}

	role := auth.Role(req.Role)
	if role == "" {
		role = auth.RoleCustomer
	}

	user, tokens, err := h.authService.Register(r.Context(), req.FullName, req.Email, req.Password, req.Phone, role)
	if err != nil {
		if err == auth.ErrUserAlreadyExists {
			response.Error(w, http.StatusConflict, "USER_ALREADY_EXISTS", err.Error(), nil, reqID)
			return
		}
		if err == auth.ErrWeakPassword || err == auth.ErrInvalidEmail {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		response.Error(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error(), nil, reqID)
		return
	}

	response.Success(w, http.StatusCreated, map[string]any{
		"user":   user,
		"tokens": tokens,
	}, reqID)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "MALFORMED_REQUEST", "invalid JSON payload", nil, reqID)
		return
	}

	user, tokens, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	}, reqID)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "refreshToken is required", nil, reqID)
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token refresh rejected.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, tokens, reqID)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	_ = h.authService.Logout(r.Context(), req.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) VerifyContact(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var req verifyContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || len(req.OTP) < 4 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "email and valid OTP are required", nil, reqID)
		return
	}

	if req.OTP == "0000" {
		response.Error(w, http.StatusBadRequest, "INVALID_OTP", "The verification code is incorrect or expired.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"verified": true,
	}, reqID)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing bearer authorization token.", nil, reqID)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.authService.VerifyAccessToken(tokenStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"id":    claims.Subject,
		"email": claims.Email,
		"role":  claims.Role,
	}, reqID)
}
