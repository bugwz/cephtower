package cluster

import (
	"context"
	"sync"
	"testing"
	"time"

	cephprovider "cephtower/backend/internal/integration/ceph"
)

type cancellationProbe struct {
	started  chan struct{}
	canceled chan struct{}
	release  chan struct{}
}

func (p cancellationProbe) Probe(ctx context.Context, _ cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	close(p.started)
	<-ctx.Done()
	close(p.canceled)
	<-p.release
	return cephprovider.ProbeResult{}, ctx.Err()
}

func TestStopCancelsAndDrainsScheduledProbes(t *testing.T) {
	p := cancellationProbe{make(chan struct{}), make(chan struct{}), make(chan struct{})}
	s, _, row := clusterTestServices(t, p)
	// Release before the service cleanup even if an assertion fails.
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(p.release) }) }
	t.Cleanup(release)
	s.scheduleProbe(row)
	select {
	case <-p.started:
	case <-time.After(time.Second):
		t.Fatal("probe did not start")
	}
	done := make(chan struct{})
	go func() { s.Stop(); close(done) }()
	select {
	case <-p.canceled:
	case <-time.After(time.Second):
		t.Fatal("stop did not cancel probe")
	}
	select {
	case <-done:
		t.Fatal("stop returned before probe exited")
	default:
	}
	// A stopped service must not admit another background worker.
	s.scheduleProbe(row)
	release()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stop did not drain probe")
	}
	s.Stop()
}
