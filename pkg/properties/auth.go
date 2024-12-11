package properties

import (
	"context"

	services "knoway.dev/api/service/v1alpha1"
)

const (
	enabledAuthFilterKey = "filters.auth.enabled"
	authInfoKey          = "auth.info"
	apiKeyKey            = "auth.apiKey" //nolint:gosec
)

func SetAuthInfoToCtx(ctx context.Context, info *services.APIKeyAuthResponse) error {
	return SetProperty(ctx, authInfoKey, info)
}

func GetAuthInfoFromCtx(ctx context.Context) (*services.APIKeyAuthResponse, bool) {
	return GetProperty[*services.APIKeyAuthResponse](ctx, authInfoKey)
}

func SetEnabledAuthFilterToCtx(ctx context.Context, enabled bool) error {
	return SetProperty(ctx, enabledAuthFilterKey, enabled)
}

func EnabledAuthFilterFromCtx(ctx context.Context) bool {
	value, ok := GetProperty[bool](ctx, enabledAuthFilterKey)
	return value && ok
}

func SetAPIKeyToCtx(ctx context.Context, apiKey string) error {
	return SetProperty(ctx, apiKeyKey, apiKey)
}

func GetAPIKeyFromCtx(ctx context.Context) (string, bool) {
	return GetProperty[string](ctx, apiKeyKey)
}
