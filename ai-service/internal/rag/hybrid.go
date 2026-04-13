package rag

import (
	"sort"
	"strings"
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

// RetrieveWithContext runs TF-IDF semantic search with entity boosts, then lexical fallback.
func RetrieveWithContext(query string, ctx RetrievalContext, topK int) []KnowledgeChunk {
	if topK <= 0 {
		return nil
	}

	idx := semanticIndex()
	candidates := idx.search(query, len(idx.chunks))
	boosted := applyEntityBoost(candidates, ctx)

	sort.Slice(boosted, func(i, j int) bool { return boosted[i].similarity > boosted[j].similarity })

	best := 0.0
	if len(boosted) > 0 {
		best = boosted[0].similarity
	}

	useSemantic := len(boosted) > 0 && best >= semanticMinSimilarity
	if useSemantic {
		return trimChunks(boosted, topK)
	}

	fallback := RetrieveLexical(query, topK)
	if len(fallback) == 0 {
		return nil
	}
	return fallback
}

func applyEntityBoost(hits []scoredChunk, ctx RetrievalContext) []scoredChunk {
	if len(hits) == 0 {
		return nil
	}
	syms := make([]string, 0, 1+len(ctx.SecondaryTickers))
	if ctx.PrimaryTicker != "" {
		syms = append(syms, strings.ToUpper(ctx.PrimaryTicker))
	}
	for _, s := range ctx.SecondaryTickers {
		syms = append(syms, strings.ToUpper(s))
	}
	out := make([]scoredChunk, len(hits))
	copy(out, hits)
	for i := range out {
		ch := &out[i].chunk
		text := strings.ToUpper(ch.Content + " " + ch.Topic + " " + strings.Join(ch.Tags, " "))
		for _, sym := range syms {
			if sym != "" && strings.Contains(text, sym) {
				out[i].similarity += 0.12
			}
		}
		for _, term := range ctx.ExtraTerms {
			t := strings.ToLower(term)
			if t == "" {
				continue
			}
			if strings.Contains(strings.ToLower(ch.Content+" "+ch.Topic), t) {
				out[i].similarity += 0.06
			}
		}
	}
	return out
}

func trimChunks(scored []scoredChunk, topK int) []KnowledgeChunk {
	out := make([]KnowledgeChunk, 0, topK)
	for i, sc := range scored {
		if i >= topK {
			break
		}
		out = append(out, sc.chunk)
	}
	return out
}
