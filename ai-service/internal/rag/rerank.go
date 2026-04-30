package rag

import (
	"math"
	"sort"
	"strings"
)

// RankWeights tunes the hybrid ranking formula:
//
//	final = Semantic*sem + Lexical*lex + Entity*ent + Metadata*meta
//
// Weights should sum to 1.0 so the resulting score stays in [0, 1] and is
// directly interpretable in demos and debugging logs.
type RankWeights struct {
	Semantic float64
	Lexical  float64
	Entity   float64
	Metadata float64
}

// DefaultRankWeights are the production defaults. Semantic dominates, lexical
// catches keyword-overlap that the TF-IDF cosine misses on short queries,
// entity boosts ticker/company hits, metadata weights priority/source.
var DefaultRankWeights = RankWeights{
	Semantic: 0.55,
	Lexical:  0.20,
	Entity:   0.15,
	Metadata: 0.10,
}

// rankedChunk is the internal representation of a candidate after rerank.
// Sub-scores are kept so the layer is explainable (debug/inspection use).
type rankedChunk struct {
	chunk      KnowledgeChunk
	finalScore float64
	semantic   float64
	lexical    float64
	entity     float64
	metadata   float64
}

// Rerank fuses semantic, lexical, entity and metadata signals into a single
// final score per candidate, then orders by final score descending.
//
// A candidate is eligible only when at least one of the textual signals is
// non-trivial (semantic >= semanticMinSimilarity, or lexical_norm > 0, or
// entity_boost > 0). This guarantees that pure metadata floor never lifts a
// non-matching chunk above the fold (no-match queries still return empty).
func Rerank(query string, candidates []scoredChunk, ctx RetrievalContext, w RankWeights) []rankedChunk {
	if len(candidates) == 0 {
		return nil
	}
	qTerms := uniqueLexicalTerms(tokenize(query))
	out := make([]rankedChunk, 0, len(candidates))
	for _, c := range candidates {
		sem := clamp01(c.similarity)
		lex := lexicalNormalized(qTerms, c.chunk)
		ent := entityBoost(c.chunk, ctx)
		meta := metadataBoost(c.chunk)

		eligible := sem >= semanticMinSimilarity || lex > 0 || ent > 0
		if !eligible {
			continue
		}

		final := w.Semantic*sem + w.Lexical*lex + w.Entity*ent + w.Metadata*meta
		out = append(out, rankedChunk{
			chunk:      c.chunk,
			finalScore: final,
			semantic:   sem,
			lexical:    lex,
			entity:     ent,
			metadata:   meta,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].finalScore > out[j].finalScore
	})
	return out
}

func trimRanked(scored []rankedChunk, topK int) []KnowledgeChunk {
	if topK <= 0 {
		return nil
	}
	if topK > len(scored) {
		topK = len(scored)
	}
	out := make([]KnowledgeChunk, 0, topK)
	for i := 0; i < topK; i++ {
		out = append(out, scored[i].chunk)
	}
	return out
}

// uniqueLexicalTerms keeps only terms with len >= 3 (drops noise like "is",
// "an", short stop-words). Deduplicates so repeated query words don't bias.
func uniqueLexicalTerms(toks []string) []string {
	seen := make(map[string]struct{}, len(toks))
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		if len(t) < 3 {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// lexicalNormalized = unique query terms found in (content+topic+tags) divided
// by total unique query terms, in [0, 1]. Length-normalised so chatty long
// chunks don't get an unfair edge.
func lexicalNormalized(qTerms []string, ch KnowledgeChunk) float64 {
	if len(qTerms) == 0 {
		return 0
	}
	text := strings.ToLower(ch.Content + " " + ch.Topic + " " + strings.Join(ch.Tags, " "))
	hits := 0
	for _, t := range qTerms {
		if strings.Contains(text, t) {
			hits++
		}
	}
	return clamp01(float64(hits) / float64(len(qTerms)))
}

// entityBoost weighs ticker / company evidence saturating to [0, 1].
//
// Saturation curve 1 - exp(-raw/2):
//
//	raw  0   -> 0.00
//	raw  1   -> 0.39
//	raw  2   -> 0.63
//	raw  3   -> 0.78
//	raw  5   -> 0.92
//
// This prevents a chunk that mentions a ticker many times from steamrolling
// the rest of the signals, while still rewarding multi-source confirmation.
func entityBoost(ch KnowledgeChunk, ctx RetrievalContext) float64 {
	if ctx.PrimaryTicker == "" && len(ctx.SecondaryTickers) == 0 && len(ctx.ExtraTerms) == 0 {
		return 0
	}
	upperContent := strings.ToUpper(ch.Content + " " + ch.Topic)
	upperTags := strings.ToUpper(strings.Join(ch.Tags, " "))
	lowerHay := strings.ToLower(ch.Content + " " + ch.Topic + " " + strings.Join(ch.Tags, " "))

	var raw float64
	if sym := strings.ToUpper(strings.TrimSpace(ctx.PrimaryTicker)); sym != "" {
		if strings.Contains(upperTags, sym) {
			raw += 1.5
		}
		if strings.Contains(upperContent, sym) {
			raw += 1.0
		}
	}
	for _, s := range ctx.SecondaryTickers {
		sym := strings.ToUpper(strings.TrimSpace(s))
		if sym == "" {
			continue
		}
		if strings.Contains(upperTags, sym) {
			raw += 1.0
		}
		if strings.Contains(upperContent, sym) {
			raw += 0.5
		}
	}
	for _, term := range ctx.ExtraTerms {
		t := strings.ToLower(strings.TrimSpace(term))
		if t == "" {
			continue
		}
		if strings.Contains(lowerHay, t) {
			raw += 0.5
		}
	}
	if raw == 0 {
		return 0
	}
	return clamp01(1 - math.Exp(-raw/2.0))
}

// metadataBoost extracts a priority/source signal from chunk metadata.
// Falls back to a neutral 0.50 baseline so metadata never penalises chunks
// that simply lack the field (e.g., the static KB).
func metadataBoost(ch KnowledgeChunk) float64 {
	if ch.Metadata == nil {
		return 0.50
	}
	switch strings.ToLower(strings.TrimSpace(ch.Metadata["priority"])) {
	case "high":
		return 0.85
	case "medium":
		return 0.50
	case "low":
		return 0.20
	}
	return 0.50
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
