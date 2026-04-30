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
		"[Live Web Context]",
		"[Portfolio Context]",
		"[System Context]",
	} {
		if !strings.Contains(p, needle) {
			t.Fatalf("system prompt missing %q", needle)
		}
	}
}

func TestGetSystemPrompt_NaturalProseStyleNotRigidTemplate(t *testing.T) {
	p := getSystemPrompt()
	if strings.Contains(p, "**Explanation**") {
		t.Fatal("system prompt should not mandate rigid **Explanation** template")
	}
	if strings.Contains(p, "ANSWER FORMAT (when answering") {
		t.Fatal("old ANSWER FORMAT block should be replaced with prose guidance")
	}
	for _, needle := range []string{
		"natural paragraphs",
		"Description:",
		"HOW TO WRITE YOUR REPLY",
		"Investing in the stock market involves risk. WealthScope does not guarantee",
	} {
		if !strings.Contains(p, needle) {
			t.Fatalf("system prompt missing %q", needle)
		}
	}
}
