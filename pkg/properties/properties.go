package properties

import (
	"context"
	"net/http"

	v1alpha12 "knoway.dev/api/service/v1alpha1"
)

type RequestProperties struct {
	EnabledAuthFilter bool

	AuthInfo *v1alpha12.APIKeyAuthResponse

	RequestModel  string
	ResponseModel string

	StatusCode   int
	ErrorMessage string
}

// RequestPropertiesFromCtx retrieves RequestProperties from context
// Note: The returned pointer allows direct access and modification of the underlying RequestProperties
// Be careful when modifying the properties as they are shared across the request context
func RequestPropertiesFromCtx(ctx context.Context) *RequestProperties {
	props, pok := ctx.Value(propertiesKey{}).(*property)
	if !pok {
		return nil
	}

	return props.request
}

func InitPropertiesContext(request *http.Request) context.Context {
	return context.WithValue(request.Context(), propertiesKey{}, &property{
		request: &RequestProperties{},
	})
}

type propertiesKey struct{}

type property struct {
	request *RequestProperties
}
