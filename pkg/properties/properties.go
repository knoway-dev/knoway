package properties

import "context"

type propertiesKey struct{}

type pp map[string]interface{}

// NewPropertiesContext 初始化 PropertiesContext
func NewPropertiesContext(ctx context.Context) context.Context {
	props := make(pp)
	return context.WithValue(ctx, propertiesKey{}, props)
}

// fromPropertiesContext 从 context 中获取所有属性
func fromPropertiesContext(ctx context.Context) pp {
	props, ok := ctx.Value(propertiesKey{}).(pp)
	if !ok {
		return nil
	}
	return props
}

// SetProperty 设置任意类型的值到 context 中
func SetProperty(ctx context.Context, key string, value interface{}) context.Context {
	props := fromPropertiesContext(ctx)
	if props == nil {
		props = make(pp)
	}
	// update old ctx
	props[key] = value

	newProps := make(pp)
	copyP(newProps, props)
	return context.WithValue(ctx, propertiesKey{}, newProps)
}

func copyP(dest, src pp) {
	for k, v := range src {
		dest[k] = v
	}
}

// GetProperty 获取任意类型的值从 context 中
func GetProperty[T any](ctx context.Context, key string) (T, bool) {
	var zero T
	props := fromPropertiesContext(ctx)
	if props == nil {
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
