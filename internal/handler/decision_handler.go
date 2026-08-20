package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/service"
	"github.com/jackc/pgx/v5"
)

type DecisionHandler struct {
	service *service.DecisionService
}

func NewDecisionHandler(service *service.DecisionService) *DecisionHandler {
	return &DecisionHandler{service: service}
}

// Check evaluates and persists a decision for the requested rule, subject,
// knowledge time, and effective time.
func (h *DecisionHandler) Check(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ruleKey := query.Get("rule_key")
	subjectID := query.Get("subject_id")
	timeWhen, err := requiredTime(query.Get("time_when_to_check"), "time_when_to_check")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	timeWhere, err := requiredTime(query.Get("time_where_to_check"), "time_where_to_check")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if ruleKey == "" || subjectID == "" {
		http.Error(w, "rule_key and subject_id are required", http.StatusBadRequest)
		return
	}

	result, err := h.service.Decide(r.Context(), ruleKey, subjectID, timeWhen, timeWhere)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "encode decision response: "+err.Error(), http.StatusInternalServerError)
	}
}

func requiredTime(value, name string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New(name + " is required")
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.New(name + " must be RFC3339")
	}
	return parsed, nil
}
