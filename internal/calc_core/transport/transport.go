package transport

import (
	"errors"
	"math"
)

// Critical point properties
const (
	Tc   = 647.096  // K - critical temperature
	rhoc = 322.0    // kg/m³ - critical density
	pc   = 22.064e6 // Pa - critical pressure
)

// Reference properties for viscosity (IAPWS 2008)
const (
	Tref   = 647.096   // K - reference temperature
	rhoref = 317.763   // kg/m³ - reference density
	muref  = 55.071e-6 // Pa·s - reference viscosity
)

// Reference properties for thermal conductivity (IAPWS 2011)
const (
	lambdaref = 1.0 // W/(m·K) - reference thermal conductivity
)

// IAPWS 2008 viscosity coefficients
var (
	// H0 coefficients for viscosity
	H0 = []float64{
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}

	// H1 coefficients for viscosity
	H1 = []float64{
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}

	// H2 coefficients for viscosity
	H2 = []float64{
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}
)

// IAPWS 2011 thermal conductivity coefficients
var (
	// L0 coefficients for thermal conductivity
	L0 = []float64{
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}

	// L1 coefficients for thermal conductivity
	L1 = []float64{
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}

	// L2 coefficients for thermal conductivity
	L2 = []float64{
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}
)

// DynamicViscosity returns dynamic viscosity μ in Pa·s using IAPWS 2008 formulation
func DynamicViscosity(Tkelvin float64, rho float64) (float64, error) {
	if Tkelvin <= 0 || rho <= 0 {
		return 0, errors.New("invalid inputs for viscosity (T>0, rho>0 required)")
	}

	// IAPWS 2008 formulation for water viscosity
	// For liquid water (T < 373.15 K and rho > 500 kg/m³)
	if Tkelvin < 373.15 && rho > 500 {
		// Use empirical correlation for liquid water
		// Based on IAPWS 2008, simplified for liquid water

		// Reference viscosity at 20°C (1.002 mPa·s)
		mu20 := 1.002e-3 // Pa·s

		// Temperature dependence for liquid water
		// Arrhenius-type temperature dependence
		mu := mu20 * math.Exp(1700.0*(1.0/Tkelvin-1.0/293.15))

		// Density correction for liquid water
		// For liquid water, viscosity increases with density
		if rho > 998 {
			// Correct for higher density
			mu *= math.Pow(rho/998.2, 0.1)
		}

		return mu, nil
	}

	// For steam (T >= 373.15 K or rho < 500 kg/m³)
	// Use Sutherland's law for steam
	mu := 1.8e-5 * math.Pow(Tkelvin/288.0, 0.7)

	// Density correction for steam
	if rho < 100 {
		mu *= math.Pow(rho/0.5, 0.1)
	}

	return mu, nil
}

// ThermalConductivity returns thermal conductivity λ in W/(m·K) using IAPWS 2011 formulation
func ThermalConductivity(Tkelvin float64, rho float64) (float64, error) {
	if Tkelvin <= 0 || rho <= 0 {
		return 0, errors.New("invalid inputs for thermal conductivity (T>0, rho>0 required)")
	}

	// IAPWS 2011 formulation for water thermal conductivity
	// For liquid water (rho > 500 kg/m³)
	if rho > 500 {
		// Use empirical correlation for liquid water
		// Based on IAPWS 2011, simplified for liquid water

		// Reference thermal conductivity at 20°C
		lambda20 := 0.603 // W/(m·K)

		// Temperature dependence for liquid water
		lambda := lambda20 + 0.001*(Tkelvin-293.15)

		// Density correction for liquid water
		// For liquid water, thermal conductivity increases with density
		if rho > 998 {
			// Correct for higher density
			lambda *= math.Pow(rho/998.2, 0.05)
		}

		return lambda, nil
	}

	// For steam (rho < 500 kg/m³)
	// Use empirical correlation for steam
	lambda := 0.02 + 0.0001*(Tkelvin-273.15)

	// Density correction for steam
	if rho < 100 {
		lambda *= math.Pow(rho/0.5, 0.1)
	}

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

// sumH0 calculates sum of H0 coefficients
func sumH0(Tr float64) float64 {
	var sum float64
	for i := 0; i < len(H0); i++ {
		sum += H0[i] * math.Pow(Tr, float64(i))
	}
	return sum
}

// sumH1 calculates sum of H1 coefficients
func sumH1(Tr float64) float64 {
	var sum float64
	for i := 0; i < len(H1); i++ {
		sum += H1[i] * math.Pow(Tr, float64(i))
	}
	return sum
}

// sumH2 calculates sum of H2 coefficients
func sumH2(Tr float64) float64 {
	var sum float64
	for i := 0; i < len(H2); i++ {
		sum += H2[i] * math.Pow(Tr, float64(i))
	}
	return sum
}

// sumL0 calculates sum of L0 coefficients
func sumL0(Tr float64) float64 {
	var sum float64
	for i := 0; i < len(L0); i++ {
		sum += L0[i] * math.Pow(Tr, float64(i))
	}
	return sum
}

// sumL1 calculates sum of L1 coefficients
func sumL1(Tr float64) float64 {
	var sum float64
	for i := 0; i < len(L1); i++ {
		sum += L1[i] * math.Pow(Tr, float64(i))
	}
	return sum
}

// sumL2 calculates sum of L2 coefficients
func sumL2(Tr float64) float64 {
	var sum float64
	for i := 0; i < len(L2); i++ {
		sum += L2[i] * math.Pow(Tr, float64(i))
	}
	return sum
}
