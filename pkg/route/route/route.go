package route

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/samber/lo"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/bootkit"
	clustermanager "knoway.dev/pkg/clusters/manager"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/config"
	"knoway.dev/pkg/route"
	"knoway.dev/pkg/route/loadbalance"
	"knoway.dev/pkg/types/openai"
	"knoway.dev/pkg/utils"
)

const (
	defaultRouteFallbackMaxRetries uint64 = 3
)

var _ route.Route = (*routeDefault)(nil)

type routeDefault struct {
	cfg                  *routev1alpha1.Route
	nsMap                map[string]string
	loadBalancer         loadbalance.LoadBalancer
	routeFilters         filters.RequestFilters
	reversedRouteFilters filters.RequestFilters
}

func NewWithConfig(cfg *routev1alpha1.Route, lifecycle bootkit.LifeCycle) (route.Route, error) {
	rm := &routeDefault{
		cfg:          cfg,
		nsMap:        buildBackendNsMap(cfg),
		loadBalancer: loadbalance.New(cfg),
	}

	for _, fc := range cfg.GetFilters() {
		f, err := config.NewRequestFilterWithConfig(fc.GetName(), fc.GetConfig(), lifecycle)
		if err != nil {
			return nil, err
		}

		rm.routeFilters = append(rm.routeFilters, f)
	}

	rm.reversedRouteFilters = utils.Clone(rm.routeFilters)

	return rm, nil
}

func (m *routeDefault) GetRouteConfig() *routev1alpha1.Route {
	return m.cfg
}

func (m *routeDefault) Match(ctx context.Context, request object.LLMRequest) bool {
	matches := m.GetRouteConfig().GetMatches()
	if len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		modelNameMatch := match.GetModel()
		if modelNameMatch == nil {
			continue
		}

		exactMatch := modelNameMatch.GetExact()
		if exactMatch == "" {
			continue
		}

		if request.GetModel() != exactMatch {
			continue
		}

		if len(m.GetRouteConfig().GetTargets()) == 0 {
			continue
		}

		return true
	}

	return false
}

func (m *routeDefault) HandleRequest(ctx context.Context, request object.LLMRequest) (object.LLMResponse, error) {
	rMeta := metadata.RequestMetadataFromCtx(ctx)

	if m.GetRouteConfig() == nil {
		return nil, object.NewErrorModelNotFoundOrNotAccessible(request.GetModel())
	}

	switch request.GetRequestType() {
	case object.RequestTypeChatCompletions, object.RequestTypeCompletions:
		for _, f := range m.routeFilters.OnCompletionRequestFilters() {
			fResult := f.OnCompletionRequest(ctx, request, request.GetRawRequest())
			if fResult.IsFailed() {
				return nil, fResult.Error
			}
		}
	case object.RequestTypeImageGenerations:
		for _, f := range m.routeFilters.OnImageGenerationsRequestFilters() {
			fResult := f.OnImageGenerationsRequest(ctx, request, request.GetRawRequest())
			if fResult.IsFailed() {
				return nil, fResult.Error
			}
		}
	}

	var retriedCount uint64

	// Fallback loop
	for {
		var clusterName string

		// default lb policy
		if m.cfg.GetLoadBalancePolicy() == routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED {
			clusterName = m.cfg.GetTargets()[0].GetDestination().GetCluster()
		} else {
			clusterName = m.loadBalancer.Next(ctx, request)
		}

		if m.cfg.GetFallback() != nil && m.cfg.GetFallback().GetPreDelay() != nil && retriedCount > 0 {
			time.Sleep(m.cfg.GetFallback().GetPreDelay().AsDuration())
		}

		resp, err := clustermanager.HandleRequest(ctx, clusterName, request)

		switch request.GetRequestType() {
		case object.RequestTypeChatCompletions, object.RequestTypeCompletions:
			if !request.IsStream() && !lo.IsNil(resp) {
				for _, f := range m.reversedRouteFilters.OnCompletionResponseFilters() {
					fResult := f.OnCompletionResponse(ctx, request, resp)
					if fResult.IsFailed() {
						slog.Error("error occurred during invoking of OnCompletionResponse filters", "error", fResult.Error)
					}
				}
			}
		case object.RequestTypeImageGenerations:
			if !lo.IsNil(resp) {
				for _, f := range m.reversedRouteFilters.OnImageGenerationsResponseFilters() {
					fResult := f.OnImageGenerationsResponse(ctx, request, resp)
					if fResult.IsFailed() {
						slog.Error("error occurred during invoking of OnImageGenerationsResponse filters", "error", fResult.Error)
					}
				}
			}
		}

		if !lo.IsNil(resp) && resp.IsStream() {
			if streamResp, ok := resp.(object.LLMStreamResponse); ok {
				streamResp.OnChunk(func(ctx context.Context, stream object.LLMStreamResponse, chunk object.LLMChunkResponse) {
					for _, f := range m.reversedRouteFilters.OnCompletionStreamResponseFilters() {
						fResult := f.OnCompletionStreamResponse(ctx, request, streamResp, chunk)
						if fResult.IsFailed() {
							// REVIEW: ignore? Or should fResult be returned?
							// Related topics: moderation, censorship, or filter keywords from the response
							slog.Error("error occurred during invoking of OnCompletionStreamResponse filters", "error", fResult.Error)
						}
					}
				})
			}
		}

		rMeta.ResponseModel = request.GetModel()

		if err == nil || errors.Is(err, openai.SkipStreamResponse) {
			return resp, err
		}

		if m.cfg.GetFallback() == nil {
			return resp, err
		}
		if m.cfg.GetFallback().GetPostDelay() != nil {
			time.Sleep(m.cfg.GetFallback().GetPostDelay().AsDuration())
		}
		if m.cfg.GetFallback().MaxRetries != nil {
			if retriedCount >= lo.CoalesceOrEmpty(m.cfg.GetFallback().GetMaxRetries(), defaultRouteFallbackMaxRetries) {
				return resp, err
			}

			retriedCount++

			continue
		}
	}
}

func buildBackendNsMap(cfg *routev1alpha1.Route) map[string]string {
	nsMap := make(map[string]string)

	for _, target := range cfg.GetTargets() {
		if target.GetDestination() == nil {
			continue
		}

		nsMap[target.GetDestination().GetBackend()] = target.GetDestination().GetNamespace()
	}

	return nsMap
}
