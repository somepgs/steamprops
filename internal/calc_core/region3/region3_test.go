package region3

import (
	"testing"
)

func TestPressureFromHS_3a(t *testing.T) {
	// s below sc => 3a
	h := 1800.0 // kJ/kg
	s := 3.5    // kJ/(kg*K)
	p, err := PressureFromHS(h, s)
	if err != nil {
		t.Fatalf("PressureFromHS error: %v", err)
	}
	if p <= 0 {
		t.Fatalf("pressure must be positive, got %g", p)
	}
	// 3a typically within tens of MPa; allow broad sanity range
	if p > 1.0e9 { // >1000 MPa is unreasonable for IF-97 range
		t.Fatalf("pressure too large: %g Pa", p)
	}
}

func TestPressureFromHS_3b(t *testing.T) {
	// s above sc => 3b
	h := 2600.0 // kJ/kg
	s := 5.5    // kJ/(kg*K)
	p, err := PressureFromHS(h, s)
	if err != nil {
		t.Fatalf("PressureFromHS error: %v", err)
	}
	if p <= 0 {
		t.Fatalf("pressure must be positive, got %g", p)
	}
	if p > 1.0e9 {
		t.Fatalf("pressure too large: %g Pa", p)
	}
}
