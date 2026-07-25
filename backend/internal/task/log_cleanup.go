package task

import "context"

func LogCleanupJob(cleanup func(context.Context) error) Job {
	return func(ctx context.Context) error { return cleanup(ctx) }
}
