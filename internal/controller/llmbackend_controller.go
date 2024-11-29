/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"

	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/registry/route"

	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"knoway.dev/api/clusters/v1alpha1"
	v1alpha12 "knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/registry/cluster"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

// LLMBackendReconciler reconciles a LLMBackend object
type LLMBackendReconciler struct {
	client.Client

	Scheme    *runtime.Scheme
	LifeCycle bootkit.LifeCycle
}

// +kubebuilder:rbac:groups=knoway.dev.knoway.dev,resources=llmbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=knoway.dev.knoway.dev,resources=llmbackends/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=knoway.dev.knoway.dev,resources=llmbackends/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LLMBackend object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *LLMBackendReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	llmBackend := &knowaydevv1alpha1.LLMBackend{}
	if err := r.Get(ctx, req.NamespacedName, llmBackend); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Log.Info("LLMBackend modelName", "modelName", llmBackend.Spec.ModelName)

	clusterCfg := llmBackendToClusterCfg(llmBackend)
	if clusterCfg != nil {
		err := cluster.UpsertAndRegisterCluster(clusterCfg, r.LifeCycle)
		if err != nil {
			log.Log.Error(err, "Failed to upsert cluster", "cluster", clusterCfg, "error", err)
			return ctrl.Result{}, nil
		}

		mName := llmBackend.Spec.ModelName
		if err = route.RegisterRouteWithConfig(route.InitDirectModelRoute(mName)); err != nil {
			log.Log.Error(err, "Failed to register route", "route", mName, "error", err)
			return ctrl.Result{}, nil
		}
	}

	//todo Maintain status states, such as model health checks, and configure validate
	return ctrl.Result{}, nil
}

func stringToUpstreamMethod(method string) v1alpha1.Upstream_Method {
	upperMethod := strings.ToUpper(method)
	if value, ok := v1alpha1.Upstream_Method_value[upperMethod]; ok {
		return v1alpha1.Upstream_Method(value)
	}

	return v1alpha1.Upstream_METHOD_UNSPECIFIED
}

func llmBackendToClusterCfg(backend *knowaydevv1alpha1.LLMBackend) *v1alpha1.Cluster {
	if backend == nil {
		return nil
	}

	mName := backend.Spec.ModelName

	// Upstream
	server := backend.Spec.Upstream.Server
	url := fmt.Sprintf("%s:%s", server.Address, server.API)

	hs := make([]*v1alpha1.Upstream_Header, 0)
	for _, h := range backend.Spec.Upstream.Headers {
		// todo ValueFrom to value
		hs = append(hs, &v1alpha1.Upstream_Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	// filters
	var filters []*v1alpha1.ClusterFilter

	for _, fc := range backend.Spec.Filters {
		var fcConfig *anypb.Any

		switch {
		case fc.UsageStats != nil:
			c := &v1alpha12.UsageStatsConfig{
				StatsServer: &v1alpha12.UsageStatsConfig_StatsServer{
					Url: fc.UsageStats.Address,
				},
			}

			us, err := anypb.New(c)
			if err != nil {
				log.Log.Error(err, "Failed to create Any from UsageStatsConfig")
			} else {
				fcConfig = us
			}

			log.Log.Info("Discovered filter during registration of cluster", "type", "UsageStats", "cluster", backend.Name, "modelName", mName, "filter_name", fc.Name)
		case fc.Custom != nil:
			// TODO: Implement custom filter
			log.Log.Info("Discovered filter during registration of cluster", "type", "Custom", "cluster", backend.Name, "modelName", mName)
		default:
			// TODO: Implement unknown filter
			log.Log.Info("Discovered filter during registration of cluster", "type", "Unknown", "cluster", backend.Name, "modelName", mName)
		}

		if fcConfig != nil {
			filters = append(filters, &v1alpha1.ClusterFilter{
				Config: fcConfig,
			})
		}
	}

	return &v1alpha1.Cluster{
		Name:     mName,
		Provider: backend.Spec.Provider,
		Created:  backend.GetCreationTimestamp().Unix(),

		// todo configurable to replace hard config
		LoadBalancePolicy: v1alpha1.LoadBalancePolicy_ROUND_ROBIN,

		Upstream: &v1alpha1.Upstream{
			Url:     url,
			Method:  stringToUpstreamMethod(server.Method),
			Headers: hs,
			Timeout: backend.Spec.Upstream.Timeout,
		},
		Filters: filters,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMBackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knowaydevv1alpha1.LLMBackend{}).
		Complete(r)
}
