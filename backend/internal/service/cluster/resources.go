package cluster

import (
	"context"

	"cephtower/backend/internal/store"
)

func DeleteCephClusterResources(ctx context.Context, db *store.Database, clusterID uint) error {
	return db.DeleteClusterResources(ctx, clusterID)
}
