package middleware

import (
	"github.com/pocketbase/pocketbase/core"
)

// SecurityHeadersMiddleware sets security-related HTTP headers on all responses.
func SecurityHeadersMiddleware(e *core.RequestEvent) error {
	h := e.Response.Header()

	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	return e.Next()
}
