package usage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/samber/lo"

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
)

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.UsageStatsConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	address := c.GetStatsServer().GetUrl()
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

	authClient := service.NewUsageStatsServiceClient(conn)

	return &UsageFilter{
		config:      c,
		conn:        conn,
		usageClient: authClient,
	}, nil
}

var _ filters.RequestFilter = (*UsageFilter)(nil)
var _ filters.OnCompletionResponseFilter = (*UsageFilter)(nil)
var _ filters.OnCompletionStreamResponseFilter = (*UsageFilter)(nil)

type UsageFilter struct {
	filters.IsRequestFilter

	config      *v1alpha1.UsageStatsConfig
	conn        *grpc.ClientConn
	usageClient service.UsageStatsServiceClient
}

func (f *UsageFilter) usageReport(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error {
	usage := response.GetUsage()
	if lo.IsNil(usage) {
		slog.Warn("no usage in response", "model", request.GetModel())
		return nil
	}

	var apiKeyID string
	authInfo, ok := properties.GetAuthInfoFromCtx(ctx)
	if !ok {
		slog.Warn("no auth info in context")

		return nil
	} else {
		apiKeyID = authInfo.GetApiKeyId()
	}

	_, err := f.usageClient.UsageReport(context.TODO(), &service.UsageReportRequest{
		ApiKeyId:          apiKeyID,
		UserModelName:     request.GetModel(),
		UpstreamModelName: response.GetModel(),
		Usage: &service.UsageReportRequest_Usage{
			InputTokens:  usage.GetPromptTokens(),
			OutputTokens: usage.GetCompletionTokens(),
		},
		Mode: service.UsageReportRequest_MODE_PER_REQUEST,
	})
	if err != nil {
		slog.Warn("failed to report usage", "error", err)
		return nil
	}
	slog.Info("report usage", "model", request.GetModel(), "input_tokens", usage.GetPromptTokens(), "output_tokens", usage.GetCompletionTokens())

	return nil
}

func (f *UsageFilter) OnCompletionResponse(ctx context.Context, request object.LLMRequest, response object.LLMResponse) filters.RequestFilterResult {
	err := f.usageReport(ctx, request, response)
	if err == nil {
		return filters.NewOK()
	}

	return filters.NewFailed(err)
}

func (f *UsageFilter) OnCompletionStreamResponse(ctx context.Context, request object.LLMRequest, response object.LLMStreamResponse, responseChunk object.LLMChunkResponse) filters.RequestFilterResult {
	if responseChunk.IsUsage() {
		err := f.usageReport(ctx, request, response)
		if err == nil {
			return filters.NewOK()
		}

		return filters.NewFailed(err)
	}
	return filters.NewOK()
}
