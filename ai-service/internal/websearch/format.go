package websearch

import (
	"fmt"
	"strings"
)

// FormatForPrompt builds the body string for the chatprompt [Live Web Context]
// section. Returns "" when no results so callers know to emit the "no data
// attached" note.
//
// The format is intentionally compact and source-attributed so the LLM can
// quote it without inventing details:
//
//	1. <Title> — <source> (<published>): <snippet>
//	2. ...
func FormatForPrompt(results []Result) string {
	if len(results) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range results {
		title := strings.TrimSpace(r.Title)
		snippet := strings.TrimSpace(r.Snippet)
		if title == "" || snippet == "" {
			continue
		}
		source := strings.TrimSpace(r.Source)
		if source == "" {
			source = hostnameOf(r.URL, "")
		}
		fmt.Fprintf(&b, "%d. %s — %s", i+1, title, source)
		if pub := strings.TrimSpace(r.PublishedAt); pub != "" {
			fmt.Fprintf(&b, " (%s)", pub)
		}
		b.WriteString(": ")
		b.WriteString(snippet)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}
