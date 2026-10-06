package ceph

import "math"

func osdCapacityRatio(value *float64) *float64 {
	if value == nil || math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 || *value > 1 {
		return nil
	}
	return value
}
