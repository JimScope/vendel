package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// Ensure core import is used (for template ownership test setup)
var _ = core.CollectionNameSuperusers

func TestSendSMS(t *testing.T) {
	userToken := generateTestToken(t, "users", "user@test.com")

	scenarios := []tests.ApiScenario{
		{
			Name:            "guest cannot send SMS",
			Method:          http.MethodPost,
			URL:             "/api/sms/send",
			Body:            stringBody(`{"recipients":["+1234567890"],"body":"Hello"}`),
			ExpectedStatus:  401,
			ExpectedContent: []string{`"message"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "missing body returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/send",
			Body:   stringBody(`{"recipients":["+1234567890"]}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "missing recipients returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/send",
			Body:   stringBody(`{"body":"Hello"}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "invalid E.164 number returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/send",
			Body:   stringBody(`{"recipients":["not-a-number"],"body":"Hello"}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`Invalid phone number`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "valid send returns accepted",
			Method: http.MethodPost,
			URL:    "/api/sms/send",
			Body:   stringBody(`{"recipients":["+1234567890"],"body":"Hello world"}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"status":"accepted"`, `"batch_id"`, `"message_ids"`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestSendSMSTemplate(t *testing.T) {
	userToken := generateTestToken(t, "users", "user@test.com")

	// Get the template ID from the seeded data
	app, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := app.FindFirstRecordByFilter("sms_templates", "name = 'Welcome Template'")
	if err != nil {
		t.Fatal(err)
	}
	templateID := tmpl.Id
	app.Cleanup()

	scenarios := []tests.ApiScenario{
		{
			Name:            "guest cannot send template SMS",
			Method:          http.MethodPost,
			URL:             "/api/sms/send-template",
			Body:            stringBody(`{"recipients":["+1234567890"],"template_id":"fake"}`),
			ExpectedStatus:  401,
			ExpectedContent: []string{`"message"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "missing template_id returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/send-template",
			Body:   stringBody(`{"recipients":["+1234567890"]}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`Template_id required`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "non-existent template returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/send-template",
			Body:   stringBody(`{"recipients":["+1234567890"],"template_id":"nonexistent123456"}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`not found`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "missing custom variable returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/send-template",
			Body:   stringBody(`{"recipients":["+1234567890"],"template_id":"` + templateID + `"}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`Missing variable: code`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "valid template send returns accepted",
			Method: http.MethodPost,
			URL:    "/api/sms/send-template",
			Body:   stringBody(`{"recipients":["+1234567890"],"template_id":"` + templateID + `","variables":{"code":"1234"}}`),
			Headers: map[string]string{
				"Authorization": userToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"status":"accepted"`, `"batch_id"`, `"message_ids"`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestSendSMSReport(t *testing.T) {
	scenarios := []tests.ApiScenario{
		{
			Name:            "guest cannot report",
			Method:          http.MethodPost,
			URL:             "/api/sms/report",
			Body:            stringBody(`{"message_id":"fake","status":"sent"}`),
			ExpectedStatus:  401,
			ExpectedContent: []string{`"message"`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "invalid status returns error",
			Method: http.MethodPost,
			URL:    "/api/sms/report",
			Body:   stringBody(`{"message_id":"fake","status":"invalid_status"}`),
			Headers: map[string]string{
				"X-API-Key": "test_device_key_123",
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`Invalid status`},
			TestAppFactory:  setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

// stringBody is a helper that wraps a JSON string for ApiScenario.Body.
func stringBody(s string) *strings.Reader {
	return strings.NewReader(s)
}
