package region5

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"github.com/somepgs/steamprops/internal/calc_core"
	"math"
	"strconv"
	"strings"
)

const (
	referP = 1.0      // MPa
	referT = 1000.0   // K
	referR = 0.461526 // kJ/kg*K
)

//go:embed iapws-if97-region5_0.csv
var coeffIdeal embed.FS

//go:embed iapws-if97-region5.csv
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
	f, err := coeffIdeal.Open("iapws-if97-region5_0.csv")
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
		parts := strings.Split(line, ",")
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
	f, err := coeffResidual.Open("iapws-if97-region5.csv")
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
		parts := strings.Split(line, ",")
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

// Calculate computes Region 5 properties for T in Celsius and P in Pascals.
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
	// IF-97 Region 5 applicability: 1073.15 K <= T <= 2273.15 K, p <= 50 MPa
	if T < 1073.15 || T > 2273.15 {
		return calc_core.Properties{}, fmt.Errorf("Region 5 not applicable: T=%.2f K out of [1073.15, 2273.15] K", T)
	}
	if pPascal > 50e6 {
		return calc_core.Properties{}, fmt.Errorf("Region 5 not applicable: p=%.0f Pa exceeds 50 MPa", pPascal)
	}

	PMPa := pPascal / 1_000_000.0
	pi := PMPa / referP
	tau := referT / T

	// Ideal part g0
	var g0, g0Tau, g0TauTau float64
	g0 = math.Log(pi)
	for _, r := range idealRows {
		g0 += r.N * math.Pow(tau, r.J)
		g0Tau += r.N * r.J * math.Pow(tau, r.J-1)
		g0TauTau += r.N * r.J * (r.J - 1) * math.Pow(tau, r.J-2)
	}
	g0Pi := 1.0 / pi
	g0PiPi := -1.0 / (pi * pi)

	// Residual part gr
	var gr, grPi, grPiPi, grTau, grTauTau, grPiTau float64
	for _, r := range residRows {
		gr += r.N * math.Pow(pi, r.I) * math.Pow(tau-1.0, r.J)
		grPi += r.N * r.I * math.Pow(pi, r.I-1) * math.Pow(tau-1.0, r.J)
		grPiPi += r.N * r.I * (r.I - 1) * math.Pow(pi, r.I-2) * math.Pow(tau-1.0, r.J)
		grTau += r.N * r.J * math.Pow(pi, r.I) * math.Pow(tau-1.0, r.J-1)
		grTauTau += r.N * r.J * (r.J - 1) * math.Pow(pi, r.I) * math.Pow(tau-1.0, r.J-2)
		grPiTau += r.N * r.I * r.J * math.Pow(pi, r.I-1) * math.Pow(tau-1.0, r.J-1)
	}

	gPi := g0Pi + grPi
	gPiPi := g0PiPi + grPiPi
	gTau := g0Tau + grTau
	gTauTau := g0TauTau + grTauTau
	gPiTau := grPiTau // g0PiTau is zero
	g := g0 + gr

	R := referR
	pKPa := pPascal / 1000.0

	v := pi * gPi * (R * T / pKPa)
	ro := pKPa / (R * T * (pi * gPi))
	u := R * T * (tau*gTau - pi*gPi)
	s := R * (tau*gTau - g)
	h := R * T * tau * gTau
	cv := R * (-(tau*tau)*gTauTau + (math.Pow(gPi-tau*gPiTau, 2) / gPiPi))
	cp := R * (-(tau*tau)*gTauTau + (gPi-tau*gPiTau)*(gPi-tau*gPiTau)/gPiPi)
	w := math.Sqrt(R * 1000.0 * T * (gPi * gPi / ((math.Pow(gPi-tau*gPiTau, 2) / (tau * tau * gTauTau)) - gPiPi)))

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
