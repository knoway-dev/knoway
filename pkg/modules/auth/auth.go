package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	services "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/properties"
)

const (
	enabledAuthFilterKey = "enabledAuthFilter"
	authInfoKey          = "authInfo"
)

func SetAuthInfoToCtx(ctx context.Context, info *services.APIKeyAuthResponse) error {
	return properties.SetProperty(ctx, authInfoKey, info)
}

func GetAuthInfoFromCtx(ctx context.Context) (*services.APIKeyAuthResponse, bool) {
	return properties.GetProperty[*services.APIKeyAuthResponse](ctx, authInfoKey)
}

func SetEnabledAuthFilterToCtx(ctx context.Context, enabled bool) error {
	return properties.SetProperty(ctx, enabledAuthFilterKey, enabled)
}

func EnabledAuthFilterFromCtx(ctx context.Context) bool {
	value, ok := properties.GetProperty[bool](ctx, enabledAuthFilterKey)
	return value && ok
}

func BearerMarshal(request *http.Request) (string, error) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, prefix)
	if token == "" {
		return "", errors.New("missing API Key in Authorization header")
	}

	return token, nil
}

/*
CanAccessModel determines whether the user can access the specified model.

The rules defined in allowModels follows the spec of the following:

- if * is provided, means that all public models can be accessed, except the ones with /.

- if u-kebe/* is provided, means that all models under the u-kebe namespace can be accessed, if we define u- means all individual users, then u-kebe/* means that all models under the kebe user can be accessed.

- if *\/* is provided, means that all models can be accessed.
*/
func CanAccessModel(allowModels []string, requestModel string) bool {
	for _, rule := range allowModels {
		// allow all models
		if rule == "*/*" {
			return true
		}

		// allow specific model
		if rule == requestModel {
			return true
		}

		// allow all public models
		if rule == "*" && !strings.Contains(requestModel, "/") {
			return true
		}

		// allow all models under a specific namespace, e.g. ns/*
		if strings.HasSuffix(rule, "/*") {
			if strings.HasPrefix(requestModel, strings.TrimSuffix(rule, "/*")+"/") {
				return true
			}
		}
	}

	return false
}
