package context

import "context"

// WithProperty 添加单个键值对到 context 中
func WithProperty(ctx context.Context, key string, value interface{}) context.Context {
	props := GetProperties(ctx)
	if props == nil {
		props = make(map[string]interface{})
	}
	props[key] = value
	return context.WithValue(ctx, propertiesKey, props)
}

func InitProperties(ctx context.Context) context.Context {
	props := make(map[string]interface{})
	return context.WithValue(ctx, propertiesKey, props)
}

// GetProperty 从 context 中获取指定键的值
func GetProperty(ctx context.Context, key string) (interface{}, bool) {
	props := GetProperties(ctx)
	if props == nil {
		return nil, false
	}
	value, ok := props[key]
	return value, ok
}

// GetProperties 从 context 中获取所有属性
func GetProperties(ctx context.Context) map[string]interface{} {
	props, ok := ctx.Value(propertiesKey).(map[string]interface{})
	if !ok {
		return nil
	}
	return props
}

// DeleteProperty 从 context 中删除指定键
func DeleteProperty(ctx context.Context, key string) context.Context {
	props := GetProperties(ctx)
	if props == nil {
		return ctx
	}
	delete(props, key)
	return context.WithValue(ctx, propertiesKey, props)
}

// SetValue 设置任意类型的值到 context 中
func SetValue(ctx context.Context, key string, value interface{}) context.Context {
	return WithProperty(ctx, key, value)
}

// GetValue 获取任意类型的值从 context 中
func GetValue[T any](ctx context.Context, key string) (T, bool) {
	value, exists := GetProperty(ctx, key)
	if !exists {
		var zero T
		return zero, false
	}
	v, ok := value.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return v, true
}

// DeleteValue 从 context 中删除指定键的值
func DeleteValue(ctx context.Context, key string) context.Context {
	return DeleteProperty(ctx, key)
}
