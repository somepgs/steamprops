package validation

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// InputValidator provides comprehensive input validation
type InputValidator struct {
	// Temperature limits (in Celsius)
	MinTemperature float64
	MaxTemperature float64

	// Pressure limits (in Pa)
	MinPressure float64
	MaxPressure float64

	// Enthalpy limits (in kJ/kg)
	MinEnthalpy float64
	MaxEnthalpy float64

	// Entropy limits (in kJ/(kg·K))
	MinEntropy float64
	MaxEntropy float64
}

// NewInputValidator creates a new input validator with default limits
func NewInputValidator() *InputValidator {
	return &InputValidator{
		MinTemperature: -273.15, // Absolute zero
		MaxTemperature: 2000.0,  // Region 5 upper limit
		MinPressure:    611.657, // Triple point pressure
		MaxPressure:    100e6,   // 100 MPa
		MinEnthalpy:    0.0,
		MaxEnthalpy:    5000.0,
		MinEntropy:     0.0,
		MaxEntropy:     15.0,
	}
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(message string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, message)
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// ValidateTemperature validates temperature input
func (iv *InputValidator) ValidateTemperature(tCelsius float64) *ValidationResult {
	result := NewValidationResult()

	// Check for NaN or Inf
	if math.IsNaN(tCelsius) {
		result.AddError("Temperature cannot be NaN")
		return result
	}
	if math.IsInf(tCelsius, 0) {
		result.AddError("Temperature cannot be infinite")
		return result
	}

	// Check range
	if tCelsius < iv.MinTemperature {
		result.AddError(fmt.Sprintf("Temperature %.2f°C is below minimum %.2f°C",
			tCelsius, iv.MinTemperature))
	}
	if tCelsius > iv.MaxTemperature {
		result.AddError(fmt.Sprintf("Temperature %.2f°C is above maximum %.2f°C",
			tCelsius, iv.MaxTemperature))
	}

	// Add warnings for extreme values
	if tCelsius < 0 && tCelsius >= iv.MinTemperature {
		result.AddWarning("Temperature below freezing point")
	}
	if tCelsius > 1000 && tCelsius <= iv.MaxTemperature {
		result.AddWarning("Very high temperature - ensure Region 5 applicability")
	}

	return result
}

// ValidatePressure validates pressure input
func (iv *InputValidator) ValidatePressure(pPascal float64) *ValidationResult {
	result := NewValidationResult()

	// Check for NaN or Inf
	if math.IsNaN(pPascal) {
		result.AddError("Pressure cannot be NaN")
		return result
	}
	if math.IsInf(pPascal, 0) {
		result.AddError("Pressure cannot be infinite")
		return result
	}

	// Check range
	if pPascal <= 0 {
		result.AddError("Pressure must be positive")
		return result
	}
	if pPascal < iv.MinPressure {
		result.AddError(fmt.Sprintf("Pressure %.0f Pa is below minimum %.0f Pa",
			pPascal, iv.MinPressure))
	}
	if pPascal > iv.MaxPressure {
		result.AddError(fmt.Sprintf("Pressure %.0f Pa is above maximum %.0f Pa",
			pPascal, iv.MaxPressure))
	}

	// Add warnings for extreme values
	if pPascal < 1000 && pPascal >= iv.MinPressure {
		result.AddWarning("Very low pressure - ensure Region 2 applicability")
	}
	if pPascal > 50e6 && pPascal <= iv.MaxPressure {
		result.AddWarning("Very high pressure - ensure Region 1/3 applicability")
	}

	return result
}

// ValidateEnthalpy validates enthalpy input
func (iv *InputValidator) ValidateEnthalpy(h float64) *ValidationResult {
	result := NewValidationResult()

	// Check for NaN or Inf
	if math.IsNaN(h) {
		result.AddError("Enthalpy cannot be NaN")
		return result
	}
	if math.IsInf(h, 0) {
		result.AddError("Enthalpy cannot be infinite")
		return result
	}

	// Check range
	if h < iv.MinEnthalpy {
		result.AddError(fmt.Sprintf("Enthalpy %.2f kJ/kg is below minimum %.2f kJ/kg",
			h, iv.MinEnthalpy))
	}
	if h > iv.MaxEnthalpy {
		result.AddError(fmt.Sprintf("Enthalpy %.2f kJ/kg is above maximum %.2f kJ/kg",
			h, iv.MaxEnthalpy))
	}

	return result
}

// ValidateEntropy validates entropy input
func (iv *InputValidator) ValidateEntropy(s float64) *ValidationResult {
	result := NewValidationResult()

	// Check for NaN or Inf
	if math.IsNaN(s) {
		result.AddError("Entropy cannot be NaN")
		return result
	}
	if math.IsInf(s, 0) {
		result.AddError("Entropy cannot be infinite")
		return result
	}

	// Check range
	if s < iv.MinEntropy {
		result.AddError(fmt.Sprintf("Entropy %.2f kJ/(kg·K) is below minimum %.2f kJ/(kg·K)",
			s, iv.MinEntropy))
	}
	if s > iv.MaxEntropy {
		result.AddError(fmt.Sprintf("Entropy %.2f kJ/(kg·K) is above maximum %.2f kJ/(kg·K)",
			s, iv.MaxEntropy))
	}

	return result
}

// ValidateTemperaturePressure validates temperature and pressure combination
func (iv *InputValidator) ValidateTemperaturePressure(tCelsius, pPascal float64) *ValidationResult {
	result := NewValidationResult()

	// Validate individual inputs
	tempResult := iv.ValidateTemperature(tCelsius)
	if !tempResult.IsValid {
		result.Errors = append(result.Errors, tempResult.Errors...)
		result.IsValid = false
	}
	result.Warnings = append(result.Warnings, tempResult.Warnings...)

	pressResult := iv.ValidatePressure(pPascal)
	if !pressResult.IsValid {
		result.Errors = append(result.Errors, pressResult.Errors...)
		result.IsValid = false
	}
	result.Warnings = append(result.Warnings, pressResult.Warnings...)

	// Additional checks for combination
	if result.IsValid {
		tKelvin := tCelsius + 273.15

		// Check for supercritical conditions
		if tKelvin > 647.096 && pPascal > 22.064e6 {
			result.AddWarning("Supercritical conditions - ensure Region 3 applicability")
		}

		// Check for near-critical conditions
		if math.Abs(tKelvin-647.096) < 1.0 && math.Abs(pPascal-22.064e6) < 1e6 {
			result.AddWarning("Near critical point - calculations may be sensitive")
		}
	}

	return result
}

// ValidateEnthalpyEntropy validates enthalpy and entropy combination
func (iv *InputValidator) ValidateEnthalpyEntropy(h, s float64) *ValidationResult {
	result := NewValidationResult()

	// Validate individual inputs
	hResult := iv.ValidateEnthalpy(h)
	if !hResult.IsValid {
		result.Errors = append(result.Errors, hResult.Errors...)
		result.IsValid = false
	}
	result.Warnings = append(result.Warnings, hResult.Warnings...)

	sResult := iv.ValidateEntropy(s)
	if !sResult.IsValid {
		result.Errors = append(result.Errors, sResult.Errors...)
		result.IsValid = false
	}
	result.Warnings = append(result.Warnings, sResult.Warnings...)

	return result
}

// ValidateStringInput validates string input and converts to float64
func (iv *InputValidator) ValidateStringInput(input string, inputType string) (float64, *ValidationResult) {
	result := NewValidationResult()

	// Check for empty input
	if strings.TrimSpace(input) == "" {
		result.AddError(fmt.Sprintf("%s cannot be empty", inputType))
		return 0, result
	}

	// Check for valid number format
	matched, _ := regexp.MatchString(`^-?\d*\.?\d+([eE][+-]?\d+)?$`, strings.TrimSpace(input))
	if !matched {
		result.AddError(fmt.Sprintf("Invalid %s format: %s", inputType, input))
		return 0, result
	}

	// Convert to float64
	value, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		result.AddError(fmt.Sprintf("Cannot parse %s: %s", inputType, err.Error()))
		return 0, result
	}

	return value, result
}

