package region1

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/somepgs/steamprops/internal/calc_core"
	"github.com/somepgs/steamprops/internal/calc_core/region4"
)

const (
	referP = 16.53    // MPa
	referT = 1386.0   // K
	referR = 0.461526 // kJ/kg*K
)

//go:embed iapws-if97-region1.csv
var coeffRegion1 embed.FS

type tableData struct {
	I  int
	Ii float64
	Ji float64
	Ni float64
}

var (
	rows           []tableData
	rowsLoadedOnce bool
)

func loadRowsOnce() error {
	if rowsLoadedOnce {
		return nil
	}
	f, err := coeffRegion1.Open("iapws-if97-region1.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ";")
		if len(parts) != 4 {
			return fmt.Errorf("invalid line in CSV: %s", line)
		}
		i, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return err
		}
		ii, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		ji, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			return err
		}
		ni, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return err
		}
		rows = append(rows, tableData{i, ii, ji, ni})
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	rowsLoadedOnce = true
	return nil
}

// Calculate computes Region 1 properties for T in Celsius and P in Pascals.
func Calculate(tCelsius, pPascal float64) (calc_core.Properties, error) {
	if err := loadRowsOnce(); err != nil {
		return calc_core.Properties{}, err
	}
	if tCelsius < -273.15 {
		return calc_core.Properties{}, errors.New("temperature below absolute zero")
	}
	if pPascal <= 0 {
		return calc_core.Properties{}, errors.New("pressure must be positive")
	}

	T := tCelsius + 273.15
	// Region 1 applicability (simplified): T <= 623.15 K, p <= 100 MPa, p >= psat(T)
	if T > 623.15 {
		return calc_core.Properties{}, fmt.Errorf("Region 1 not applicable: T=%.2f K exceeds 623.15 K", T)
	}
	if pPascal > 100e6 {
		return calc_core.Properties{}, fmt.Errorf("Region 1 not applicable: p=%.0f Pa exceeds 100 MPa", pPascal)
	}
	if T < 647.096 {
		if ps, err := region4.SaturationPressure(T); err == nil {
			if pPascal < ps {
				return calc_core.Properties{}, fmt.Errorf("Region 1 not applicable: p < psat(%.2f K)", T)
			}
		}
	}

	PMPa := pPascal / 1_000_000.0

	pi := PMPa / referP
	tau := referT / T

	var g, gPi, gPiPi, gTau, gTauTau, gPiTau float64
	for _, row := range rows {
		powPi := math.Pow(7.1-pi, row.Ii)
		powTau := math.Pow(tau-1.222, row.Ji)
		g += row.Ni * powPi * powTau
	}
	for _, row := range rows {
		gPi += (-row.Ni) * row.Ii * math.Pow(7.1-pi, row.Ii-1) * math.Pow(tau-1.222, row.Ji)
	}
	for _, row := range rows {
		gPiPi += row.Ni * row.Ii * (row.Ii - 1) * math.Pow(7.1-pi, row.Ii-2) * math.Pow(tau-1.222, row.Ji)
	}
	for _, row := range rows {
		gTau += row.Ni * row.Ji * math.Pow(7.1-pi, row.Ii) * math.Pow(tau-1.222, row.Ji-1)
	}
	for _, row := range rows {
		gTauTau += row.Ni * row.Ji * (row.Ji - 1) * math.Pow(7.1-pi, row.Ii) * math.Pow(tau-1.222, row.Ji-2)
	}
	for _, row := range rows {
		gPiTau += (-row.Ni) * row.Ii * row.Ji * math.Pow(7.1-pi, row.Ii-1) * math.Pow(tau-1.222, row.Ji-1)
	}

	R := referR
	pKPa := pPascal / 1000.0

	v := pi * gPi * (R * T / pKPa)
	ro := pKPa / (R * T * (pi * gPi))
	u := R * T * (tau*gTau - pi*gPi)
	s := R * (tau*gTau - g)
	h := R * T * tau * gTau
	cv := R * (-(tau*tau)*gTauTau + (math.Pow(gPi-tau*gPiTau, 2) / gPiPi))
	cp := R * (-(tau*tau)*gTauTau + (math.Pow(gPi-tau*gPiTau, 2) / gPiPi))
	// Calculate speed of sound using alternative IF-97 formula
	// w² = R * T * gPi² / (gPiPi - (gPi - tau*gPiTau)² / (tau² * gTauTau))
	// But if denominator is negative, use absolute value (common in some implementations)
	denominator := gPiPi - (math.Pow(gPi-tau*gPiTau, 2) / (tau * tau * gTauTau))

	if math.Abs(denominator) < 1e-10 {
		return calc_core.Properties{}, fmt.Errorf("Region 1: speed of sound calculation failed (denominator too small: %.6f)", denominator)
	}
	w := math.Sqrt(R * 1000.0 * T * (gPi * gPi / math.Abs(denominator)))

	// Sanity validation
	if !finiteAll(v, ro, u, s, h, cv, cp, w) {
		return calc_core.Properties{}, errors.New("Region 1 calculation produced non-finite values")
	}
	if ro <= 0 || v <= 0 || w <= 0 {
		return calc_core.Properties{}, errors.New("Region 1 invalid physical result (negative density/volume/speed)")
	}
	// cp and cv should be positive in Region 1
	if cp <= 0 || cv <= 0 {
		return calc_core.Properties{}, errors.New("Region 1 heat capacities are non-positive; inputs may be out of applicability")
	}

	return calc_core.Properties{
		SpecificVolume:                v,
		Density:                       ro,
		SpecificInternalEnergy:        u,
		SpecificEntropy:               s,
		SpecificEnthalpy:              h,
		SpecificIsochoricHeatCapacity: cv,
		SpecificIsobaricHeatCapacity:  cp,
		SpeedOfSound:                  w,
	}, nil
}

func finiteAll(vals ...float64) bool {
	for _, x := range vals {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return false
		}
	}
	return true
}
