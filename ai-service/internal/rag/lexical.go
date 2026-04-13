package rag

import "strings"

type lexicalHit struct {
	chunk     KnowledgeChunk
	relevance float64
}

// RetrieveLexical scores chunks by overlapping query terms (legacy behavior).
func RetrieveLexical(query string, topK int) []KnowledgeChunk {
	query = strings.ToLower(query)
	scores := make([]lexicalHit, 0)

	for _, ch := range allChunks() {
		score := scoreLexical(query, ch)
		if score > 0 {
			scores = append(scores, lexicalHit{chunk: ch, relevance: score})
		}
	}

	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].relevance > scores[i].relevance {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	out := make([]KnowledgeChunk, 0, topK)
	for i, h := range scores {
		if i >= topK {
			break
		}
		out = append(out, h.chunk)
	}
	return out
}

func scoreLexical(query string, ch KnowledgeChunk) float64 {
	content := strings.ToLower(ch.Content + " " + ch.Topic)
	for _, t := range ch.Tags {
		content += " " + strings.ToLower(t)
	}
	words := strings.Fields(query)
	score := 0.0

	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		if strings.Contains(content, word) {
			score += 1.0
		}
	}
	return score
}
