package handlers

import (
	"vendel/services"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// handleServiceError maps known service error types to appropriate HTTP responses.
// Returns nil if the error was not a known type (caller should handle the fallback).
func handleServiceError(e *core.RequestEvent, err error) error {
	if qe, ok := err.(*services.QuotaError); ok {
		return e.JSON(qe.StatusCode, qe.Body)
	}
	return apis.NewBadRequestError(err.Error(), nil)
}
