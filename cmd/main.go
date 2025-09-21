package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/somepgs/steamprops/internal/calc_core"
	"github.com/somepgs/steamprops/internal/calc_core/region1"
	"github.com/somepgs/steamprops/internal/calc_core/region2"
	"github.com/somepgs/steamprops/internal/calc_core/region3"
	"github.com/somepgs/steamprops/internal/calc_core/region5"
)

func main() {
	mode := flag.String("mode", "tp", "Режим: tp (по T и p) или hs (по h и s → p)")
	tC := flag.Float64("t", 200.0, "Температура, ℃")
	pPa := flag.Float64("p", 40_000_000.0, "Давление, Па")
	h := flag.Float64("h", 2000.0, "Энтальпия, кДж/кг (для режима hs)")
	s := flag.Float64("s", 5.0, "Энтропия, кДж/(кг*К) (для режима hs)")
	region := flag.String("region", "auto", "Регион IF-97: auto, 1, 2, 3, 5")
	flag.Parse()

	switch strings.ToLower(*mode) {
	case "hs":
		p, err := region3.PressureFromHS(*h, *s)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Давление по (h,s): %.6f Па\n", p)
		return
	case "tp":
		// fallthrough to existing tp flow
	default:
		if *mode != "tp" {
			log.Fatal("некорректный режим --mode: ожидается tp или hs")
		}
	}

	var props calc_core.Properties
	var err error

	switch strings.ToLower(*region) {
	case "1":
		props, err = region1.Calculate(*tC, *pPa)
	case "2":
		props, err = region2.Calculate(*tC, *pPa)
	case "3":
		props, err = region3.Calculate(*tC, *pPa)
	case "5":
		props, err = region5.Calculate(*tC, *pPa)
	case "auto":
		T := *tC + 273.15
		reg := calc_core.RegionFromTP(T, *pPa)
		switch reg {
		case calc_core.Region1:
			props, err = region1.Calculate(*tC, *pPa)
		case calc_core.Region2:
			props, err = region2.Calculate(*tC, *pPa)
		case calc_core.Region5:
			props, err = region5.Calculate(*tC, *pPa)
		case calc_core.Region3:
			props, err = region3.Calculate(*tC, *pPa)
		default:
			props, err = region2.Calculate(*tC, *pPa)
		}
	default:
		log.Fatal("некорректное значение --region: ожидается auto, 1, 2, 3 или 5")
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Удельный объем: %.12f м3/кг\n", props.SpecificVolume)
	fmt.Printf("Плотность: %.12f кг/м3\n", props.Density)
	fmt.Printf("Удельная внутренняя энергия: %.12f кДж/кг\n", props.SpecificInternalEnergy)
	fmt.Printf("Удельная энтропия: %.12f кДж/кг*К\n", props.SpecificEntropy)
	fmt.Printf("Удельная энтальпия: %.12f кДж/кг\n", props.SpecificEnthalpy)
	fmt.Printf("Удельная изохорная теплоемкость: %.12f кДж/кг*К\n", props.SpecificIsochoricHeatCapacity)
	fmt.Printf("Удельная изобарная теплоемкость: %.12f кДж/кг*К\n", props.SpecificIsobaricHeatCapacity)
	fmt.Printf("Скорость звука: %.12f м/с\n", props.SpeedOfSound)
}
