package properties

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"knoway.dev/api/clusters/v1alpha1"
	v1alpha12 "knoway.dev/api/service/v1alpha1"
)

type RequestProperties struct {
	EnabledAuthFilter bool
	AuthInfo          *v1alpha12.APIKeyAuthResponse
	APIKey            string

	RequestModel  string
	ResponseModel string
	Cluster       *v1alpha1.Cluster

	StatusCode   int
	ErrorMessage string
}

func GetRequestFromCtx(ctx context.Context) *RequestProperties {
	props, pok := fromPropertiesContext(ctx)
	if !pok {
		return nil
	}

	return props.request
}

func InitPropertiesContext(request *http.Request) context.Context {
	return context.WithValue(request.Context(), propertiesKey{}, &property{
		mutex:   sync.RWMutex{},
		pp:      make(map[string]any),
		request: &RequestProperties{},
	})
}

type propertiesKey struct{}

type property struct {
	mutex   sync.RWMutex
	pp      map[string]any
	request *RequestProperties
}

func (p *property) Set(key string, value interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.pp[key] = value
}

func (p *property) Get(key string) (any, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	value, exists := p.pp[key]
	if !exists {
		return nil, false
	}

	return value, true
}

func fromPropertiesContext(ctx context.Context) (*property, bool) {
	props, ok := ctx.Value(propertiesKey{}).(*property)
	return props, ok
}

func setProperty(ctx context.Context, key string, value interface{}) error {
	props, ok := fromPropertiesContext(ctx)
	if !ok {
		return errors.New("context does not have properties space, please use NewPropertiesContext to initialize it")
	}
	// update old ctx
	props.Set(key, value)

	return nil
}

func getProperty[T any](ctx context.Context, key string) (T, bool) {
	var zero T

	props, pok := fromPropertiesContext(ctx)
	if !pok {
		return zero, false
	}

	value, exists := props.Get(key)
	if !exists {
		return zero, false
	}

	v, ok := value.(T)
	if !ok {
		return zero, false
	}

	return v, true
}
