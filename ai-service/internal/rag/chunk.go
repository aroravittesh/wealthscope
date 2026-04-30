package rag

// KnowledgeChunk is one retrievable unit for RAG grounding.
//
// Metadata carries optional structured signals (priority, source_type,
// difficulty, category, ticker) that the hybrid reranker consumes. KB chunks
// can omit it; the reranker treats missing metadata as a neutral baseline.
type KnowledgeChunk struct {
	ID       string
	Topic    string
	Content  string
	Tags     []string
	Metadata map[string]string
}

func chunkFromDoc(doc FinancialDocument) KnowledgeChunk {
	tags := append([]string(nil), doc.Tags...)
	if doc.Topic != "" {
		tags = append(tags, doc.Topic)
	}
	return KnowledgeChunk{
		ID:      doc.ID,
		Topic:   doc.Topic,
		Content: doc.Content,
		Tags:    tags,
	}
}

func allChunks() []KnowledgeChunk {
	out := make([]KnowledgeChunk, 0, len(KnowledgeBase))
	for _, d := range KnowledgeBase {
		out = append(out, chunkFromDoc(d))
	}
	return out
}
