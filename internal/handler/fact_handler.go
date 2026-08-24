package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/service"
	"github.com/google/uuid"
)

// Doing dependency injection
// Added the tenant id based logic
type FactHandler struct {
	service service.FactService
}

func NewFactHandler(service service.FactService) *FactHandler {
	return &FactHandler{
		service: service,
	}
}

func (h *FactHandler) Get(w http.ResponseWriter, r *http.Request) {
	factInformationID, err := uuid.Parse(
		r.PathValue("factInformationID"),
	)
	if err != nil {
		http.Error(
			w,
			"invalid fact information id",
			http.StatusBadRequest,
		)
		return
	}

	factInformation, err := h.service.Get(
		r.Context(),
		factInformationID,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(factInformation); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}

// Now the fact insertion
func (h *FactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var factInformation model.FactInformation
	if err := json.NewDecoder(r.Body).Decode(&factInformation); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	// Getting and parsing the idempotent key
	idempotentKeyString := r.Header.Get("Idempotency-Key")

	if idempotentKeyString == "" {
		http.Error(
			w,
			"missing Idempotency-Key header",
			http.StatusBadRequest,
		)
		return
	}

	idempotentKey, err := uuid.Parse(idempotentKeyString)
	if err != nil {
		http.Error(
			w,
			"invalid Idempotency-Key",
			http.StatusBadRequest,
		)
		return
	}
	createdFact, err := h.service.Create(
		r.Context(),
		&factInformation,
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdFact); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}

// Fetching the fact handler
func (h *FactHandler) FindByEffectiveTime(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query()

	factKey := query.Get("fact_key")
	if factKey == "" {
		http.Error(
			w,
			"missing fact_key",
			http.StatusBadRequest,
		)
		return
	}

	tenantID := query.Get("tenant_id")
	if tenantID == "" {
		http.Error(
			w,
			"missing tenant Id",
			http.StatusBadRequest,
		)
		return
	}

	subjectID := query.Get("subject_id")
	if subjectID == "" {
		http.Error(
			w,
			"missing subject_id",
			http.StatusBadRequest,
		)
		return
	}

	timeWhereToCheckString := query.Get("time_where_to_check")
	if timeWhereToCheckString == "" {
		http.Error(
			w,
			"missing time_where_to_check",
			http.StatusBadRequest,
		)
		return
	}

	timeWhenToCheckString := query.Get("time_when_to_check")
	if timeWhenToCheckString == "" {
		http.Error(
			w,
			"missing time_when_to_check",
			http.StatusBadRequest,
		)
		return
	}

	timeWhereToCheck, err := time.Parse(
		time.RFC3339,
		timeWhereToCheckString,
	)
	if err != nil {
		http.Error(
			w,
			"invalid time_where_to_check",
			http.StatusBadRequest,
		)
		return
	}

	timeWhenToCheck, err := time.Parse(
		time.RFC3339,
		timeWhenToCheckString,
	)
	if err != nil {
		http.Error(
			w,
			"invalid time_when_to_check",
			http.StatusBadRequest,
		)
		return
	}

	facts, err := h.service.FindByEffectiveTime(
		r.Context(),
		factKey,
		subjectID,
		tenantID,
		timeWhereToCheck,
		timeWhenToCheck,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(facts); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}
