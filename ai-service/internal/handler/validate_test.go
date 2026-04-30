package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeSessionID_DefaultWhenEmpty(t *testing.T) {
	if got := NormalizeSessionID(""); got != "default" {
		t.Fatalf("got %q want default", got)
	}
	if got := NormalizeSessionID("   "); got != "default" {
		t.Fatalf("got %q want default", got)
	}
	if got := NormalizeSessionID("sess-1"); got != "sess-1" {
		t.Fatalf("got %q want sess-1", got)
	}
}

func TestRequiredTrimmed_Responds400OnEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", func(c *gin.Context) {
		_, ok := RequiredTrimmed(c, "   ", "value required")
		if ok {
			t.Fatal("expected !ok")
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRequireNonEmptySlice_Responds400OnEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", func(c *gin.Context) {
		ok := RequireNonEmptySlice(c, []int{}, "list required")
		if ok {
			t.Fatal("expected !ok")
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

