package service

import (
	"fmt"
	"log"
	"strings"
)

const maxLogPreviewLen = 120

// logChatEvent prints a compact key=value line suitable for grepping and demo
// walkthroughs. Keep payloads small and flat (no large objects/prompts).
func logChatEvent(event string, kv ...any) {
	var b strings.Builder
	b.WriteString("component=chat")
	b.WriteString(" event=")
	b.WriteString(sanitizeForLog(event))

	for i := 0; i+1 < len(kv); i += 2 {
		key := sanitizeForLog(fmt.Sprint(kv[i]))
		val := sanitizeForLog(fmt.Sprint(kv[i+1]))
		if key == "" {
			continue
		}
		b.WriteByte(' ')
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(val)
	}

	log.Print(b.String())
}

func messagePreview(message string) string {
	s := strings.TrimSpace(message)
	if s == "" {
		return ""
	}
	// Keep log lines single-line and compact.
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxLogPreviewLen {
		return s[:maxLogPreviewLen] + "…"
	}
	return s
}

func sanitizeForLog(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), "_")
	return s
}

// LogChatRequestStart logs the HTTP-level request start event.
func LogChatRequestStart(sessionID, message string) {
	logChatEvent("chat_request_start",
		"session_id", sessionID,
		"message_preview", messagePreview(message),
	)
}

// LogChatRequestFailed logs the HTTP-level request failure event.
func LogChatRequestFailed(sessionID, message string, err error) {
	logChatEvent("chat_request_failed",
		"session_id", sessionID,
		"message_preview", messagePreview(message),
		"error", err,
	)
}

// LogChatRequestComplete logs the HTTP-level request completion.
func LogChatRequestComplete(sessionID, message, response string) {
	logChatEvent("chat_request_complete",
		"session_id", sessionID,
		"message_preview", messagePreview(message),
		"response_chars", len(response),
	)
}

