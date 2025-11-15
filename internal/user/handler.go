package user

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.writeError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	profile, err := h.service.GetProfile(userID)
	if err != nil {
		h.writeError(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		// Получаем email пользователя из контекста
		email, _ := r.Context().Value("userEmail").(string)
		profile = &Profile{
			ID:    userID,
			Email: email,
		}
	}

	h.writeJSON(w, profile, http.StatusOK)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.writeError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile := &Profile{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Address:   req.Address,
	}

	err := h.service.UpdateProfile(userID, profile)
	if err != nil {
		h.writeError(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "profile updated"}, http.StatusOK)
}

// Вспомогательные методы
func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	h.writeJSON(w, ErrorResponse{Error: message}, statusCode)
}
