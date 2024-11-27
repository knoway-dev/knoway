package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

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
		return nil, fmt.Errorf("invalid auth server url")
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

func BearerMarshal(request *http.Request) (token string, err error) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("invalid Authorization header format")
	}

	token = strings.TrimPrefix(authHeader, prefix)
	if token == "" {
		return "", errors.New("missing API Key in Authorization header")
	}
	return
}

func (a *AuthFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHttpRequest *http.Request) filters.RequestFilterResult {
	slog.Debug(fmt.Sprintf("starting auth filter OnCompletionRequest"))
	// parse apikey
	apiKey, err := BearerMarshal(sourceHttpRequest)
	if err != nil {
		return filters.NewFailed(openai.NewErrorIncorrectAPIKey())
	}
	request.SetApiKey(apiKey)

	// check apikey
	slog.Debug(fmt.Sprintf("auth filter: rpc APIKeyAuth"))
	response, err := a.authClient.APIKeyAuth(ctx, &v1alpha12.APIKeyAuthRequest{
		ApiKey: apiKey,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("auth filter: APIKeyAuth error: %s", err))
		return filters.NewFailed(err)
	}

	if response != nil {
		request.SetAuthInfo(
			response.IsValid,
			response.UserId,
			response.AllowModels)
	}

	if !response.IsValid {
		slog.Debug(fmt.Sprintf("auth filter: user %s apikey invalid", response.UserId))
		return filters.NewFailed(openai.NewErrorIncorrectAPIKey())
	}

	accessModel := request.GetModel()
	if accessModel != "" && !request.CanAccessModel(accessModel) {
		slog.Debug(fmt.Sprintf("auth filter: user %s can not access model %s", response.UserId, accessModel))
		return filters.NewFailed(openai.NewErrorModelNotFoundOrNotAccessible(accessModel))
	}
	slog.Debug(fmt.Sprintf("auth filter: user %s authorization succeeds: allowMoldes %v", response.UserId, response.AllowModels))
	return filters.NewOK()
}

func (a *AuthFilter) OnCompletionResponse(ctx context.Context, response object.LLMResponse) filters.RequestFilterResult {
	return filters.NewOK()
}

func (a *AuthFilter) OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) filters.RequestFilterResult {
	return filters.NewOK()
}
