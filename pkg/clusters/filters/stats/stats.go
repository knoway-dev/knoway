package stats

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	service "knoway.dev/api/service/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clusterfilters "knoway.dev/pkg/clusters/filters"
	"knoway.dev/pkg/modules/auth"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"
)

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (clusterfilters.ClusterFilter, error) {
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

	return &usageStatsFilter{
		cfg:         c,
		usageClient: service.NewUsageStatsServiceClient(conn),
	}, nil
}

var _ clusterfilters.ClusterFilter = (*usageStatsFilter)(nil)
var _ clusterfilters.ClusterFilterResponseHandler = (*usageStatsFilter)(nil)

type usageStatsFilter struct {
	clusterfilters.IsClusterFilter

	cfg         *v1alpha1.UsageStatsConfig
	usageClient service.UsageStatsServiceClient
}

func (f *usageStatsFilter) ResponseComplete(ctx context.Context, request object.LLMRequest, response object.LLMResponse) error {
	usage := response.GetUsage()
	if usage == nil {
		slog.Warn("no usage in response")

		return nil
	}

	var apiKeyID string

	authInfo, ok := auth.GetAuthInfoFromCtx(ctx)
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

	return nil
}
