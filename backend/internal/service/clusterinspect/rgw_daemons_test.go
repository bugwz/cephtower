package clusterinspect

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRGWDaemonServiceMap(t *testing.T) {
	body := `{"services":{"rgw":{"daemons":{"summary":"2 daemons","91":{"metadata":{"id":"a","hostname":"host","ceph_version":"ceph version 20","realm_name":"r","zonegroup_name":"g","zonegroup_id":"gid","zone_name":"z","frontend_config#0":"private-secret"}},"12":{"metadata":{"id":"a"}}}},"other":null}}`
	rows, err := decodeRGWDaemons([]byte(body))
	if err != nil || len(rows) != 2 || rows[0].ServiceMapID != "12" || rows[1].ServiceMapID != "91" || rows[1].Hostname != "host" || rows[1].ZonegroupID != "gid" {
		t.Fatal("invalid projection", err)
	}
	encoded, _ := json.Marshal(rows)
	if strings.Contains(string(encoded), "private-secret") {
		t.Fatal("metadata leaked")
	}
	for _, body := range []string{`{"services":{}}`, `{"services":{"rgw":{"daemons":{"summary":""}}}}`} {
		rows, err := decodeRGWDaemons([]byte(body))
		if err != nil || rows == nil || len(rows) != 0 {
			t.Fatal("empty map rejected")
		}
	}
	for _, body := range []string{`null`, `{}`, `{"services":null}`, `{"services":{"rgw":null}}`, `{"services":{"rgw":{"daemons":{"1":{"metadata":{}}}}}}`, `{"services":{"rgw":{"daemons":{"1":{"metadata":{"id":1}}}}}}`, `{"services":{}} {}`} {
		if _, err := decodeRGWDaemons([]byte(body)); err == nil {
			t.Fatal("invalid map accepted")
		}
	}
	s, runner, id := testInspection(t)
	for _, mode := range []string{"success", "error", "exit", "cancel"} {
		ctx, cancel := context.WithCancel(context.Background())
		output := []byte(body)
		diagnostic := []byte("private-diagnostic")
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if spec.Binary != executor.BinaryCeph || spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"service", "dump", "--format", "json"}) {
				t.Fatal("incorrect command")
			}
			result := executor.CommandResult{Stdout: output, Stderr: diagnostic}
			if mode == "error" {
				return result, errors.New("private-error")
			}
			if mode == "exit" {
				result.ExitCode = 1
			}
			if mode == "cancel" {
				cancel()
			}
			return result, nil
		}
		_, err := s.RGWDaemons(ctx, id)
		cancel()
		if (err == nil) != (mode == "success") || err != nil && strings.Contains(err.Error(), "private-") {
			t.Fatal("unsafe result", err)
		}
		for _, buffer := range [][]byte{output, diagnostic} {
			for _, b := range buffer {
				if b != 0 {
					t.Fatal("output not cleared")
				}
			}
		}
	}
}

func TestRGWDaemonMultisiteIdentity(t *testing.T) {
	rows, err := decodeRGWDaemons([]byte(`{"services":{"rgw":{"daemons":{"1":{"metadata":{"id":"a","realm_name":"same","realm_id":"realm-a","zone_name":"same","zone_id":"zone-a"}},"2":{"metadata":{"id":"b","realm_name":"same","realm_id":"realm-b","zone_name":"same","zone_id":"zone-b"}},"3":{"metadata":{"id":"c"}}}}}}`))
	if err != nil || len(rows) != 3 {
		t.Fatal("invalid registrations", err)
	}
	if rows[0].RealmID != "realm-a" || rows[1].RealmID != "realm-b" || rows[0].ZoneID != "zone-a" || rows[1].ZoneID != "zone-b" || rows[2].RealmID != "" || rows[2].ZoneID != "" {
		t.Fatal("multisite identity lost or inferred")
	}
	encoded, _ := json.Marshal(rows[0])
	if !strings.Contains(string(encoded), `"realm_id":"realm-a"`) || !strings.Contains(string(encoded), `"zone_id":"zone-a"`) {
		t.Fatal("identity not exposed")
	}
}

func TestRGWDaemonSystemMetadata(t *testing.T) {
	rows, err := decodeRGWDaemons([]byte(`{"services":{"rgw":{"daemons":{"1":{"metadata":{"id":"a","os":"Linux","kernel_version":"6.1","arch":"aarch64","pod_name":"rgw-pod","num_handles":"9007199254740993","password":"private","frontend_config#0":"private","unknown":"private"}},"2":{"metadata":{"id":"b"}}}}}}`))
	if err != nil || len(rows) != 2 {
		t.Fatal(err)
	}
	want := map[string]string{"os": "Linux", "kernel_version": "6.1", "arch": "aarch64", "pod_name": "rgw-pod", "num_handles": "9007199254740993"}
	if !reflect.DeepEqual(rows[0].Metadata, want) || rows[1].Metadata == nil || len(rows[1].Metadata) != 0 {
		t.Fatal("metadata projection failed")
	}
}
