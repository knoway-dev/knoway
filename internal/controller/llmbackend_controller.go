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
	"net/url"
	"strings"
	"time"

	"github.com/stoewer/go-strcase"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knoway.dev/pkg/clusters/manager"

	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/registry/route"

	"github.com/hashicorp/go-multierror"
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

// +kubebuilder:rbac:groups=llm.knoway.dev,resources=llmbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llm.knoway.dev,resources=llmbackends/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llm.knoway.dev,resources=llmbackends/finalizers,verbs=update

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
		log.Log.Error(err, "reconcile LLMBackend", "name", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Log.Info("reconcile LLMBackend modelName", "modelName", llmBackend.Spec.ModelName)

	rrs := r.getReconciles()
	if isDeleted(llmBackend) {
		rrs = r.getDeleteReconciles()
	}
	llmBackend.Status.Conditions = nil
	for _, rr := range rrs {
		typ := rr.typ
		err := rr.reconciler(ctx, llmBackend)
		if err != nil {
			if isDeleted(llmBackend) && shouldForceDelete(llmBackend) {
				continue
			}
			log.Log.Error(err, "llmBackend reconcile error", "name", llmBackend.Name, "type", typ)

			setStatusCondition(llmBackend, typ, false, err.Error())
			break
		} else {
			setStatusCondition(llmBackend, typ, true, "")
		}
	}

	_ = r.reconcilePhase(ctx, llmBackend)

	var after time.Duration
	if llmBackend.Status.Status == knowaydevv1alpha1.Failed {
		after = 30 * time.Second
	}

	if err := r.Status().Update(ctx, llmBackend); err != nil {
		log.Log.Error(err, "update llmBackend status error", "name", llmBackend.GetName())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *LLMBackendReconciler) reconcileRegister(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	mName := llmBackend.Spec.ModelName

	removeBackendFunc := func() {
		if mName != "" {
			cluster.RemoveCluster(&v1alpha1.Cluster{
				Name: mName,
			})
			route.RemoveRoute(mName)
		}
	}
	if isDeleted(llmBackend) {
		removeBackendFunc()
		return nil
	}

	clusterCfg, err := llmBackendToClusterCfg(llmBackend)
	if err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	routeCfg := route.InitDirectModelRoute(mName)

	mulErrs := &multierror.Error{}
	if clusterCfg != nil {
		err = cluster.UpsertAndRegisterCluster(clusterCfg, r.LifeCycle)
		if err != nil {
			log.Log.Error(err, "Failed to upsert llmBackend", "cluster", clusterCfg)
			multierror.Append(mulErrs, fmt.Errorf("failed to upsert llmBackend %s: %v", llmBackend.GetName(), err))
		}

		if err = route.RegisterRouteWithConfig(routeCfg); err != nil {
			log.Log.Error(err, "Failed to register route", "route", mName)
			multierror.Append(mulErrs, fmt.Errorf("failed to upsert llmBackend %s route: %v", llmBackend.GetName(), err))
		}
	}

	if mulErrs.ErrorOrNil() != nil {
		removeBackendFunc()
	}

	return mulErrs.ErrorOrNil()
}

func (r *LLMBackendReconciler) reconcileValidator(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	if llmBackend.Spec.ModelName == "" {
		return fmt.Errorf("modelName cannot be empty")
	}

	if llmBackend.Spec.Upstream.BaseURL == "" {
		return fmt.Errorf("upstream.baseURL cannot be empty")
	}

	_, err := url.Parse(llmBackend.Spec.Upstream.BaseURL)
	if err != nil {
		return fmt.Errorf("upstream.baseURL parse error: %v", err)
	}

	// validator cluster filter by new
	clusterCfg, err := llmBackendToClusterCfg(llmBackend)
	if err != nil {
		return fmt.Errorf("failed to convert LLMBackend to cluster config: %w", err)
	}
	_, err = manager.NewWithConfigs(clusterCfg, nil)
	if err != nil {
		return fmt.Errorf("invalid cluster configuration: %w", err)
	}

	return nil
}

func (r *LLMBackendReconciler) reconcileUpstreamHealthy(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	// todo use model list api ?
	return nil
}

func (r *LLMBackendReconciler) reconcilePhase(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	llmBackend.Status.Status = knowaydevv1alpha1.Unknown
	if isDeleted(llmBackend) {
		llmBackend.Status.Status = knowaydevv1alpha1.Healthy
		return nil
	}

	for _, cond := range llmBackend.Status.Conditions {
		if cond.Status == metav1.ConditionFalse {
			llmBackend.Status.Status = knowaydevv1alpha1.Failed
			return nil
		}
	}
	return nil
}

func isDeleted(c *knowaydevv1alpha1.LLMBackend) bool {
	return c.DeletionTimestamp != nil
}

func setStatusCondition(llmBackend *knowaydevv1alpha1.LLMBackend, typ string, ready bool, message string) {
	cs := metav1.ConditionFalse
	if ready {
		cs = metav1.ConditionTrue
	}
	index := -1
	newCond := metav1.Condition{
		Type:               typ,
		Reason:             typ,
		Message:            message,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Status:             cs,
	}
	for i, cond := range llmBackend.Status.Conditions {
		if cond.Type == typ {
			index = i
			break
		}
	}
	if index == -1 {
		llmBackend.Status.Conditions = append(llmBackend.Status.Conditions, newCond)
	} else {
		old := llmBackend.Status.Conditions[index]
		if old.Status == newCond.Status && old.Message == newCond.Message {
			return
		}
		llmBackend.Status.Conditions[index] = newCond
	}
}

type reconcileHandler struct {
	typ        string
	reconciler func(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error
}

const (
	KnowayFinalzer = "knoway.dev"

	deleteCondPrefix = "delete-"
)

const (
	condConfig          = "config"
	condValidator       = "validator"
	condUpstreamHealthy = "upstreamHealthy"
	condRegister        = "register"
	condFinalDelete     = "finalDelete"
)

func (r *LLMBackendReconciler) getReconciles() []reconcileHandler {
	rhs := []reconcileHandler{
		{
			typ:        condConfig,
			reconciler: r.reconcileConfig,
		},
		{
			typ:        condValidator,
			reconciler: r.reconcileValidator,
		},
		{
			typ:        condUpstreamHealthy,
			reconciler: r.reconcileUpstreamHealthy,
		},
		{
			typ:        condRegister,
			reconciler: r.reconcileRegister,
		},
	}
	if reconcilesHook != nil {
		return reconcilesHook(rhs)
	}
	return rhs
}

func (r *LLMBackendReconciler) getDeleteReconciles() []reconcileHandler {
	rhs := []reconcileHandler{
		{
			typ:        condConfig,
			reconciler: r.reconcileConfig,
		},
		{
			typ:        strcase.LowerCamelCase(deleteCondPrefix + condRegister),
			reconciler: r.reconcileRegister,
		},
		{
			typ:        condFinalDelete,
			reconciler: r.reconcileFinalDelete,
		},
	}
	if reconcilesHook != nil {
		return reconcilesHook(rhs)
	}
	return rhs
}

// just for test
var reconcilesHook func([]reconcileHandler) []reconcileHandler

func (r *LLMBackendReconciler) reconcileConfig(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	if len(llmBackend.Finalizers) == 0 {
		llmBackend.Finalizers = []string{KnowayFinalzer}
		if err := r.Update(ctx, llmBackend.DeepCopy()); err != nil {
			log.Log.Error(err, "update cluster finalizer error")
			return err
		}
	}
	return nil
}

const graceDeletePeriod = time.Minute * 10

func shouldForceDelete(llmBackend *knowaydevv1alpha1.LLMBackend) bool {
	if llmBackend.DeletionTimestamp == nil {
		return false
	}
	return llmBackend.DeletionTimestamp.Add(graceDeletePeriod).Before(time.Now())
}

func (r *LLMBackendReconciler) reconcileFinalDelete(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	canDelete := true
	for _, con := range llmBackend.Status.Conditions {
		if strings.Contains(con.Type, deleteCondPrefix) && con.Status == metav1.ConditionFalse {
			canDelete = false
		}
	}

	if !canDelete && !shouldForceDelete(llmBackend) {
		return fmt.Errorf("have delete condition not ready")
	}
	llmBackend.Finalizers = nil
	if err := r.Update(ctx, llmBackend); err != nil {
		log.Log.Error(err, "update llmBackend finalizer error")
		return err
	}
	log.Log.Info("remove llmBackend finalizer", "name", llmBackend.GetName())
	return nil
}

func llmBackendToClusterCfg(backend *knowaydevv1alpha1.LLMBackend) (*v1alpha1.Cluster, error) {
	if backend == nil {
		return nil, nil
	}

	mName := backend.Spec.ModelName

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
			// todo override request param
			Url:     backend.Spec.Upstream.BaseURL,
			Headers: hs,
			Timeout: backend.Spec.Upstream.Timeout,
		},
		Filters: filters,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMBackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knowaydevv1alpha1.LLMBackend{}).
		Complete(r)
}
