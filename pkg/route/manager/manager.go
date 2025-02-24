package manager

import (
	"context"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knoway.dev/pkg/filters/lbfilter/loadbanlance"

	"knoway.dev/pkg/clients"
	"knoway.dev/pkg/clients/gvr"

	"knoway.dev/api/route/v1alpha1"
	llmv1alpha1 "knoway.dev/api/v1alpha1"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/route"
)

type routeManager struct {
	cfg *v1alpha1.Route
	// filters []filters.RequestFilter
	route.Route
	lb    loadbanlance.LoadBalancer
	nsMap map[string]string
}

func NewWithConfig(cfg *v1alpha1.Route) (route.Route, error) {
	rm := &routeManager{
		cfg:   cfg,
		lb:    loadbanlance.New(cfg),
		nsMap: buildBackendNsMap(cfg),
	}

	return rm, nil
}

func (m *routeManager) Match(ctx context.Context, request object.LLMRequest) (string, bool) {
	var (
		clusterName string
		found       bool
	)

	defer func() {
		if found {
			m.lb.Done()
		}
	}()

	matches := m.cfg.GetMatches()
	if len(matches) == 0 {
		return "", false
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

		if !isModelRouteConfiguration(m.cfg) {
			continue
		}

		if backend := m.lb.Next(request); backend != "" {
			modelName := m.modelFromLlmBackend(backend)
			if modelName != "" {
				slog.Debug("found model route", "model", modelName, "backend", backend)
				clusterName = modelName
				found = true

				break
			}
		}
	}

	return clusterName, found
}

func isModelRouteConfiguration(cfg *v1alpha1.Route) bool {
	return cfg.GetLoadBalancePolicy() != v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED && len(cfg.GetTargets()) != 0
}

func (m *routeManager) modelFromLlmBackend(backend string) string {
	client, err := clients.GetClients().DynamicClient()
	if err != nil {
		slog.Error("failed to get dynamic client", "err", err.Error())
		return ""
	}

	utd, err := client.Resource(gvr.From(&llmv1alpha1.LLMBackend{})).Namespace(m.nsMap[backend]).Get(context.Background(), backend, metav1.GetOptions{})
	if err != nil {
		slog.Error("failed to get model route", "backend", backend, "err", err.Error())
		return ""
	}

	llmBackend, err := clients.FromUnstructured[llmv1alpha1.LLMBackend](utd)
	if err != nil {
		slog.Error("failed to convert unstructured to LLMBackend", "err", err.Error())
		return ""
	}
	if llmBackend.Spec.ModelName != nil {
		return *llmBackend.Spec.ModelName
	}

	return fmt.Sprintf("%s/%s", llmBackend.Namespace, llmBackend.Name)
}

func buildBackendNsMap(cfg *v1alpha1.Route) map[string]string {
	nsMap := make(map[string]string)
	if isModelRouteConfiguration(cfg) {
		for _, target := range cfg.GetTargets() {
			if target.GetDestination() == nil {
				continue
			}

			ns := target.GetDestination().GetNamespace()
			if ns == "" {
				ns = "public"
			}
			nsMap[target.GetDestination().GetBackend()] = ns
		}
	}

	return nsMap
}
