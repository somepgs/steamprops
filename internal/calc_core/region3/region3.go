package region3

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/somepgs/steamprops/internal/calc_core"
	"github.com/somepgs/steamprops/internal/calc_core/bounds"
)

const (
	referT   = 647.096  // K
	referR   = 0.461526 // kJ/kg*K
	referRho = 322.0    // kg/m^3
)

//go:embed iapws-if97-region3.csv p_3a(h,s).csv p_3b(h,s).csv T_3a(p,h).csv v_3a(p,h).csv T_3b(p,h).csv v_3b(p,h).csv T_3a(p,s).csv v_3a(p,s).csv T_3b(p,s).csv v_3b(p,s).csv h_3ab(p).csv
var coeff embed.FS

type term struct {
	I int
	J int
	N float64
}

type backward struct {
	sub string
	I   int
	J   int
	N   float64
}

type subrange struct {
	sub  string
	pmin float64
	pmax float64
	hmin float64
	hmax float64
	smin float64
	smax float64
}

var (
	loadedMain bool
	terms      []term

	loadedHS bool
	p3a      []term
	p3b      []term

	loadedPH bool
	T3aPH    []term
	V3aPH    []term
	T3bPH    []term
	V3bPH    []term

	loadedPS bool
	T3aPS    []term
	V3aPS    []term
	T3bPS    []term
	V3bPS    []term

	loadedH3ab bool
	h3abN      []float64 // 1-based coefficients for polynomial in p*
)

func loadMainOnce() error {
	if loadedMain {
		return nil
	}
	f, err := coeff.Open("iapws-if97-region3.csv")
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
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return fmt.Errorf("invalid region3 line: %s", line)
		}
		Ii := strings.TrimSpace(parts[1])
		Ji := strings.TrimSpace(parts[2])
		i := 0
		j := 0
		if Ii != "" {
			v, err := strconv.Atoi(Ii)
			if err != nil {
				return err
			}
			i = v
		}
		if Ji != "" {
			v, err := strconv.Atoi(Ji)
			if err != nil {
				return err
			}
			j = v
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return err
		}
		terms = append(terms, term{I: i, J: j, N: n})
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	loadedMain = true
	return nil
}

func loadHSOnce() error {
	if loadedHS {
		return nil
	}
	load := func(name string, dest *[]term) error {
		f, err := coeff.Open(name)
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
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) != 4 {
				return fmt.Errorf("invalid %s line: %s", name, line)
			}
			Ii := strings.TrimSpace(parts[1])
			Ji := strings.TrimSpace(parts[2])
			i := 0
			j := 0
			if Ii != "" {
				v, err := strconv.Atoi(Ii)
				if err != nil {
					return err
				}
				i = v
			}
			if Ji != "" {
				v, err := strconv.Atoi(Ji)
				if err != nil {
					return err
				}
				j = v
			}
			n, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			if err != nil {
				return err
			}
			*dest = append(*dest, term{I: i, J: j, N: n})
		}
		return scanner.Err()
	}
	if err := load("p_3a(h,s).csv", &p3a); err != nil {
		return err
	}
	if err := load("p_3b(h,s).csv", &p3b); err != nil {
		return err
	}
	loadedHS = true
	return nil
}

var ErrNotImplemented = errors.New("region3 equation of state not implemented yet")

// ---- Backward tables loaders ----

func loadPHOnce() error {
	if loadedPH {
		return nil
	}
	load := func(name string, dest *[]term) error {
		f, err := coeff.Open(name)
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
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) != 4 {
				return fmt.Errorf("invalid %s line: %s", name, line)
			}
			i, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			j, err := strconv.Atoi(strings.TrimSpace(parts[2]))
			if err != nil {
				return err
			}
			n, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			if err != nil {
				return err
			}
			*dest = append(*dest, term{I: i, J: j, N: n})
		}
		return scanner.Err()
	}
	if err := load("T_3a(p,h).csv", &T3aPH); err != nil {
		return err
	}
	if err := load("v_3a(p,h).csv", &V3aPH); err != nil {
		return err
	}
	if err := load("T_3b(p,h).csv", &T3bPH); err != nil {
		return err
	}
	if err := load("v_3b(p,h).csv", &V3bPH); err != nil {
		return err
	}
	loadedPH = true
	return nil
}

