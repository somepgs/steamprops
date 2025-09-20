package region4

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

//go:embed iapws-if97-region4.csv
var coeff embed.FS

var (
	loaded bool
	n      []float64 // 1-based, n[1..10]
)

func loadOnce() error {
	if loaded {
		return nil
	}
	f, err := coeff.Open("iapws-if97-region4.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	n = make([]float64, 11)
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid region4 line: %s", line)
		}
		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return err
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		n[idx] = val
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	loaded = true
	return nil
}

// SaturationPressure returns saturation pressure (Pa) for given temperature (K)
// Uses IF-97 Region 4 formulation with provided coefficients
func SaturationPressure(T float64) (float64, error) {
	if err := loadOnce(); err != nil {
		return 0, err
	}
	if T <= 0 {
		return 0, errors.New("invalid temperature")
	}
	// Standard IF-97 form:
	// theta = T + n9/(T - n10)
	// A = theta^2 + n1*theta + n2
	// B = n3*theta^2 + n4*theta + n5
	// C = n6*theta^2 + n7*theta + n8
	// p = 1e6 * (2*C / (-B + sqrt(B^2 - 4*A*C)))^4
	theta := T + n[9]/(T-n[10])
	A := theta*theta + n[1]*theta + n[2]
	B := n[3]*theta*theta + n[4]*theta + n[5]
	C := n[6]*theta*theta + n[7]*theta + n[8]
	disc := B*B - 4*A*C
	if disc < 0 {
		return 0, errors.New("region4: negative discriminant")
	}
	sqrtDisc := math.Sqrt(disc)
	x := (2 * C) / (-B + sqrtDisc)
	p := 1e6 * math.Pow(x, 4)
	return p, nil
}

// SaturationTemperature returns saturation temperature (K) for given pressure (Pa)
// Inverts the same equation per IF-97 recommended inversion
func SaturationTemperature(p float64) (float64, error) {
	if err := loadOnce(); err != nil {
		return 0, err
	}
	if p <= 0 {
		return 0, errors.New("invalid pressure")
	}
	// IF-97 inversion:
	// beta = p^(1/4)
	// E = beta^2 + n3*beta + n6
	// F = n1*beta^2 + n4*beta + n7
	// G = n2*beta^2 + n5*beta + n8
	// D = 2*G/(-F - sqrt(F^2 - 4*E*G))
	// T = 0.5*(n10 + D - sqrt((n10 + D)^2 - 4*(n9 + n10*D)))
	beta := math.Pow(p/1e6, 0.25)
	E := beta*beta + n[3]*beta + n[6]
	F := n[1]*beta*beta + n[4]*beta + n[7]
	G := n[2]*beta*beta + n[5]*beta + n[8]
	disc := F*F - 4*E*G
	if disc < 0 {
		return 0, errors.New("region4 inversion: negative discriminant")
	}
	sqrtDisc := math.Sqrt(disc)
	D := (2 * G) / (-F - sqrtDisc)
	y := n[10] + D
	inner := y*y - 4*(n[9]+n[10]*D)
	if inner < 0 {
		return 0, errors.New("region4 inversion: negative inner discriminant")
	}
	T := 0.5 * (y - math.Sqrt(inner))
	return T, nil
}
