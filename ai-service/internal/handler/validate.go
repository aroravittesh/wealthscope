package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultSessionID = "default"

// BindJSONOrRespond binds JSON into dst and writes a standardized 400 response
// if parsing fails. Returns true when binding succeeded.
func BindJSONOrRespond(c *gin.Context, dst any, invalidMessage string) bool {
	if err := c.BindJSON(dst); err != nil {
		RespondBadRequest(c, "Request failed", invalidMessage)
		return false
	}
	return true
}

// RequiredTrimmed ensures s is non-empty after trim. It writes a 400 error on
// failure and returns ("", false). On success it returns (trimmed, true).
func RequiredTrimmed(c *gin.Context, s string, errMessage string) (string, bool) {
	out := strings.TrimSpace(s)
	if out == "" {
		RespondBadRequest(c, "Request failed", errMessage)
		return "", false
	}
	return out, true
}

// NormalizeSessionID trims session_id and applies the default when empty.
func NormalizeSessionID(raw string) string {
	id := strings.TrimSpace(raw)
	if id == "" {
		return defaultSessionID
	}
	return id
}

// RequirePathParam validates a required route param and returns trimmed value.
func RequirePathParam(c *gin.Context, name string, errMessage string) (string, bool) {
	return RequiredTrimmed(c, c.Param(name), errMessage)
}

// RequireNonEmptySlice validates a required non-empty slice.
func RequireNonEmptySlice[T any](c *gin.Context, s []T, errMessage string) bool {
	if len(s) == 0 {
		RespondBadRequest(c, "Request failed", errMessage)
		return false
	}
	return true
}

