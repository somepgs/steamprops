package region3

import (
	"testing"
)

func TestRegion3TPBasic(t *testing.T) {
	// Representative Region 3 point: T=650 K, p=25 MPa
	T := 650.0 - 273.15 // in Celsius for API
	p := 25e6           // Pa
	props, err := Calculate(T, p)
	if err != nil {
		t.Fatalf("Region3.Calculate error: %v", err)
	}
	if !(props.SpecificVolume > 0 && props.Density > 0 && props.SpeedOfSound > 0) {
		t.Fatalf("non-positive physical property: v=%g, rho=%g, w=%g", props.SpecificVolume, props.Density, props.SpeedOfSound)
	}
	if !(props.SpecificIsobaricHeatCapacity > 0 && props.SpecificIsochoricHeatCapacity > 0) {
		t.Fatalf("heat capacities should be positive: cp=%g, cv=%g", props.SpecificIsobaricHeatCapacity, props.SpecificIsochoricHeatCapacity)
	}
}
