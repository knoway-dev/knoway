package context

import (
	"context"

	"knoway.dev/api/service/v1alpha1"
)

func SetAuthInfo(ctx context.Context, info *v1alpha1.APIKeyAuthResponse) context.Context {
	return SetValue(ctx, authInfoKey, info)
}

func GetAuthInfo(ctx context.Context) (*v1alpha1.APIKeyAuthResponse, bool) {
	return GetValue[*v1alpha1.APIKeyAuthResponse](ctx, authInfoKey)
}

func SetEnabledAuthFilter(ctx context.Context, enabled bool) context.Context {
	return SetValue(ctx, enabledAuthFilterKey, enabled)
}

func EnabledAuthFilter(ctx context.Context) bool {
	value, ok := GetValue[bool](ctx, enabledAuthFilterKey)
	return value && ok
}
