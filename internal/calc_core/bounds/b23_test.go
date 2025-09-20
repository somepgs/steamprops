package bounds

import (
	"math"
	"testing"
)

func almostEqual(a, b, rel float64) bool {
	den := math.Max(1.0, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= rel*den
}

func TestB23Roundtrip(t *testing.T) {
	press := []float64{1.0, 5.0, 10.0, 20.0} // MPa
	for _, p := range press {
		T, err := B23T(p)
		if err != nil {
			t.Fatalf("B23T(%g MPa) error: %v", p, err)
		}
		p2, err := B23P(T)
		if err != nil {
			t.Fatalf("B23P(%g K) error: %v", T, err)
		}
		if !almostEqual(p, p2, 1e-10) {
			t.Fatalf("roundtrip p->T->p failed: p=%g, p2=%g, T=%g", p, p2, T)
		}
	}
}