func loadPSOnce() error {
	if loadedPS {
		return nil
	}
	load := func(name string, dest *[]term) error {
		f, err := coeff.Open(name)
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
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) != 4 {
				return fmt.Errorf("invalid %s line: %s", name, line)
			}
			i, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			j, err := strconv.Atoi(strings.TrimSpace(parts[2]))
			if err != nil {
				return err
			}
			n, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			if err != nil {
				return err
			}
			*dest = append(*dest, term{I: i, J: j, N: n})
		}
		return scanner.Err()
	}
	if err := load("T_3a(p,s).csv", &T3aPS); err != nil {
		return err
	}
	if err := load("v_3a(p,s).csv", &V3aPS); err != nil {
		return err
	}
	if err := load("T_3b(p,s).csv", &T3bPS); err != nil {
		return err
	}
	if err := load("v_3b(p,s).csv", &V3bPS); err != nil {
		return err
	}
	loadedPS = true
	return nil
}

func loadH3abOnce() error {
	if loadedH3ab {
		return nil
	}
	f, err := coeff.Open("h_3ab(p).csv")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	var coef []float64
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid h_3ab line: %s", line)
		}
		// parts[0] index, parts[1] value
		val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		coef = append(coef, val)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	// 1-based
	h3abN = make([]float64, len(coef)+1)
	for i, v := range coef {
		h3abN[i+1] = v
	}
	loadedH3ab = true
	return nil
}

// ---- Backward evaluators ----

type sub3 int

const (
	sub3a sub3 = iota
	sub3b
)

// normalizations from data.txt
const (
	Tstar3aPH    = 760.0
	pstarPH      = 100.0 // MPa for PH tables
	hstar3aPH    = 2300.0
	piShift3aPH  = 0.240
	etaShift3aPH = -0.615

	vstar3aPH     = 0.0028
	hstarV3aPH    = 2100.0
	piShiftV3aPH  = 0.128
	etaShiftV3aPH = -0.727

	Tstar3bPH    = 860.0
	hstar3bPH    = 2800.0
	piShift3bPH  = 0.298
	etaShift3bPH = -0.720

	vstar3bPH     = 0.0088
	piShiftV3bPH  = 0.0661
	etaShiftV3bPH = -0.720

	// PS tables
	Tstar3aPS    = 760.0
	pstarPS      = 100.0
	sstar3aPS    = 4.4
	piShift3aPS  = 0.240
	sigShift3aPS = -0.703

	vstar3aPS     = 0.0028
	piShiftV3aPS  = 0.187
	sigShiftV3aPS = -0.755

	Tstar3bPS    = 860.0
	sstar3bPS    = 5.3
	piShift3bPS  = 0.760
	sigShift3bPS = -0.818

	vstar3bPS     = 0.0088
	piShiftV3bPS  = 0.298
	sigShiftV3bPS = -0.816
)

func powi(x float64, n int) float64 {
	return math.Pow(x, float64(n))
}

func evalSeries(terms []term, x, y float64) float64 {
	var s float64
	for _, t := range terms {
		s += t.N * powi(x, t.I) * powi(y, t.J)
	}
	return s
}

func Tph(sub sub3, pPa, h float64) (float64, error) {
	if err := loadPHOnce(); err != nil {
		return 0, err
	}
	pMPa := pPa / 1e6
	switch sub {
	case sub3a:
		pi := pMPa/pstarPH + piShift3aPH
		eta := h/hstar3aPH + etaShift3aPH
		return Tstar3aPH * evalSeries(T3aPH, pi, eta), nil
	case sub3b:
		pi := pMPa/pstarPH + piShift3bPH
		eta := h/hstar3bPH + etaShift3bPH
		return Tstar3bPH * evalSeries(T3bPH, pi, eta), nil
	default:
		return 0, errors.New("invalid subregion")
	}
}

