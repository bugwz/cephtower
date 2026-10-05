package ceph

import "testing"

func TestMonitorMetricTypeMatchesReferenceMask(t *testing.T) {
	for _, tc := range []struct {
		kind int
		want string
	}{
		{0, "gauge"}, {1, "gauge"}, {2, "gauge"}, {3, "gauge"},
		{4, "counter"}, {5, "counter"}, {6, "counter"}, {8, "counter"}, {9, "counter"}, {10, "counter"},
		{16, "histogram"}, {17, "histogram"}, {18, "histogram"},
		{12, ""}, {20, ""}, {24, ""}, {32, ""}, {-1, ""},
	} {
		if got := dashboardMetricType(tc.kind); got != tc.want {
			t.Fatalf("type=%d got=%s want=%s", tc.kind, got, tc.want)
		}
		if tc.want != "counter" && dashboardPerfUnit("bytes", tc.want) != "" {
			t.Fatal("non-counter labeled as rate")
		}
	}
}
