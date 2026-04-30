package chatprompt

import (
	"fmt"
	"strings"
)

// Section headers for grounded context (stable for prompts and tests).
const (
	SectionKnowledge   = "[Relevant Financial Knowledge]"
	SectionQAKnowledge = "[Relevant QA Knowledge]"
	SectionLiveMarket  = "[Live Market Data]"
	SectionNews        = "[News Context]"
	SectionWebContext  = "[Live Web Context]"
	SectionPortfolio   = "[Portfolio Context]"
	SectionSystem      = "[System Context]"
)

// UserReplyStyleFooter closes the grounded-context block with a reminder that
// the model's visible reply should be natural prose (see system prompt). It
// does not add facts—only style guidance for the assistant turn.
const UserReplyStyleFooter = "\n\n--- End of grounded context ---\nWhen you reply to the user, write in natural paragraphs with a blank line between paragraphs. The bracketed sections above are internal reference only—do not echo their titles or use rigid labels (Description:, Summary:, Value:, Risk:, etc.) in your answer."

// EnvelopeInput carries optional context blocks for the user message sent to the LLM.
type EnvelopeInput struct {
	UserMessage      string
	KnowledgeLines   []string // each full line e.g. "[topic] content"
	QAKnowledgeLines []string // curated Q&A CSV retrieval lines
	LiveMarketBody   string   // preformatted quote + fundamentals, or empty
	NewsBody         string   // preformatted headlines, or empty
	WebContextBody   string   // preformatted live web search results, or empty
	PortfolioBody    string   // empty → default "no portfolio" line
	Intent           string
	Ticker           string
	Sentiment        string
	IntentConfidence float64
}

// BuildUserContent assembles the user turn: original message + labeled grounding sections.
func BuildUserContent(in EnvelopeInput) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(in.UserMessage))
	b.WriteString("\n\n--- Grounded context (cite only facts from these sections; if a section says no data was provided, state that clearly) ---")

	writeSection := func(title, body string, emptyNote string) {
		b.WriteString("\n\n")
		b.WriteString(title)
		b.WriteByte('\n')
		if strings.TrimSpace(body) == "" {
			b.WriteString(emptyNote)
		} else {
			b.WriteString(strings.TrimSpace(body))
		}
	}

	// Knowledge
	b.WriteString("\n\n")
	b.WriteString(SectionKnowledge)
	b.WriteByte('\n')
	if len(in.KnowledgeLines) == 0 {
		b.WriteString("(No curated knowledge snippets were retrieved for this query.)")
	} else {
		for _, line := range in.KnowledgeLines {
			b.WriteString("- ")
			b.WriteString(strings.TrimSpace(line))
			b.WriteByte('\n')
		}
	}

	b.WriteString("\n\n")
	b.WriteString(SectionQAKnowledge)
	b.WriteByte('\n')
	if len(in.QAKnowledgeLines) == 0 {
		b.WriteString("(No matching Q&A knowledge rows were retrieved for this query.)")
	} else {
		for _, line := range in.QAKnowledgeLines {
			b.WriteString("- ")
			b.WriteString(strings.TrimSpace(line))
			b.WriteByte('\n')
		}
	}

	writeSection(SectionLiveMarket, in.LiveMarketBody,
		"(No live market data was attached for this request.)")

	writeSection(SectionNews, in.NewsBody,
		"(No news headlines were attached for this request.)")

	writeSection(SectionWebContext, in.WebContextBody,
		"(No live web search results were attached for this request.)")

	portfolio := strings.TrimSpace(in.PortfolioBody)
	if portfolio == "" {
		portfolio = "(No portfolio holdings or allocation were provided in this request.)"
	}
	writeSection(SectionPortfolio, portfolio, "")

	b.WriteString("\n\n")
	b.WriteString(SectionSystem)
	b.WriteString(fmt.Sprintf(
		"\nIntent: %s | Primary ticker (if any): %s | Message sentiment (lexical): %s | Intent confidence: %.2f",
		in.Intent,
		strings.ToUpper(strings.TrimSpace(in.Ticker)),
		in.Sentiment,
		in.IntentConfidence,
	))

	b.WriteString(UserReplyStyleFooter)
	return b.String()
}
