package handlers

import (
	"bufio"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var phoneCleanRegex = regexp.MustCompile(`[^\d+]`)

// RegisterContactRoutes registers contact-related API routes.
func RegisterContactRoutes(se *core.ServeEvent) {
	// POST /api/contacts/import — import contacts from vCard (.vcf) file
	se.Router.POST("/api/contacts/import", func(e *core.RequestEvent) error {
		userId := e.Auth.Id

		file, _, err := e.Request.FormFile("file")
		if err != nil {
			return apis.NewBadRequestError("File is required", nil)
		}
		defer file.Close()

		groupId := e.Request.FormValue("group_id")

		contacts := parseVCard(file)
		if len(contacts) == 0 {
			return apis.NewBadRequestError("No valid contacts found in file", nil)
		}

		collection, err := e.App.FindCollectionByNameOrId("contacts")
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "contacts collection not found", nil)
		}

		imported, skipped := 0, 0
		for _, c := range contacts {
			// Skip if phone already exists for this user
			existing, _ := e.App.FindFirstRecordByFilter(
				"contacts",
				"user = {:userId} && phone_number = {:phone}",
				dbx.Params{"userId": userId, "phone": c.phone},
			)
			if existing != nil {
				skipped++
				continue
			}

			record := core.NewRecord(collection)
			record.Set("user", userId)
			record.Set("name", c.name)
			record.Set("phone_number", c.phone)
			if groupId != "" {
				record.Set("groups", []string{groupId})
			}

			if err := e.App.Save(record); err != nil {
				skipped++
				continue
			}
			imported++
		}

		return e.JSON(http.StatusOK, map[string]any{
			"imported": imported,
			"skipped":  skipped,
			"total":    len(contacts),
		})
	}).Bind(apis.RequireAuth("users"))
}

type vcardEntry struct {
	name  string
	phone string
}

// parseVCard extracts name + phone from a vCard file.
// Supports multiple contacts in a single .vcf file.
func parseVCard(r io.Reader) []vcardEntry {
	var entries []vcardEntry
	var current vcardEntry
	inCard := false

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case line == "BEGIN:VCARD":
			inCard = true
			current = vcardEntry{}

		case line == "END:VCARD":
			if inCard && current.name != "" && current.phone != "" {
				entries = append(entries, current)
			}
			inCard = false

		case inCard && strings.HasPrefix(line, "FN:"):
			current.name = strings.TrimPrefix(line, "FN:")

		case inCard && strings.Contains(line, "TEL"):
			// TEL:+123... or TEL;TYPE=CELL:+123... or TEL;TYPE=CELL;VALUE=uri:tel:+123...
			phone := line
			if idx := strings.LastIndex(phone, ":"); idx >= 0 {
				phone = phone[idx+1:]
			}
			// Clean non-digit chars except +
			phone = phoneCleanRegex.ReplaceAllString(phone, "")
			if e164Regex.MatchString(phone) && current.phone == "" {
				current.phone = phone
			}
		}
	}

	return entries
}
