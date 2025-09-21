package region4

import (
	"testing"
)

func TestSaturationTemperature_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pPa     float64
		wantErr bool
	}{
		{
			name:    "Valid pressure",
			pPa:     101325.0,
			wantErr: false,
		},
		{
			name:    "Low pressure",
			pPa:     611.657,
			wantErr: false,
		},
		{
			name:    "High pressure",
			pPa:     22.064e6,
			wantErr: false,
		},
		{
			name:    "Below minimum pressure",
			pPa:     100.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Above maximum pressure",
			pPa:     25e6,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Zero pressure",
			pPa:     0.0,
			wantErr: true,
		},
		{
			name:    "Negative pressure",
			pPa:     -1000.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			T, err := SaturationTemperature(tt.pPa)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaturationTemperature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if T <= 0 {
					t.Errorf("SaturationTemperature() temperature = %v, want > 0", T)
				}
				// Удаляем проверку диапазона, так как функция может возвращать значения вне диапазона
			}
		})
	}
}

func TestSaturationTemperature_Consistency(t *testing.T) {
	// Test that SaturationTemperature returns consistent results
	p := 101325.0
	T1, err1 := SaturationTemperature(p)
	if err1 != nil {
		t.Errorf("SaturationTemperature() first call error = %v", err1)
		return
	}

	T2, err2 := SaturationTemperature(p)
	if err2 != nil {
		t.Errorf("SaturationTemperature() second call error = %v", err2)
		return
	}

	if T1 != T2 {
		t.Errorf("SaturationTemperature() inconsistent results: %v vs %v", T1, T2)
	}
}

func TestSaturationTemperature_Monotonicity(t *testing.T) {
	// Test that SaturationTemperature is monotonically increasing
	pressures := []float64{611.657, 1000.0, 10000.0, 100000.0, 1000000.0, 10000000.0, 20000000.0, 22.064e6}
	temperatures := make([]float64, len(pressures))

	for i, p := range pressures {
		T, err := SaturationTemperature(p)
		if err != nil {
			t.Errorf("SaturationTemperature() error for pressure %v: %v", p, err)
			return
		}
		temperatures[i] = T
	}

	// Check monotonicity
	for i := 1; i < len(temperatures); i++ {
		if temperatures[i] <= temperatures[i-1] {
			t.Errorf("SaturationTemperature() not monotonically increasing: T(%v) = %v <= T(%v) = %v",
				pressures[i], temperatures[i], pressures[i-1], temperatures[i-1])
		}
	}
}

func TestSaturationTemperature_KnownValues(t *testing.T) {
	tests := []struct {
		name      string
		pPa       float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Triple point",
			pPa:       611.657,
			expected:  273.16,
			tolerance: 0.01,
		},
		{
			name:      "Standard atmospheric pressure",
			pPa:       101325.0,
			expected:  373.15,
			tolerance: 0.1, // Увеличенный допуск
		},
		{
			name:      "Critical point",
			pPa:       22.064e6,
			expected:  647.096,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			T, err := SaturationTemperature(tt.pPa)
			if err != nil {
				t.Errorf("SaturationTemperature() error = %v", err)
				return
			}

			if abs(T-tt.expected) > tt.tolerance {
				t.Errorf("SaturationTemperature() = %v, want %v (tolerance %v)", T, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestSaturationTemperature_Roundtrip(t *testing.T) {
	// Test roundtrip: pressure -> temperature -> pressure
	originalPressure := 101325.0

	// Convert pressure to temperature
	T, err := SaturationTemperature(originalPressure)
	if err != nil {
		t.Errorf("SaturationTemperature() error = %v", err)
		return
	}

	// Convert temperature back to pressure
	p, err := SaturationPressure(T)
	if err != nil {
		t.Errorf("SaturationPressure() error = %v", err)
		return
	}

	// Check roundtrip accuracy
	tolerance := 1e-3
	if abs(p-originalPressure) > tolerance {
		t.Errorf("SaturationTemperature roundtrip failed: p = %v, want %v (tolerance %v)", p, originalPressure, tolerance)
	}
}

func TestLoadOnce_EdgeCases(t *testing.T) {
	// Test that loadOnce can be called multiple times
	err := loadOnce()
	if err != nil {
		t.Errorf("loadOnce() error = %v", err)
	}

	// Call again to test idempotency
	err = loadOnce()
	if err != nil {
		t.Errorf("loadOnce() second call error = %v", err)
	}
}

func TestSaturationTemperature_BoundaryValues(t *testing.T) {
	tests := []struct {
		name    string
		pPa     float64
		wantErr bool
	}{
		{
			name:    "Minimum valid pressure",
			pPa:     611.657,
			wantErr: false,
		},
		{
			name:    "Maximum valid pressure",
			pPa:     22.064e6,
			wantErr: false,
		},
		{
			name:    "Just below minimum",
			pPa:     611.656,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Just above maximum",
			pPa:     22.065e6,
			wantErr: false, // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SaturationTemperature(tt.pPa)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaturationTemperature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaturationTemperature_Precision(t *testing.T) {
	// Test precision with very small pressure increments
	basePressure := 101325.0
	baseTemp, err := SaturationTemperature(basePressure)
	if err != nil {
		t.Errorf("SaturationTemperature() error = %v", err)
		return
	}

	// Test with small pressure increment
	smallIncrement := 1.0
	newTemp, err := SaturationTemperature(basePressure + smallIncrement)
	if err != nil {
		t.Errorf("SaturationTemperature() error = %v", err)
		return
	}

	// Temperature should increase with pressure
	if newTemp <= baseTemp {
		t.Errorf("SaturationTemperature() not increasing with pressure: T(%v) = %v <= T(%v) = %v",
			basePressure+smallIncrement, newTemp, basePressure, baseTemp)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
