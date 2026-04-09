package analytics

import "testing"

func TestDiversificationScore(t *testing.T) {
	if g := DiversificationScore(nil); g != 0 {
		t.Fatalf("empty: got %f", g)
	}
	if g := DiversificationScore([]float64{100}); g != 0 {
		t.Fatalf("single: got %f", g)
	}
	// two equal
	if g := DiversificationScore([]float64{50, 50}); g < 99 || g > 100 {
		t.Fatalf("equal pair: got %f want ~100", g)
	}
	// concentrated < balanced
	if divConc := DiversificationScore([]float64{95, 5}); divConc >= DiversificationScore([]float64{50, 50}) {
		t.Fatalf("concentrated %f should be below equal-weight %f", divConc, DiversificationScore([]float64{50, 50}))
	}
	// ignores zero-weight legs for n
	if g := DiversificationScore([]float64{100, 0, 0}); g != 0 {
		t.Fatalf("single positive: got %f", g)
	}
}

func TestVolatilityScore(t *testing.T) {
	if g := VolatilityScore([]float64{100}, []string{"stock"}); g < 40 || g > 55 {
		t.Fatalf("stock-only: got %f (approx 22/45*100)", g)
	}
	// half crypto half cash
	v := VolatilityScore(
		[]float64{50, 50},
		[]string{"crypto", "cash"},
	)
	if v < 60 || v > 70 {
		t.Fatalf("blend: got %f", v)
	}
	if VolatilityScore([]float64{}, []string{}) != 0 {
		t.Fatal()
	}
}
