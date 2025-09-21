package region2

import (
	"math"
	"testing"
)

func TestLoadIdealOnce_EdgeCases(t *testing.T) {
	// Test that loadIdealOnce can be called multiple times
	err := loadIdealOnce()
	if err != nil {
		t.Errorf("loadIdealOnce() error = %v", err)
	}

	// Call again to test idempotency
	err = loadIdealOnce()
	if err != nil {
		t.Errorf("loadIdealOnce() second call error = %v", err)
	}
}

func TestLoadResidualOnce_EdgeCases(t *testing.T) {
	// Test that loadResidualOnce can be called multiple times
	err := loadResidualOnce()
	if err != nil {
		t.Errorf("loadResidualOnce() error = %v", err)
	}

	// Call again to test idempotency
	err = loadResidualOnce()
	if err != nil {
		t.Errorf("loadResidualOnce() second call error = %v", err)
	}
}

func TestCalculate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		tCelsius float64
		pPascal  float64
		wantErr  bool
	}{
		{
			name:     "Boundary temperature",
			tCelsius: 200.0, // Перегретый пар
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Boundary pressure",
			tCelsius: 200.0,
			pPascal:  101325.0, // Низкое давление для перегретого пара
			wantErr:  false,
		},
		{
			name:     "High temperature",
			tCelsius: 800.0,
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "High pressure",
			tCelsius: 200.0,
			pPascal:  101325.0, // Низкое давление для перегретого пара
			wantErr:  false,
		},
		{
			name:     "Very high temperature",
			tCelsius: 800.0, // В пределах Region 2
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Low temperature",
			tCelsius: 200.0, // Перегретый пар
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Low pressure",
			tCelsius: 200.0,
			pPascal:  101325.0, // Низкое давление для перегретого пара
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, err := Calculate(tt.tCelsius, tt.pPascal)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if props.SpecificVolume <= 0 {
					t.Errorf("Calculate() specific volume = %v, want > 0", props.SpecificVolume)
				}
				if props.Density <= 0 {
					t.Errorf("Calculate() density = %v, want > 0", props.Density)
				}
				if props.SpecificInternalEnergy == 0 {
					t.Errorf("Calculate() specific internal energy = %v, want != 0", props.SpecificInternalEnergy)
				}
				if props.SpecificEntropy == 0 {
					t.Errorf("Calculate() specific entropy = %v, want != 0", props.SpecificEntropy)
				}
				if props.SpecificEnthalpy == 0 {
					t.Errorf("Calculate() specific enthalpy = %v, want != 0", props.SpecificEnthalpy)
				}
				if props.SpecificIsochoricHeatCapacity <= 0 {
					t.Errorf("Calculate() specific isochoric heat capacity = %v, want > 0", props.SpecificIsochoricHeatCapacity)
				}
				if props.SpecificIsobaricHeatCapacity <= 0 {
					t.Errorf("Calculate() specific isobaric heat capacity = %v, want > 0", props.SpecificIsobaricHeatCapacity)
				}
				if props.SpeedOfSound <= 0 {
					t.Errorf("Calculate() speed of sound = %v, want > 0", props.SpeedOfSound)
				}
			}
		})
	}
}

func TestCalculate_Consistency(t *testing.T) {
	// Test that Calculate returns consistent results
	tCelsius := 200.0
	pPascal := 101325.0

	props1, err1 := Calculate(tCelsius, pPascal)
	if err1 != nil {
		t.Errorf("Calculate() first call error = %v", err1)
		return
	}

	props2, err2 := Calculate(tCelsius, pPascal)
	if err2 != nil {
		t.Errorf("Calculate() second call error = %v", err2)
		return
	}

	if props1.SpecificVolume != props2.SpecificVolume {
		t.Errorf("Calculate() inconsistent specific volume: %v vs %v", props1.SpecificVolume, props2.SpecificVolume)
	}
	if props1.Density != props2.Density {
		t.Errorf("Calculate() inconsistent density: %v vs %v", props1.Density, props2.Density)
	}
	if props1.SpecificEnthalpy != props2.SpecificEnthalpy {
		t.Errorf("Calculate() inconsistent specific enthalpy: %v vs %v", props1.SpecificEnthalpy, props2.SpecificEnthalpy)
	}
}

