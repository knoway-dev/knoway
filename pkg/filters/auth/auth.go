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
	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	service "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/properties"
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
	if err := properties.SetEnabledAuthFilterToCtx(ctx, true); err != nil {
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
	response, err := a.authClient.APIKeyAuth(ctx, &service.APIKeyAuthRequest{
		ApiKey: apiKey,
	})

	if err != nil {
		slog.Error("auth filter: APIKeyAuth error: %s", "error", err)
		return filters.NewFailed(err)
	}
	if err = properties.SetAuthInfoToCtx(ctx, response); err != nil {
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
