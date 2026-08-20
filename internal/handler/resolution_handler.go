package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/service"
	"github.com/google/uuid"
)

type ResolutionHandler struct {
	service service.ResolutionService
}

func NewResolutionHandler(
	service service.ResolutionService,
) *ResolutionHandler {
	return &ResolutionHandler{
		service: service,
	}
}

// Get resolution by resolution ID.
func (h *ResolutionHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	resolutionID, err := uuid.Parse(
		r.PathValue("resolutionID"),
	)

	if err != nil {
		http.Error(
			w,
			"invalid resolution id",
			http.StatusBadRequest,
		)
		return
	}

	resolution, err := h.service.Get(
		r.Context(),
		resolutionID,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(resolution); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}

// Create a new resolution version.
func (h *ResolutionHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req dto.CreateResolutionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	// Get Idempotency-Key
	idempotentKeyString := r.Header.Get(
		"Idempotency-Key",
	)

	if idempotentKeyString == "" {
		http.Error(
			w,
			"missing Idempotency-Key header",
			http.StatusBadRequest,
		)
		return
	}

	idempotentKey, err := uuid.Parse(
		idempotentKeyString,
	)

	if err != nil {
		http.Error(
			w,
			"invalid Idempotency-Key",
			http.StatusBadRequest,
		)
		return
	}

	// Create resolution

	err = h.service.Create(
		r.Context(),
		&req,
		idempotentKey,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// Response

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"message":           "resolution created successfully",
		"rule_key":          req.RuleKey,
		"tenant_id":         req.TenantID,
		"resolution_policy": req.ResolutionPolicy,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}
