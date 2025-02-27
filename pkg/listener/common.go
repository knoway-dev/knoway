package listener

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"

	"knoway.dev/pkg/clusters"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/registry/route"
	"knoway.dev/pkg/types/openai"
)

const (
	defaultRouteFallbackMaxRetries uint64 = 3
)

func CommonListenerHandler(
	filters filters.RequestFilters,
	reversedFilters filters.RequestFilters,
	parseRequest func(request *http.Request) (object.LLMRequest, error),
	doRequest func(ctx context.Context, cluster clusters.Cluster, writer http.ResponseWriter, request *http.Request, llmRequest object.LLMRequest) (object.LLMResponse, error),
) func(writer http.ResponseWriter, request *http.Request) (any, error) {
	return func(writer http.ResponseWriter, request *http.Request) (any, error) {
		var err error

		for _, f := range filters.OnRequestPreFilters() {
			fResult := f.OnRequestPre(request.Context(), request)
			if fResult.IsFailed() {
				return nil, fResult.Error
			}
		}

		var resp object.LLMResponse

		defer func() {
			for _, f := range reversedFilters.OnResponsePostFilters() {
				f.OnResponsePost(request.Context(), request, resp, err)
			}
		}()

		llmRequest, err := parseRequest(request)
		if err != nil {
			return nil, err
		}

		rMeta := metadata.RequestMetadataFromCtx(request.Context())
		rMeta.RequestModel = llmRequest.GetModel()

		matchedRoute := route.FindRoute(request.Context(), llmRequest)
		if matchedRoute == nil || matchedRoute.GetRouteConfig() == nil {
			return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
		}

		rMeta.MatchRoute = matchedRoute
		routeConfig := matchedRoute.GetRouteConfig()

		switch llmRequest.GetRequestType() {
		case object.RequestTypeChatCompletions, object.RequestTypeCompletions:
			for _, f := range filters.OnCompletionRequestFilters() {
				fResult := f.OnCompletionRequest(request.Context(), llmRequest, request)
				if fResult.IsFailed() {
					return nil, fResult.Error
				}
			}
		case object.RequestTypeImageGenerations:
			for _, f := range filters.OnImageGenerationsRequestFilters() {
				fResult := f.OnImageGenerationsRequest(request.Context(), llmRequest, request)
				if fResult.IsFailed() {
					return nil, fResult.Error
				}
			}
		}

		var retriedCount uint64

		// Fallback loop
		for {
			// Re-select cluster from route (by Load Balancer)
			selected, err := matchedRoute.SelectCluster(request.Context(), llmRequest)
			if err != nil {
				return nil, err
			}
			if selected == nil {
				return nil, openai.NewErrorModelNotFoundOrNotAccessible(llmRequest.GetModel())
			}

			rMeta.SelectedCluster = mo.Some(selected)

			if routeConfig.GetFallback() != nil && routeConfig.GetFallback().GetPreDelay() != nil && retriedCount > 0 {
				time.Sleep(routeConfig.GetFallback().GetPreDelay().AsDuration())
			}

			resp, err = doRequest(request.Context(), rMeta.SelectedCluster.MustGet(), writer, request, llmRequest)

			switch llmRequest.GetRequestType() {
			case object.RequestTypeChatCompletions, object.RequestTypeCompletions:
				if !llmRequest.IsStream() && !lo.IsNil(resp) {
					for _, f := range reversedFilters.OnCompletionResponseFilters() {
						fResult := f.OnCompletionResponse(request.Context(), llmRequest, resp)
						if fResult.IsFailed() {
							slog.Error("error occurred during invoking of OnCompletionResponse filters", "error", fResult.Error)
						}
					}
				}
			case object.RequestTypeImageGenerations:
				if !lo.IsNil(resp) {
					for _, f := range reversedFilters.OnImageGenerationsResponseFilters() {
						fResult := f.OnImageGenerationsResponse(request.Context(), llmRequest, resp)
						if fResult.IsFailed() {
							slog.Error("error occurred during invoking of OnImageGenerationsResponse filters", "error", fResult.Error)
						}
					}
				}
			}

			rMeta.ResponseModel = llmRequest.GetModel()
			if err == nil {
				return resp, err
			}

			if routeConfig.GetFallback() == nil {
				return resp, err
			}
			if routeConfig.GetFallback().GetPostDelay() != nil {
				time.Sleep(routeConfig.GetFallback().GetPostDelay().AsDuration())
			}
			if routeConfig.GetFallback().MaxRetries != nil {
				if retriedCount >= lo.CoalesceOrEmpty(routeConfig.GetFallback().GetMaxRetries(), defaultRouteFallbackMaxRetries) {
					return resp, err
				}

				retriedCount++

				continue
			}
		}
	}
}
