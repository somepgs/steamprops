package transport

import (
	"errors"
	"math"
)

// DynamicViscosity returns dynamic viscosity μ in Pa·s using simplified IAPWS formulation.
// Inputs:
//   - Tkelvin: temperature in K
//   - rho: density in kg/m^3
func DynamicViscosity(Tkelvin float64, rho float64) (float64, error) {
	if Tkelvin <= 0 || rho <= 0 {
		return 0, errors.New("invalid inputs for viscosity (T>0, rho>0 required)")
	}

	// Simplified IAPWS formulation for water viscosity
	// Based on empirical correlations for water

	// Reference viscosity at 20°C
	mu20 := 1.002e-3 // Pa·s

	// Temperature dependence (simplified)
	// For liquid water (T < 373.15 K)
	if Tkelvin < 373.15 {
		// Arrhenius-type temperature dependence
		mu := mu20 * math.Exp(1700.0*(1.0/Tkelvin-1.0/293.15))
		return mu, nil
	}

	// For steam (T >= 373.15 K)
	// Simplified Sutherland's law
	mu := 1.8e-5 * math.Pow(Tkelvin/288.0, 0.7)
	return mu, nil
}

// ThermalConductivity returns thermal conductivity λ in W/(m·K) using simplified formulation.
// Inputs:
//   - Tkelvin: temperature in K
//   - rho: density in kg/m^3
func ThermalConductivity(Tkelvin float64, rho float64) (float64, error) {
	if Tkelvin <= 0 || rho <= 0 {
		return 0, errors.New("invalid inputs for thermal conductivity (T>0, rho>0 required)")
	}

	// Simplified thermal conductivity for water/steam

	// For liquid water (high density)
	if rho > 500 {
		// Temperature dependence for liquid water
		lambda := 0.6 + 0.001*(Tkelvin-273.15)
		return lambda, nil
	}

	// For steam (low density)
	// Temperature dependence for steam
	lambda := 0.02 + 0.0001*(Tkelvin-273.15)
	return lambda, nil
}

// KinematicViscosity returns kinematic viscosity ν in m²/s
func KinematicViscosity(Tkelvin float64, rho float64) (float64, error) {
	mu, err := DynamicViscosity(Tkelvin, rho)
	if err != nil {
		return 0, err
	}

	if rho <= 0 {
		return 0, errors.New("density must be positive for kinematic viscosity")
	}

	return mu / rho, nil
}
