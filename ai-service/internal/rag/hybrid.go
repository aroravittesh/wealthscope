package rag

import (
	"sync"

	"wealthscope-ai/internal/entity"
)

var (
	indexOnce sync.Once
	corpusIdx *tfidfIndex
)

func semanticIndex() *tfidfIndex {
	indexOnce.Do(func() {
		corpusIdx = newTFIDFIndex(allChunks())
	})
	return corpusIdx
}

// RetrievalContext carries lightweight hints for grounding (tickers, extra terms).
type RetrievalContext struct {
	PrimaryTicker    string
	SecondaryTickers []string
	ExtraTerms       []string
}

// RetrievalContextFromEntity builds hints from entity extraction output.
func RetrievalContextFromEntity(ent entity.EntityResult) RetrievalContext {
	return RetrievalContext{
		PrimaryTicker:    ent.PrimaryTicker,
		SecondaryTickers: ent.SecondaryTickers,
		ExtraTerms:       ent.CompanyMatches,
	}
}

// RetrieveWithContext returns the top-K knowledge chunks for a query, ranked
// by the hybrid reranker (semantic + lexical + entity + metadata). If the
// reranker filters out every candidate (no textual overlap at all), it falls
// back to legacy pure-lexical retrieval to preserve old behavior.
func RetrieveWithContext(query string, ctx RetrievalContext, topK int) []KnowledgeChunk {
	if topK <= 0 {
		return nil
	}

	idx := semanticIndex()
	candidates := idx.search(query, len(idx.chunks))
	ranked := Rerank(query, candidates, ctx, DefaultRankWeights)
	if len(ranked) > 0 {
		return trimRanked(ranked, topK)
	}

	fallback := RetrieveLexical(query, topK)
	if len(fallback) == 0 {
		return nil
	}
	return fallback
}
