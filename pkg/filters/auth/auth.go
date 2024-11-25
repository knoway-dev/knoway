package auth

import (
	context "context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

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

func (a *AuthFilter) OnCompletionRequest(ctx context.Context, request object.LLMRequest) filters.RequestFilterResult {
	return filters.OK
}

func (a *AuthFilter) OnCompletionResponse(ctx context.Context, response object.LLMResponse) filters.RequestFilterResult {
	return filters.OK
}

func (a *AuthFilter) OnCompletionStreamResponse(ctx context.Context, response object.LLMRequest, endStream bool) filters.RequestFilterResult {
	return filters.OK
}

func (a *AuthFilter) APIKeyAuth(ctx context.Context, apikey string) (*v1alpha12.APIKeyAuthResponse, error) {
	req := &v1alpha12.APIKeyAuthRequest{
		ApiKey: apikey,
	}
	return a.authClient.APIKeyAuth(ctx, req)
}
