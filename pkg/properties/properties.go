package properties

import (
	"context"
	"errors"
)

type propertiesKey struct{}

type pp map[string]interface{}

// fromPropertiesContext 从 context 中获取所有属性
func fromPropertiesContext(ctx context.Context) (pp, bool) {
	props, ok := ctx.Value(propertiesKey{}).(pp)
	return props, ok
}

// NewPropertiesContext 初始化 PropertiesContext
func NewPropertiesContext(ctx context.Context) context.Context {
	props := make(pp)
	return context.WithValue(ctx, propertiesKey{}, props)
}

// SetProperty 设置任意类型的值到 context 中
func SetProperty(ctx context.Context, key string, value interface{}) error {
	props, ok := fromPropertiesContext(ctx)
	if !ok {
		return errors.New("context does not have properties space, please use NewPropertiesContext to initialize it")
	}
	// update old ctx
	props[key] = value

	return nil
}

// GetProperty 获取任意类型的值从 context 中
func GetProperty[T any](ctx context.Context, key string) (T, bool) {
	var zero T

	props, pok := fromPropertiesContext(ctx)
	if !pok {
		return zero, false
	}

	value, exists := props[key]
	if !exists {
		return zero, false
	}

	v, ok := value.(T)
	if !ok {
		return zero, false
	}

	return v, true
}
