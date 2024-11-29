package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	service "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/modules/auth"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
	"knoway.dev/pkg/types/openai"
)

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.APIKeyAuthConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	address := c.GetAuthServer().GetUrl()
	if address == "" {
		return nil, errors.New("invalid auth server url")
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	authClient := service.NewAuthServiceClient(conn)

	return &AuthFilter{
		config:     c,
		conn:       conn,
		authClient: authClient,
	}, nil
}

var _ filters.RequestFilter = (*AuthFilter)(nil)
var _ filters.OnCompletionRequestFilter = (*AuthFilter)(nil)

type AuthFilter struct {
	filters.IsRequestFilter

	config     *v1alpha1.APIKeyAuthConfig
	conn       *grpc.ClientConn
	authClient service.AuthServiceClient
}

func (a *AuthFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	slog.Debug("starting auth filter OnCompletionRequest")
	if err := auth.SetEnabledAuthFilterToCtx(ctx, true); err != nil {
		return filters.NewFailed(err)
	}

	// parse apikey
	apiKey, err := auth.BearerMarshal(sourceHTTPRequest)
	if err != nil {
		// todo added generic error handling, non-Hardcode openai error
		return filters.NewFailed(openai.NewErrorIncorrectAPIKey())
	}

	request.SetAPIKey(apiKey)

	// check apikey
	slog.Debug("auth filter: rpc APIKeyAuth")
	response, err := a.authClient.APIKeyAuth(ctx, &service.APIKeyAuthRequest{
		ApiKey: apiKey,
	})

	if err != nil {
		slog.Error("auth filter: APIKeyAuth error: %s", "error", err)
		return filters.NewFailed(err)
	}
	if err = auth.SetAuthInfoToCtx(ctx, response); err != nil {
		return filters.NewFailed(err)
	}

	request.SetUser(response.GetUserId())

	if !response.GetIsValid() {
		slog.Debug("auth filter: user apikey invalid", "user", response.GetUserId())
		return filters.NewFailed(openai.NewErrorIncorrectAPIKey())
	}

	accessModel := request.GetModel()
	if accessModel != "" && !auth.CanAccessModel(response.GetAllowModels(), accessModel) {
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
