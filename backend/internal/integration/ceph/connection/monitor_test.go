package connection

import "testing"

func TestParseMonitorAddresses(t *testing.T) {
	cases := []struct {
		value string
		count int
	}{
		{"mon-a:6789", 1},
		{"v2:192.0.2.1:3300/0", 1},
		{"[v2:192.0.2.1:3300/0,v1:192.0.2.1:6789/0],[v2:[2001:db8::1]:3300/0,v1:[2001:db8::1]:6789/0]", 4},
		{"mon-a:6789 mon-b:6789", 2},
	}
	for _, tc := range cases {
		got, err := ParseMonitorAddresses(tc.value)
		if err != nil || len(got) != tc.count {
			t.Fatalf("ParseMonitorAddresses(%q) = %#v, %v", tc.value, got, err)
		}
	}
}
func TestParseMonitorAddressesRejectsInvalid(t *testing.T) {
	for _, value := range []string{"", "[v2:a:3300", "v3:a:1", "a:0", "[v2:a:3300,]", "2001:db8::1:6789"} {
		if _, err := ParseMonitorAddresses(value); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}
