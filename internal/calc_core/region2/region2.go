package region2

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"github.com/somepgs/steamprops/internal/calc_core"
	"github.com/somepgs/steamprops/internal/calc_core/region4"
	"math"
	"strconv"
	"strings"
)

const (
	referP = 1.0      // MPa
	referT = 540.0    // K
	referR = 0.461526 // kJ/kg*K
)

//go:embed iapws-if97-region2-0.csv
var coeffIdeal embed.FS

//go:embed iapws-if97-region2-r.csv
var coeffResidual embed.FS

type idealRow struct {
	J float64
	N float64
}

type residualRow struct {
	I float64
	J float64
	N float64
}

var (
	idealLoaded bool
	residLoaded bool
	idealRows   []idealRow
	residRows   []residualRow
)

func loadIdealOnce() error {
	if idealLoaded {
		return nil
	}
	f, err := coeffIdeal.Open("iapws-if97-region2-0.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue // skip header
		}
		parts := strings.Split(line, ";")
		if len(parts) != 3 {
			return fmt.Errorf("invalid ideal line: %s", line)
		}
		j, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			return err
		}
		idealRows = append(idealRows, idealRow{J: j, N: n})
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	idealLoaded = true
	return nil
}

func loadResidualOnce() error {
	if residLoaded {
		return nil
	}
	f, err := coeffResidual.Open("iapws-if97-region2-r.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue // header
		}
		parts := strings.Split(line, ";")
		if len(parts) != 4 {
			return fmt.Errorf("invalid residual line: %s", line)
		}
		i, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		j, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			return err
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return err
		}
		residRows = append(residRows, residualRow{I: i, J: j, N: n})
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	residLoaded = true
	return nil
}

// Calculate computes Region 2 properties for T in Celsius and P in Pascals.
func Calculate(tCelsius, pPascal float64) (calc_core.Properties, error) {
	if tCelsius < -273.15 {
		return calc_core.Properties{}, errors.New("temperature below absolute zero")
	}
	if pPascal <= 0 {
		return calc_core.Properties{}, errors.New("pressure must be positive")
	}
	if err := loadIdealOnce(); err != nil {
		return calc_core.Properties{}, err
	}
	if err := loadResidualOnce(); err != nil {
		return calc_core.Properties{}, err
	}

	T := tCelsius + 273.15
	// Region 2 applicability (simplified): T >= 273.15 K and <= 1073.15 K, p <= 100 MPa, p <= psat(T) below critical
	if T < 273.15 {
		return calc_core.Properties{}, fmt.Errorf("Region 2 not applicable: T=%.2f K below 273.15 K", T)
	}
	if T > 1073.15 {
		return calc_core.Properties{}, fmt.Errorf("Region 2 not applicable: T=%.2f K exceeds 1073.15 K", T)
	}
	if pPascal > 100e6 {
		return calc_core.Properties{}, fmt.Errorf("Region 2 not applicable: p=%.0f Pa exceeds 100 MPa", pPascal)
	}
	if T < 647.096 {
		if ps, err := region4.SaturationPressure(T); err == nil {
			if pPascal > ps {
				return calc_core.Properties{}, fmt.Errorf("Region 2 not applicable: p > psat(%.2f K)", T)
			}
		}
	}

	Tval := T
	PMPa := pPascal / 1_000_000.0
	pi := PMPa / referP
	tau := referT / Tval

	var g0, g0Tau, g0TauTau float64
	g0 = math.Log(pi)
	for _, r := range idealRows {
		g0 += r.N * math.Pow(tau, r.J)
		g0Tau += r.N * r.J * math.Pow(tau, r.J-1)
		g0TauTau += r.N * r.J * (r.J - 1) * math.Pow(tau, r.J-2)
	}
	g0Pi := 1.0 / pi
	g0PiPi := -1.0 / (pi * pi)

	var gr, grPi, grPiPi, grTau, grTauTau, grPiTau float64
	for _, r := range residRows {
		gr += r.N * math.Pow(pi, r.I) * math.Pow(tau-0.5, r.J)
		grPi += r.N * r.I * math.Pow(pi, r.I-1) * math.Pow(tau-0.5, r.J)
		grPiPi += r.N * r.I * (r.I - 1) * math.Pow(pi, r.I-2) * math.Pow(tau-0.5, r.J)
		grTau += r.N * r.J * math.Pow(pi, r.I) * math.Pow(tau-0.5, r.J-1)
		grTauTau += r.N * r.J * (r.J - 1) * math.Pow(pi, r.I) * math.Pow(tau-0.5, r.J-2)
		grPiTau += r.N * r.I * r.J * math.Pow(pi, r.I-1) * math.Pow(tau-0.5, r.J-1)
	}

	gPi := g0Pi + grPi
	gPiPi := g0PiPi + grPiPi
	gTau := g0Tau + grTau
	gTauTau := g0TauTau + grTauTau
	gPiTau := grPiTau
	g := g0 + gr

	R := referR
	pKPa := pPascal / 1000.0

	v := pi * gPi * (R * Tval / pKPa)
	ro := pKPa / (R * Tval * (pi * gPi))
	u := R * Tval * (tau*gTau - pi*gPi)
	s := R * (tau*gTau - g)
	h := R * Tval * tau * gTau
	cv := R * (-(tau*tau)*gTauTau + (math.Pow(gPi-tau*gPiTau, 2) / gPiPi))
	cp := R * (-(tau*tau)*gTauTau + (gPi-tau*gPiTau)*(gPi-tau*gPiTau)/gPiPi)
	w := math.Sqrt(R * 1000.0 * Tval * (gPi * gPi / ((math.Pow(gPi-tau*gPiTau, 2) / (tau * tau * gTauTau)) - gPiPi)))

	// Sanity validation
	if !finiteAll(v, ro, u, s, h, cv, cp, w) {
		return calc_core.Properties{}, errors.New("Region 2 calculation produced non-finite values")
	}
	if ro <= 0 || v <= 0 || w <= 0 {
		return calc_core.Properties{}, errors.New("Region 2 invalid physical result (negative density/volume/speed)")
	}
	if cp <= 0 || cv <= 0 {
		return calc_core.Properties{}, errors.New("Region 2 heat capacities are non-positive; inputs may be out of applicability")
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
