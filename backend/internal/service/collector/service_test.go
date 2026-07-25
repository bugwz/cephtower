package collector

import (
	"testing"
	"time"
)

func TestServiceDeduplicatesClusterModuleKey(t *testing.T) {
	service := NewService(nil)
	if !service.startRun("7:health") {
		t.Fatal("first run was rejected")
	}
	if service.startRun("7:health") {
		t.Fatal("duplicate cluster/module run was accepted")
	}
	service.finishRun("7:health")
	if !service.startRun("7:health") {
		t.Fatal("released cluster/module key was not reusable")
	}
}

func TestDeterministicJitterIsStableAndBounded(t *testing.T) {
	first := deterministicJitter(7, "health", 30)
	second := deterministicJitter(7, "health", 30)
	if first != second {
		t.Fatalf("jitter changed: %v != %v", first, second)
	}
	if first < 0 || first > 30*time.Second {
		t.Fatalf("jitter = %v, want [0,30s]", first)
	}
}
