package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/handler"
	"wealthscope-ai/internal/service"
)

type mockService struct {
	reply string
	err   error
}

func (m *mockService) ProcessMessage(sessionID string, message string) (string, error) {
	return m.reply, m.err
}

func setupRouter(svc service.ChatServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/chat", handler.ChatHandlerWithService(svc))
	return router
}

func TestChatHandler_Success(t *testing.T) {
	router := setupRouter(&mockService{reply: "AAPL is a technology stock."})

	body := map[string]string{
		"message":    "Tell me about AAPL",
		"session_id": "test123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["response"] != "AAPL is a technology stock." {
		t.Fatalf("unexpected response: %s", resp["response"])
	}
}

func TestChatHandler_EmptyMessage(t *testing.T) {
	router := setupRouter(&mockService{})

	body := map[string]string{"message": ""}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestChatHandler_BadJSON(t *testing.T) {
	router := setupRouter(&mockService{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat",
		bytes.NewBufferString("not valid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestChatHandler_ServiceError(t *testing.T) {
	router := setupRouter(&mockService{err: fmt.Errorf("openai down")})

	body := map[string]string{
		"message":    "What is a stock?",
		"session_id": "test123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", w.Code)
	}
}

func TestClearChatHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/chat/session/:session_id", handler.ClearChatHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/chat/session/sess-xyz", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["message"] == "" {
		t.Fatalf("expected message in body: %v", resp)
	}
}

func TestChatHandler_DefaultSession(t *testing.T) {
	router := setupRouter(&mockService{reply: "ok"})

	body := map[string]string{"message": "What is a stock?"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["session_id"] != "default" {
		t.Fatalf("expected session_id 'default' got %s", resp["session_id"])
	}
}
