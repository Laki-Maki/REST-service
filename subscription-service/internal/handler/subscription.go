// internal/handler/subscription.go
package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"subscription-service/internal/db"
	"subscription-service/internal/model"
	"subscription-service/internal/service"

	"github.com/gorilla/mux"
)

// Helper parse "MM-YYYY" -> time.Time (first day of month)
func parseMonthYear(param string) (time.Time, error) {
	t, err := time.Parse("01-2006", param)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

type Handler struct {
	Store *db.Store
}

func NewHandler(store *db.Store) *Handler {
	return &Handler{Store: store}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Главная страница
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Subscription service is running"))
	}).Methods("GET")

	// Подписки (важно: Aggregate перед {id})
	r.HandleFunc("/subscriptions/aggregate", h.Aggregate).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.GetSubscription).Methods("GET")
	r.HandleFunc("/subscriptions", h.ListSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions", h.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions/{id}", h.UpdateSubscription).Methods("PUT")
	r.HandleFunc("/subscriptions/{id}", h.DeleteSubscription).Methods("DELETE")
}

// Вспомогательные функции
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CreateSubscription
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ServiceName string  `json:"service_name"`
		Price       int     `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if in.ServiceName == "" || in.Price < 0 || in.UserID == "" || in.StartDate == "" {
		http.Error(w, "missing or invalid fields", http.StatusBadRequest)
		return
	}
	start, err := parseMonthYear(in.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
		return
	}
	var end *time.Time
	if in.EndDate != nil && *in.EndDate != "" {
		t, err := parseMonthYear(*in.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
		end = &t
		if end.Before(start) {
			http.Error(w, "end_date cannot be before start_date", http.StatusBadRequest)
			return
		}
	}

	sub := &model.Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	if err := h.Store.CreateSubscription(r.Context(), sub); err != nil {
		log.Printf("create subscription error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, sub)
}

// GetSubscription
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sub, err := h.Store.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("get subscription error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, sub)
}

// ListSubscriptions
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := nullableString(q.Get("user_id"))
	serviceName := nullableString(q.Get("service_name"))
	limit := 50
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}
	offset := 0
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	subs, err := h.Store.ListSubscriptions(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		log.Printf("list subscriptions error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, subs)
}

// UpdateSubscription
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var in struct {
		ServiceName string  `json:"service_name"`
		Price       int     `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if in.ServiceName == "" || in.Price < 0 || in.UserID == "" || in.StartDate == "" {
		http.Error(w, "missing or invalid fields", http.StatusBadRequest)
		return
	}
	start, err := parseMonthYear(in.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
		return
	}
	var end *time.Time
	if in.EndDate != nil && *in.EndDate != "" {
		t, err := parseMonthYear(*in.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
		end = &t
		if end.Before(start) {
			http.Error(w, "end_date cannot be before start_date", http.StatusBadRequest)
			return
		}
	}
	sub := &model.Subscription{
		ID:          id,
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   start,
		EndDate:     end,
	}
	if err := h.Store.UpdateSubscription(r.Context(), sub); err != nil {
		log.Printf("update subscription error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, sub)
}

// DeleteSubscription
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.Store.DeleteSubscription(r.Context(), id); err != nil {
		log.Printf("delete subscription error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Aggregate
func (h *Handler) Aggregate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, "from and to are required (MM-YYYY)", http.StatusBadRequest)
		return
	}

	from, err := parseMonthYear(fromStr)
	if err != nil {
		http.Error(w, "invalid from format, expected MM-YYYY", http.StatusBadRequest)
		return
	}
	to, err := parseMonthYear(toStr)
	if err != nil {
		http.Error(w, "invalid to format, expected MM-YYYY", http.StatusBadRequest)
		return
	}
	if from.After(to) {
		http.Error(w, "from cannot be after to", http.StatusBadRequest)
		return
	}

	// Пустые значения превращаются в nil
	var userID *string
	if v := q.Get("user_id"); v != "" {
		userID = &v
	}

	var serviceName *string
	if v := q.Get("service_name"); v != "" {
		serviceName = &v
	}

	log.Printf("Aggregate: searching subscriptions from=%v to=%v user_id=%v service_name=%v",
		from, to, userID, serviceName)

	subs, err := h.Store.FindSubscriptionsOverlapping(r.Context(), from, to, userID, serviceName)
	if err != nil {
		log.Printf("aggregate find error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("Aggregate: found %d subscriptions", len(subs))
	total := service.CalculateTotal(subs, from, to)
	log.Printf("Aggregate: calculated total=%v", total)

	writeJSON(w, map[string]interface{}{
		"from":  fromStr,
		"to":    toStr,
		"total": total,
	})
}
