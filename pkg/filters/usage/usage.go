package usage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	defaultUsageServerTimeout = 3 * time.Second
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

	if c.GetStatsServer().GetTimeout().AsDuration() <= 0 {
		c.StatsServer.Timeout = durationpb.New(defaultUsageServerTimeout)
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
var _ filters.OnImageGenerationsResponseFilter = (*UsageFilter)(nil)

type UsageFilter struct {
	filters.IsRequestFilter

	config      *v1alpha1.UsageStatsConfig
	conn        *grpc.ClientConn
	usageClient service.UsageStatsServiceClient
}

func (f *UsageFilter) usageReport(ctx context.Context, request object.LLMRequest, response object.LLMResponse) {
	usage := response.GetUsage()
	if lo.IsNil(usage) {
		slog.Warn("no usage in response", "model", request.GetModel())
		return
	}

	var apiKeyID string

	rMeta := metadata.RequestMetadataFromCtx(ctx)
	if rMeta != nil && rMeta.AuthInfo != nil {
		apiKeyID = rMeta.AuthInfo.GetApiKeyId()
	} else {
		slog.Warn("no auth info in context")
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), f.config.GetStatsServer().GetTimeout().AsDuration())
	defer cancel()

	switch request.GetRequestType() {
	case
		object.RequestTypeChatCompletions,
		object.RequestTypeCompletions:
		tokensUsage, ok := object.AsLLMTokensUsage(usage)
		if !ok {
			slog.Warn("failed to cast usage to LLMUsageTokens")
			break
		}

		_, err := f.usageClient.UsageReport(ctx, &service.UsageReportRequest{
			ApiKeyId:          apiKeyID,
			UserModelName:     request.GetModel(),
			UpstreamModelName: response.GetModel(),
			Usage: &service.UsageReportRequest_Usage{
				InputTokens:  tokensUsage.GetPromptTokens(),
				OutputTokens: tokensUsage.GetCompletionTokens(),
			},
			Mode: service.UsageReportRequest_MODE_PER_REQUEST,
		})
		if err != nil {
			slog.Warn("failed to report usage", slog.Any("error", err))
			return
		}

		slog.Info("report usage",
			slog.String("model", request.GetModel()),
			slog.Uint64("input_tokens", tokensUsage.GetPromptTokens()),
			slog.Uint64("output_tokens", tokensUsage.GetCompletionTokens()),
		)
	case
		object.RequestTypeImageGenerations:
		imagesUsage, ok := object.AsLLMImagesUsage(usage)
		if !ok {
			slog.Warn("failed to cast usage to LLMUsageImage")
			break
		}
		if len(imagesUsage.GetOutputImages()) == 0 {
			break
		}

		usageImage := &service.UsageReportRequest_UsageImage{
			Width:   imagesUsage.GetOutputImages()[0].GetWidth(),
			Height:  imagesUsage.GetOutputImages()[0].GetHeight(),
			Numbers: uint64(len(imagesUsage.GetOutputImages())), // REVIEW: what if n != len(imagesUsage.GetOutputImages()),
			Style:   imagesUsage.GetOutputImages()[0].GetStyle(),
			Quality: imagesUsage.GetOutputImages()[0].GetQuality(),
		}

		_, err := f.usageClient.UsageReport(ctx, &service.UsageReportRequest{
			ApiKeyId:          apiKeyID,
			UserModelName:     request.GetModel(),
			UpstreamModelName: response.GetModel(),
			Usage:             &service.UsageReportRequest_Usage{OutputImages: usageImage},
			Mode:              service.UsageReportRequest_MODE_PER_REQUEST,
		})
		if err != nil {
			slog.Warn("failed to report usage", slog.Any("error", err))
			return
		}

		slog.Info("report usage",
			slog.String("model", request.GetModel()),
			slog.Uint64("output_images", usageImage.GetNumbers()),
			slog.Uint64("width", usageImage.GetWidth()),
			slog.Uint64("height", usageImage.GetHeight()),
		)
	}
}

func (f *UsageFilter) OnCompletionResponse(ctx context.Context, request object.LLMRequest, response object.LLMResponse) filters.RequestFilterResult {
	f.usageReport(ctx, request, response)

	return filters.NewOK()
}

func (f *UsageFilter) OnCompletionStreamResponse(ctx context.Context, request object.LLMRequest, response object.LLMStreamResponse, responseChunk object.LLMChunkResponse) filters.RequestFilterResult {
	if !responseChunk.IsUsage() {
		return filters.NewOK()
	}

	f.usageReport(ctx, request, response)

	return filters.NewOK()
}

func (f *UsageFilter) OnImageGenerationsResponse(ctx context.Context, request object.LLMRequest, response object.LLMResponse) filters.RequestFilterResult {
	f.usageReport(ctx, request, response)

	return filters.NewOK()
}
