package rag

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
)

// QADatasetPathOverride, when non-empty, is used instead of env and default search (tests only).
var QADatasetPathOverride string

var overrideMu sync.Mutex

// SetQADatasetPathForTest sets the CSV path override and resets the lazy loader (serialise tests that mutate globals).
func SetQADatasetPathForTest(path string) {
	overrideMu.Lock()
	QADatasetPathOverride = path
	overrideMu.Unlock()
	ResetQALoaderForTest()
}

// ClearQADatasetPathOverride clears the test override and resets the loader.
func ClearQADatasetPathOverride() {
	overrideMu.Lock()
	QADatasetPathOverride = ""
	overrideMu.Unlock()
	ResetQALoaderForTest()
}

var (
	qaMu       sync.RWMutex
	qaLoaded   bool
	qaChunks   []KnowledgeChunk
	qaIndex    *tfidfIndex
	qaInitErr  error
	qaPathUsed string
)

var qaHeader = []string{
	"id", "category", "sub_category", "question", "answer", "keywords",
	"ticker", "difficulty", "source_type", "priority", "last_updated",
}

// ResetQALoaderForTest clears the lazy QA corpus so the next retrieval reloads from disk.
func ResetQALoaderForTest() {
	qaMu.Lock()
	defer qaMu.Unlock()
	qaLoaded = false
	qaChunks = nil
	qaIndex = nil
	qaInitErr = nil
	qaPathUsed = ""
}

func resolveQADatasetPath() string {
	overrideMu.Lock()
	o := QADatasetPathOverride
	overrideMu.Unlock()
	if o != "" {
		return o
	}
	if p := strings.TrimSpace(os.Getenv("WEALTHSCOPE_QA_DATASET_PATH")); p != "" {
		return p
	}
	candidates := []string{
		"../data/qa_dataset.csv",
		"../../data/qa_dataset.csv",
		"data/qa_dataset.csv",
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return "../data/qa_dataset.csv"
}

// LoadQADatasetFromPath reads qa_dataset.csv and returns retrieval chunks (no global cache).
func LoadQADatasetFromPath(path string) ([]KnowledgeChunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("qa csv header: %w", err)
	}
	if len(header) != len(qaHeader) {
		return nil, fmt.Errorf("qa csv: expected %d columns, got %d", len(qaHeader), len(header))
	}
	for i := range qaHeader {
		if strings.TrimSpace(header[i]) != qaHeader[i] {
			return nil, fmt.Errorf("qa csv: bad header col %d: want %q got %q", i, qaHeader[i], header[i])
		}
	}

	var out []KnowledgeChunk
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("qa csv row: %w", err)
		}
		if len(rec) != len(qaHeader) {
			return nil, fmt.Errorf("qa csv: row has %d fields, want %d", len(rec), len(qaHeader))
		}
		ch, err := qaRowToChunk(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("qa csv: no data rows")
	}
	return out, nil
}

func qaRowToChunk(rec []string) (KnowledgeChunk, error) {
	id := strings.TrimSpace(rec[0])
	category := strings.TrimSpace(rec[1])
	sub := strings.TrimSpace(rec[2])
	question := strings.TrimSpace(rec[3])
	answer := strings.TrimSpace(rec[4])
	keywords := strings.TrimSpace(rec[5])
	ticker := strings.TrimSpace(rec[6])
	diff := strings.TrimSpace(rec[7])
	src := strings.TrimSpace(rec[8])
	pri := strings.TrimSpace(rec[9])
	if id == "" || question == "" || answer == "" {
		return KnowledgeChunk{}, fmt.Errorf("qa csv: empty id/question/answer in row %q", id)
	}

	topic := fmt.Sprintf("%s / %s", category, sub)
	content := fmt.Sprintf(
		"Question: %s\nAnswer: %s\nCategory: %s\nSub-category: %s\nKeywords: %s",
		question, answer, category, sub, keywords,
	)

	tags := []string{
		"qa_dataset",
		id,
		category,
		strings.ReplaceAll(sub, " ", "_"),
		diff,
		src,
		pri,
	}
	for _, p := range strings.Split(keywords, ",") {
		p = strings.TrimSpace(strings.ReplaceAll(p, "_", " "))
		if p != "" {
			tags = append(tags, p)
		}
	}
	if ticker != "" {
		tags = append(tags, strings.ToUpper(ticker))
	}

	return KnowledgeChunk{
		ID:      id,
		Topic:   topic,
		Content: content,
		Tags:    tags,
	}, nil
}

func ensureQACorpus() {
	path := resolveQADatasetPath()
	qaMu.Lock()
	defer qaMu.Unlock()
	if qaLoaded {
		return
	}
	qaLoaded = true
	qaPathUsed = path
	chunks, err := LoadQADatasetFromPath(path)
	if err != nil {
		qaInitErr = err
		qaChunks = nil
		qaIndex = nil
		return
	}
	qaChunks = chunks
	qaIndex = newTFIDFIndex(chunks)
}

// QADatasetPathUsed returns the path last attempted by the lazy loader (empty if not yet loaded).
func QADatasetPathUsed() string {
	qaMu.RLock()
	defer qaMu.RUnlock()
	return qaPathUsed
}

// QADatasetLoadError returns the load error from the last lazy init (nil if OK or not yet loaded).
func QADatasetLoadError() error {
	qaMu.RLock()
	defer qaMu.RUnlock()
	return qaInitErr
}

// RetrieveQAWithContext runs TF-IDF + entity boost on the QA CSV corpus, with lexical fallback (QA-only).
func RetrieveQAWithContext(query string, ctx RetrievalContext, topK int) []KnowledgeChunk {
	if topK <= 0 {
		return nil
	}
	ensureQACorpus()
	qaMu.RLock()
	idx := qaIndex
	chunks := qaChunks
	loadErr := qaInitErr
	qaMu.RUnlock()

	if loadErr != nil || idx == nil || len(chunks) == 0 {
		return nil
	}

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
	return retrieveQALexical(query, topK, chunks)
}

func retrieveQALexical(query string, topK int, chunks []KnowledgeChunk) []KnowledgeChunk {
	query = strings.ToLower(query)
	type hit struct {
		ch    KnowledgeChunk
		score float64
	}
	var scores []hit
	for _, ch := range chunks {
		s := scoreLexical(query, ch)
		if s > 0 {
			scores = append(scores, hit{ch: ch, score: s})
		}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	out := make([]KnowledgeChunk, 0, topK)
	for i, h := range scores {
		if i >= topK {
			break
		}
		out = append(out, h.ch)
	}
	return out
}

// FormatQAKnowledgeLine renders one chunk for the chat envelope (compact, single line per bullet).
func FormatQAKnowledgeLine(ch KnowledgeChunk, question, answer string) string {
	q := strings.TrimSpace(question)
	a := strings.TrimSpace(answer)
	if len(q) > 220 {
		q = q[:217] + "..."
	}
	if len(a) > 400 {
		a = a[:397] + "..."
	}
	return fmt.Sprintf("[%s | %s] Q: %s | A: %s", ch.ID, ch.Topic, q, a)
}

// ChunkQAPair extracts question/answer strings from chunk Content (loader format).
func ChunkQAPair(ch KnowledgeChunk) (question, answer string) {
	lines := strings.Split(ch.Content, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(ln, "Question: "):
			question = strings.TrimPrefix(ln, "Question: ")
		case strings.HasPrefix(ln, "Answer: "):
			answer = strings.TrimPrefix(ln, "Answer: ")
		}
	}
	return question, answer
}
