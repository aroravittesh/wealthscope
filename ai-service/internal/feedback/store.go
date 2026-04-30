package feedback

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Store is the abstraction the handler talks to. JSONL is the v1 implementation;
// swap for Postgres / Kafka / S3 by writing another Store later.
type Store interface {
	// Append validates, sanitises, fills ID + Timestamp, and persists.
	Append(f Feedback) (Feedback, error)
	// List returns matching records newest-first.
	List(filter ListFilter) ([]Feedback, error)
	// Count returns the total number of records matching filter.
	Count(filter ListFilter) (int, error)
	// Export streams the raw underlying log to w. For NDJSON downloads.
	Export(w io.Writer) error
}

// JSONLStore is an append-only file-backed store with a single mutex.
// Suitable for single-instance dev/demo deployments. For multi-instance
// production, swap for a database-backed Store.
type JSONLStore struct {
	mu   sync.Mutex
	path string
	now  func() time.Time
	idFn func() string
}

// NewJSONLStore constructs a JSONL store at path. The parent directory is
// created on first append if missing.
func NewJSONLStore(path string) *JSONLStore {
	return &JSONLStore{
		path: path,
		now:  time.Now,
		idFn: defaultID,
	}
}

// SetClock is a test hook that overrides the timestamp source.
func (s *JSONLStore) SetClock(fn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn == nil {
		s.now = time.Now
	} else {
		s.now = fn
	}
}

// SetIDFn is a test hook that overrides the ID generator.
func (s *JSONLStore) SetIDFn(fn func() string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn == nil {
		s.idFn = defaultID
	} else {
		s.idFn = fn
	}
}

// Path returns the file path being used (helpful for diagnostics).
func (s *JSONLStore) Path() string { return s.path }

// Append validates, sanitises, fills ID + Timestamp, and writes one JSON line.
func (s *JSONLStore) Append(f Feedback) (Feedback, error) {
	f.Sanitize()
	if err := f.Validate(); err != nil {
		return Feedback{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if f.ID == "" {
		f.ID = s.idFn()
	}
	if f.Timestamp.IsZero() {
		f.Timestamp = s.now().UTC()
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return Feedback{}, fmt.Errorf("feedback: ensure dir: %w", err)
	}
	fp, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return Feedback{}, fmt.Errorf("feedback: open log: %w", err)
	}
	defer fp.Close()

	enc, err := json.Marshal(f)
	if err != nil {
		return Feedback{}, fmt.Errorf("feedback: marshal: %w", err)
	}
	if _, err := fp.Write(append(enc, '\n')); err != nil {
		return Feedback{}, fmt.Errorf("feedback: write: %w", err)
	}
	return f, nil
}

// List parses the entire log and returns matches newest-first. Suitable for
// the v1 inspection endpoint; for high volume, swap for a real DB query.
func (s *JSONLStore) List(filter ListFilter) ([]Feedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}
	out := applyFilter(all, filter)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Timestamp.After(out[j].Timestamp) })
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, nil
}

// Count returns the number of records matching the filter (no Limit applied).
func (s *JSONLStore) Count(filter ListFilter) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readAll()
	if err != nil {
		return 0, err
	}
	filter.Limit = 0
	return len(applyFilter(all, filter)), nil
}

// Export streams the raw JSONL file to w. Returns nil if the file does not
// exist yet so that consumers can rely on "no records" being represented as
// an empty body.
func (s *JSONLStore) Export(w io.Writer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fp, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer fp.Close()
	_, err = io.Copy(w, fp)
	return err
}

