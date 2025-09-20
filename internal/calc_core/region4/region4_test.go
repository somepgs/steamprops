package region4

import (
	"math"
	"testing"
)

func almostEqual(a, b, rel float64) bool {
	den := math.Max(1.0, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= rel*den
}

func TestSaturationRoundtrip(t *testing.T) {
	cases := []float64{273.16, 300.0, 373.15, 450.0, 600.0}
	for _, T := range cases {
		p, err := SaturationPressure(T)
		if err != nil {
			t.Fatalf("psat(%g K) error: %v", T, err)
		}
		T2, err := SaturationTemperature(p)
		if err != nil {
			t.Fatalf("Tsat(%g Pa) error: %v", p, err)
		}
		if !almostEqual(T, T2, 5e-6) { // ~5 ppm relative
			t.Fatalf("roundtrip T->p->T failed: T=%g, T2=%g, p=%g", T, T2, p)
		}
	}
}
