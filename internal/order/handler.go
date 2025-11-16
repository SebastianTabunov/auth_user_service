package order

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.writeError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.writeError(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(orderID, userID)
	if err != nil {
		h.writeError(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	if order == nil {
		h.writeError(w, "Order not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, order, http.StatusOK)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.writeError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Простая валидация
	if req.Title == "" {
		h.writeError(w, "Title is required", http.StatusBadRequest)
		return
	}

	if req.Price <= 0 {
		h.writeError(w, "Price must be positive", http.StatusBadRequest)
		return
	}

	order, err := h.service.CreateOrder(userID, req.Title, req.Description, req.Price)
	if err != nil {
		h.writeError(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, order, http.StatusCreated)
}

func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.writeError(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetUserOrders(userID)
	if err != nil {
		h.writeError(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, orders, http.StatusOK)
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
