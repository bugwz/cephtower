package clusterinspect

import (
	"reflect"
	"testing"
)

func TestRGWListeners(t *testing.T) {
	for _, test := range []struct {
		config   string
		ports    []uint16
		tls      []bool
		complete bool
	}{
		{"beast port=8080 ssl_port=8443 ssl_certificate=private", []uint16{8080, 8443}, []bool{false, true}, true},
		{"beast endpoint=192.0.2.1:8000 ssl_endpoint=[::1]:9443", []uint16{8000, 9443}, []bool{false, true}, true},
		{"beast endpoint=192.0.2.1 ssl_endpoint=[::1]", []uint16{80, 443}, []bool{false, true}, true},
		{"beast port=8080 port=8081", []uint16{8080, 8081}, []bool{false, false}, true},
		{"beast port=0", nil, nil, false}, {"beast port=65536", nil, nil, false},
		{"beast endpoint=host:80", nil, nil, false}, {"beast ssl_endpoint=::1", nil, nil, false},
		{"beast endpoint=[192.0.2.1]:80", nil, nil, false},
		{"beast endpoint=[192.0.2.1]", nil, nil, false},
		{"beast endpoint=[192.0.2.1", nil, nil, false},
		{"beast endpoint=192.0.2.1]", nil, nil, false},
		{"beast endpoint=[::1]garbage", nil, nil, false},
		{"beast endpoint=[::ffff:192.0.2.1]:8080", []uint16{8080}, []bool{false}, true},
		{"beast port=80garbage", nil, nil, false}, {"civetweb port=80", nil, nil, false},
		{"beast ssl_certificate=private", nil, nil, false}, {"beast port=8080 ssl_port=invalid", []uint16{8080}, []bool{false}, false},
	} {
		rows, complete := rgwListeners(map[string]string{"frontend_config#0": test.config})
		var ports []uint16
		var tls []bool
		for _, row := range rows {
			ports = append(ports, row.Port)
			tls = append(tls, row.TLS)
			if row.Frontend != "frontend_config#0" {
				t.Fatal("frontend identity lost")
			}
		}
		if complete != test.complete || !reflect.DeepEqual(ports, test.ports) || !reflect.DeepEqual(tls, test.tls) {
			t.Fatalf("unexpected projection for %q", test.config)
		}
	}
	rows, complete := rgwListeners(nil)
	if rows == nil || len(rows) != 0 || complete {
		t.Fatal("missing configuration not unknown")
	}
	rows, complete = rgwListeners(map[string]string{"frontend_config#1": "beast ssl_port=443", "frontend_config#0": "beast port=80"})
	if !complete || len(rows) != 2 || rows[0].Port != 80 {
		t.Fatal("multiple frontends lost")
	}
}
