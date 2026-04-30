package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/feedback"
	"wealthscope-ai/internal/handler"
)

type feedbackEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   any             `json:"error"`
}

func decodeFeedbackEnvelopeData(t *testing.T, body []byte, out any) feedbackEnvelope {
	t.Helper()
	var env feedbackEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatal(err)
	}
	if out != nil {
		if err := json.Unmarshal(env.Data, out); err != nil {
			t.Fatal(err)
		}
	}
	return env
}

// errorStore returns an unexpected error from Append, simulating disk failure.
type errorStore struct{ feedback.MemoryStore }

func (e *errorStore) Append(f feedback.Feedback) (feedback.Feedback, error) {
	return feedback.Feedback{}, errors.New("disk full")
}

func newFeedbackRouter(store feedback.Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/feedback", handler.RecordFeedbackHandlerWithStore(store))
	r.GET("/feedback", handler.ListFeedbackHandlerWithStore(store))
	r.GET("/feedback/export", handler.ExportFeedbackHandlerWithStore(store))
	return r
}

func postJSON(t *testing.T, r *gin.Engine, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestFeedbackHandler_RecordValid(t *testing.T) {
	store := feedback.NewMemoryStore()
	r := newFeedbackRouter(store)

	w := postJSON(t, r, "/feedback", map[string]any{
		"session_id":     "sess-1",
		"message_id":     "m-42",
		"query":          "Compare Apple and Microsoft",
		"response_type":  "chat",
		"feedback":       "not_helpful",
		"reason":         "too generic",
		"predicted_intent": "compare_companies",
		"intent_confidence": 0.81,
		"endpoint":       "/chat",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Ok bool   `json:"ok"`
		ID string `json:"id"`
	}
	env := decodeFeedbackEnvelopeData(t, w.Body.Bytes(), &resp)
	if !env.Success {
		t.Fatalf("expected success envelope, got %+v", env)
	}
	if !resp.Ok || resp.ID == "" {
		t.Fatalf("expected ok+id, got %+v", resp)
	}

	got, _ := store.List(feedback.ListFilter{})
	if len(got) != 1 {
		t.Fatalf("expected 1 stored record, got %d", len(got))
	}
	if got[0].PredictedIntent != "COMPARE_COMPANIES" {
		t.Fatalf("intent should be normalised, got %q", got[0].PredictedIntent)
	}
}

func TestFeedbackHandler_RecordValidationFailure(t *testing.T) {
	r := newFeedbackRouter(feedback.NewMemoryStore())

	cases := []struct {
		name    string
		payload map[string]any
	}{
		{"missing session", map[string]any{"response_type": "chat", "feedback": "helpful"}},
		{"bad response_type", map[string]any{"session_id": "s", "response_type": "wat", "feedback": "helpful"}},
		{"bad feedback", map[string]any{"session_id": "s", "response_type": "chat", "feedback": "kinda"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := postJSON(t, r, "/feedback", c.payload)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 got %d", w.Code)
			}
		})
	}
}

func TestFeedbackHandler_RecordBadJSON(t *testing.T) {
	r := newFeedbackRouter(feedback.NewMemoryStore())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/feedback", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestFeedbackHandler_RecordStoreError(t *testing.T) {
	r := newFeedbackRouter(&errorStore{})
	w := postJSON(t, r, "/feedback", map[string]any{
		"session_id":    "s",
		"response_type": "chat",
		"feedback":      "helpful",
	})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", w.Code, w.Body.String())
	}
}

func TestFeedbackHandler_ListFiltersAndLimit(t *testing.T) {
	store := feedback.NewMemoryStore()
	r := newFeedbackRouter(store)

	postJSON(t, r, "/feedback", map[string]any{"session_id": "a", "response_type": "chat", "feedback": "helpful"})
	postJSON(t, r, "/feedback", map[string]any{"session_id": "a", "response_type": "chat", "feedback": "not_helpful"})
	postJSON(t, r, "/feedback", map[string]any{"session_id": "b", "response_type": "retrieval", "feedback": "poor_retrieval"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/feedback?response_type=chat", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	var got struct {
		Count int                 `json:"count"`
		Items []feedback.Feedback `json:"items"`
	}
	env := decodeFeedbackEnvelopeData(t, w.Body.Bytes(), &got)
	if !env.Success {
		t.Fatalf("expected success envelope, got %+v", env)
	}
	if got.Count != 2 {
		t.Fatalf("expected 2 chat records, got %d", got.Count)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/feedback?feedback=poor_retrieval", nil)
	r.ServeHTTP(w, req)
	decodeFeedbackEnvelopeData(t, w.Body.Bytes(), &got)
	if got.Count != 1 || got.Items[0].SessionID != "b" {
		t.Fatalf("filter by label failed: %+v", got)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/feedback?limit=1", nil)
	r.ServeHTTP(w, req)
	decodeFeedbackEnvelopeData(t, w.Body.Bytes(), &got)
	if len(got.Items) != 1 {
		t.Fatalf("limit=1 ignored: %d", len(got.Items))
	}
}

func TestFeedbackHandler_ListBadSinceReturns400(t *testing.T) {
	r := newFeedbackRouter(feedback.NewMemoryStore())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/feedback?since=not-a-date", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestFeedbackHandler_ExportNDJSON(t *testing.T) {
	store := feedback.NewMemoryStore()
	r := newFeedbackRouter(store)
	postJSON(t, r, "/feedback", map[string]any{"session_id": "x", "response_type": "chat", "feedback": "helpful"})
	postJSON(t, r, "/feedback", map[string]any{"session_id": "y", "response_type": "chat", "feedback": "incorrect"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/feedback/export", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Fatalf("content-type: want application/x-ndjson got %s", ct)
	}
	body, _ := io.ReadAll(w.Body)
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 ndjson lines got %d (body=%q)", len(lines), string(body))
	}
}
