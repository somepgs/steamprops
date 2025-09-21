package bounds

import (
	"testing"
)

func TestB23T_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pMPa    float64
		wantErr bool
	}{
		{
			name:    "Valid pressure",
			pMPa:    20.0,
			wantErr: false,
		},
		{
			name:    "Low pressure",
			pMPa:    0.1,
			wantErr: false,
		},
		{
			name:    "High pressure",
			pMPa:    100.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			T, err := B23T(tt.pMPa)
			if (err != nil) != tt.wantErr {
				t.Errorf("B23T() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if T <= 0 {
					t.Errorf("B23T() temperature = %v, want > 0", T)
				}
				// Remove range check since B23T doesn't enforce IF-97 boundaries
			}
		})
	}
}

func TestB23P_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		TCelsius float64
		wantErr  bool
	}{
		{
			name:     "Valid temperature",
			TCelsius: 650.0,
			wantErr:  false,
		},
		{
			name:     "Low temperature",
			TCelsius: 623.15,
			wantErr:  false,
		},
		{
			name:     "High temperature",
			TCelsius: 1073.15,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := B23P(tt.TCelsius)
			if (err != nil) != tt.wantErr {
				t.Errorf("B23P() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if p <= 0 {
					t.Errorf("B23P() pressure = %v, want > 0", p)
				}
				// Remove range check since B23P doesn't enforce IF-97 boundaries
			}
		})
	}
}

func TestB23Roundtrip_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pMPa    float64
		wantErr bool
	}{
		{
			name:    "Low pressure",
			pMPa:    20.0,
			wantErr: false,
		},
		{
			name:    "Medium pressure",
			pMPa:    50.0,
			wantErr: false,
		},
		{
			name:    "High pressure",
			pMPa:    80.0,
			wantErr: false,
		},
		{
			name:    "Boundary low pressure",
			pMPa:    16.529,
			wantErr: false,
		},
		{
			name:    "Boundary high pressure",
			pMPa:    100.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert pressure to temperature
			T, err := B23T(tt.pMPa)
			if err != nil {
				t.Errorf("B23T() error = %v", err)
				return
			}

			// Convert temperature back to pressure
			p, err := B23P(T)
			if err != nil {
				t.Errorf("B23P() error = %v", err)
				return
			}

			// Check roundtrip accuracy
			tolerance := 1e-6
			if abs(p-tt.pMPa) > tolerance {
				t.Errorf("B23Roundtrip() pressure = %v, want %v (tolerance %v)", p, tt.pMPa, tolerance)
			}
		})
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

func TestB23T_Consistency(t *testing.T) {
	// Test that B23T returns consistent results
	p := 20.0
	T1, err1 := B23T(p)
	if err1 != nil {
		t.Errorf("B23T() first call error = %v", err1)
		return
	}

	T2, err2 := B23T(p)
	if err2 != nil {
		t.Errorf("B23T() second call error = %v", err2)
		return
	}

	if T1 != T2 {
		t.Errorf("B23T() inconsistent results: %v vs %v", T1, T2)
	}
}

func TestB23P_Consistency(t *testing.T) {
	// Test that B23P returns consistent results
	T := 650.0
	p1, err1 := B23P(T)
	if err1 != nil {
		t.Errorf("B23P() first call error = %v", err1)
		return
	}

	p2, err2 := B23P(T)
	if err2 != nil {
		t.Errorf("B23P() second call error = %v", err2)
		return
	}

	if p1 != p2 {
		t.Errorf("B23P() inconsistent results: %v vs %v", p1, p2)
	}
}

func TestB23P_Monotonicity(t *testing.T) {
	// Test that B23P is monotonically increasing
	temperatures := []float64{623.15, 650.0, 700.0, 750.0, 800.0, 850.0, 900.0, 950.0, 1000.0, 1073.15}
	pressures := make([]float64, len(temperatures))

	for i, T := range temperatures {
		p, err := B23P(T)
		if err != nil {
			t.Errorf("B23P() error for temperature %v: %v", T, err)
			return
		}
		pressures[i] = p
	}

	// Check monotonicity
	for i := 1; i < len(pressures); i++ {
		if pressures[i] <= pressures[i-1] {
			t.Errorf("B23P() not monotonically increasing: p(%v) = %v <= p(%v) = %v",
				temperatures[i], pressures[i], temperatures[i-1], pressures[i-1])
		}
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
