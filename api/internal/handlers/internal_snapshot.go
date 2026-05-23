package handlers

import (
	"net/http"

	"github.com/isaquebock/flags-api/internal/middleware"
	"github.com/isaquebock/flags-api/internal/snapshot"
)

type FlagsHandler struct {
	store snapshot.Store
}

func NewFlagsHandler(store snapshot.Store) *FlagsHandler {
	return &FlagsHandler{store: store}
}

func (h *FlagsHandler) InternalSnapshot(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.ClientIDFromContext(r.Context())

	snap, err := h.store.Get(r.Context(), clientID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch snapshot")
		return
	}

	if snap == nil || (len(snap.Flags) == 0 && snap.GeneratedAt == nil) {
		WriteError(w, http.StatusNotFound, "snapshot_not_found", "No snapshot found for this client")
		return
	}

	WriteJSON(w, http.StatusOK, snap)
}
