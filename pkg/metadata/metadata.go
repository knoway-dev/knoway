package metadata

import (
	"context"
	"net/http"
	"time"

	"github.com/samber/mo"

	"knoway.dev/api/clusters/v1alpha1"
	v1alpha12 "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/object"
)

type RequestMetadata struct {
	// RequestModel is the requested model name from user side,
	// used to route to the correct cluster and corresponding model.
	// Much similar to server_name in nginx or vHost in Apache.
	RequestModel string
	RequestAt    time.Time
	// ResponseModel is the model name that the user expects to receive.
	// In many scenarios, this is the same as RequestModel, except for
	// auto-routed models, where RequestModel could be `auto`, and
	// the actual selected Cluster of model name will be selected based on
	// the request payload / inference difficulty.
	ResponseModel string
	RespondAt     time.Time
	// Egress related metadata
	StatusCode   int
	ErrorMessage string

	// Auth related metadata
	EnabledAuthFilter bool                          // Set in AuthFilter
	AuthInfo          *v1alpha12.APIKeyAuthResponse // Set in AuthFilter

	// SelectedCluster is the cluster that the request is routed to
	SelectedCluster mo.Option[clusters.Cluster]

	// Upstream related metadata
	UpstreamProvider             v1alpha1.ClusterProvider // Set in Cluster Manager
	UpstreamResponseStatusCode   int                      // Set in Cluster Manager
	UpstreamResponseHeader       mo.Option[http.Header]   // Set in Cluster Manager
	UpstreamResponseErrorMessage string                   // Set in Cluster Manager
	// UpstreamRequestModel is the model name that the gateway will send to
	// upstream provider, generally the same as how Cluster overrides `model`
	// parameter in the request payload.
	UpstreamRequestModel string    // Set in Cluster Manager
	UpstreamRequestAt    time.Time // Set in Cluster Manager
	// UpstreamResponseModel is the model name that the upstream provider
	// will respond with. Same as explained in ResponseModel, when
	// UpstreamRequestModel set to `auto`, the actual model name will be
	// different from the UpstreamRequestModel since the load-balancing or
	// generic model routing will be done by the upstream provider.
	UpstreamResponseModel string    // Set in Cluster Manager
	UpstreamRespondAt     time.Time // Set in Cluster Manager
	// Setting in Listener is because when reading and handling the stream
	// of data, the response has been made and processed by Cluster, which
	// leaves the scope of Cluster Manager, and marshalling and writing to
	// Connection IO writer is done by Listener, thus the only actor that
	// knows when the first valid chunk of data is received.
	UpstreamFirstValidChunkAt time.Time // Set in Listener

	// Overall usage consumption
	LLMUpstreamTokensUsage mo.Option[object.LLMTokensUsage]
	LLMUpstreamImagesUsage mo.Option[object.LLMImagesUsage]
}

// RequestMetadataFromCtx retrieves RequestMetadata from context
// Note: The returned pointer allows direct access and modification of the underlying RequestMetadata
// Be careful when modifying the properties as they are shared across the request context
func RequestMetadataFromCtx(ctx context.Context) *RequestMetadata {
	props, pok := ctx.Value(metadataKey{}).(*metadata)
	if !pok {
		return nil
	}

	return props.request
}

func InitMetadataContext(request *http.Request) context.Context {
	return context.WithValue(request.Context(), metadataKey{}, &metadata{
		request: &RequestMetadata{},
	})
}

type metadataKey struct{}

type metadata struct {
	request *RequestMetadata
}
