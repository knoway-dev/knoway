package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"knoway.dev/pkg/properties"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"knoway.dev/pkg/types/openai"

	"google.golang.org/protobuf/types/known/anypb"

	v1alpha12 "knoway.dev/api/service/v1alpha1"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewWithConfig(cfg *anypb.Any) (filters.RequestFilter, error) {
	c, err := protoutils.FromAny[*v1alpha1.APIKeyAuthConfig](cfg, &v1alpha1.APIKeyAuthConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	address := c.GetAuthServer().GetUrl()
	if address == "" {
		return nil, errors.New("invalid auth server url")
	}
	// todo how to close connect when process exist
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	authClient := v1alpha12.NewAuthServiceClient(conn)

	return &AuthFilter{
		config:     c,
		conn:       conn,
		authClient: authClient,
	}, nil
}

type AuthFilter struct {
	config     *v1alpha1.APIKeyAuthConfig
	conn       *grpc.ClientConn
	authClient v1alpha12.AuthServiceClient
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

// CanAccessModel 判断是否可以访问指定的模型
// 规范如下：
//
//	如果是 *，代表可以访问所有公开模型，无法匹配 /。
//	如果是 u-kebe/* 代表可以访问 kebe 的所有模型。
//	如果是 */* 表示可以访问任意模型。
func CanAccessModel(allowModels []string, requestModel string) bool {
	for _, rule := range allowModels {
		// 处理 "*/*"，允许访问任意模型
		if rule == "*/*" {
			return true
		}

		// 处理具体模型名称的匹配
		if rule == requestModel {
			return true
		}

		// 处理 "*"，允许访问所有公开模型
		if rule == "*" && !strings.Contains(requestModel, "/") {
			return true
		}

		// 处理 "ns/*"，允许访问特定 ns 下的所有模型
		if strings.HasSuffix(rule, "/*") {
			if strings.HasPrefix(requestModel, strings.TrimSuffix(rule, "/*")+"/") {
				return true
			}
		}
	}

	return false
}

func (a *AuthFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	slog.Debug("starting auth filter OnCompletionRequest")
	if err := SetEnabledAuthFilterToCtx(ctx, true); err != nil {
		return filters.NewFailed(err)
	}

	// parse apikey
	apiKey, err := BearerMarshal(sourceHTTPRequest)
	if err != nil {
		// todo added generic error handling, non-Hardcode openai error
		return filters.NewFailed(openai.NewErrorIncorrectAPIKey())
	}

	request.SetAPIKey(apiKey)

	// check apikey
	slog.Debug("auth filter: rpc APIKeyAuth")
	response, err := a.authClient.APIKeyAuth(ctx, &v1alpha12.APIKeyAuthRequest{
		ApiKey: apiKey,
	})

	if err != nil {
		slog.Error("auth filter: APIKeyAuth error: %s", "error", err)
		return filters.NewFailed(err)
	}
	if err = SetAuthInfoToCtx(ctx, response); err != nil {
		return filters.NewFailed(err)
	}

	request.SetUser(response.GetUserId())

	if !response.GetIsValid() {
		slog.Debug("auth filter: user apikey invalid", "user", response.GetUserId())
		return filters.NewFailed(openai.NewErrorIncorrectAPIKey())
	}

	accessModel := request.GetModel()
	if accessModel != "" && !CanAccessModel(response.GetAllowModels(), accessModel) {
		slog.Debug("auth filter: user can not access model", "user", response.GetUserId(), "model", accessModel)
		return filters.NewFailed(openai.NewErrorModelNotFoundOrNotAccessible(accessModel))
	}

	slog.Debug("auth filter: user authorization succeeds", "user", response.GetUserId(), "allow models", response.GetAllowModels())

	return filters.NewOK()
}

func (a *AuthFilter) OnCompletionResponse(ctx context.Context, response object.LLMResponse) filters.RequestFilterResult {
	return filters.NewOK()
}

func (a *AuthFilter) OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) filters.RequestFilterResult {
	return filters.NewOK()
}

const (
	enabledAuthFilterKey = "enabledAuthFilter"
	authInfoKey          = "authInfo"
)

func SetAuthInfoToCtx(ctx context.Context, info *v1alpha12.APIKeyAuthResponse) error {
	return properties.SetProperty(ctx, authInfoKey, info)
}

func GetAuthInfoFromCtx(ctx context.Context) (*v1alpha12.APIKeyAuthResponse, bool) {
	return properties.GetProperty[*v1alpha12.APIKeyAuthResponse](ctx, authInfoKey)
}

func SetEnabledAuthFilterToCtx(ctx context.Context, enabled bool) error {
	return properties.SetProperty(ctx, enabledAuthFilterKey, enabled)
}

func EnabledAuthFilterFromCtx(ctx context.Context) bool {
	value, ok := properties.GetProperty[bool](ctx, enabledAuthFilterKey)
	return value && ok
}
