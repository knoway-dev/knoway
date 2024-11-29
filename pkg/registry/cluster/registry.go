package cluster

import (
	"log/slog"
	"sync"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusters2 "knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/manager"
)

var clusterRegister *Register

func FindClusterByName(name string) (clusters2.Cluster, bool) {
	return clusterRegister.FindClusterByName(name)
}

func RemoveCluster(cluster *v1alpha1.Cluster) {
	clusterRegister.DeleteCluster(cluster.GetName())
}

func UpsertAndRegisterCluster(cluster *v1alpha1.Cluster, lifecycle bootkit.LifeCycle) error {
	return clusterRegister.UpsertAndRegisterCluster(cluster, lifecycle)
}

func ListModels() []*v1alpha1.Cluster {
	if clusterRegister == nil {
		return nil
	}

	return clusterRegister.ListModels()
}

func init() { //nolint:gochecknoinits
	if clusterRegister == nil {
		InitClusterRegister()
	}
}

type Register struct {
	clusters        map[string]clusters2.Cluster
	clustersDetails map[string]*v1alpha1.Cluster
	clustersLock    sync.RWMutex
}

type RegisterOptions struct {
	DevConfig bool
}

func NewClusterRegister() *Register {
	r := &Register{
		clusters:        make(map[string]clusters2.Cluster),
		clustersDetails: make(map[string]*v1alpha1.Cluster),
		clustersLock:    sync.RWMutex{},
	}

	return r
}

func InitClusterRegister() {
	c := NewClusterRegister()
	clusterRegister = c
}

func (cr *Register) DeleteCluster(name string) {
	cr.clustersLock.Lock()
	defer cr.clustersLock.Unlock()

	delete(cr.clusters, name)
	delete(cr.clustersDetails, name)
	slog.Info("remove cluster", "name", name)
}

func (cr *Register) FindClusterByName(name string) (clusters2.Cluster, bool) {
	cr.clustersLock.RLock()
	defer cr.clustersLock.RUnlock()

	c, ok := cr.clusters[name]

	return c, ok
}

func (cr *Register) UpsertAndRegisterCluster(cluster *v1alpha1.Cluster, lifecycle bootkit.LifeCycle) error {
	cr.clustersLock.Lock()
	defer cr.clustersLock.Unlock()

	name := cluster.GetName()

	c, err := manager.NewWithConfigs(cluster, lifecycle)
	if err != nil {
		return err
	}
	cr.clustersDetails[cluster.GetName()] = cluster
	cr.clusters[name] = c

	slog.Info("register cluster", "name", name)

	return nil
}

func (cr *Register) ListModels() []*v1alpha1.Cluster {
	cr.clustersLock.RLock()
	defer cr.clustersLock.RUnlock()

	clusters := make([]*v1alpha1.Cluster, 0, len(cr.clusters))
	for _, cluster := range cr.clustersDetails {
		clusters = append(clusters, cluster)
	}

	return clusters
}
