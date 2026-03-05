package vendel

import "fmt"

// VendelError is the base error returned by the SDK.
type VendelError struct {
	StatusCode int
	Message    string
	Detail     map[string]any
}

func (e *VendelError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

// QuotaError is returned when a quota limit is exceeded (HTTP 429).
type QuotaError struct {
	VendelError
	Limit     int
	Used      int
	Available int
}

// IsQuotaError returns true if err is a *QuotaError.
func IsQuotaError(err error) bool {
	_, ok := err.(*QuotaError)
	return ok
}

// IsAPIError returns true if err is a *VendelError (or *QuotaError).
func IsAPIError(err error) bool {
	_, ok := err.(*VendelError)
	if ok {
		return true
	}
	return IsQuotaError(err)
}
