package app

import "context"

type App struct {
	LogRetentionCleanup func(context.Context)
}

var Global App
