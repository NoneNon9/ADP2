package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type OrderHandler struct {
	usecase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

func (h *OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		ItemName   string `json:"item_name"`
		Amount     int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")

	order, err := h.usecase.CreateOrder(req.CustomerID, req.ItemName, req.Amount, idemKey)

	if err == domain.ErrPaymentFailed {
		http.Error(w, `{"error": "Payment service unavailable"}`, http.StatusServiceUnavailable)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}
	id := parts[2]

	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid Order ID format", http.StatusBadRequest)
		return
	}

	order, err := h.usecase.GetOrder(r.Context(), id)
	if err != nil {
		if err.Error() == "order not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] != "cancel" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[2]

	err := h.usecase.CancelOrder(id)
	if err == domain.ErrCannotCancelPaid {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Order cancelled successfully"}`))
}

func NewRouter(handler *OrderHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.HandleCreateOrder(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.HandleGetOrder(w, r)
		case http.MethodPatch:
			handler.HandleCancelOrder(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
