package region5

import (
	"math"
	"testing"
)

func almostEqual(a, b, rel float64) bool {
	den := math.Max(1.0, math.Max(math.Abs(a), math.Max(math.Abs(b), 1.0)))
	return math.Abs(a-b) <= rel*den
}

func TestRegion5Applicability(t *testing.T) {
	tests := []struct {
		name    string
		T       float64
		p       float64
		wantErr bool
	}{
		{
			name:    "Valid Region 5 point",
			T:       1200.0, // 1200°C
			p:       10e6,   // 10 MPa
			wantErr: false,
		},
		{
			name:    "Temperature too low",
			T:       799.0, // 799°C (below 1073.15 K)
			p:       10e6,
			wantErr: true,
		},
		{
			name:    "Temperature too high",
			T:       2100.0, // 2100°C (above 2273.15 K)
			p:       10e6,
			wantErr: true,
		},
		{
			name:    "Pressure too high",
			T:       1200.0,
			p:       60e6, // 60 MPa (above 50 MPa)
			wantErr: true,
		},
		{
			name:    "Negative pressure",
			T:       1200.0,
			p:       -10e6,
			wantErr: true,
		},
		{
			name:    "Zero pressure",
			T:       1200.0,
			p:       0,
			wantErr: true,
		},
		{
			name:    "Temperature below absolute zero",
			T:       -300.0,
			p:       10e6,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Calculate(tt.T, tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegion5Properties(t *testing.T) {
	// Test at high temperature and pressure typical for Region 5
	T := 1200.0 // 1200°C
	p := 10e6   // 10 MPa

	props, err := Calculate(T, p)
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	// Check that all properties are positive and finite
	if props.SpecificVolume <= 0 {
		t.Errorf("SpecificVolume = %v, expected positive", props.SpecificVolume)
	}
	if props.Density <= 0 {
		t.Errorf("Density = %v, expected positive", props.Density)
	}
	if props.SpecificEnthalpy <= 0 {
		t.Errorf("SpecificEnthalpy = %v, expected positive", props.SpecificEnthalpy)
	}
	if props.SpecificEntropy <= 0 {
		t.Errorf("SpecificEntropy = %v, expected positive", props.SpecificEntropy)
	}
	if props.SpecificIsobaricHeatCapacity <= 0 {
		t.Errorf("SpecificIsobaricHeatCapacity = %v, expected positive", props.SpecificIsobaricHeatCapacity)
	}
	if props.SpecificIsochoricHeatCapacity <= 0 {
		t.Errorf("SpecificIsochoricHeatCapacity = %v, expected positive", props.SpecificIsochoricHeatCapacity)
	}
	if props.SpeedOfSound <= 0 {
		t.Errorf("SpeedOfSound = %v, expected positive", props.SpeedOfSound)
	}

	// Check for NaN or Inf values
	if math.IsNaN(props.SpecificVolume) || math.IsInf(props.SpecificVolume, 0) {
		t.Errorf("SpecificVolume = %v, expected finite value", props.SpecificVolume)
	}
	if math.IsNaN(props.Density) || math.IsInf(props.Density, 0) {
		t.Errorf("Density = %v, expected finite value", props.Density)
	}
	if math.IsNaN(props.SpecificEnthalpy) || math.IsInf(props.SpecificEnthalpy, 0) {
		t.Errorf("SpecificEnthalpy = %v, expected finite value", props.SpecificEnthalpy)
	}
	if math.IsNaN(props.SpecificEntropy) || math.IsInf(props.SpecificEntropy, 0) {
		t.Errorf("SpecificEntropy = %v, expected finite value", props.SpecificEntropy)
	}
	if math.IsNaN(props.SpecificIsobaricHeatCapacity) || math.IsInf(props.SpecificIsobaricHeatCapacity, 0) {
		t.Errorf("SpecificIsobaricHeatCapacity = %v, expected finite value", props.SpecificIsobaricHeatCapacity)
	}
	if math.IsNaN(props.SpecificIsochoricHeatCapacity) || math.IsInf(props.SpecificIsochoricHeatCapacity, 0) {
		t.Errorf("SpecificIsochoricHeatCapacity = %v, expected finite value", props.SpecificIsochoricHeatCapacity)
	}
	if math.IsNaN(props.SpeedOfSound) || math.IsInf(props.SpeedOfSound, 0) {
		t.Errorf("SpeedOfSound = %v, expected finite value", props.SpeedOfSound)
	}
}

func TestRegion5Consistency(t *testing.T) {
	// Test consistency between density and specific volume
	T := 1200.0 // 1200°C
	p := 10e6   // 10 MPa

	props, err := Calculate(T, p)
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	// Density should be inverse of specific volume
	expectedDensity := 1.0 / props.SpecificVolume
	if !almostEqual(props.Density, expectedDensity, 1e-10) {
		t.Errorf("Density consistency check failed: Density = %v, 1/SpecificVolume = %v",
			props.Density, expectedDensity)
	}
}

func TestRegion5HeatCapacityRatio(t *testing.T) {
	// Test that cp > cv (always true for real fluids)
	T := 1200.0 // 1200°C
	p := 10e6   // 10 MPa

	props, err := Calculate(T, p)
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	if props.SpecificIsobaricHeatCapacity < props.SpecificIsochoricHeatCapacity {
		t.Errorf("cp < cv: cp = %v, cv = %v (cp should be greater than or equal to cv)",
			props.SpecificIsobaricHeatCapacity, props.SpecificIsochoricHeatCapacity)
	}
}

func TestRegion5SpeedOfSound(t *testing.T) {
	// Test speed of sound at different conditions
	tests := []struct {
		name string
		T    float64
		p    float64
		minW float64 // minimum expected speed of sound
		maxW float64 // maximum expected speed of sound
	}{
		{
			name: "High temperature, low pressure",
			T:    1200.0,
			p:    1e6, // 1 MPa
			minW: 400,
			maxW: 1000,
		},
		{
			name: "High temperature, high pressure",
			T:    1200.0,
			p:    50e6, // 50 MPa
			minW: 600,
			maxW: 1200,
		},
		{
			name: "Very high temperature",
			T:    2000.0,
			p:    10e6,
			minW: 500,
			maxW: 1200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, err := Calculate(tt.T, tt.p)
			if err != nil {
				t.Fatalf("Calculate() error: %v", err)
			}

			if props.SpeedOfSound < tt.minW || props.SpeedOfSound > tt.maxW {
				t.Errorf("SpeedOfSound = %v, expected between %v and %v",
					props.SpeedOfSound, tt.minW, tt.maxW)
			}
		})
	}
}

func TestRegion5BoundaryConditions(t *testing.T) {
	// Test at boundary conditions
	tests := []struct {
		name string
		T    float64
		p    float64
	}{
		{
			name: "Lower temperature boundary",
			T:    800.0, // Just above 1073.15 K
			p:    10e6,
		},
		{
			name: "Upper temperature boundary",
			T:    2000.0, // Just below 2273.15 K
			p:    10e6,
		},
		{
			name: "Upper pressure boundary",
			T:    1200.0,
			p:    50e6, // 50 MPa
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, err := Calculate(tt.T, tt.p)
			if err != nil {
				t.Fatalf("Calculate() error: %v", err)
			}

			// All properties should be positive and finite
			if props.SpecificVolume <= 0 || math.IsNaN(props.SpecificVolume) || math.IsInf(props.SpecificVolume, 0) {
				t.Errorf("SpecificVolume = %v, expected positive finite value", props.SpecificVolume)
			}
			if props.Density <= 0 || math.IsNaN(props.Density) || math.IsInf(props.Density, 0) {
				t.Errorf("Density = %v, expected positive finite value", props.Density)
			}
			if props.SpeedOfSound <= 0 || math.IsNaN(props.SpeedOfSound) || math.IsInf(props.SpeedOfSound, 0) {
				t.Errorf("SpeedOfSound = %v, expected positive finite value", props.SpeedOfSound)
			}
		})
	}
}
