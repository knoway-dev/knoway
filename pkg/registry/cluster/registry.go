package cluster

import (
	"knoway.dev/api/clusters/v1alpha1"
	clusters2 "knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/manager"
	"sync"
)

var (
	clusters     = map[string]clusters2.Cluster{}
	clustersLock sync.RWMutex
)

// RegisterCluster registers a cluster
func RegisterClusterWithConfig(name string, cluster *v1alpha1.Cluster) error {
	clustersLock.Lock()
	defer clustersLock.Unlock()
	c, err := manager.NewWithConfigs(cluster)
	if err != nil {
		return err
	}
	clusters[name] = c
	return nil
}

func FindClusterByName(name string) (clusters2.Cluster, bool) {
	clustersLock.RLock()
	defer clustersLock.RUnlock()
	c, ok := clusters[name]
	return c, ok
}
