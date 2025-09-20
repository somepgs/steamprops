package transport

import (
	"math"
	"testing"
)

func almostEqual(a, b, rel float64) bool {
	den := math.Max(1.0, math.Max(math.Abs(a), math.Max(math.Abs(b), 1.0)))
	return math.Abs(a-b) <= rel*den
}

func TestDynamicViscosity(t *testing.T) {
	tests := []struct {
		name      string
		T         float64
		rho       float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Water at 20°C",
			T:         293.15,
			rho:       998.2,
			expected:  1.002e-3,
			tolerance: 0.1,
		},
		{
			name:      "Water at 100°C",
			T:         373.15,
			rho:       958.4,
			expected:  0.282e-3,
			tolerance: 0.2,
		},
		{
			name:      "Steam at 200°C, 1 bar",
			T:         473.15,
			rho:       0.46,
			expected:  15.8e-6,
			tolerance: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, err := DynamicViscosity(tt.T, tt.rho)
			if err != nil {
				t.Fatalf("DynamicViscosity error: %v", err)
			}

			if !almostEqual(mu, tt.expected, tt.tolerance) {
				t.Errorf("DynamicViscosity() = %v, expected %v (tolerance %v)", mu, tt.expected, tt.tolerance)
			}

			// Sanity checks
			if mu <= 0 {
				t.Errorf("DynamicViscosity() = %v, expected positive value", mu)
			}
			if math.IsNaN(mu) || math.IsInf(mu, 0) {
				t.Errorf("DynamicViscosity() = %v, expected finite value", mu)
			}
		})
	}
}

func TestThermalConductivity(t *testing.T) {
	tests := []struct {
		name      string
		T         float64
		rho       float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Water at 20°C",
			T:         293.15,
			rho:       998.2,
			expected:  0.598,
			tolerance: 0.2,
		},
		{
			name:      "Water at 100°C",
			T:         373.15,
			rho:       958.4,
			expected:  0.682,
			tolerance: 0.2,
		},
		{
			name:      "Steam at 200°C, 1 bar",
			T:         473.15,
			rho:       0.46,
			expected:  0.033,
			tolerance: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lambda, err := ThermalConductivity(tt.T, tt.rho)
			if err != nil {
				t.Fatalf("ThermalConductivity error: %v", err)
			}

			if !almostEqual(lambda, tt.expected, tt.tolerance) {
				t.Errorf("ThermalConductivity() = %v, expected %v (tolerance %v)", lambda, tt.expected, tt.tolerance)
			}

			// Sanity checks
			if lambda <= 0 {
				t.Errorf("ThermalConductivity() = %v, expected positive value", lambda)
			}
			if math.IsNaN(lambda) || math.IsInf(lambda, 0) {
				t.Errorf("ThermalConductivity() = %v, expected finite value", lambda)
			}
		})
	}
}

func TestKinematicViscosity(t *testing.T) {
	tests := []struct {
		name      string
		T         float64
		rho       float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Water at 20°C",
			T:         293.15,
			rho:       998.2,
			expected:  1.004e-6,
			tolerance: 0.01,
		},
		{
			name:      "Water at 100°C",
			T:         373.15,
			rho:       958.4,
			expected:  0.294e-6,
			tolerance: 0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nu, err := KinematicViscosity(tt.T, tt.rho)
			if err != nil {
				t.Fatalf("KinematicViscosity error: %v", err)
			}

			if !almostEqual(nu, tt.expected, tt.tolerance) {
				t.Errorf("KinematicViscosity() = %v, expected %v (tolerance %v)", nu, tt.expected, tt.tolerance)
			}

			// Sanity checks
			if nu <= 0 {
				t.Errorf("KinematicViscosity() = %v, expected positive value", nu)
			}
			if math.IsNaN(nu) || math.IsInf(nu, 0) {
				t.Errorf("KinematicViscosity() = %v, expected finite value", nu)
			}
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		T    float64
		rho  float64
	}{
		{"Negative temperature", -100, 1000},
		{"Zero temperature", 0, 1000},
		{"Negative density", 300, -100},
		{"Zero density", 300, 0},
		{"Both negative", -100, -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, err := DynamicViscosity(tt.T, tt.rho)
			if err == nil {
				t.Errorf("DynamicViscosity() = %v, expected error", mu)
			}

			lambda, err := ThermalConductivity(tt.T, tt.rho)
			if err == nil {
				t.Errorf("ThermalConductivity() = %v, expected error", lambda)
			}

			nu, err := KinematicViscosity(tt.T, tt.rho)
			if err == nil {
				t.Errorf("KinematicViscosity() = %v, expected error", nu)
			}
		})
	}
}

func TestCriticalPoint(t *testing.T) {
	// Test near critical point (T=647.096 K, ρ=322 kg/m³)
	T := 647.096
	rho := 322.0

	mu, err := DynamicViscosity(T, rho)
	if err != nil {
		t.Fatalf("DynamicViscosity at critical point error: %v", err)
	}

	lambda, err := ThermalConductivity(T, rho)
	if err != nil {
		t.Fatalf("ThermalConductivity at critical point error: %v", err)
	}

	// Critical point values should be finite and positive
	if mu <= 0 || math.IsNaN(mu) || math.IsInf(mu, 0) {
		t.Errorf("DynamicViscosity at critical point = %v, expected finite positive value", mu)
	}

	if lambda <= 0 || math.IsNaN(lambda) || math.IsInf(lambda, 0) {
		t.Errorf("ThermalConductivity at critical point = %v, expected finite positive value", lambda)
	}
}
