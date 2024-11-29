package properties

import (
	"context"

	"knoway.dev/api/clusters/v1alpha1"
)

var (
	clusterKey = "cluster"
)

func SetClusterToContext(ctx context.Context, cluster *v1alpha1.Cluster) error {
	return SetProperty(ctx, clusterKey, cluster)
}

func GetClusterFromContext(ctx context.Context) (*v1alpha1.Cluster, bool) {
	return GetProperty[*v1alpha1.Cluster](ctx, clusterKey)
}
