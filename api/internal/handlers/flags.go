package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/isaquebock/flags-api/internal/middleware"
	"github.com/isaquebock/flags-api/internal/snapshot"
	"github.com/isaquebock/flags-api/internal/validation"
)

type CreateFlagRequest struct {
	Key         string `json:"key"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type UpdateFlagRequest struct {
	Enabled     *bool  `json:"enabled"`
	Description *string `json:"description"`
}

func (h *FlagsHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.ClientIDFromContext(r.Context())

	snap, err := h.store.Get(r.Context(), clientID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list flags")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"flags": snap.Flags,
	})
}

func (h *FlagsHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.ClientIDFromContext(r.Context())

	var req CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if err := validation.ValidateFlagKey(req.Key); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_key", err.Error())
		return
	}

	if err := validation.ValidateDescription(req.Description); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_description", err.Error())
		return
	}

	snap, err := h.store.Mutate(r.Context(), clientID, func(s *snapshot.Snapshot) error {
		return s.AddFlag(req.Key, req.Enabled, req.Description)
	})

	if err == snapshot.ErrFlagExists {
		WriteError(w, http.StatusConflict, "flag_already_exists", "Flag with this key already exists")
		return
	}

	if err == snapshot.ErrTooManyRetries {
		WriteError(w, http.StatusServiceUnavailable, "concurrent_write_retry_exhausted", "Too many concurrent writes, please retry")
		return
	}

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create flag")
		return
	}

	flag := snap.Flags[req.Key]
	WriteJSON(w, http.StatusCreated, flag)
}

func (h *FlagsHandler) Get(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.ClientIDFromContext(r.Context())
	key := chi.URLParam(r, "key")

	snap, err := h.store.Get(r.Context(), clientID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get flag")
		return
	}

	flag, exists := snap.Flags[key]
	if !exists {
		WriteError(w, http.StatusNotFound, "flag_not_found", "Flag not found")
		return
	}

	WriteJSON(w, http.StatusOK, flag)
}

func (h *FlagsHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.ClientIDFromContext(r.Context())
	key := chi.URLParam(r, "key")

	var req UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if req.Description != nil {
		if err := validation.ValidateDescription(*req.Description); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid_description", err.Error())
			return
		}
	}

	snap, err := h.store.Mutate(r.Context(), clientID, func(s *snapshot.Snapshot) error {
		return s.UpdateFlag(key, req.Enabled, req.Description)
	})

	if err == snapshot.ErrFlagNotFound {
		WriteError(w, http.StatusNotFound, "flag_not_found", "Flag not found")
		return
	}

	if err == snapshot.ErrTooManyRetries {
		WriteError(w, http.StatusServiceUnavailable, "concurrent_write_retry_exhausted", "Too many concurrent writes, please retry")
		return
	}

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to update flag")
		return
	}

	flag := snap.Flags[key]
	WriteJSON(w, http.StatusOK, flag)
}

func (h *FlagsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.ClientIDFromContext(r.Context())
	key := chi.URLParam(r, "key")

	_, err := h.store.Mutate(r.Context(), clientID, func(s *snapshot.Snapshot) error {
		return s.RemoveFlag(key)
	})

	if err == snapshot.ErrFlagNotFound {
		WriteError(w, http.StatusNotFound, "flag_not_found", "Flag not found")
		return
	}

	if err == snapshot.ErrTooManyRetries {
		WriteError(w, http.StatusServiceUnavailable, "concurrent_write_retry_exhausted", "Too many concurrent writes, please retry")
		return
	}

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete flag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
