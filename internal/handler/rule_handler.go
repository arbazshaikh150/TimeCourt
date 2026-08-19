package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RuleHandler struct {
	service *service.RuleService
}

func NewRuleHandler(service *service.RuleService) *RuleHandler {
	return &RuleHandler{
		service: service,
	}
}

func (h *RuleHandler) Create(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Rule request received")
	var createRuleRequest dto.CreateRuleRequest
	// Decoding the req body
	err := json.NewDecoder(req.Body).Decode(&createRuleRequest)
	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	//
	// Getting idempotency key
	//

	idempotencyKey := req.Header.Get("Idempotency-Key")
	// Present in string format --> but i want in uuid
	if idempotencyKey == "" {
		http.Error(
			w,
			"Idempotency-Key header is required",
			http.StatusBadRequest,
		)
		return
	}

	idempotentUUID, err := uuid.Parse(idempotencyKey)
	if err != nil {
		http.Error(
			w,
			"invalid Idempotency-Key",
			http.StatusBadRequest,
		)
		return
	}

	// Calling service
	err = h.service.Create(
		req.Context(),
		&createRuleRequest,
		idempotentUUID,
	)

	fmt.Println("The error is ", err)
	if err != nil {
		http.Error(
			w,
			"failed to create rule",
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]any{
		"message": "rule created successfully",
	}); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}

func (h *RuleHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()

	ruleKey := query.Get("rule_key")
	subjectID := query.Get("subject_id")

	timeWhereToCheckStr := query.Get("timewheretocheck")
	timeWhenToCheckStr := query.Get("timewhentocheck")

	if ruleKey == "" {
		http.Error(
			w,
			"rule_key is required",
			http.StatusBadRequest,
		)
		return
	}

	if subjectID == "" {
		http.Error(
			w,
			"subject_id is required",
			http.StatusBadRequest,
		)
		return
	}

	if timeWhereToCheckStr == "" {
		http.Error(
			w,
			"timewheretocheck is required",
			http.StatusBadRequest,
		)
		return
	}

	if timeWhenToCheckStr == "" {
		http.Error(
			w,
			"timewhentocheck is required",
			http.StatusBadRequest,
		)
		return
	}

	timeWhereToCheck, err := time.Parse(
		time.RFC3339,
		timeWhereToCheckStr,
	)

	if err != nil {
		http.Error(
			w,
			"invalid timewhere; expected RFC3339",
			http.StatusBadRequest,
		)
		return
	}

	timeWhenToCheck, err := time.Parse(
		time.RFC3339,
		timeWhenToCheckStr,
	)

	if err != nil {
		http.Error(
			w,
			"invalid knowledge_time; expected RFC3339",
			http.StatusBadRequest,
		)
		return
	}

	result, err := h.service.Find(
		ctx,
		ruleKey,
		subjectID,
		timeWhenToCheck,
		timeWhereToCheck,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(
				w,
				"rule not found",
				http.StatusNotFound,
			)
			return
		}
		fmt.Println("Error is :", err)
		http.Error(
			w,
			"failed to find rule",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		return
	}
}
