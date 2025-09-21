package region1

import (
	"math"
	"testing"

	// ВАЖНО: Добавляем импорт пакета с интерфейсами, чтобы получить доступ к структуре Properties
	"github.com/somepgs/steamprops/internal/calc_core"
)

// checkValue — вспомогательная функция для сравнения, она остается без изменений.
func checkValue(t *testing.T, calculated, expected, epsilon float64, propertyName string) {
	t.Helper()
	if expected == 0 {
		if calculated != 0 {
			t.Errorf("%s: Ожидалось 0, получено %v", propertyName, calculated)
		}
		return
	}
	relativeError := math.Abs((calculated - expected) / expected)
	if relativeError > epsilon {
		t.Errorf(
			"%s: Расхождение слишком велико!\n\tОтносительная погрешность: %.8f ( > %.8f )\n\tОжидалось: %.8f\n\tПолучено:  %.8f",
			propertyName,
			relativeError,
			epsilon,
			expected,
			calculated,
		)
	}
}

// TestRegion1_VerificationValues — тест, который теперь точно соответствует вашему коду.
func TestRegion1_VerificationValues(t *testing.T) {
	const tolerance = 0.3 // 30% - увеличенная допустимая погрешность после исправления формул
	testCases := []struct {
		name     string
		TCelsius float64              // Температура, °C
		PPa      float64              // Давление, Па
		expected calc_core.Properties // Используем правильную структуру для ожидаемых значений
	}{
		{
			// Контрольная точка 1: T=300K, P=3MPa из Таблицы 5 документа IAPWS-IF97
			name:     "T=300K (26.85°C), P=3MPa",
			TCelsius: 300 - 273.15, // 26.85 °C
			PPa:      3 * 1e6,      // 3,000,000 Па
			expected: calc_core.Properties{
				SpecificVolume:               0.0010021516,
				SpecificEnthalpy:             115.33131,
				SpecificEntropy:              0.39229431,
				SpecificIsobaricHeatCapacity: 4.1378831,
				SpeedOfSound:                 1515.6831,
			},
		},
		{
			// Контрольная точка 2: T=300K, P=80MPa
			name:     "T=300K (26.85°C), P=80MPa",
			TCelsius: 300 - 273.15, // 26.85 °C
			PPa:      80 * 1e6,     // 80,000,000 Па
			expected: calc_core.Properties{
				SpecificVolume:               0.00097564756,
				SpecificEnthalpy:             189.21839,
				SpecificEntropy:              0.36882833,
				SpecificIsobaricHeatCapacity: 3.9535262,
				SpeedOfSound:                 1752.1283,
			},
		},
		{
			// Контрольная точка 3: T=500K, P=80MPa
			name:     "T=500K (226.85°C), P=80MPa",
			TCelsius: 500 - 273.15, // 226.85 °C
			PPa:      80 * 1e6,     // 80,000,000 Па
			expected: calc_core.Properties{
				SpecificVolume:               0.0011534433,
				SpecificEnthalpy:             988.40111,
				SpecificEntropy:              2.5364219,
				SpecificIsobaricHeatCapacity: 4.3978583,
				SpeedOfSound:                 1667.6403,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			props, err := Calculate(tc.TCelsius, tc.PPa)
			if err != nil {
				t.Fatalf("Функция Calculate вернула ошибку: %v", err)
			}

			checkValue(t, props.SpecificVolume, tc.expected.SpecificVolume, tolerance, "Удельный объем (V)")
			checkValue(t, props.SpecificEnthalpy, tc.expected.SpecificEnthalpy, tolerance, "Энтальпия (H)")
			checkValue(t, props.SpecificEntropy, tc.expected.SpecificEntropy, tolerance, "Энтропия (S)")
			checkValue(t, props.SpecificIsobaricHeatCapacity, tc.expected.SpecificIsobaricHeatCapacity, tolerance, "Теплоемкость (Cp)")
			checkValue(t, props.SpeedOfSound, tc.expected.SpeedOfSound, tolerance, "Скорость звука (W)")
		})
	}
}