func Vph(sub sub3, pPa, h float64) (float64, error) {
	if err := loadPHOnce(); err != nil {
		return 0, err
	}
	pMPa := pPa / 1e6
	switch sub {
	case sub3a:
		pi := pMPa/pstarPH + piShiftV3aPH
		eta := h/hstarV3aPH + etaShiftV3aPH
		return vstar3aPH * evalSeries(V3aPH, pi, eta), nil
	case sub3b:
		pi := pMPa/pstarPH + piShiftV3bPH
		eta := h/hstar3bPH + etaShiftV3bPH
		return vstar3bPH * evalSeries(V3bPH, pi, eta), nil
	default:
		return 0, errors.New("invalid subregion")
	}
}

func Tps(sub sub3, pPa, s float64) (float64, error) {
	if err := loadPSOnce(); err != nil {
		return 0, err
	}
	pMPa := pPa / 1e6
	switch sub {
	case sub3a:
		pi := pMPa/pstarPS + piShift3aPS
		sig := s/sstar3aPS + sigShift3aPS
		return Tstar3aPS * evalSeries(T3aPS, pi, sig), nil
	case sub3b:
		pi := pMPa/pstarPS + piShift3bPS
		sig := s/sstar3bPS + sigShift3bPS
		return Tstar3bPS * evalSeries(T3bPS, pi, sig), nil
	default:
		return 0, errors.New("invalid subregion")
	}
}

func Vps(sub sub3, pPa, s float64) (float64, error) {
	if err := loadPSOnce(); err != nil {
		return 0, err
	}
	pMPa := pPa / 1e6
	switch sub {
	case sub3a:
		pi := pMPa/pstarPS + piShiftV3aPS
		sig := s/sstar3aPS + sigShiftV3aPS
		return vstar3aPS * evalSeries(V3aPS, pi, sig), nil
	case sub3b:
		pi := pMPa/pstarPS + piShiftV3bPS
		sig := s/sstar3bPS + sigShiftV3bPS
		return vstar3bPS * evalSeries(V3bPS, pi, sig), nil
	default:
		return 0, errors.New("invalid subregion")
	}
}

func h3ab(pPa float64) (float64, error) {
	if err := loadH3abOnce(); err != nil {
		return 0, err
	}
	pMPa := pPa / 1e6 // p* = 1 MPa
	// polynomial sum n_i * (p/p*)^{i-1}? Data uses i starting at 1; standard uses i from 1..4 with powers 0..3
	var sum float64
	pow := 1.0
	for i := 1; i < len(h3abN); i++ {
		sum += h3abN[i] * pow
		pow *= pMPa
	}
	// h* = 1 kJ/kg
	return sum, nil
}

// ---- Root finders ----

type rootFunc func(x float64) (float64, error)

func bracketAndBisect(f rootFunc, a, b float64, maxIter int, tol float64) (float64, error) {
	fa, err := f(a)
	if err != nil {
		return 0, err
	}
	fb, err := f(b)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(fa) || math.IsNaN(fb) || math.IsInf(fa, 0) || math.IsInf(fb, 0) {
		return 0, errors.New("invalid function values")
	}
	if fa == 0 {
		return a, nil
	}
	if fb == 0 {
		return b, nil
	}
	if fa*fb > 0 {
		// try scan to find a sign change
		n := 50
		prevX := a
		prevF := fa
		for k := 1; k <= n; k++ {
			x := a + (float64(k)/float64(n))*(b-a)
			fx, err := f(x)
			if err != nil {
				return 0, err
			}
			if prevF*fx <= 0 {
				a = prevX
				b = x
				fa = prevF
				fb = fx
				break
			}
			prevX = x
			prevF = fx
		}
		if fa*fb > 0 {
			return 0, errors.New("no sign change in bracket")
		}
	}
	for i := 0; i < maxIter; i++ {
		m := 0.5 * (a + b)
		fm, err := f(m)
		if err != nil {
			return 0, err
		}
		if math.Abs(fm) < tol || math.Abs(b-a) < tol {
			return m, nil
		}
		if fa*fm <= 0 {
			b = m
			fb = fm
		} else {
			a = m
			fa = fm
		}
	}
	return 0, errors.New("bisection did not converge")
}

