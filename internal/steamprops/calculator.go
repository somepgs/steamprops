package steamprops

import (
	"fmt"
	"math"

	"github.com/somepgs/steamprops/internal/calc_core"
	"github.com/somepgs/steamprops/internal/calc_core/region1"
	"github.com/somepgs/steamprops/internal/calc_core/region2"
	"github.com/somepgs/steamprops/internal/calc_core/region3"
	"github.com/somepgs/steamprops/internal/calc_core/region4"
	"github.com/somepgs/steamprops/internal/calc_core/region5"
	"github.com/somepgs/steamprops/internal/calc_core/transport"
)

// Calculator представляет основной калькулятор SteamProps
type Calculator struct {
	// Используем функции напрямую
}

// NewCalculator создает новый калькулятор
func NewCalculator() *Calculator {
	return &Calculator{}
}

// InputData представляет входные данные для расчета
type InputData struct {
	Mode        string  // "TP" или "HS"
	Temperature float64 // °C
	Pressure    float64 // Pa
	Enthalpy    float64 // кДж/кг
	Entropy     float64 // кДж/(кг·К)
}

// Validate проверяет корректность входных данных с улучшенной валидацией
func (i *InputData) Validate() error {
	if i.Mode != "TP" && i.Mode != "HS" {
		return fmt.Errorf("неверный режим расчета: %s", i.Mode)
	}

	if i.Mode == "TP" {
		// Проверка на NaN и Inf
		if math.IsNaN(i.Temperature) || math.IsInf(i.Temperature, 0) {
			return fmt.Errorf("температура содержит недопустимое значение: %v", i.Temperature)
		}
		if math.IsNaN(i.Pressure) || math.IsInf(i.Pressure, 0) {
			return fmt.Errorf("давление содержит недопустимое значение: %v", i.Pressure)
		}

		// Проверка физических границ
		if i.Temperature < -273.15 {
			return fmt.Errorf("температура %.2f°C ниже абсолютного нуля", i.Temperature)
		}
		if i.Pressure <= 0 {
			return fmt.Errorf("давление %.0f Па должно быть положительным", i.Pressure)
		}

		// Проверка границ IF-97
		if i.Temperature > 2000 {
			return fmt.Errorf("температура %.2f°C превышает максимальную для IF-97 (2000°C)", i.Temperature)
		}
		if i.Pressure > 100e6 {
			return fmt.Errorf("давление %.0f Па превышает максимальное для IF-97 (100 МПа)", i.Pressure)
		}

		// Проверка минимальных границ IF-97
		if i.Temperature < -0.01 {
			return fmt.Errorf("температура %.2f°C ниже минимальной для IF-97 (-0.01°C)", i.Temperature)
		}
		if i.Pressure < 611.657 {
			return fmt.Errorf("давление %.0f Па ниже минимального для IF-97 (611.657 Па)", i.Pressure)
		}

	} else { // HS режим
		// Проверка на NaN и Inf
		if math.IsNaN(i.Enthalpy) || math.IsInf(i.Enthalpy, 0) {
			return fmt.Errorf("энтальпия содержит недопустимое значение: %v", i.Enthalpy)
		}
		if math.IsNaN(i.Entropy) || math.IsInf(i.Entropy, 0) {
			return fmt.Errorf("энтропия содержит недопустимое значение: %v", i.Entropy)
		}

		// Проверка физических границ
		if i.Enthalpy < 0 {
			return fmt.Errorf("энтальпия %.2f кДж/кг не может быть отрицательной", i.Enthalpy)
		}
		if i.Entropy < 0 {
			return fmt.Errorf("энтропия %.2f кДж/(кг·К) не может быть отрицательной", i.Entropy)
		}

		// Проверка разумных границ для IF-97
		if i.Enthalpy > 5000 {
			return fmt.Errorf("энтальпия %.2f кДж/кг превышает разумный максимум для IF-97", i.Enthalpy)
		}
		if i.Entropy > 15 {
			return fmt.Errorf("энтропия %.2f кДж/(кг·К) превышает разумный максимум для IF-97", i.Entropy)
		}
	}

	return nil
}

// Result представляет результат расчета
type Result struct {
	Properties     calc_core.Properties
	Region         calc_core.Region
	Phase          string
	TransportProps map[string]string
	Temperature    float64 // °C
	Pressure       float64 // Pa
}

// Calculate выполняет расчет свойств
func (c *Calculator) Calculate(inputs *InputData) (*Result, error) {
	var props calc_core.Properties
	var region calc_core.Region
	var err error

	var tKelvin float64
	var temperatureC float64
	var pressurePa float64

	if inputs.Mode == "TP" {
		// Расчет по температуре и давлению
		props, region, err = c.calculateFromTP(inputs.Temperature, inputs.Pressure)
		if err != nil {
			return nil, fmt.Errorf("ошибка расчета по T,P: %w", err)
		}
		temperatureC = inputs.Temperature
		tKelvin = temperatureC + 273.15
		pressurePa = inputs.Pressure
	} else {
		// Расчет по энтальпии и энтропии (Region 3 обратные зависимости)
		pHS, TK, pr, err := region3.PropertiesFromHS(inputs.Enthalpy, inputs.Entropy)
		if err != nil {
			// Пробуем подсказать возможный регион
			guess := c.guessRegionFromHS(inputs.Enthalpy, inputs.Entropy)
			return nil, fmt.Errorf("ошибка расчета по h,s: %w (предполагаемый регион: %d)", err, int(guess))
		}
		props = pr
		region = calc_core.Region3
		tKelvin = TK
		temperatureC = TK - 273.15
		pressurePa = pHS
	}

	// Определяем фазу вещества
	phase := c.determinePhase(props, region)

	// Рассчитываем транспортные свойства при фактической температуре
	transportProps := c.calculateTransportProperties(tKelvin, props)

	return &Result{
		Properties:     props,
		Region:         region,
		Phase:          phase,
		TransportProps: transportProps,
		Temperature:    temperatureC,
		Pressure:       pressurePa,
	}, nil
}