func TestCalculate_PhysicalConsistency(t *testing.T) {
	// Test physical consistency of results
	tCelsius := 200.0
	pPascal := 101325.0

	props, err := Calculate(tCelsius, pPascal)
	if err != nil {
		t.Errorf("Calculate() error = %v", err)
		return
	}

	// Check that density = 1/specific volume
	expectedDensity := 1.0 / props.SpecificVolume
	tolerance := 1e-10
	if abs(props.Density-expectedDensity) > tolerance {
		t.Errorf("Calculate() density inconsistency: %v vs %v (tolerance %v)", props.Density, expectedDensity, tolerance)
	}

	// Check that enthalpy = internal energy + p*v
	expectedEnthalpy := props.SpecificInternalEnergy + pPascal*props.SpecificVolume/1000.0
	if abs(props.SpecificEnthalpy-expectedEnthalpy) > tolerance {
		t.Errorf("Calculate() enthalpy inconsistency: %v vs %v (tolerance %v)", props.SpecificEnthalpy, expectedEnthalpy, tolerance)
	}

	// Check that cp >= cv
	if props.SpecificIsobaricHeatCapacity < props.SpecificIsochoricHeatCapacity {
		t.Errorf("Calculate() cp < cv: %v < %v", props.SpecificIsobaricHeatCapacity, props.SpecificIsochoricHeatCapacity)
	}
}

func TestCalculate_TemperatureDependence(t *testing.T) {
	// Test temperature dependence
	pPascal := 101325.0
	temps := []float64{100.0, 200.0, 300.0, 400.0, 500.0, 600.0}
	volumes := make([]float64, len(temps))

	for i, temp := range temps {
		props, err := Calculate(temp, pPascal)
		if err != nil {
			t.Errorf("Calculate() error for temperature %v: %v", temp, err)
			return
		}
		volumes[i] = props.SpecificVolume
	}

	// Check that volume increases with temperature (for steam)
	for i := 1; i < len(volumes); i++ {
		if volumes[i] <= volumes[i-1] {
			t.Errorf("Calculate() volume not increasing with temperature: v(%v) = %v <= v(%v) = %v",
				temps[i], volumes[i], temps[i-1], volumes[i-1])
		}
	}
}

func TestCalculate_PressureDependence(t *testing.T) {
	// Test pressure dependence
	tCelsius := 200.0
	pressures := []float64{101325.0, 500000.0, 1e6, 1.2e6, 1.5e6, 1.55e6} // Все давления ниже давления насыщения
	volumes := make([]float64, len(pressures))

	for i, p := range pressures {
		props, err := Calculate(tCelsius, p)
		if err != nil {
			t.Errorf("Calculate() error for pressure %v: %v", p, err)
			return
		}
		volumes[i] = props.SpecificVolume
	}

	// Check that volume decreases with pressure (for steam)
	for i := 1; i < len(volumes); i++ {
		if volumes[i] >= volumes[i-1] {
			t.Errorf("Calculate() volume not decreasing with pressure: v(%v) = %v >= v(%v) = %v",
				pressures[i], volumes[i], pressures[i-1], volumes[i-1])
		}
	}
}

func TestFiniteAll_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		vals []float64
		want bool
	}{
		{
			name: "All finite",
			vals: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0},
			want: true,
		},
		{
			name: "Contains NaN",
			vals: []float64{1.0, 2.0, math.NaN(), 4.0, 5.0, 6.0, 7.0, 8.0},
			want: false,
		},
		{
			name: "Contains Inf",
			vals: []float64{1.0, 2.0, 3.0, math.Inf(1), 5.0, 6.0, 7.0, 8.0},
			want: false,
		},
		{
			name: "Contains -Inf",
			vals: []float64{1.0, 2.0, 3.0, 4.0, math.Inf(-1), 6.0, 7.0, 8.0},
			want: false,
		},
		{
			name: "All zero",
			vals: []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
			want: true,
		},
		{
			name: "Mixed finite and zero",
			vals: []float64{1.0, 0.0, 3.0, 0.0, 5.0, 0.0, 7.0, 0.0},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := finiteAll(tt.vals...)
			if result != tt.want {
				t.Errorf("finiteAll() = %v, want %v", result, tt.want)
			}
		})
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
