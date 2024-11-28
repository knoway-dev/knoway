package cluster

import (
	"log/slog"
	"sync"

	v1alpha3 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/registry/route"

	"knoway.dev/api/clusters/v1alpha1"
	clusters2 "knoway.dev/pkg/clusters"
	"knoway.dev/pkg/clusters/manager"
)

var clusterRegister *Register

func FindClusterByName(name string) (clusters2.Cluster, bool) {
	return clusterRegister.FindClusterByName(name)
}

func RemoveCluster(cluster *v1alpha1.Cluster) {
	clusterRegister.DeleteCluster(cluster.Name)
}

func UpsertAndRegisterCluster(cluster *v1alpha1.Cluster) error {
	return clusterRegister.UpsertAndRegisterCluster(cluster)
}

func ListModels() []*v1alpha1.Cluster {
	if clusterRegister == nil {
		return nil
	}

	return clusterRegister.ListModels()
}

func init() {
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

	route.RemoveRoute(name)
	slog.Info("remove cluster and direct route: %s", "name", name)
}

func (cr *Register) FindClusterByName(name string) (clusters2.Cluster, bool) {
	cr.clustersLock.RLock()
	defer cr.clustersLock.RUnlock()

	c, ok := cr.clusters[name]

	return c, ok
}

func (cr *Register) UpsertAndRegisterCluster(cluster *v1alpha1.Cluster) error {
	cr.clustersLock.Lock()
	defer cr.clustersLock.Unlock()

	name := cluster.Name

	c, err := manager.NewWithConfigs(cluster)
	if err != nil {
		return err
	}
	cr.clustersDetails[cluster.Name] = cluster
	cr.clusters[name] = c

	rConfig := &v1alpha3.Route{
		Name: name,
		Matches: []*v1alpha3.Match{
			{
				Model: &v1alpha3.StringMatch{
					Match: &v1alpha3.StringMatch_Exact{
						Exact: name,
					},
				},
			},
		},
		ClusterName: name,
		Filters:     nil, // todo future
	}
	if err = route.RegisterRouteWithConfig(rConfig); err != nil {
		return err
	}
	slog.Info("register cluster and direct route: %s", "name", name)
	return nil
}

func StaticRegisterClusters(clusterDetails map[string]*v1alpha1.Cluster) error {
	for _, cluster := range clusterDetails {
		if err := UpsertAndRegisterCluster(cluster); err != nil {
			return err
		}
	}

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
