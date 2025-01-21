package clusters

import (
	"knoway.dev/api/clusters/v1alpha1"
	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

var (
	mapClusterProviderBackendProvider = map[v1alpha1.ClusterProvider]knowaydevv1alpha1.Provider{}
	mapBackendProviderClusterProvider = map[knowaydevv1alpha1.Provider]v1alpha1.ClusterProvider{}
)

func MapClusterProviderToBackendProvider(provider v1alpha1.ClusterProvider) knowaydevv1alpha1.Provider {
	return mapClusterProviderBackendProvider[provider]
}

func MapBackendProviderToClusterProvider(provider knowaydevv1alpha1.Provider) v1alpha1.ClusterProvider {
	return mapBackendProviderClusterProvider[provider]
}
