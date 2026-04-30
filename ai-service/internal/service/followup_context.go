package service

import (
	"regexp"
	"strings"

	"wealthscope-ai/internal/entity"
	"wealthscope-ai/internal/openai"
)

var pronounRefRE = regexp.MustCompile(`\b(it|its|that|this|those|them)\b`)

type followUpContext struct {
	CarriedTicker string
	Reason        string
}

func resolveFollowUpContext(sessionID, message string) followUpContext {
	msg := strings.ToLower(strings.TrimSpace(message))
	if msg == "" || !looksReferential(msg) || sessionID == "" {
		return followUpContext{}
	}
	history := openai.SessionMessages(sessionID)
	if len(history) == 0 {
		return followUpContext{}
	}

	// Walk newest -> oldest and carry the first primary ticker we can recover.
	for i := len(history) - 1; i >= 0; i-- {
		turn := strings.TrimSpace(history[i].Content)
		if turn == "" {
			continue
		}
		ent := entity.Extract(turn)
		if ent.PrimaryTicker != "" {
			return followUpContext{CarriedTicker: ent.PrimaryTicker, Reason: "referential_followup"}
		}
	}
	return followUpContext{}
}

func applyFollowUpCarryover(message string, ctx followUpContext) string {
	raw := strings.TrimSpace(message)
	if raw == "" || ctx.CarriedTicker == "" {
		return raw
	}
	lower := strings.ToLower(raw)
	if strings.Contains(lower, strings.ToLower(ctx.CarriedTicker)) {
		return raw
	}

	switch {
	case containsAny(lower, "compare it", "compare that", "compare this", "compare with", "compare"):
		// "compare it with Microsoft" -> "compare AAPL with Microsoft"
		return strings.Replace(raw, "it", ctx.CarriedTicker, 1)
	case containsAny(lower, "latest news", "news too", "news", "what about its risk", "its risk", "that stock", "the company", "show me more"):
		return raw + " about " + ctx.CarriedTicker
	case containsAny(lower, "how does that affect my portfolio", "that affect my portfolio", "my portfolio"):
		return raw + " for " + ctx.CarriedTicker
	default:
		return raw + " regarding " + ctx.CarriedTicker
	}
}

func looksReferential(lowerMessage string) bool {
	if pronounRefRE.MatchString(lowerMessage) {
		return true
	}
	return containsAny(lowerMessage,
		"that stock", "that company", "the company", "that one",
		"show me more", "latest news too", "what about", "how does that",
		"compare it", "compare that",
	)
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

