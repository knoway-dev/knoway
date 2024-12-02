package properties

import (
	"context"
	"errors"
	"sync"
)

type propertiesKey struct{}

type property struct {
	mutex sync.RWMutex
	pp    map[string]any
}

func newProperty() *property {
	return &property{
		pp: make(map[string]any),
	}
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

// fromPropertiesContext 从 context 中获取所有属性
func fromPropertiesContext(ctx context.Context) (*property, bool) {
	props, ok := ctx.Value(propertiesKey{}).(*property)
	return props, ok
}

// NewPropertiesContext 初始化 PropertiesContext
func NewPropertiesContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, propertiesKey{}, newProperty())
}

// SetProperty 设置任意类型的值到 context 中
func SetProperty(ctx context.Context, key string, value interface{}) error {
	props, ok := fromPropertiesContext(ctx)
	if !ok {
		return errors.New("context does not have properties space, please use NewPropertiesContext to initialize it")
	}
	// update old ctx
	props.Set(key, value)

	return nil
}

// GetProperty 获取任意类型的值从 context 中
func GetProperty[T any](ctx context.Context, key string) (T, bool) {
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
