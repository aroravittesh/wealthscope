package rag

// Retrieve runs hybrid semantic + lexical retrieval without entity hints (backward compatible).
func Retrieve(query string, topK int) []FinancialDocument {
	return ToFinancialDocuments(RetrieveWithContext(query, RetrievalContext{}, topK))
}

// ToFinancialDocuments maps chunks to the legacy document type for existing callers.
func ToFinancialDocuments(chunks []KnowledgeChunk) []FinancialDocument {
	out := make([]FinancialDocument, 0, len(chunks))
	for _, ch := range chunks {
		out = append(out, FinancialDocument{
			ID:      ch.ID,
			Topic:   ch.Topic,
			Content: ch.Content,
			Tags:    ch.Tags,
		})
	}
	return out
}
