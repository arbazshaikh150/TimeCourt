package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/service"
	"github.com/google/uuid"
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
