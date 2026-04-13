package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

const semanticMinSimilarity = 0.08

type scoredChunk struct {
	chunk      KnowledgeChunk
	similarity float64
}

type tfidfIndex struct {
	chunks []KnowledgeChunk
	vecs   [][]float64
	vocab  []string
	idf    []float64
}

func newTFIDFIndex(chunks []KnowledgeChunk) *tfidfIndex {
	docToks := make([][]string, len(chunks))
	allTerms := make(map[string]int)
	for i, ch := range chunks {
		toks := tokenize(ch.Content + " " + ch.Topic + " " + strings.Join(ch.Tags, " "))
		docToks[i] = toks
		seen := make(map[string]bool)
		for _, t := range toks {
			if !seen[t] {
				seen[t] = true
				allTerms[t]++
			}
		}
	}

	vocab := make([]string, 0, len(allTerms))
	for t := range allTerms {
		vocab = append(vocab, t)
	}
	sort.Strings(vocab)

	idf := make([]float64, len(vocab))
	n := float64(len(chunks))
	for i, term := range vocab {
		df := float64(allTerms[term])
		idf[i] = math.Log(1.0 + n/(1.0+df))
	}

	vecs := make([][]float64, len(chunks))
	for i, toks := range docToks {
		vecs[i] = termFreqIDF(toks, vocab, idf)
		l2Normalize(vecs[i])
	}

	return &tfidfIndex{
		chunks: chunks,
		vecs:   vecs,
		vocab:  vocab,
		idf:    idf,
	}
}

func (idx *tfidfIndex) search(query string, topK int) []scoredChunk {
	if idx == nil || len(idx.chunks) == 0 {
		return nil
	}
	qToks := tokenize(query)
	if len(qToks) == 0 {
		return nil
	}
	qVec := termFreqIDF(qToks, idx.vocab, idx.idf)
	l2Normalize(qVec)

	hits := make([]scoredChunk, 0, len(idx.chunks))
	for i, dv := range idx.vecs {
		sim := cosine(qVec, dv)
		hits = append(hits, scoredChunk{chunk: idx.chunks[i], similarity: sim})
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].similarity > hits[j].similarity })
	if topK > len(hits) {
		topK = len(hits)
	}
	return hits[:topK]
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var cur strings.Builder
	var out []string
	flush := func() {
		if cur.Len() >= 2 {
			out = append(out, cur.String())
		}
		cur.Reset()
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func termFreqIDF(tokens []string, vocab []string, idf []float64) []float64 {
	vec := make([]float64, len(vocab))
	tf := make(map[string]float64)
	for _, t := range tokens {
		tf[t]++
	}
	for i, term := range vocab {
		if c, ok := tf[term]; ok {
			vec[i] = c * idf[i]
		}
	}
	return vec
}

func l2Normalize(v []float64) {
	var s float64
	for _, x := range v {
		s += x * x
	}
	if s == 0 {
		return
	}
	inv := 1.0 / math.Sqrt(s)
	for i := range v {
		v[i] *= inv
	}
}

func cosine(a, b []float64) float64 {
	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}