func (s *JSONLStore) readAll() ([]Feedback, error) {
	fp, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer fp.Close()

	out := make([]Feedback, 0, 64)
	sc := bufio.NewScanner(fp)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var f Feedback
		if err := json.Unmarshal(line, &f); err != nil {
			// Skip corrupt lines so one bad row doesn't block reads. In a
			// production path we'd surface a metric here.
			continue
		}
		out = append(out, f)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func applyFilter(all []Feedback, filter ListFilter) []Feedback {
	out := make([]Feedback, 0, len(all))
	for _, f := range all {
		if filter.Feedback != "" && f.Feedback != filter.Feedback {
			continue
		}
		if filter.ResponseType != "" && f.ResponseType != filter.ResponseType {
			continue
		}
		if filter.SessionID != "" && f.SessionID != filter.SessionID {
			continue
		}
		if !filter.Since.IsZero() && f.Timestamp.Before(filter.Since) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// MemoryStore is a lock-protected in-memory implementation of Store, used by
// tests and any caller that wants ephemeral feedback (e.g. e2e fixtures).
type MemoryStore struct {
	mu   sync.Mutex
	data []Feedback
	now  func() time.Time
	idFn func() string
}

// NewMemoryStore builds an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{now: time.Now, idFn: defaultID}
}

// SetClock overrides the timestamp source (tests).
func (s *MemoryStore) SetClock(fn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn == nil {
		s.now = time.Now
	} else {
		s.now = fn
	}
}

// SetIDFn overrides the ID generator (tests).
func (s *MemoryStore) SetIDFn(fn func() string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn == nil {
		s.idFn = defaultID
	} else {
		s.idFn = fn
	}
}

// Append mirrors JSONLStore.Append behaviour.
func (s *MemoryStore) Append(f Feedback) (Feedback, error) {
	f.Sanitize()
	if err := f.Validate(); err != nil {
		return Feedback{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.ID == "" {
		f.ID = s.idFn()
	}
	if f.Timestamp.IsZero() {
		f.Timestamp = s.now().UTC()
	}
	s.data = append(s.data, f)
	return f, nil
}

// List mirrors JSONLStore.List behaviour.
func (s *MemoryStore) List(filter ListFilter) ([]Feedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := applyFilter(s.data, filter)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Timestamp.After(out[j].Timestamp) })
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, nil
}

// Count mirrors JSONLStore.Count behaviour.
func (s *MemoryStore) Count(filter ListFilter) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	filter.Limit = 0
	return len(applyFilter(s.data, filter)), nil
}

// Export writes the in-memory records as JSONL to w.
func (s *MemoryStore) Export(w io.Writer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.data {
		b, err := json.Marshal(f)
		if err != nil {
			return err
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return nil
}

// ---- process-wide default store (env-configurable) ----

const (
	envFeedbackPath  = "WEALTHSCOPE_FEEDBACK_PATH"
	defaultLogPath   = "data/feedback.jsonl"
)

var (
	defaultStoreOnce sync.Once
	defaultStore     Store
	defaultStoreMu   sync.RWMutex
)

// DefaultStore returns the process-wide store, lazily initialising it from
// env (`WEALTHSCOPE_FEEDBACK_PATH`) on first use. Override with SetDefaultStore
// in tests.
func DefaultStore() Store {
	defaultStoreOnce.Do(func() {
		defaultStoreMu.Lock()
		defer defaultStoreMu.Unlock()
		if defaultStore != nil {
			return
		}
		path := strings.TrimSpace(os.Getenv(envFeedbackPath))
		if path == "" {
			path = defaultLogPath
		}
		defaultStore = NewJSONLStore(path)
	})
	defaultStoreMu.RLock()
	defer defaultStoreMu.RUnlock()
	return defaultStore
}

// SetDefaultStore swaps the process-wide store. Used by tests to inject the
// MemoryStore so HTTP-level tests stay hermetic.
func SetDefaultStore(s Store) {
	defaultStoreMu.Lock()
	defer defaultStoreMu.Unlock()
	defaultStore = s
	// Mark as initialised so DefaultStore() does not overwrite this on first
	// call from another goroutine.
	defaultStoreOnce.Do(func() {})
}

// ---- helpers ----

func defaultID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("fb_%d", time.Now().UnixNano())
	}
	return "fb_" + hex.EncodeToString(b[:])
}
