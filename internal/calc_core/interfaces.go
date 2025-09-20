package calc_core

import (
	"fmt"
	"steamprops/internal/calc_core/bounds"
	"steamprops/internal/calc_core/region4"
)

// Properties represents thermodynamic properties of water/steam
type Properties struct {
	SpecificVolume                float64 // m^3/kg
	Density                       float64 // kg/m^3
	SpecificInternalEnergy        float64 // kJ/kg
	SpecificEntropy               float64 // kJ/(kg*K)
	SpecificEnthalpy              float64 // kJ/kg
	SpecificIsochoricHeatCapacity float64 // kJ/(kg*K)
	SpecificIsobaricHeatCapacity  float64 // kJ/(kg*K)
	SpeedOfSound                  float64 // m/s
}

// Region represents IF-97 regions
type Region int

const (
	RegionAuto Region = iota
	Region1
	Region2
	Region3
	Region4
	Region5
)

// RegionCalculator defines the interface for region calculations
type RegionCalculator interface {
	// Calculate computes thermodynamic properties for given temperature and pressure
	Calculate(tCelsius, pPascal float64) (Properties, error)

	// IsApplicable checks if the region is applicable for given conditions
	IsApplicable(tCelsius, pPascal float64) bool

	// GetRegion returns the region identifier
	GetRegion() Region
}

// BackwardCalculator defines interface for backward calculations
type BackwardCalculator interface {
	// CalculateFromHS computes properties from enthalpy and entropy
	CalculateFromHS(h, s float64) (Properties, error)

	// PressureFromHS computes pressure from enthalpy and entropy
	PressureFromHS(h, s float64) (float64, error)
}

// TransportCalculator defines interface for transport properties
type TransportCalculator interface {
	// DynamicViscosity returns dynamic viscosity in Pa·s
	DynamicViscosity(Tkelvin, rho float64) (float64, error)

	// ThermalConductivity returns thermal conductivity in W/(m·K)
	ThermalConductivity(Tkelvin, rho float64) (float64, error)

	// KinematicViscosity returns kinematic viscosity in m²/s
	KinematicViscosity(Tkelvin, rho float64) (float64, error)
}

// RegionFactory manages region calculators
type RegionFactory struct {
	calculators map[Region]RegionCalculator
}

// NewRegionFactory creates a new region factory
func NewRegionFactory() *RegionFactory {
	return &RegionFactory{
		calculators: make(map[Region]RegionCalculator),
	}
}

// RegisterCalculator registers a region calculator
func (rf *RegionFactory) RegisterCalculator(region Region, calculator RegionCalculator) {
	rf.calculators[region] = calculator
}

// GetCalculator returns the calculator for the specified region
func (rf *RegionFactory) GetCalculator(region Region) (RegionCalculator, error) {
	calculator, exists := rf.calculators[region]
	if !exists {
		return nil, &RegionNotSupportedError{Region: region}
	}
	return calculator, nil
}

// CalculateProperties calculates properties for given conditions
func (rf *RegionFactory) CalculateProperties(tCelsius, pPascal float64, region Region) (Properties, error) {
	if region == RegionAuto {
		region = RegionFromTP(tCelsius+273.15, pPascal)
	}

	calculator, err := rf.GetCalculator(region)
	if err != nil {
		return Properties{}, err
	}

	if !calculator.IsApplicable(tCelsius, pPascal) {
		return Properties{}, &RegionNotApplicableError{
			Region: region,
			T:      tCelsius,
			P:      pPascal,
		}
	}

	return calculator.Calculate(tCelsius, pPascal)
}

// RegionNotSupportedError represents an error when a region is not supported
type RegionNotSupportedError struct {
	Region Region
}

func (e *RegionNotSupportedError) Error() string {
	return fmt.Sprintf("region %d is not supported", e.Region)
}

// RegionNotApplicableError represents an error when a region is not applicable
type RegionNotApplicableError struct {
	Region Region
	T      float64
	P      float64
}

func (e *RegionNotApplicableError) Error() string {
	return fmt.Sprintf("region %d is not applicable for T=%.2f°C, p=%.0f Pa", e.Region, e.T, e.P)
}

// RegionFromTP returns a best-effort region guess using Region4 saturation and B23 boundary.
// Inputs: T in K, p in Pa.
func RegionFromTP(T float64, p float64) Region {
	// Very high T: Region 5 by IF-97 (T > 1073.15 K up to 2273.15 K, p up to 50 MPa)
	if T > 1073.15 {
		return Region5
	}
	// Saturation line separates regions 1 and 2 below ~ 647K
	if T < 647.096 {
		psat, err := region4.SaturationPressure(T)
		if err == nil {
			if p >= psat {
				return Region1
			}
			return Region2
		}
	}
	// Between 2 and 3 use B23
	T23, err := bounds.B23T(p / 1e6) // p in MPa
	if err == nil {
		if T >= T23 {
			return Region2
		}
		return Region3
	}
	// Fallback heuristic
	if p >= 16.5292e6 {
		return Region3
	}
	return Region2
}

// ValidationService provides input validation
type ValidationService struct {
	maxTemperature float64
	minTemperature float64
	maxPressure    float64
	minPressure    float64
}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	return &ValidationService{
		maxTemperature: 2273.15, // Region 5 upper limit
		minTemperature: 273.15,  // Triple point
		maxPressure:    100e6,   // 100 MPa
		minPressure:    611.657, // Triple point pressure
	}
}

// ValidateTemperature validates temperature input
func (vs *ValidationService) ValidateTemperature(tCelsius float64) error {
	tKelvin := tCelsius + 273.15
	if tKelvin < vs.minTemperature {
		return fmt.Errorf("temperature %.2f°C is below minimum %.2f°C", tCelsius, vs.minTemperature-273.15)
	}
	if tKelvin > vs.maxTemperature {
		return fmt.Errorf("temperature %.2f°C is above maximum %.2f°C", tCelsius, vs.maxTemperature-273.15)
	}
	return nil
}

// ValidatePressure validates pressure input
func (vs *ValidationService) ValidatePressure(pPascal float64) error {
	if pPascal < vs.minPressure {
		return fmt.Errorf("pressure %.0f Pa is below minimum %.0f Pa", pPascal, vs.minPressure)
	}
	if pPascal > vs.maxPressure {
		return fmt.Errorf("pressure %.0f Pa is above maximum %.0f Pa", pPascal, vs.maxPressure)
	}
	return nil
}

// ValidateEnthalpy validates enthalpy input
func (vs *ValidationService) ValidateEnthalpy(h float64) error {
	if h < 0 || h > 5000 {
		return fmt.Errorf("enthalpy %.2f kJ/kg is outside valid range [0, 5000]", h)
	}
	return nil
}

// ValidateEntropy validates entropy input
func (vs *ValidationService) ValidateEntropy(s float64) error {
	if s < 0 || s > 15 {
		return fmt.Errorf("entropy %.2f kJ/(kg·K) is outside valid range [0, 15]", s)
	}
	return nil
}