// calculateFromTP рассчитывает свойства по температуре и давлению
func (c *Calculator) calculateFromTP(temperature, pressure float64) (calc_core.Properties, calc_core.Region, error) {
	// Определяем регион с помощью общего определения (ожидает T в K)
	region := calc_core.RegionFromTP(temperature+273.15, pressure)

	var props calc_core.Properties
	var err error

	switch region {
	case calc_core.Region1:
		props, err = region1.Calculate(temperature, pressure)
	case calc_core.Region2:
		props, err = region2.Calculate(temperature, pressure)
	case calc_core.Region3:
		props, err = region3.Calculate(temperature, pressure)
	case calc_core.Region4:
		// Region 4 - линия насыщения: выберем сторону по давлению относительно линии насыщения
		psat, e := region4.SaturationPressure(temperature + 273.15)
		if e == nil && pressure >= psat {
			props, err = region1.Calculate(temperature, pressure)
		} else {
			props, err = region2.Calculate(temperature, pressure)
		}
	case calc_core.Region5:
		props, err = region5.Calculate(temperature, pressure)
	default:
		return calc_core.Properties{}, calc_core.RegionAuto, fmt.Errorf("неопределенный регион")
	}

	if err != nil {
		return calc_core.Properties{}, region, err
	}

	return props, region, nil
}

// calculateFromHS рассчитывает свойства по энтальпии и энтропии
func (c *Calculator) calculateFromHS(enthalpy, entropy float64) (calc_core.Properties, calc_core.Region, error) {
	// Пока используем только Region 3 для HS расчетов
	// В будущем можно расширить для всех регионов
	_, _, props, err := region3.PropertiesFromHS(enthalpy, entropy)
	if err != nil {
		// Пробуем определить, в каком регионе должна быть точка
		region := c.guessRegionFromHS(enthalpy, entropy)
		return calc_core.Properties{}, region, fmt.Errorf("не удалось найти решение для h=%.2f кДж/кг, s=%.2f кДж/(кг·К). Возможно, точка находится в Region %d", enthalpy, entropy, int(region))
	}

	return props, calc_core.Region3, nil
}

// determineRegion определяет регион IF-97 по температуре и давлению
// Использует улучшенную логику определения региона
func (c *Calculator) determineRegion(temperature, pressure float64) calc_core.Region {
	// Используем улучшенную функцию из calc_core
	return calc_core.RegionFromTP(temperature, pressure)
}

// guessRegionFromHS пытается определить регион по энтальпии и энтропии
func (c *Calculator) guessRegionFromHS(enthalpy, entropy float64) calc_core.Region {
	// Критические значения (приблизительные)
	const h_c = 2084.264 // кДж/кг
	const s_c = 4.412    // кДж/(кг·К)

	if enthalpy < h_c && entropy < s_c {
		return calc_core.Region1 // Сжатая жидкость
	} else if enthalpy > h_c || entropy > s_c {
		return calc_core.Region2 // Пар
	} else {
		return calc_core.Region3 // Критическая область
	}
}

// determinePhase определяет фазу вещества
func (c *Calculator) determinePhase(props calc_core.Properties, region calc_core.Region) string {
	switch region {
	case calc_core.Region1:
		return "Сжатая жидкость"
	case calc_core.Region2:
		return "Перегретый пар"
	case calc_core.Region3:
		return "Критическая/сверхкритическая область"
	case calc_core.Region4:
		return "Двухфазная область"
	case calc_core.Region5:
		return "Высокотемпературный газ"
	default:
		return "Неопределенная фаза"
	}
}

// calculateTransportProperties рассчитывает транспортные свойства при заданной температуре (K)
func (c *Calculator) calculateTransportProperties(Tkelvin float64, props calc_core.Properties) map[string]string {
	dynamicViscosity, _ := transport.DynamicViscosity(Tkelvin, props.Density)
	thermalConductivity, _ := transport.ThermalConductivity(Tkelvin, props.Density)
	kinematicViscosity, _ := transport.KinematicViscosity(Tkelvin, props.Density)

	return map[string]string{
		"dynamic_viscosity":    fmt.Sprintf("%.2e Па·с", dynamicViscosity),
		"kinematic_viscosity":  fmt.Sprintf("%.2e м²/с", kinematicViscosity),
		"thermal_conductivity": fmt.Sprintf("%.3f Вт/(м·К)", thermalConductivity),
	}
}
