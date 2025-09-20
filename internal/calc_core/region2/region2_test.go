package region2

import "testing"

func TestRegion2Applicability(t *testing.T) {
	T := 100.0 // C
	// above saturation roughly at 1 atm
	if _, err := Calculate(T, 500_000.0); err == nil {
		t.Fatalf("expected error for p > psat(T) in Region 2")
	}
	// below saturation
	if _, err := Calculate(T, 50_000.0); err != nil {
		t.Fatalf("unexpected error for valid Region2 point: %v", err)
	}
}
