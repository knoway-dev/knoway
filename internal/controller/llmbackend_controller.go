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

	Scheme *runtime.Scheme
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
		cluster.UpsertAndRegisterCluster(*clusterCfg)
	}

	// todo Maintain status states, such as model health checks, and configure validate
	return ctrl.Result{}, nil
}

func llmBackendToClusterCfg(backend *knowaydevv1alpha1.LLMBackend) *v1alpha1.Cluster {
	if backend == nil {
		return nil
	}
	name := backend.GetName()

	// todo upstream, headers ....

	// filters
	var filters []*v1alpha1.ClusterFilter
	for _, fc := range backend.Spec.Filters {
		var fcConfig *anypb.Any
		if fc.UsageStats != nil {
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
		} else if fc.ModelRewrite != nil {
			rf := &v1alpha12.OpenAIModelNameRewriteConfig{
				ModelName: fc.ModelRewrite.ModelName,
			}
			us, err := anypb.New(rf)
			if err != nil {
				log.Log.Error(err, "Failed to create Any from UsageStatsConfig")
			} else {
				fcConfig = us
			}
		}
		if fcConfig != nil {
			filters = append(filters, &v1alpha1.ClusterFilter{
				Config: fcConfig,
			})
		}
	}
	return &v1alpha1.Cluster{
		Name:     name,
		Filters:  filters,
		Provider: backend.Spec.Provider,
		Created:  backend.GetCreationTimestamp().Unix(),
		// todo configurable to replace hard config
		LoadBalancePolicy: v1alpha1.LoadBalancePolicy_ROUND_ROBIN,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMBackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knowaydevv1alpha1.LLMBackend{}).
		Complete(r)
}
