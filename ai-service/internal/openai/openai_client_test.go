package openai

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallOpenAI_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")

	prev := defaultStore
	t.Cleanup(func() { defaultStore = prev })
	defaultStore = NewStore(DefaultStoreConfig())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method %s", r.Method)
		}
		if !strings.Contains(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("missing bearer: %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"Hello from model"}}]}`))
	}))
	t.Cleanup(srv.Close)

	cleanup := SetChatHTTPTestConfig(srv.Client(), srv.URL)
	t.Cleanup(cleanup)

	reply, err := CallOpenAI("sess-openai-success", "user text")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "Hello from model" {
		t.Fatalf("reply: %q", reply)
	}
}

func TestCallOpenAI_NonOKStatus(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")

	prev := defaultStore
	t.Cleanup(func() { defaultStore = prev })
	defaultStore = NewStore(DefaultStoreConfig())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cleanup := SetChatHTTPTestConfig(srv.Client(), srv.URL)
	t.Cleanup(cleanup)

	_, err := CallOpenAI("sess-openai-rate", "x")
	if err == nil || !strings.Contains(err.Error(), "OpenAI API error") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestCallOpenAI_EmptyChoices(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")

	prev := defaultStore
	t.Cleanup(func() { defaultStore = prev })
	defaultStore = NewStore(DefaultStoreConfig())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	t.Cleanup(srv.Close)

	cleanup := SetChatHTTPTestConfig(srv.Client(), srv.URL)
	t.Cleanup(cleanup)

	_, err := CallOpenAI("sess-openai-empty", "x")
	if err == nil || !strings.Contains(err.Error(), "no response from OpenAI") {
		t.Fatalf("expected empty choices error, got %v", err)
	}
}

func TestCallOpenAI_InvalidJSONBody(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")

	prev := defaultStore
	t.Cleanup(func() { defaultStore = prev })
	defaultStore = NewStore(DefaultStoreConfig())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-json`))
	}))
	t.Cleanup(srv.Close)

	cleanup := SetChatHTTPTestConfig(srv.Client(), srv.URL)
	t.Cleanup(cleanup)

	_, err := CallOpenAI("sess-openai-badjson", "x")
	if err == nil {
		t.Fatal("expected decode error")
	}
}
