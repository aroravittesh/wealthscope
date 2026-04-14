package handler

import (
	"errors"
	"net/http"
	"testing"

	"wealthscope-ai/internal/compare"
)

func TestHTTPStatusForCompareError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid count", compare.ErrInvalidSymbolCount, http.StatusBadRequest},
		{"empty symbol", compare.ErrEmptySymbol, http.StatusBadRequest},
		{"wrapped bad request", errors.Join(compare.ErrEmptySymbol, errors.New("detail")), http.StatusBadRequest},
		{"quote upstream", errors.New("AAPL: quote: timeout"), http.StatusBadGateway},
		{"overview upstream", errors.New("MSFT: overview: rate limit"), http.StatusBadGateway},
		{"generic internal", errors.New("unexpected failure"), http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HTTPStatusForCompareError(tc.err)
			if got != tc.want {
				t.Fatalf("got %d want %d for %v", got, tc.want, tc.err)
			}
		})
	}
}
