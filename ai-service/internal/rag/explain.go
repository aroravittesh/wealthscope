package rag

import (
	"fmt"
	"strings"

	"wealthscope-ai/internal/explain"
)

// RetrievalHit pairs a retrieved chunk with the breakdown of its hybrid score
// and a structured explanation of why it was selected. This is intended for
// debugging, /chat grounding diagnostics, and the demo-friendly "why this
// chunk?" panel — it is *not* required by the legacy retrieval callers.
type RetrievalHit struct {
	Chunk       KnowledgeChunk      `json:"chunk"`
	FinalScore  float64             `json:"final_score"`
	Semantic    float64             `json:"semantic"`
	Lexical     float64             `json:"lexical"`
	Entity      float64             `json:"entity"`
	Metadata    float64             `json:"metadata"`
	Reason      string              `json:"reason"`
	Explanation explain.Explanation `json:"explanation"`
}

// RetrieveWithExplanation runs the same hybrid pipeline as RetrieveWithContext
// but exposes the per-signal sub-scores and a structured explanation per hit.
//
// The base RetrieveWithContext / Retrieve helpers continue to work unchanged;
// this is a non-breaking, opt-in inspection surface.
func RetrieveWithExplanation(query string, ctx RetrievalContext, topK int) []RetrievalHit {
	if topK <= 0 {
		return nil
	}
	idx := semanticIndex()
	candidates := idx.search(query, len(idx.chunks))
	ranked := Rerank(query, candidates, ctx, DefaultRankWeights)
	if len(ranked) == 0 {
		return nil
	}
	if topK > len(ranked) {
		topK = len(ranked)
	}
	out := make([]RetrievalHit, 0, topK)
	for i := 0; i < topK; i++ {
		r := ranked[i]
		reason := retrievalReason(r)
		hit := RetrievalHit{
			Chunk:       r.chunk,
			FinalScore:  r.finalScore,
			Semantic:    r.semantic,
			Lexical:     r.lexical,
			Entity:      r.entity,
			Metadata:    r.metadata,
			Reason:      reason,
			Explanation: buildRetrievalExplanation(r, reason),
		}
		out = append(out, hit)
	}
	return out
}

func retrievalReason(r rankedChunk) string {
	parts := make([]string, 0, 4)
	if r.semantic >= semanticMinSimilarity {
		parts = append(parts, "semantic match")
	}
	if r.lexical > 0 {
		parts = append(parts, "lexical overlap")
	}
	if r.entity > 0 {
		parts = append(parts, "entity boost")
	}
	if r.metadata > 0.5 {
		parts = append(parts, "metadata priority boost")
	}
	if len(parts) == 0 {
		return "neutral baseline"
	}
	return strings.Join(parts, " + ")
}

func buildRetrievalExplanation(r rankedChunk, reason string) explain.Explanation {
	signals := []explain.Signal{
		{Code: "RAG_SEMANTIC", Label: "TF-IDF cosine", Score: r.semantic, Detail: "Semantic similarity from the TF-IDF index."},
		{Code: "RAG_LEXICAL", Label: "Keyword overlap", Score: r.lexical, Detail: "Length-normalised keyword overlap."},
		{Code: "RAG_ENTITY", Label: "Entity boost", Score: r.entity, Detail: "Saturating boost from primary/secondary tickers and aliases."},
		{Code: "RAG_METADATA", Label: "Metadata priority", Score: r.metadata, Detail: "Source priority / weighting from chunk metadata."},
	}
	return explain.Explanation{
		Code:    "RAG_HYBRID_RANKING",
		Source:  "hybrid_retriever",
		Summary: fmt.Sprintf("Chunk %q retrieved with hybrid score %.2f (%s).", r.chunk.Topic, r.finalScore, reason),
		Reasons: []string{
			"Hybrid ranking combines semantic, lexical, entity and metadata signals.",
			"Top-1 driver is " + dominantSignal(r) + ".",
		},
		TopSignals: signals,
	}
}

func dominantSignal(r rankedChunk) string {
	type kv struct {
		label string
		score float64
	}
	all := []kv{
		{"semantic similarity", r.semantic * DefaultRankWeights.Semantic},
		{"lexical overlap", r.lexical * DefaultRankWeights.Lexical},
		{"entity boost", r.entity * DefaultRankWeights.Entity},
		{"metadata priority", r.metadata * DefaultRankWeights.Metadata},
	}
	bestLabel := all[0].label
	bestScore := all[0].score
	for _, x := range all[1:] {
		if x.score > bestScore {
			bestScore = x.score
			bestLabel = x.label
		}
	}
	return bestLabel
}