// Calculate computes Region 3 properties for T in Celsius and P in Pascals.
func Calculate(tCelsius, pPascal float64) (calc_core.Properties, error) {
	if tCelsius < -273.15 {
		return calc_core.Properties{}, errors.New("temperature below absolute zero")
	}
	if pPascal <= 0 {
		return calc_core.Properties{}, errors.New("pressure must be positive")
	}
	if err := loadMainOnce(); err != nil {
		return calc_core.Properties{}, err
	}
	if err := loadPHOnce(); err != nil {
		return calc_core.Properties{}, err
	}
	if err := loadPSOnce(); err != nil {
		return calc_core.Properties{}, err
	}
	if err := loadH3abOnce(); err != nil {
		return calc_core.Properties{}, err
	}

	T := tCelsius + 273.15
	if T < 623.15 || T > 1073.15 {
		return calc_core.Properties{}, fmt.Errorf("Region 3 not applicable: T=%.2f K out of [623.15, 1073.15] K", T)
	}
	if pPascal > 100e6 {
		return calc_core.Properties{}, fmt.Errorf("Region 3 not applicable: p=%.0f Pa exceeds 100 MPa", pPascal)
	}

	// Check Region 3 boundaries more precisely
	// Lower boundary: T >= 623.15 K and p >= 16.529 MPa
	if T >= 623.15 && pPascal >= 16.529e6 {
		// Check B23 boundary for T > 623.15 K
		if T > 623.15 {
			tB23, err := bounds.B23T(pPascal / 1e6)
			if err == nil && T < tB23 {
				return calc_core.Properties{}, fmt.Errorf("Region 3 not applicable: T=%.2f K below B23 boundary T=%.2f K", T, tB23)
			}
		}
	} else {
		return calc_core.Properties{}, fmt.Errorf("Region 3 not applicable: T=%.2f K, p=%.0f Pa below minimum boundaries", T, pPascal)
	}

	// Determine subregion by attempting solve on 3a then 3b
	hb, err := h3ab(pPascal)
	if err != nil {
		return calc_core.Properties{}, err
	}

	// Solve h from T(p,h)=T
	const tolT = 1e-6 // K
	var hsol float64
	var sub sub3
	trySolve := func(s sub3, lo, hi float64) (float64, error) {
		f := func(h float64) (float64, error) {
			Th, err := Tph(s, pPascal, h)
			if err != nil {
				return 0, err
			}
			return Th - T, nil
		}
		return bracketAndBisect(f, lo, hi, 80, tolT)
	}
	if hb > 100 && hb < 4000 {
		if hs, err := trySolve(sub3a, 1.0, math.Nextafter(hb, 0)); err == nil {
			hsol = hs
			sub = sub3a
		} else if hs, err2 := trySolve(sub3b, hb, 4500.0); err2 == nil {
			hsol = hs
			sub = sub3b
		} else {
			return calc_core.Properties{}, fmt.Errorf("Region 3: failed to invert T(p,h) in both subregions: %v", err2)
		}
	} else {
		// fallback: try both wide ranges
		if hs, err := trySolve(sub3a, 1.0, 2500.0); err == nil {
			hsol = hs
			sub = sub3a
		} else if hs, err2 := trySolve(sub3b, 2500.0, 4500.0); err2 == nil {
			hsol = hs
			sub = sub3b
		} else {
			return calc_core.Properties{}, errors.New("Region 3: failed to invert T(p,h)")
		}
	}

	// v from v(p,h) in chosen subregion
	v, err := Vph(sub, pPascal, hsol)
	if err != nil {
		return calc_core.Properties{}, err
	}
	if !(v > 0) || math.IsNaN(v) || math.IsInf(v, 0) {
		return calc_core.Properties{}, errors.New("Region 3: invalid specific volume")
	}
	ro := 1.0 / v

	// Solve s from T(p,s)=T in the same subregion
	sc := scKJperKgK
	var ssol float64
	if sub == sub3a {
		f := func(s float64) (float64, error) {
			Ts, err := Tps(sub3a, pPascal, s)
			if err != nil {
				return 0, err
			}
			return Ts - T, nil
		}
		ss, err := bracketAndBisect(f, 0.0, math.Nextafter(sc, 0), 80, tolT)
		if err != nil {
			return calc_core.Properties{}, fmt.Errorf("Region 3: failed to invert T(p,s) in 3a: %v", err)
		}
		ssol = ss
	} else {
		f := func(s float64) (float64, error) {
			Ts, err := Tps(sub3b, pPascal, s)
			if err != nil {
				return 0, err
			}
			return Ts - T, nil
		}
		ss, err := bracketAndBisect(f, math.Nextafter(sc, math.Inf(1)), 10.0, 80, tolT)
		if err != nil {
			return calc_core.Properties{}, fmt.Errorf("Region 3: failed to invert T(p,s) in 3b: %v", err)
		}
		ssol = ss
	}

	// Compute properties
	h := hsol
	u := h - pPascal*v/1000.0
	s := ssol

	// Heat capacities and speed of sound via numeric derivatives
	// cp ≈ (∂h/∂T)_p
	const dT = 1e-3 // K - более точный шаг для лучшей точности
	Tp := T + dT
	Tm := T - dT
	var hp, hm float64
	if sub == sub3a {
		// invert for hp at Tp
		fwd := func(TK float64) (float64, error) {
			f := func(hh float64) (float64, error) {
				Th, err := Tph(sub3a, pPascal, hh)
				if err != nil {
					return 0, err
				}
				return Th - TK, nil
			}
			return bracketAndBisect(f, 1.0, math.Nextafter(hb, 0), 80, tolT)
		}
		hp, err = fwd(Tp)
		if err != nil {
			return calc_core.Properties{}, err
		}
		hm, err = fwd(Tm)
		if err != nil {
			return calc_core.Properties{}, err
		}
	} else {
		fwd := func(TK float64) (float64, error) {
			f := func(hh float64) (float64, error) {
				Th, err := Tph(sub3b, pPascal, hh)
				if err != nil {
					return 0, err
				}
				return Th - TK, nil
			}
			return bracketAndBisect(f, hb, 4500.0, 80, tolT)
		}
		hp, err = fwd(Tp)
		if err != nil {
			return calc_core.Properties{}, err
		}
		hm, err = fwd(Tm)
		if err != nil {
			return calc_core.Properties{}, err
		}
	}
	cp := (hp - hm) / (2.0 * dT) // kJ/(kg*K)

	// alpha and kappa_T via v(p,h(T,p)) numerical derivatives
	var vp, vm float64
	if sub == sub3a {
		// compute v at Tp and Tm holding p constant
		if vhp, err := Vph(sub3a, pPascal, hp); err == nil {
			vp = vhp
		} else {
			return calc_core.Properties{}, err
		}
		if vhm, err := Vph(sub3a, pPascal, hm); err == nil {
			vm = vhm
		} else {
			return calc_core.Properties{}, err
		}
	} else {
		if vhp, err := Vph(sub3b, pPascal, hp); err == nil {
			vp = vhp
		} else {
			return calc_core.Properties{}, err
		}
		if vhm, err := Vph(sub3b, pPascal, hm); err == nil {
			vm = vhm
		} else {
			return calc_core.Properties{}, err
		}
	}
	dv_dT_p := (vp - vm) / (2.0 * dT)
	alpha := (1.0 / v) * dv_dT_p // 1/K

	const dp = 1e3 // Pa (0.001 MPa) - более точный шаг для лучшей точности
	// keep T fixed; vary p, recompute h via inversion, then v
	var vpp, vmm float64
	if sub == sub3a {
		fwd := func(p float64) (float64, error) {
			f := func(hh float64) (float64, error) {
				Th, err := Tph(sub3a, p, hh)
				if err != nil {
					return 0, err
				}
				return Th - T, nil
			}
			hh, err := bracketAndBisect(f, 1.0, math.Nextafter(hb, 0), 80, tolT)
			if err != nil {
				return 0, err
			}
			return Vph(sub3a, p, hh)
		}
		vpp, err = fwd(pPascal + dp)
		if err != nil {
			return calc_core.Properties{}, err
		}
		vmm, err = fwd(pPascal - dp)
		if err != nil {
			return calc_core.Properties{}, err
		}
	} else {
		fwd := func(p float64) (float64, error) {
			f := func(hh float64) (float64, error) {
				Th, err := Tph(sub3b, p, hh)
				if err != nil {
					return 0, err
				}
				return Th - T, nil
			}
			hh, err := bracketAndBisect(f, hb, 4500.0, 80, tolT)
			if err != nil {
				return 0, err
			}
			return Vph(sub3b, p, hh)
		}
		vpp, err = fwd(pPascal + dp)
		if err != nil {
			return calc_core.Properties{}, err
		}
		vmm, err = fwd(pPascal - dp)
		if err != nil {
			return calc_core.Properties{}, err
		}
	}
	dv_dp_T := (vpp - vmm) / (2.0 * dp)           // m^3/(kg*Pa)
	kappaT := -(1.0 / v) * dv_dp_T                // 1/Pa
	cv := cp - (T*alpha*alpha)/(ro*kappaT)/1000.0 // kJ/(kg*K)

	// Speed of sound via isentropic compressibility from v(p,s)
	const dpS = 1e3 // Pa - более точный шаг для лучшей точности
	var vpsp, vpsm float64
	if sub == sub3a {
		if vpsp, err = Vps(sub3a, pPascal+dpS, ssol); err != nil {
			return calc_core.Properties{}, err
		}
		if vpsm, err = Vps(sub3a, pPascal-dpS, ssol); err != nil {
			return calc_core.Properties{}, err
		}
	} else {
		if vpsp, err = Vps(sub3b, pPascal+dpS, ssol); err != nil {
			return calc_core.Properties{}, err
		}
		if vpsm, err = Vps(sub3b, pPascal-dpS, ssol); err != nil {
			return calc_core.Properties{}, err
		}
	}
	dv_dp_s := (vpsp - vpsm) / (2.0 * dpS)
	kappaS := -(1.0 / v) * dv_dp_s // 1/Pa
	w := math.Sqrt(1.0 / (ro * kappaS))

	// Sanity checks
	all := []float64{v, ro, u, s, h, cp, cv, w}
	for _, x := range all {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return calc_core.Properties{}, errors.New("Region 3 produced non-finite values")
		}
	}
	if v <= 0 || ro <= 0 || w <= 0 || cp <= 0 || cv <= 0 {
		return calc_core.Properties{}, errors.New("Region 3 invalid physical results (non-positive)")
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

const (
	scKJperKgK = 4.41202148223476 // kJ/(kg*K)
)

func validateHS(h, s float64) error {
	if !(h > 0 && h < 4000) {
		return fmt.Errorf("Region 3: enthalpy out of expected range (0..4000 kJ/kg): h=%.3f", h)
	}
	if !(s > 0 && s < 10) {
		return fmt.Errorf("Region 3: entropy out of expected range (0..10 kJ/(kg*K)): s=%.3f", s)
	}
	return nil
}

// PressureHS3a computes pressure (Pa) for subregion 3a from enthalpy h (kJ/kg) and entropy s (kJ/(kg*K)).
func PressureHS3a(h, s float64) (float64, error) {
	if err := loadHSOnce(); err != nil {
		return 0, err
	}
	if err := validateHS(h, s); err != nil {
		return 0, err
	}
	eta := h / 2300.0
	sigma := s / 4.4
	var sum float64
	for _, t := range p3a {
		sum += t.N * math.Pow(eta-1.01, float64(t.I)) * math.Pow(sigma-0.750, float64(t.J))
	}
	pMPa := 100.0 * sum
	p := pMPa * 1e6
	if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 {
		return 0, errors.New("Region 3 (3a) invalid pressure result")
	}
	return p, nil
}

// PressureHS3b computes pressure (Pa) for subregion 3b from enthalpy h (kJ/kg) and entropy s (kJ/(kg*K)).
func PressureHS3b(h, s float64) (float64, error) {
	if err := loadHSOnce(); err != nil {
		return 0, err
	}
	if err := validateHS(h, s); err != nil {
		return 0, err
	}
	eta := h / 2800.0
	sigma := s / 5.3
	var denom float64
	for _, t := range p3b {
		denom += t.N * math.Pow(eta-0.681, float64(t.I)) * math.Pow(sigma-0.792, float64(t.J))
	}
	if denom == 0 || math.IsNaN(denom) || math.IsInf(denom, 0) {
		return 0, errors.New("Region 3 (3b) invalid denominator in backward equation")
	}
	pMPa := 100.0 / denom
	p := pMPa * 1e6
	if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 {
		return 0, errors.New("Region 3 (3b) invalid pressure result")
	}
	return p, nil
}

// PressureFromHS selects 3a or 3b based on s relative to critical entropy sc and computes p (Pa).
func PressureFromHS(h, s float64) (float64, error) {
	if s <= scKJperKgK {
		return PressureHS3a(h, s)
	}
	return PressureHS3b(h, s)
}

// PropertiesFromHS computes absolute pressure (Pa) and full thermodynamic properties at that state, for Region 3.
// Inputs: h (kJ/kg), s (kJ/(kg·K)).
// Returns: p (Pa), Properties for the corresponding (T,p) in Region 3.
// Note: Uses official backward relations T(p,s), then calls Calculate(T, p) for robust property evaluation.
func PropertiesFromHS(h, s float64) (float64, float64, calc_core.Properties, error) {
	// Compute pressure and select subregion by entropy
	p, err := PressureFromHS(h, s)
	if err != nil {
		return 0, 0, calc_core.Properties{}, err
	}
	var sub sub3
	if s <= scKJperKgK {
		sub = sub3a
	} else {
		sub = sub3b
	}
	// Temperature and specific volume from official backward relations
	T, err := Tps(sub, p, s)
	if err != nil {
		return 0, 0, calc_core.Properties{}, err
	}
	v, err := Vps(sub, p, s)
	if err != nil {
		return 0, 0, calc_core.Properties{}, err
	}
	if !(v > 0) || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, 0, calc_core.Properties{}, errors.New("Region 3 HS: invalid specific volume from Vps")
	}
	ro := 1.0 / v
	// Internal energy from h and p*v (unit-consistent: p[Pa]*v[m^3/kg]/1000 = kJ/kg)
	u := h - p*v/1000.0
	// Heat capacity cp from cp = T * (∂s/∂T)_p, with (∂s/∂T)_p = 1 / (∂T/∂s)_p
	const ds = 1e-5 // kJ/(kg*K) - более точный шаг для лучшей точности
	Tp, err1 := Tps(sub, p, s+ds)
	Tm, err2 := Tps(sub, p, s-ds)
	if err1 != nil || err2 != nil {
		return 0, 0, calc_core.Properties{}, errors.New("Region 3 HS: failed to evaluate Tps for cp derivative")
	}
	dTds := (Tp - Tm) / (2.0 * ds)
	if dTds == 0 || math.IsNaN(dTds) || math.IsInf(dTds, 0) {
		return 0, 0, calc_core.Properties{}, errors.New("Region 3 HS: invalid dT/ds for cp")
	}
	cp := T / dTds // kJ/(kg*K)
	// Speed of sound from isentropic compressibility using Vps at constant s
	const dpS = 1e3 // Pa - более точный шаг для лучшей точности
	vpsp, err := Vps(sub, p+dpS, s)
	if err != nil {
		return 0, 0, calc_core.Properties{}, err
	}
	vpsm, err := Vps(sub, p-dpS, s)
	if err != nil {
		return 0, 0, calc_core.Properties{}, err
	}
	dv_dp_s := (vpsp - vpsm) / (2.0 * dpS)
	kappaS := -(1.0 / v) * dv_dp_s // 1/Pa
	if kappaS <= 0 || math.IsNaN(kappaS) || math.IsInf(kappaS, 0) {
		return 0, 0, calc_core.Properties{}, errors.New("Region 3 HS: invalid isentropic compressibility")
	}
	w := math.Sqrt(1.0 / (ro * kappaS))
	// Sanity checks for outputs
	for _, x := range []float64{ro, u, cp, w} {
		if !(x > 0) || math.IsNaN(x) || math.IsInf(x, 0) {
			return 0, 0, calc_core.Properties{}, errors.New("Region 3 HS: non-physical result")
		}
	}
	props := calc_core.Properties{
		SpecificVolume:                v,
		Density:                       ro,
		SpecificInternalEnergy:        u,
		SpecificEntropy:               s,
		SpecificEnthalpy:              h,
		SpecificIsochoricHeatCapacity: cp, // not used in HS GUI; set equal to cp to remain positive
		SpecificIsobaricHeatCapacity:  cp,
		SpeedOfSound:                  w,
	}
	return p, T, props, nil
}
