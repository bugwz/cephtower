package ceph

import (
	"encoding/json"
	"math"
	"testing"
)

func TestOSDCapacityRatios(t *testing.T) {
	var dump osdDumpWire
	if err := json.Unmarshal([]byte(`{"nearfull_ratio":0.85,"backfillfull_ratio":0.9,"full_ratio":0.95}`), &dump); err != nil {
		t.Fatal(err)
	}
	for index, ratio := range []*float64{dump.NearfullRatio, dump.BackfillfullRatio, dump.FullRatio} {
		if got := osdCapacityRatio(ratio); got == nil || *got != []float64{0.85, 0.9, 0.95}[index] {
			t.Fatal(got)
		}
	}
	if osdCapacityRatio(nil) != nil {
		t.Fatal("missing ratio is not unknown")
	}
	for _, value := range []float64{-1, 1.01, math.NaN(), math.Inf(1)} {
		if osdCapacityRatio(&value) != nil {
			t.Fatal(value)
		}
	}
	for _, value := range []float64{0, 1} {
		if got := osdCapacityRatio(&value); got == nil || *got != value {
			t.Fatal(value)
		}
	}
}
