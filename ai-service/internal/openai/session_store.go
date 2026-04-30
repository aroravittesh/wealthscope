package openai

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// StoreConfig controls TTL and history bounds. Safe defaults for in-memory production use.
type StoreConfig struct {
	TTL time.Duration
	// MaxMessages triggers compaction: when len(Messages) exceeds this after an assistant reply, older turns fold into one summary line.
	MaxMessages int
	// KeepAfterCompact is how many recent user/assistant messages to retain after folding (summary counts as one message).
	KeepAfterCompact int
}

// DefaultStoreConfig returns 24h TTL, compact when >24 stored messages, keep 10 recent after fold.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		TTL:              24 * time.Hour,
		MaxMessages:      24,
		KeepAfterCompact: 10,
	}
}

// Session is one chat session's rolling history (system prompt is not stored here).
type Session struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []Message
}

// Store is an in-memory session map; replace with Redis/DB by swapping this type later.
type Store struct {
	mu   sync.Mutex
	cfg  StoreConfig
	byID map[string]*Session
	now  func() time.Time
}

// NewStore creates a session store.
func NewStore(cfg StoreConfig) *Store {
	return &Store{
		cfg:  cfg,
		byID: make(map[string]*Session),
		now:  time.Now,
	}
}

// SetClock overrides time source (tests).
func (s *Store) SetClock(fn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn == nil {
		s.now = time.Now
	} else {
		s.now = fn
	}
}

var defaultStore = NewStore(DefaultStoreConfig())

// SetDefaultStore replaces the process-wide session store (tests or custom wiring).
func SetDefaultStore(st *Store) {
	if st == nil {
		st = NewStore(DefaultStoreConfig())
	}
	defaultStore = st
}

func (s *Store) purgeExpired() {
	ttl := s.cfg.TTL
	if ttl <= 0 {
		return
	}
	now := s.now()
	for id, se := range s.byID {
		if now.Sub(se.UpdatedAt) > ttl {
			delete(s.byID, id)
		}
	}
}

func (s *Store) getOrCreateSession(sessionID string) *Session {
	now := s.now()
	ttl := s.cfg.TTL

	if se, ok := s.byID[sessionID]; ok {
		if ttl > 0 && now.Sub(se.UpdatedAt) > ttl {
			delete(s.byID, sessionID)
		} else {
			return se
		}
	}

	se := &Session{CreatedAt: now, UpdatedAt: now, Messages: nil}
	s.byID[sessionID] = se
	return se
}

// AddUserMessage records the user turn and returns a copy of history for the API payload.
func (s *Store) AddUserMessage(sessionID, text string) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpired()
	se := s.getOrCreateSession(sessionID)
	se.Messages = append(se.Messages, Message{Role: "user", Content: text})
	se.UpdatedAt = s.now()
	out := make([]Message, len(se.Messages))
	copy(out, se.Messages)
	return out
}

// AddAssistantMessage records the model reply and compacts if over MaxMessages.
func (s *Store) AddAssistantMessage(sessionID, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	se, ok := s.byID[sessionID]
	if !ok {
		se = s.getOrCreateSession(sessionID)
	}
	se.Messages = append(se.Messages, Message{Role: "assistant", Content: text})
	se.UpdatedAt = s.now()
	s.maybeCompact(se)
}

// Clear removes a session entirely.
func (s *Store) Clear(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byID, sessionID)
}

// Messages returns a copy of the current session messages (oldest to newest).
// If the session does not exist, it returns nil.
func (s *Store) Messages(sessionID string) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpired()
	se, ok := s.byID[sessionID]
	if !ok {
		return nil
	}
	out := make([]Message, len(se.Messages))
	copy(out, se.Messages)
	return out
}

func (s *Store) maybeCompact(se *Session) {
	max := s.cfg.MaxMessages
	keep := s.cfg.KeepAfterCompact
	if max <= 0 || keep <= 0 || len(se.Messages) <= max {
		return
	}
	se.Messages = foldOldestIntoSummary(se.Messages, keep)
}

// foldOldestIntoSummary replaces the oldest (len-keep) messages with one synthetic user note.
func foldOldestIntoSummary(msgs []Message, keepRecent int) []Message {
	if len(msgs) <= keepRecent {
		return msgs
	}
	drop := len(msgs) - keepRecent
	prefix := msgs[:drop]
	suffix := msgs[drop:]
	var b strings.Builder
	for i, m := range prefix {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(m.Role)
		b.WriteString(": ")
		c := m.Content
		if len(c) > 200 {
			c = c[:200] + "…"
		}
		b.WriteString(c)
	}
	summary := Message{
		Role: "user",
		Content: "[Earlier session context — condensed for length; not a new user request]\n" +
			"Turns summarized: " + strconv.Itoa(len(prefix)) + "\n" + b.String(),
	}
	return append([]Message{summary}, suffix...)
}
