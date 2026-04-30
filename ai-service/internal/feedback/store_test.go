package feedback

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixedClock(ts time.Time) func() time.Time { return func() time.Time { return ts } }

func TestJSONLStore_AppendAssignsIDAndTimestamp(t *testing.T) {
	dir := t.TempDir()
	s := NewJSONLStore(filepath.Join(dir, "fb.jsonl"))
	s.SetClock(fixedClock(time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)))
	s.SetIDFn(func() string { return "fb_test_1" })

	saved, err := s.Append(Feedback{
		SessionID:    "s-1",
		ResponseType: ResponseTypeChat,
		Feedback:     LabelHelpful,
		Query:        "Compare Apple and Microsoft",
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID != "fb_test_1" {
		t.Fatalf("ID: want fb_test_1 got %s", saved.ID)
	}
	if saved.Timestamp.IsZero() {
		t.Fatal("timestamp must be assigned")
	}

	listed, err := s.List(ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != "fb_test_1" {
		t.Fatalf("listed: %+v", listed)
	}
}

func TestJSONLStore_RejectsInvalid(t *testing.T) {
	dir := t.TempDir()
	s := NewJSONLStore(filepath.Join(dir, "fb.jsonl"))

	cases := []struct {
		name string
		in   Feedback
	}{
		{"missing session", Feedback{ResponseType: ResponseTypeChat, Feedback: LabelHelpful}},
		{"bad response_type", Feedback{SessionID: "s", ResponseType: "WAT", Feedback: LabelHelpful}},
		{"bad feedback", Feedback{SessionID: "s", ResponseType: ResponseTypeChat, Feedback: "lol"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := s.Append(c.in); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestJSONLStore_FilterByCategoryAndType(t *testing.T) {
	dir := t.TempDir()
	s := NewJSONLStore(filepath.Join(dir, "fb.jsonl"))
	s.SetClock(fixedClock(time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)))

	mustAppend(t, s, Feedback{SessionID: "a", ResponseType: ResponseTypeChat, Feedback: LabelNotHelpful, Reason: "too generic"})
	s.SetClock(fixedClock(time.Date(2026, 4, 29, 9, 1, 0, 0, time.UTC)))
	mustAppend(t, s, Feedback{SessionID: "b", ResponseType: ResponseTypeRetrieval, Feedback: LabelPoorRetrieval})
	s.SetClock(fixedClock(time.Date(2026, 4, 29, 9, 2, 0, 0, time.UTC)))
	mustAppend(t, s, Feedback{SessionID: "a", ResponseType: ResponseTypeChat, Feedback: LabelHelpful})

	gotChat, _ := s.List(ListFilter{ResponseType: ResponseTypeChat})
	if len(gotChat) != 2 {
		t.Fatalf("want 2 chat got %d", len(gotChat))
	}
	if gotChat[0].Feedback != LabelHelpful {
		// newest first
		t.Fatalf("expected newest first, got %s", gotChat[0].Feedback)
	}

	gotPoor, _ := s.List(ListFilter{Feedback: LabelPoorRetrieval})
	if len(gotPoor) != 1 || gotPoor[0].SessionID != "b" {
		t.Fatalf("filter by label wrong: %+v", gotPoor)
	}

	gotSession, _ := s.List(ListFilter{SessionID: "a"})
	if len(gotSession) != 2 {
		t.Fatalf("session filter wrong: %+v", gotSession)
	}

	limited, _ := s.List(ListFilter{Limit: 1})
	if len(limited) != 1 {
		t.Fatalf("limit ignored: %d", len(limited))
	}
}

func TestJSONLStore_ExportStreamsRawFile(t *testing.T) {
	dir := t.TempDir()
	s := NewJSONLStore(filepath.Join(dir, "fb.jsonl"))
	mustAppend(t, s, Feedback{SessionID: "s1", ResponseType: ResponseTypeChat, Feedback: LabelHelpful})
	mustAppend(t, s, Feedback{SessionID: "s1", ResponseType: ResponseTypeChat, Feedback: LabelNotHelpful})

	var buf bytes.Buffer
	if err := s.Export(&buf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines got %d", len(lines))
	}
	if !strings.Contains(lines[0], `"feedback":"helpful"`) {
		t.Fatalf("line 0 unexpected: %s", lines[0])
	}
}

func TestJSONLStore_SanitizesAndTruncates(t *testing.T) {
	dir := t.TempDir()
	s := NewJSONLStore(filepath.Join(dir, "fb.jsonl"))
	long := strings.Repeat("x", maxReasonLen+200)
	saved, err := s.Append(Feedback{
		SessionID:        "  S-1  ",
		ResponseType:     "  CHAT  ",
		Feedback:         "  HELPFUL  ",
		Reason:           long,
		IntentConfidence: 1.7,
		PredictedTicker:  "  aapl  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.SessionID != "S-1" {
		t.Fatalf("session id not trimmed: %q", saved.SessionID)
	}
	if saved.ResponseType != ResponseTypeChat {
		t.Fatalf("response_type not normalised: %q", saved.ResponseType)
	}
	if saved.Feedback != LabelHelpful {
		t.Fatalf("feedback label not normalised: %q", saved.Feedback)
	}
	if len(saved.Reason) != maxReasonLen {
		t.Fatalf("reason not truncated: len=%d", len(saved.Reason))
	}
	if saved.IntentConfidence != 1 {
		t.Fatalf("confidence not clamped: %f", saved.IntentConfidence)
	}
	if saved.PredictedTicker != "AAPL" {
		t.Fatalf("ticker not uppercased: %q", saved.PredictedTicker)
	}
}

func TestMemoryStore_RoundtripAndCount(t *testing.T) {
	s := NewMemoryStore()
	mustAppend(t, s, Feedback{SessionID: "x", ResponseType: ResponseTypeRisk, Feedback: LabelIncorrect})
	mustAppend(t, s, Feedback{SessionID: "y", ResponseType: ResponseTypeRisk, Feedback: LabelHelpful})

	n, _ := s.Count(ListFilter{ResponseType: ResponseTypeRisk})
	if n != 2 {
		t.Fatalf("count: want 2 got %d", n)
	}
	n, _ = s.Count(ListFilter{Feedback: LabelIncorrect})
	if n != 1 {
		t.Fatalf("count incorrect: want 1 got %d", n)
	}
}

func mustAppend(t *testing.T, s Store, f Feedback) Feedback {
	t.Helper()
	out, err := s.Append(f)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	return out
}
