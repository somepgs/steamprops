package transport

import (
	"math"
	"testing"
)

func TestTransportPropertiesAccuracy(t *testing.T) {
	tests := []struct {
		name           string
		Tkelvin        float64
		rho            float64
		expectedMu     float64
		expectedLambda float64
		tolerance      float64
	}{
		{
			name:           "Water at 20°C, 1 bar (AquaDat reference)",
			Tkelvin:        293.15,
			rho:            998.2,
			expectedMu:     1.002e-3, // 1.002 mPa·s
			expectedLambda: 0.603,    // W/(m·K)
			tolerance:      0.01,     // 1% tolerance
		},
		{
			name:           "Water at 100°C, 1 bar",
			Tkelvin:        373.15,
			rho:            958.4,
			expectedMu:     0.282e-3, // mPa·s
			expectedLambda: 0.68,     // W/(m·K)
			tolerance:      0.05,     // 5% tolerance
		},
		{
			name:           "Steam at 200°C, 1 bar",
			Tkelvin:        473.15,
			rho:            0.46,
			expectedMu:     15.8e-6, // Pa·s
			expectedLambda: 0.033,   // W/(m·K)
			tolerance:      0.1,     // 10% tolerance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test dynamic viscosity
			mu, err := DynamicViscosity(tt.Tkelvin, tt.rho)
			if err != nil {
				t.Fatalf("DynamicViscosity error: %v", err)
			}

			if math.Abs(mu-tt.expectedMu) > tt.tolerance*tt.expectedMu {
				t.Errorf("DynamicViscosity() = %v, expected %v (tolerance %v)",
					mu, tt.expectedMu, tt.tolerance)
			}

			// Test thermal conductivity
			lambda, err := ThermalConductivity(tt.Tkelvin, tt.rho)
			if err != nil {
				t.Fatalf("ThermalConductivity error: %v", err)
			}

			if math.Abs(lambda-tt.expectedLambda) > tt.tolerance*tt.expectedLambda {
				t.Errorf("ThermalConductivity() = %v, expected %v (tolerance %v)",
					lambda, tt.expectedLambda, tt.tolerance)
			}

			// Test kinematic viscosity
			nu, err := KinematicViscosity(tt.Tkelvin, tt.rho)
			if err != nil {
				t.Fatalf("KinematicViscosity error: %v", err)
			}

			expectedNu := tt.expectedMu / tt.rho
			if math.Abs(nu-expectedNu) > tt.tolerance*expectedNu {
				t.Errorf("KinematicViscosity() = %v, expected %v (tolerance %v)",
					nu, expectedNu, tt.tolerance)
			}

			// Sanity checks
			if mu <= 0 || math.IsNaN(mu) || math.IsInf(mu, 0) {
				t.Errorf("DynamicViscosity() = %v, expected positive finite value", mu)
			}
			if lambda <= 0 || math.IsNaN(lambda) || math.IsInf(lambda, 0) {
				t.Errorf("ThermalConductivity() = %v, expected positive finite value", lambda)
			}
			if nu <= 0 || math.IsNaN(nu) || math.IsInf(nu, 0) {
				t.Errorf("KinematicViscosity() = %v, expected positive finite value", nu)
			}
		})
	}
}

func TestTransportPropertiesConsistency(t *testing.T) {
	// Test that transport properties are consistent across different regions
	testCases := []struct {
		name    string
		Tkelvin float64
		rho     float64
	}{
		{"Region 1 liquid", 293.15, 998.2},
		{"Region 2 steam", 473.15, 0.46},
		{"Region 3 critical", 647.1, 322.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mu, err := DynamicViscosity(tc.Tkelvin, tc.rho)
			if err != nil {
				t.Fatalf("DynamicViscosity error: %v", err)
			}

			lambda, err := ThermalConductivity(tc.Tkelvin, tc.rho)
			if err != nil {
				t.Fatalf("ThermalConductivity error: %v", err)
			}

			nu, err := KinematicViscosity(tc.Tkelvin, tc.rho)
			if err != nil {
				t.Fatalf("KinematicViscosity error: %v", err)
			}

			// Check that kinematic viscosity = dynamic viscosity / density
			expectedNu := mu / tc.rho
			if math.Abs(nu-expectedNu) > 1e-10 {
				t.Errorf("KinematicViscosity inconsistency: got %v, expected %v", nu, expectedNu)
			}

			// Check physical bounds
			if mu <= 0 || mu > 1.0 { // Pa·s
				t.Errorf("Dynamic viscosity out of reasonable bounds: %v Pa·s", mu)
			}
			if lambda <= 0 || lambda > 10.0 { // W/(m·K)
				t.Errorf("Thermal conductivity out of reasonable bounds: %v W/(m·K)", lambda)
			}
			if nu <= 0 || nu > 1e-3 { // m²/s
				t.Errorf("Kinematic viscosity out of reasonable bounds: %v m²/s", nu)
			}
		})
	}
}
