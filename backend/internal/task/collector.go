package task

import (
	"context"

	"cephtower/backend/internal/service/collector"
)

func CollectorJob(service *collector.Service) Job {
	return func(ctx context.Context) error { return service.RunDue(ctx) }
}
