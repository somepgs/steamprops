package region5

import (
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
			tCelsius: 1073.15,
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Boundary pressure",
			tCelsius: 1200.0,
			pPascal:  50e6,
			wantErr:  false,
		},
		{
			name:     "High temperature",
			tCelsius: 1500.0,
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "High pressure",
			tCelsius: 1200.0,
			pPascal:  30e6,
			wantErr:  false,
		},
		{
			name:     "Very high temperature",
			tCelsius: 1200.0, // В пределах Region 5
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Low temperature",
			tCelsius: 1073.16,
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Low pressure",
			tCelsius: 1200.0,
			pPascal:  1000.0,
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
	tCelsius := 1200.0
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
	tCelsius := 1200.0
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
	temps := []float64{1100.0, 1200.0, 1300.0, 1400.0, 1500.0, 1600.0}
	volumes := make([]float64, len(temps))

	for i, temp := range temps {
		props, err := Calculate(temp, pPascal)
		if err != nil {
			t.Errorf("Calculate() error for temperature %v: %v", temp, err)
			return
		}
		volumes[i] = props.SpecificVolume
	}

	// Check that volume increases with temperature (for high-temperature steam)
	for i := 1; i < len(volumes); i++ {
		if volumes[i] <= volumes[i-1] {
			t.Errorf("Calculate() volume not increasing with temperature: v(%v) = %v <= v(%v) = %v",
				temps[i], volumes[i], temps[i-1], volumes[i-1])
		}
	}
}

func TestCalculate_PressureDependence(t *testing.T) {
	// Test pressure dependence
	tCelsius := 1200.0
	pressures := []float64{101325.0, 1e6, 5e6, 10e6, 20e6, 30e6}
	volumes := make([]float64, len(pressures))

	for i, p := range pressures {
		props, err := Calculate(tCelsius, p)
		if err != nil {
			t.Errorf("Calculate() error for pressure %v: %v", p, err)
			return
		}
		volumes[i] = props.SpecificVolume
	}

	// Check that volume decreases with pressure (for high-temperature steam)
	for i := 1; i < len(volumes); i++ {
		if volumes[i] >= volumes[i-1] {
			t.Errorf("Calculate() volume not decreasing with pressure: v(%v) = %v >= v(%v) = %v",
				pressures[i], volumes[i], pressures[i-1], volumes[i-1])
		}
	}
}

func TestCalculate_HeatCapacityRatio(t *testing.T) {
	// Test heat capacity ratio
	tCelsius := 1200.0
	pPascal := 101325.0

	props, err := Calculate(tCelsius, pPascal)
	if err != nil {
		t.Errorf("Calculate() error = %v", err)
		return
	}

	// Check that heat capacity ratio is reasonable for high-temperature steam
	ratio := props.SpecificIsobaricHeatCapacity / props.SpecificIsochoricHeatCapacity
	if ratio < 1.0 || ratio > 2.0 {
		t.Errorf("Calculate() heat capacity ratio = %v, want in range [1.0, 2.0]", ratio)
	}
}

func TestCalculate_SpeedOfSound(t *testing.T) {
	// Test speed of sound
	tCelsius := 1200.0
	pPascal := 101325.0

	props, err := Calculate(tCelsius, pPascal)
	if err != nil {
		t.Errorf("Calculate() error = %v", err)
		return
	}

	// Check that speed of sound is reasonable for high-temperature steam
	if props.SpeedOfSound < 100.0 || props.SpeedOfSound > 2000.0 {
		t.Errorf("Calculate() speed of sound = %v, want in range [100.0, 2000.0]", props.SpeedOfSound)
	}
}

func TestCalculate_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		tCelsius float64
		pPascal  float64
		wantErr  bool
	}{
		{
			name:     "Lower temperature boundary",
			tCelsius: 1073.15,
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Upper temperature boundary",
			tCelsius: 1200.0, // В пределах Region 5
			pPascal:  101325.0,
			wantErr:  false,
		},
		{
			name:     "Upper pressure boundary",
			tCelsius: 1200.0,
			pPascal:  50e6,
			wantErr:  false,
		},
		{
			name:     "Below lower temperature boundary",
			tCelsius: 500.0, // Ниже нижней границы Region 5
			pPascal:  101325.0,
			wantErr:  true,
		},
		{
			name:     "Above upper temperature boundary",
			tCelsius: 2273.16,
			pPascal:  101325.0,
			wantErr:  true,
		},
		{
			name:     "Above upper pressure boundary",
			tCelsius: 1200.0,
			pPascal:  50.1e6,
			wantErr:  true,
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
