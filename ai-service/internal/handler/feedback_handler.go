package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/feedback"
)

const (
	defaultListLimit = 100
	maxListLimit     = 1000
)

// RecordFeedbackHandler handles POST /feedback using the process-wide store.
func RecordFeedbackHandler(c *gin.Context) {
	RecordFeedbackHandlerWithStore(feedback.DefaultStore())(c)
}

// RecordFeedbackHandlerWithStore allows injecting a custom store (tests).
func RecordFeedbackHandlerWithStore(store feedback.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body feedback.Feedback
		if !BindJSONOrRespond(c, &body, "invalid JSON") {
			return
		}
		// Server controls these regardless of what the client sent.
		body.ID = ""
		body.Timestamp = time.Time{}

		saved, err := store.Append(body)
		if err != nil {
			RespondError(c, httpStatusForFeedbackError(err), "Request failed", err.Error())
			return
		}
		RespondSuccess(c, http.StatusOK, "Feedback recorded", gin.H{
			"ok":        true,
			"id":        saved.ID,
			"timestamp": saved.Timestamp,
		})
	}
}

// ListFeedbackHandler handles GET /feedback with optional query filters.
func ListFeedbackHandler(c *gin.Context) {
	ListFeedbackHandlerWithStore(feedback.DefaultStore())(c)
}

// ListFeedbackHandlerWithStore allows injecting a custom store (tests).
func ListFeedbackHandlerWithStore(store feedback.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := feedback.ListFilter{
			Feedback:     strings.ToLower(strings.TrimSpace(c.Query("feedback"))),
			ResponseType: strings.ToLower(strings.TrimSpace(c.Query("response_type"))),
			SessionID:    strings.TrimSpace(c.Query("session_id")),
			Limit:        parseLimit(c.Query("limit")),
		}
		if since := strings.TrimSpace(c.Query("since")); since != "" {
			t, err := time.Parse(time.RFC3339, since)
			if err != nil {
				RespondBadRequest(c, "Request failed", "since must be RFC3339")
				return
			}
			filter.Since = t
		}

		items, err := store.List(filter)
		if err != nil {
			RespondError(c, http.StatusInternalServerError, "Request failed", err.Error())
			return
		}
		RespondSuccess(c, http.StatusOK, "Feedback list retrieved", gin.H{
			"count": len(items),
			"items": items,
		})
	}
}

// ExportFeedbackHandler streams the raw NDJSON feedback log.
func ExportFeedbackHandler(c *gin.Context) {
	ExportFeedbackHandlerWithStore(feedback.DefaultStore())(c)
}

// ExportFeedbackHandlerWithStore allows injecting a custom store (tests).
func ExportFeedbackHandlerWithStore(store feedback.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/x-ndjson")
		c.Header("Content-Disposition", `attachment; filename="feedback.jsonl"`)
		if err := store.Export(c.Writer); err != nil {
			RespondError(c, http.StatusInternalServerError, "Request failed", err.Error())
			return
		}
	}
}

// httpStatusForFeedbackError maps known validation errors to 400 and
// everything else to 500.
func httpStatusForFeedbackError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	msg := err.Error()
	if strings.HasPrefix(msg, "feedback: session_id") ||
		strings.HasPrefix(msg, "feedback: response_type") ||
		strings.HasPrefix(msg, "feedback: feedback") {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func parseLimit(raw string) int {
	if raw == "" {
		return defaultListLimit
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultListLimit
	}
	if n > maxListLimit {
		return maxListLimit
	}
	return n
}