// ValidateUnitConversion validates unit conversion parameters
func (iv *InputValidator) ValidateUnitConversion(value float64, unit string, property string) *ValidationResult {
	result := NewValidationResult()

	// Check for valid unit
	validUnits := map[string][]string{
		"temperature": {"°C", "K", "°F"},
		"pressure":    {"Pa", "MPa", "bar", "atm", "кгс/см²", "мм рт.ст.", "мм вод.ст."},
		"enthalpy":    {"кДж/кг", "Дж/кг"},
		"entropy":     {"кДж/(кг·К)", "Дж/(кг·К)"},
	}

	if units, exists := validUnits[property]; exists {
		valid := false
		for _, u := range units {
			if u == unit {
				valid = true
				break
			}
		}
		if !valid {
			result.AddError(fmt.Sprintf("Invalid unit %s for %s", unit, property))
		}
	}

	// Check for reasonable values after unit conversion
	if property == "temperature" && unit == "K" && value < 0 {
		result.AddError("Temperature in Kelvin cannot be negative")
	}

	return result
}

// GetValidationSummary returns a summary of validation results
func (vr *ValidationResult) GetValidationSummary() string {
	if vr.IsValid && len(vr.Warnings) == 0 {
		return "Validation passed"
	}

	summary := ""
	if !vr.IsValid {
		summary += "Validation failed:\n"
		for _, err := range vr.Errors {
			summary += "  - " + err + "\n"
		}
	}

	if len(vr.Warnings) > 0 {
		if summary != "" {
			summary += "\n"
		}
		summary += "Warnings:\n"
		for _, warn := range vr.Warnings {
			summary += "  - " + warn + "\n"
		}
	}

	return strings.TrimSpace(summary)
}
