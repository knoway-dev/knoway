package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"knoway.dev/api/filters/v1alpha1"
	service "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

const (
	defaultAuthServerTimeout = 3 * time.Second
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

	if c.GetAuthServer().GetTimeout().AsDuration() <= 0 {
		c.AuthServer.Timeout = durationpb.New(defaultAuthServerTimeout)
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
var _ filters.OnRequestPreFilter = (*AuthFilter)(nil)
var _ filters.OnCompletionRequestFilter = (*AuthFilter)(nil)
var _ filters.OnImageGenerationsRequestFilter = (*AuthFilter)(nil)

type AuthFilter struct {
	filters.IsRequestFilter

	config     *v1alpha1.APIKeyAuthConfig
	conn       *grpc.ClientConn
	authClient service.AuthServiceClient
}

func (a *AuthFilter) OnRequestPre(ctx context.Context, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	rMeta.EnabledAuthFilter = true

	slog.Debug("starting auth filter OnCompletionRequest")

	// parse apikey
	apiKey, err := BearerMarshal(sourceHTTPRequest)
	if err != nil {
		return filters.NewFailed(object.NewErrorMissingAPIKey())
	}

	getAuthCtx, cancel := context.WithTimeout(ctx, a.config.GetAuthServer().GetTimeout().AsDuration())
	defer cancel()

	// check apikey
	slog.Debug("auth filter: rpc APIKeyAuth")

	response, err := a.authClient.APIKeyAuth(getAuthCtx, &service.APIKeyAuthRequest{
		ApiKey: apiKey,
	})
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			slog.Error("auth filter: APIKeyAuth error: %s", "error", err)
			return filters.NewFailed(err)
		}

		switch s.Code() { //nolint:exhaustive
		case codes.NotFound:
			slog.Debug("auth filter: user apikey not found", "apikey", apiKey)
			return filters.NewFailed(object.NewErrorIncorrectAPIKey(apiKey))
		case codes.Unauthenticated:
			slog.Debug("auth filter: user apikey invalid", "apikey", apiKey)
			return filters.NewFailed(object.NewErrorIncorrectAPIKey(apiKey))
		case codes.PermissionDenied:
			slog.Debug("auth filter: user apikey permission denied", "apikey", apiKey)
			return filters.NewFailed(object.NewErrorIncorrectAPIKey(apiKey))
		case codes.Unavailable:
			slog.Debug("auth filter: user apikey service unavailable", "apikey", apiKey)
			return filters.NewFailed(object.NewErrorServiceUnavailable())
		default:
			slog.Error("auth filter: APIKeyAuth error: %s", "error", err)
			return filters.NewFailed(err)
		}
	}

	rMeta.AuthInfo = response

	if !response.GetIsValid() {
		slog.Debug("auth filter: user apikey invalid", "user", response.GetUserId())
		return filters.NewFailed(object.NewErrorIncorrectAPIKey(apiKey))
	}

	slog.Debug("auth filter: user authorization succeeds", "user", response.GetUserId(), "allow models", response.GetAllowModels())

	return filters.NewOK()
}

func (a *AuthFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	if rMeta.AuthInfo == nil {
		return filters.NewFailed(errors.New("missing auth info in context"))
	}

	authInfo := rMeta.AuthInfo

	accessModel := request.GetModel()
	if accessModel == "" {
		slog.Debug("auth filter: user model not found", "user", authInfo.GetUserId())
		return filters.NewFailed(object.NewErrorMissingModel())
	}

	denied := IsDenied(accessModel, authInfo.GetDenyModels())
	granted := IsGranted(accessModel, authInfo.GetAllowModels())

	if !CanAccessModelFromValues(denied, granted) {
		if denied {
			slog.Debug("auth filter: user access model is denied", "user", authInfo.GetUserId(), "model", accessModel)
			return filters.NewFailed(object.NewErrorModelAccessDenied(accessModel))
		}
		if !granted {
			slog.Debug("auth filter: user can not access model", "user", authInfo.GetUserId(), "model", accessModel)
			return filters.NewFailed(object.NewErrorModelNotFoundOrNotAccessible(accessModel))
		}
	}

	slog.Debug("auth filter: user authorization succeeds", "user", authInfo.GetUserId(), "allow models", authInfo.GetAllowModels())

	return filters.NewOK()
}

func (a *AuthFilter) OnImageGenerationsRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	if rMeta.AuthInfo == nil {
		return filters.NewFailed(errors.New("missing auth info in context"))
	}

	authInfo := rMeta.AuthInfo

	accessModel := request.GetModel()
	if accessModel == "" {
		slog.Debug("auth filter: user model not found", "user", authInfo.GetUserId())
		return filters.NewFailed(object.NewErrorMissingModel())
	}

	denied := IsDenied(accessModel, authInfo.GetDenyModels())
	granted := IsGranted(accessModel, authInfo.GetAllowModels())

	if !CanAccessModelFromValues(denied, granted) {
		if denied {
			slog.Debug("auth filter: user access model is denied", "user", authInfo.GetUserId(), "model", accessModel)
			return filters.NewFailed(object.NewErrorModelAccessDenied(accessModel))
		}
		if !granted {
			slog.Debug("auth filter: user can not access model", "user", authInfo.GetUserId(), "model", accessModel)
			return filters.NewFailed(object.NewErrorModelNotFoundOrNotAccessible(accessModel))
		}
	}

	slog.Debug("auth filter: user authorization succeeds", "user", authInfo.GetUserId(), "allow models", authInfo.GetAllowModels())

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

func IsDenied(requestModel string, denyModels []string) bool {
	if len(denyModels) == 0 {
		return false
	}

	return lo.SomeBy(denyModels, func(rule string) bool {
		// use glob matching to match the rule
		matched, err := doublestar.Match(rule, requestModel)
		if err != nil {
			return false
		}

		return matched
	})
}

func IsGranted(requestModel string, allowModels []string) bool {
	// if allowModels is empty, means that all models can be accessed.
	// This follows our definition in the api.
	if len(allowModels) == 0 {
		return true
	}

	return lo.SomeBy(allowModels, func(rule string) bool {
		// use glob matching to match the rule
		matched, err := doublestar.Match(rule, requestModel)
		if err != nil {
			return false
		}

		return matched
	})
}

/*
CanAccessModel determines whether the user can access the specified model.

The rules defined in allowModels follows the spec of the following:

- if * is provided, means that all public models can be accessed, except the ones with /.

- if u-kebe/* is provided, means that all models under the u-kebe namespace can be accessed, if we define u- means all individual users, then u-kebe/* means that all models under the kebe user can be accessed.

- if ** is provided, means that all models can be accessed.
*/
func CanAccessModel(requestModel string, allowModels []string, denyModels []string) bool {
	denied := IsDenied(requestModel, denyModels)
	granted := IsGranted(requestModel, allowModels)

	return CanAccessModelFromValues(denied, granted)
}

func CanAccessModelFromValues(isDenied bool, isGranted bool) bool {
	return !isDenied && isGranted
}
