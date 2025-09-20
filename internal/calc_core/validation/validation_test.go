package validation

import (
	"math"
	"testing"
)

func TestInputValidator_ValidateTemperature(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		temp     float64
		wantErr  bool
		wantWarn bool
	}{
		{
			name:     "Valid temperature",
			temp:     200.0,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "Below freezing",
			temp:     -10.0,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "Very high temperature",
			temp:     1500.0,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "Below absolute zero",
			temp:     -300.0,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Above maximum",
			temp:     2500.0,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "NaN",
			temp:     math.NaN(),
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Infinity",
			temp:     math.Inf(1),
			wantErr:  true,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateTemperature(tt.temp)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateTemperature() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}

			if (len(result.Warnings) > 0) != tt.wantWarn {
				t.Errorf("ValidateTemperature() warnings = %v, wantWarn %v", result.Warnings, tt.wantWarn)
			}
		})
	}
}

func TestInputValidator_ValidatePressure(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		pressure float64
		wantErr  bool
		wantWarn bool
	}{
		{
			name:     "Valid pressure",
			pressure: 1e6,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "Very low pressure",
			pressure: 500.0, // Below minimum pressure
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Very high pressure",
			pressure: 80e6,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "Zero pressure",
			pressure: 0.0,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Negative pressure",
			pressure: -1000.0,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Above maximum",
			pressure: 150e6,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "NaN",
			pressure: math.NaN(),
			wantErr:  true,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePressure(tt.pressure)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidatePressure() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}

			if (len(result.Warnings) > 0) != tt.wantWarn {
				t.Errorf("ValidatePressure() warnings = %v, wantWarn %v", result.Warnings, tt.wantWarn)
			}
		})
	}
}

func TestInputValidator_ValidateEnthalpy(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		enthalpy float64
		wantErr  bool
	}{
		{
			name:     "Valid enthalpy",
			enthalpy: 2000.0,
			wantErr:  false,
		},
		{
			name:     "Below minimum",
			enthalpy: -100.0,
			wantErr:  true,
		},
		{
			name:     "Above maximum",
			enthalpy: 6000.0,
			wantErr:  true,
		},
		{
			name:     "NaN",
			enthalpy: math.NaN(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEnthalpy(tt.enthalpy)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateEnthalpy() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateEntropy(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name    string
		entropy float64
		wantErr bool
	}{
		{
			name:    "Valid entropy",
			entropy: 5.0,
			wantErr: false,
		},
		{
			name:    "Below minimum",
			entropy: -1.0,
			wantErr: true,
		},
		{
			name:    "Above maximum",
			entropy: 20.0,
			wantErr: true,
		},
		{
			name:    "NaN",
			entropy: math.NaN(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEntropy(tt.entropy)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateEntropy() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateTemperaturePressure(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		temp     float64
		pressure float64
		wantErr  bool
		wantWarn bool
	}{
		{
			name:     "Valid combination",
			temp:     200.0,
			pressure: 1e6,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "Supercritical conditions",
			temp:     700.0,
			pressure: 25e6,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "Near critical point",
			temp:     374.0,
			pressure: 22e6,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "Invalid temperature",
			temp:     -300.0,
			pressure: 1e6,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Invalid pressure",
			temp:     200.0,
			pressure: -1000.0,
			wantErr:  true,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateTemperaturePressure(tt.temp, tt.pressure)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateTemperaturePressure() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}

			if (len(result.Warnings) > 0) != tt.wantWarn {
				t.Errorf("ValidateTemperaturePressure() warnings = %v, wantWarn %v", result.Warnings, tt.wantWarn)
			}
		})
	}
}

func TestInputValidator_ValidateStringInput(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name      string
		input     string
		inputType string
		wantErr   bool
		wantValue float64
	}{
		{
			name:      "Valid number",
			input:     "200.5",
			inputType: "temperature",
			wantErr:   false,
			wantValue: 200.5,
		},
		{
			name:      "Scientific notation",
			input:     "1.5e6",
			inputType: "pressure",
			wantErr:   false,
			wantValue: 1.5e6,
		},
		{
			name:      "Empty input",
			input:     "",
			inputType: "temperature",
			wantErr:   true,
			wantValue: 0,
		},
		{
			name:      "Invalid format",
			input:     "abc",
			inputType: "temperature",
			wantErr:   true,
			wantValue: 0,
		},
		{
			name:      "Whitespace",
			input:     "  200.5  ",
			inputType: "temperature",
			wantErr:   false,
			wantValue: 200.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, result := validator.ValidateStringInput(tt.input, tt.inputType)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateStringInput() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}

			if !tt.wantErr && value != tt.wantValue {
				t.Errorf("ValidateStringInput() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestInputValidator_ValidateUnitConversion(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		value    float64
		unit     string
		property string
		wantErr  bool
	}{
		{
			name:     "Valid temperature unit",
			value:    200.0,
			unit:     "°C",
			property: "temperature",
			wantErr:  false,
		},
		{
			name:     "Valid pressure unit",
			value:    1.0,
			unit:     "MPa",
			property: "pressure",
			wantErr:  false,
		},
		{
			name:     "Invalid unit",
			value:    200.0,
			unit:     "invalid",
			property: "temperature",
			wantErr:  true,
		},
		{
			name:     "Negative Kelvin",
			value:    -100.0,
			unit:     "K",
			property: "temperature",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateUnitConversion(tt.value, tt.unit, tt.property)

			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateUnitConversion() errors = %v, wantErr %v", result.Errors, tt.wantErr)
			}
		})
	}
}

func TestValidationResult_GetValidationSummary(t *testing.T) {
	tests := []struct {
		name     string
		result   *ValidationResult
		expected string
	}{
		{
			name: "Valid result",
			result: &ValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
			expected: "Validation passed",
		},
		{
			name: "Result with errors",
			result: &ValidationResult{
				IsValid:  false,
				Errors:   []string{"Error 1", "Error 2"},
				Warnings: []string{},
			},
			expected: "Validation failed:\n  - Error 1\n  - Error 2",
		},
		{
			name: "Result with warnings",
			result: &ValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{"Warning 1"},
			},
			expected: "Warnings:\n  - Warning 1",
		},
		{
			name: "Result with errors and warnings",
			result: &ValidationResult{
				IsValid:  false,
				Errors:   []string{"Error 1"},
				Warnings: []string{"Warning 1"},
			},
			expected: "Validation failed:\n  - Error 1\n\nWarnings:\n  - Warning 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.GetValidationSummary()
			if summary != tt.expected {
				t.Errorf("GetValidationSummary() = %v, want %v", summary, tt.expected)
			}
		})
	}
}
