package cluster

import (
	"knoway.dev/api/clusters/v1alpha1"
	clusters2 "knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/manager"
	"knoway.dev/pkg/config"
	"log"
	"sync"
)

var clusterRegister *ClusterRegister

type ClusterRegister struct {
	configClient *config.ConfigClient
	clusters     map[string]clusters2.Cluster
	clustersLock sync.RWMutex
}

func NewClusterRegister(configFilePath string) *ClusterRegister {
	r := &ClusterRegister{
		clusters: make(map[string]clusters2.Cluster),
	}
	r.configClient = config.NewConfigClient(configFilePath, r.handleClusterUpdates)
	return r
}

func (r *ClusterRegister) handleClusterUpdates(updatedClusters map[string]*v1alpha1.Cluster) {
	r.clustersLock.Lock()
	oldClusters := r.clusters
	r.clustersLock.Unlock()

	// Register new clusters and update existing ones
	for name, cluster := range updatedClusters {
		if _, exists := oldClusters[name]; exists {
			r.DeleteCluster(name) // Remove the old cluster before updating
		}
		if err := r.RegisterClusterWithConfig(name, cluster); err != nil {
			log.Printf("Error registering/updating cluster %s: %v", name, err)
		}
	}

	// Remove clusters that are no longer present
	for name := range oldClusters {
		if _, exists := updatedClusters[name]; !exists {
			r.DeleteCluster(name)
			log.Printf("Removed cluster: %s", name)
		}
	}
}

// Start starts the config client
func (cr *ClusterRegister) Start() {
	cr.configClient.Start()
}

// RegisterCluster registers a cluster
func (cr *ClusterRegister) RegisterClusterWithConfig(name string, cluster *v1alpha1.Cluster) error {
	cr.clustersLock.Lock()
	defer cr.clustersLock.Unlock()
	c, err := manager.NewWithConfigs(cluster)
	if err != nil {
		return err
	}
	cr.clusters[name] = c
	return nil
}

func (cr *ClusterRegister) DeleteCluster(name string) {
	cr.clustersLock.Lock()
	defer cr.clustersLock.Unlock()
	delete(cr.clusters, name)
}

func (cr *ClusterRegister) FindClusterByName(name string) (clusters2.Cluster, bool) {
	cr.clustersLock.RLock()
	defer cr.clustersLock.RUnlock()
	c, ok := cr.clusters[name]
	return c, ok
}

func FindClusterByName(name string) (clusters2.Cluster, bool) {
	return clusterRegister.FindClusterByName(name)
}
