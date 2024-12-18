package metadata

import (
	"context"
	"net/http"
	"time"

	v1alpha12 "knoway.dev/api/service/v1alpha1"
)

type RequestMetadata struct {
	EnabledAuthFilter bool

	AuthInfo *v1alpha12.APIKeyAuthResponse

	RequestModel  string
	ResponseModel string

	StatusCode   int
	ErrorMessage string

	RequestAt                 time.Time
	ResponseAt                time.Time
	UpstreamRequestAt         time.Time
	UpstreamResponseAt        time.Time
	UpstreamFirstValidChunkAt time.Time
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
