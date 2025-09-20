package region1

import (
	"testing"
)

func TestRegion1Applicability(t *testing.T) {
	T := 150.0 // C
	// choose p below psat: e.g., 0.05 MPa
	if _, err := Calculate(T, 50_000.0); err == nil {
		t.Fatalf("expected error for p < psat(T) in Region 1")
	}
	// choose p high (e.g., 1 MPa) expected valid within ranges
	if _, err := Calculate(T, 1_000_000.0); err != nil {
		t.Fatalf("unexpected error for valid Region1 point: %v", err)
	}
}
