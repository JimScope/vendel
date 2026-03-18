package templates

import _ "embed"

//go:embed verification.html
var VerificationBody string

const VerificationSubject = "Verify your {APP_NAME} email"

//go:embed reset-password.html
var ResetPasswordBody string

const ResetPasswordSubject = "Reset your {APP_NAME} password"
