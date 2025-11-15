package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
	ID    int    `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Валидация
	if err := req.Validate(); err != nil {
		h.writeError(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.GenerateToken(user.ID, user.Email)
	if err != nil {
		h.writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		Email: user.Email,
		ID:    user.ID,
	}

	h.writeJSON(w, response, http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Валидация
	if err := req.Validate(); err != nil {
		h.writeError(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		h.writeError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.service.GenerateToken(user.ID, user.Email)
	if err != nil {
		h.writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		Email: user.Email,
		ID:    user.ID,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.writeError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	newToken, err := h.service.GenerateToken(user.ID, user.Email)
	if err != nil {
		h.writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: newToken,
		Email: user.Email,
		ID:    user.ID,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// В будущем можно добавить blacklist токенов в Redis
	response := map[string]string{
		"message": "Logout successful",
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			h.writeError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Убираем "Bearer " префикс
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		userID, email, err := h.service.ValidateToken(tokenString)
		if err != nil {
			h.writeError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		ctx = context.WithValue(ctx, "userEmail", email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Вспомогательные методы
func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	h.writeJSON(w, ErrorResponse{Error: message}, statusCode)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
