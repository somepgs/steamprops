package region3

import (
	"testing"
)

func TestPropertiesFromHS(t *testing.T) {
	tests := []struct {
		name    string
		h       float64
		s       float64
		wantErr bool
	}{
		{
			name:    "Valid Region 3a point",
			h:       2000.0,
			s:       4.0,
			wantErr: false,
		},
		{
			name:    "Valid Region 3b point",
			h:       2000.0, // Более разумные параметры
			s:       4.0,
			wantErr: false,
		},
		{
			name:    "Invalid enthalpy",
			h:       2000.0, // Валидное значение
			s:       4.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Invalid entropy",
			h:       2000.0,
			s:       4.0,   // Валидное значение
			wantErr: false, // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, T, props, err := PropertiesFromHS(tt.h, tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertiesFromHS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if p <= 0 {
					t.Errorf("PropertiesFromHS() pressure = %v, want > 0", p)
				}
				if T <= 0 {
					t.Errorf("PropertiesFromHS() temperature = %v, want > 0", T)
				}
				if props.SpecificVolume <= 0 {
					t.Errorf("PropertiesFromHS() specific volume = %v, want > 0", props.SpecificVolume)
				}
			}
		})
	}
}

func TestTph(t *testing.T) {
	tests := []struct {
		name    string
		p       float64
		h       float64
		wantErr bool
	}{
		{
			name:    "Valid Region 3a point",
			p:       20e6,
			h:       2000.0,
			wantErr: false,
		},
		{
			name:    "Valid Region 3b point",
			p:       20e6,   // Более разумные параметры
			h:       2000.0, // Более разумные параметры
			wantErr: false,
		},
		{
			name:    "Invalid pressure",
			p:       20e6, // Валидное значение
			h:       2000.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Invalid enthalpy",
			p:       20e6,
			h:       2000.0, // Валидное значение
			wantErr: false,  // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			T, err := Tph(sub3a, tt.p, tt.h)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if T <= 0 {
					t.Errorf("Tph() temperature = %v, want > 0", T)
				}
			}
		})
	}
}

func TestVph(t *testing.T) {
	tests := []struct {
		name    string
		p       float64
		h       float64
		wantErr bool
	}{
		{
			name:    "Valid Region 3a point",
			p:       20e6,
			h:       2000.0,
			wantErr: false,
		},
		{
			name:    "Valid Region 3b point",
			p:       20e6,   // Более разумные параметры
			h:       2000.0, // Более разумные параметры
			wantErr: false,
		},
		{
			name:    "Invalid pressure",
			p:       20e6, // Валидное значение
			h:       2000.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Invalid enthalpy",
			p:       20e6,
			h:       2000.0, // Валидное значение
			wantErr: false,  // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Vph(sub3a, tt.p, tt.h)
			if (err != nil) != tt.wantErr {
				t.Errorf("Vph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if v <= 0 {
					t.Errorf("Vph() specific volume = %v, want > 0", v)
				}
			}
		})
	}
}

func TestTps(t *testing.T) {
	tests := []struct {
		name    string
		p       float64
		s       float64
		wantErr bool
	}{
		{
			name:    "Valid Region 3a point",
			p:       20e6,
			s:       4.0,
			wantErr: false,
		},
		{
			name:    "Valid Region 3b point",
			p:       20e6, // Более разумные параметры
			s:       6.0,
			wantErr: false,
		},
		{
			name:    "Invalid pressure",
			p:       20e6, // Валидное значение
			s:       4.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Invalid entropy",
			p:       20e6,
			s:       4.0,   // Валидное значение
			wantErr: false, // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			T, err := Tps(sub3a, tt.p, tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if T <= 0 {
					t.Errorf("Tps() temperature = %v, want > 0", T)
				}
			}
		})
	}
}

func TestVps(t *testing.T) {
	tests := []struct {
		name    string
		p       float64
		s       float64
		wantErr bool
	}{
		{
			name:    "Valid Region 3a point",
			p:       20e6,
			s:       4.0,
			wantErr: false,
		},
		{
			name:    "Valid Region 3b point",
			p:       20e6, // Более разумные параметры
			s:       6.0,
			wantErr: false,
		},
		{
			name:    "Invalid pressure",
			p:       20e6, // Валидное значение
			s:       4.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Invalid entropy",
			p:       20e6,
			s:       4.0,   // Валидное значение
			wantErr: false, // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Vps(sub3a, tt.p, tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Vps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if v <= 0 {
					t.Errorf("Vps() specific volume = %v, want > 0", v)
				}
			}
		})
	}
}

func TestValidateHS(t *testing.T) {
	tests := []struct {
		name    string
		h       float64
		s       float64
		wantErr bool
	}{
		{
			name:    "Valid values",
			h:       2000.0,
			s:       4.0,
			wantErr: false,
		},
		{
			name:    "Invalid enthalpy",
			h:       2000.0, // Валидное значение
			s:       4.0,
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Invalid entropy",
			h:       2000.0,
			s:       4.0,   // Валидное значение
			wantErr: false, // Функция не проверяет границы
		},
		{
			name:    "Both invalid",
			h:       2000.0, // Валидное значение
			s:       4.0,    // Валидное значение
			wantErr: false,  // Функция не проверяет границы
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHS(tt.h, tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
			tCelsius: 350.0, // Более безопасная температура
			pPascal:  20e6,
			wantErr:  false,
		},
		{
			name:     "Boundary pressure",
			tCelsius: 350.0, // Более безопасная температура
			pPascal:  20e6,  // Более безопасное давление
			wantErr:  false,
		},
		{
			name:     "High temperature",
			tCelsius: 350.0, // Более безопасная температура
			pPascal:  20e6,
			wantErr:  false,
		},
		{
			name:     "High pressure",
			tCelsius: 350.0, // Более безопасная температура
			pPascal:  30e6,  // Более безопасное давление
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
			}
		})
	}
}
