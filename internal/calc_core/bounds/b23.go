package bounds

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

//go:embed iapws-if97-region2_3.csv
var coeffB23 embed.FS

var (
	loaded bool
	n      []float64 // 1-based: n[1], n[2], n[3] ...
)

func loadOnce() error {
	if loaded {
		return nil
	}
	f, err := coeffB23.Open("iapws-if97-region2_3.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	// Pre-size to 6 for convenience (indices 1..5 used)
	n = make([]float64, 6)
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid B23 line: %s", line)
		}
		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return err
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return err
		}
		if idx >= len(n) {
			// grow
			nnew := make([]float64, idx+1)
			copy(nnew, n)
			n = nnew
		}
		n[idx] = val
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	loaded = true
	return nil
}

// B23T returns temperature (K) on the B23 boundary for given pressure (MPa)
// Uses quadratic form T = n1 + n2*p + n3*p^2 from IF-97.
func B23T(pMPa float64) (float64, error) {
	if err := loadOnce(); err != nil {
		return 0, err
	}
	// Use first three coefficients
	T := n[1] + n[2]*pMPa + n[3]*pMPa*pMPa
	return T, nil
}

// B23P returns pressure (MPa) on the B23 boundary for given temperature (K)
// Inverts the same quadratic.
func B23P(TK float64) (float64, error) {
	if err := loadOnce(); err != nil {
		return 0, err
	}
	// Solve n3*p^2 + n2*p + (n1 - T) = 0
	A := n[3]
	B := n[2]
	C := n[1] - TK
	if math.Abs(A) < 1e-18 {
		if math.Abs(B) < 1e-18 {
			return 0, errors.New("invalid B23 coefficients: degenerate equation")
		}
		return -C / B, nil
	}
	d := B*B - 4*A*C
	if d < 0 {
		return 0, errors.New("no real solution for B23P")
	}
	sqrtD := math.Sqrt(d)
	// Choose physically meaningful positive root
	p1 := (-B + sqrtD) / (2 * A)
	p2 := (-B - sqrtD) / (2 * A)
	p := p1
	if p < 0 || (p2 > 0 && p2 < p) {
		p = p2
	}
	return p, nil
}
