package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/service"
	"github.com/google/uuid"
)

// Doing dependency injection
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
	err = h.service.Create(
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

	if err := json.NewEncoder(w).Encode(factInformation); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}
