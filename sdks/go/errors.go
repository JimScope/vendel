package ender

import "fmt"

// EnderError is the base error returned by the SDK.
type EnderError struct {
	StatusCode int
	Message    string
	Detail     map[string]any
}

func (e *EnderError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

// QuotaError is returned when a quota limit is exceeded (HTTP 429).
type QuotaError struct {
	EnderError
	Limit     int
	Used      int
	Available int
}

// IsQuotaError returns true if err is a *QuotaError.
func IsQuotaError(err error) bool {
	_, ok := err.(*QuotaError)
	return ok
}

// IsAPIError returns true if err is an *EnderError (or *QuotaError).
func IsAPIError(err error) bool {
	_, ok := err.(*EnderError)
	if ok {
		return true
	}
	return IsQuotaError(err)
}
