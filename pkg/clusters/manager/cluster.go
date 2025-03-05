package manager

import (
	"context"
	"log/slog"
	"sync"

	"github.com/samber/lo"
	"github.com/samber/mo"

	"knoway.dev/api/clusters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusters2 "knoway.dev/pkg/clusters"
	cluster "knoway.dev/pkg/clusters/cluster"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
)

var clusterRegister *Register

func HandleRequest(ctx context.Context, clusterName string, request object.LLMRequest) (object.LLMResponse, error) {
	foundCluster, ok := clusterRegister.FindClusterByName(clusterName)
	if !ok {
		return nil, object.NewErrorModelNotFoundOrNotAccessible(request.GetModel())
	}

	rMeta := metadata.RequestMetadataFromCtx(ctx)
	rMeta.SelectedCluster = mo.Some(foundCluster)

	resp, err := foundCluster.DoUpstreamRequest(ctx, request)
	if err != nil {
		// Cluster will ensure that error will always be LLMError
		return resp, err
	}
	if resp.GetError() != nil {
		return resp, resp.GetError()
	}

	return resp, err
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

func (cr *Register) UpsertAndRegisterCluster(c *v1alpha1.Cluster, lifecycle bootkit.LifeCycle) error {
	cr.clustersLock.Lock()
	defer cr.clustersLock.Unlock()

	name := c.GetName()

	newCluster, err := cluster.NewWithConfigs(c, lifecycle)
	if err != nil {
		return err
	}

	cr.clustersDetails[c.GetName()] = c
	cr.clusters[name] = newCluster

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

func (cr *Register) dumpAllClusters() []*v1alpha1.Cluster {
	cr.clustersLock.RLock()
	defer cr.clustersLock.RUnlock()

	return lo.Values(clusterRegister.clustersDetails)
}

func DebugDumpAllClusters() []*v1alpha1.Cluster {
	return clusterRegister.dumpAllClusters()
}
