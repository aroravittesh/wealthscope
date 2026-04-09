package market

import "testing"

func TestDeterministicMultiplier_Bounds(t *testing.T) {
	for _, sym := range []string{"", "A", "AAPL", "BTC", "VERYLONGSYMBOL"} {
		m := DeterministicMultiplier(sym)
		if m < 0.92 || m > 1.08 {
			t.Fatalf("%q: mult %f out of range", sym, m)
		}
	}
}

func TestSimulated_UnitPrice(t *testing.T) {
	var s Simulated
	p := s.UnitPrice("AAPL", 100)
	if p < 92 || p > 108 {
		t.Fatalf("expected scaled price in [92,108], got %f", p)
	}
}

func TestPassthrough_UnitPrice(t *testing.T) {
	var p Passthrough
	if got := p.UnitPrice("X", 12.5); got != 12.5 {
		t.Fatalf("got %f", got)
	}
}
