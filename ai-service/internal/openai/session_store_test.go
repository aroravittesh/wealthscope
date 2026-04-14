package openai

import (
	"strings"
	"testing"
	"time"
)

func TestStore_SessionCreation(t *testing.T) {
	st := NewStore(DefaultStoreConfig())
	h := st.AddUserMessage("sess-new", "hello")
	if len(h) != 1 || h[0].Role != "user" || h[0].Content != "hello" {
		t.Fatalf("got %+v", h)
	}
}

func TestStore_SessionReuse(t *testing.T) {
	st := NewStore(DefaultStoreConfig())
	st.AddUserMessage("s1", "first")
	st.AddAssistantMessage("s1", "reply one")
	h := st.AddUserMessage("s1", "second")
	if len(h) != 3 {
		t.Fatalf("want 3 messages, got %d", len(h))
	}
	if h[2].Content != "second" {
		t.Fatalf("last user message wrong")
	}
}

func TestStore_ClearSession(t *testing.T) {
	st := NewStore(DefaultStoreConfig())
	st.AddUserMessage("to-clear", "x")
	st.Clear("to-clear")
	h := st.AddUserMessage("to-clear", "after")
	if len(h) != 1 {
		t.Fatalf("after clear expected fresh session with 1 msg, got %d", len(h))
	}
}

func TestStore_SessionExpiration(t *testing.T) {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	st := NewStore(StoreConfig{TTL: time.Hour, MaxMessages: 100, KeepAfterCompact: 20})
	st.SetClock(func() time.Time { return start })

	st.AddUserMessage("exp", "old")
	st.AddAssistantMessage("exp", "old reply")

	st.SetClock(func() time.Time { return start.Add(2 * time.Hour) })
	h := st.AddUserMessage("exp", "new question")

	if len(h) != 1 {
		t.Fatalf("expired session should start clean, got len=%d: %+v", len(h), h)
	}
	if h[0].Content != "new question" {
		t.Fatalf("unexpected content %q", h[0].Content)
	}
}

func TestStore_CompactionSummarizesOldest(t *testing.T) {
	st := NewStore(StoreConfig{
		TTL:              time.Hour,
		MaxMessages:      4,
		KeepAfterCompact: 2,
	})
	st.AddUserMessage("c", "u1")
	st.AddAssistantMessage("c", "a1")
	st.AddUserMessage("c", "u2")
	st.AddAssistantMessage("c", "a2")
	// 4 messages; next assistant pushes to 5 -> compact
	st.AddUserMessage("c", "u3")
	st.AddAssistantMessage("c", "a3")

	st.mu.Lock()
	se := st.byID["c"]
	if se == nil {
		t.Fatal("missing session")
	}
	msgs := se.Messages
	st.mu.Unlock()

	if len(msgs) > 4 {
		t.Fatalf("expected bounded history after compact, got len=%d", len(msgs))
	}
	foundSummary := false
	for _, m := range msgs {
		if m.Role == "user" && len(m.Content) > 20 && strings.Contains(m.Content, "Earlier session context") {
			foundSummary = true
			break
		}
	}
	if !foundSummary {
		t.Fatalf("expected compaction summary in messages: %#v", msgs)
	}
	last := msgs[len(msgs)-1]
	if last.Role != "assistant" || last.Content != "a3" {
		t.Fatalf("expected last message assistant a3, got %+v", last)
	}
}

func TestFoldOldestIntoSummary_Shape(t *testing.T) {
	msgs := []Message{
		{Role: "user", Content: "one"},
		{Role: "assistant", Content: "two"},
		{Role: "user", Content: "three"},
		{Role: "assistant", Content: "four"},
	}
	out := foldOldestIntoSummary(msgs, 2)
	if len(out) != 3 {
		t.Fatalf("want 1 summary + 2 kept, got %d", len(out))
	}
	if !strings.Contains(out[0].Content, "Earlier session context") {
		t.Fatalf("first should be summary: %q", out[0].Content)
	}
	if out[1].Content != "three" || out[2].Content != "four" {
		t.Fatalf("suffix wrong: %+v", out)
	}
}

func TestDefaultStore_ClearIntegration(t *testing.T) {
	prev := defaultStore
	t.Cleanup(func() { defaultStore = prev })
	defaultStore = NewStore(DefaultStoreConfig())

	ClearSession("integration-clear")
	defaultStore.AddUserMessage("integration-clear", "hi")
	ClearSession("integration-clear")
	h := defaultStore.AddUserMessage("integration-clear", "again")
	if len(h) != 1 {
		t.Fatalf("want 1 got %d", len(h))
	}
}
