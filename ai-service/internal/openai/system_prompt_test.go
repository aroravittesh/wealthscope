package openai

import (
	"strings"
	"testing"
)

func TestGetSystemPrompt_MentionsGroundingSections(t *testing.T) {
	p := getSystemPrompt()
	for _, needle := range []string{
		"[Relevant Financial Knowledge]",
		"[Relevant QA Knowledge]",
		"[Live Market Data]",
		"[News Context]",
		"[Portfolio Context]",
		"[System Context]",
	} {
		if !strings.Contains(p, needle) {
			t.Fatalf("system prompt missing %q", needle)
		}
	}
}
