package rag

import "strings"

type RetrievalResult struct {
    Document  FinancialDocument
    Relevance float64
}

func Retrieve(query string, topK int) []FinancialDocument {
    query = strings.ToLower(query)
    scores := make([]RetrievalResult, 0)

    for _, doc := range KnowledgeBase {
        score := scoreDocument(query, doc)
        if score > 0 {
            scores = append(scores, RetrievalResult{
                Document:  doc,
                Relevance: score,
            })
        }
    }

    for i := 0; i < len(scores); i++ {
        for j := i + 1; j < len(scores); j++ {
            if scores[j].Relevance > scores[i].Relevance {
                scores[i], scores[j] = scores[j], scores[i]
            }
        }
    }

    results := make([]FinancialDocument, 0)
    for i, r := range scores {
        if i >= topK {
            break
        }
        results = append(results, r.Document)
    }
    return results
}

func scoreDocument(query string, doc FinancialDocument) float64 {
    content := strings.ToLower(doc.Content + " " + doc.Topic)
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
