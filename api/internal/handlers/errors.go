package handlers

import (
	"net/http"

	"github.com/isaquebock/flags-api/internal/respond"
)

func WriteError(w http.ResponseWriter, status int, code, message string) {
	respond.Error(w, status, code, message)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	respond.JSON(w, status, data)
}
