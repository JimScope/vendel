package templates

import _ "embed"

//go:embed verification.html
var VerificationBody string

const VerificationSubject = "Verify your {APP_NAME} email"
