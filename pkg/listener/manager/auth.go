package manager

import (
	"context"
	"errors"
	"net/http"
	"strings"

	v1alpha12 "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/filters/auth"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/types/openai"
)

func UnmarshalLLMRequestHeader(ctx context.Context, request *http.Request) (object.RequestHeader, error) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		return object.RequestHeader{}, errors.New("missing Authorization header")
	}

	// 检查是否为 Bearer 格式
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return object.RequestHeader{}, errors.New("invalid Authorization header format")
	}

	// 提取 API Key
	apiKey := strings.TrimPrefix(authHeader, prefix)
	if apiKey == "" {
		return object.RequestHeader{}, errors.New("missing API Key in Authorization header")
	}

	return object.RequestHeader{APIKey: apiKey}, nil
}

type AuthZOption struct {
	CanAccessModel string
}

func Auth(authFilter *auth.AuthFilter, request *http.Request, opt AuthZOption) (*v1alpha12.APIKeyAuthResponse, error) {
	head, err := UnmarshalLLMRequestHeader(request.Context(), request)
	if err != nil {
		return nil, openai.NewErrorIncorrectAPIKey()
	}

	authResp, err := authFilter.APIKeyAuth(request.Context(), head.APIKey)
	if err != nil {
		return nil, err
	}

	if !authResp.IsValid {
		return nil, openai.NewErrorIncorrectAPIKey()
	}

	if opt.CanAccessModel != "" && !authResp.CanAccessModel(opt.CanAccessModel) {
		return nil, openai.NewErrorModelNotFoundOrNotAccessible(opt.CanAccessModel)
	}

	return authResp, nil
}
