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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"knoway.dev/api/clusters/v1alpha1"
	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/clusters/cluster"
	clustermanager "knoway.dev/pkg/clusters/manager"
	routemanager "knoway.dev/pkg/route/manager"
)

// ImageGenerationBackendReconciler reconciles a ImageGenerationBackend object
type ImageGenerationBackendReconciler struct {
	client.Client

	Scheme    *runtime.Scheme
	LifeCycle bootkit.LifeCycle
}

// +kubebuilder:rbac:groups=llm.knoway.dev,resources=imagegenerationbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llm.knoway.dev,resources=imagegenerationbackends/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llm.knoway.dev,resources=imagegenerationbackends/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ImageGenerationBackend object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *ImageGenerationBackendReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	currentBackend := &knowaydevv1alpha1.ImageGenerationBackend{}
	if err := r.Get(ctx, req.NamespacedName, currentBackend); err != nil {
		log.Log.Error(err, "reconcile ImageGenerationBackend", "name", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Log.Info("reconcile ImageGenerationBackend modelName", "modelName", modelNameOrNamespacedName(currentBackend))

	rrs := r.getReconciles()
	if isBackendDeleted(BackendFromImageGenerationBackend(currentBackend)) {
		rrs = r.getDeleteReconciles()
	}

	currentBackend.Status.Conditions = nil

	for _, rr := range rrs {
		typ := rr.typ

		err := rr.reconciler(ctx, currentBackend)
		if err != nil {
			if isBackendDeleted(BackendFromImageGenerationBackend(currentBackend)) &&
				shouldForceDeleteBackend(BackendFromImageGenerationBackend(currentBackend)) {
				continue
			}

			log.Log.Error(err, "ImageGenerationBackend reconcile error", "name", currentBackend.Name, "type", typ)
			setStatusCondition(BackendFromImageGenerationBackend(currentBackend), typ, false, err.Error())

			break
		} else {
			setStatusCondition(BackendFromImageGenerationBackend(currentBackend), typ, true, "")
		}
	}

	r.reconcilePhase(ctx, currentBackend)

	var after time.Duration
	if currentBackend.Status.Status == knowaydevv1alpha1.Failed {
		after = 30 * time.Second //nolint:mnd
	}

	newBackend := &knowaydevv1alpha1.ImageGenerationBackend{}
	if err := r.Get(ctx, req.NamespacedName, newBackend); err != nil {
		log.Log.Error(err, "reconcile ImageGenerationBackend", "name", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !statusEqual(BackendFromImageGenerationBackend(currentBackend).GetStatus(), BackendFromImageGenerationBackend(newBackend).GetStatus()) {
		newBackend.Status = currentBackend.Status
		if err := r.Status().Update(ctx, newBackend); err != nil {
			log.Log.Error(err, "update ImageGenerationBackend status error", "name", currentBackend.GetName())
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *ImageGenerationBackendReconciler) reconcileRegister(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) error {
	modelName := modelNameOrNamespacedName(backend)

	removeBackendFunc := func() {
		if modelName != "" {
			clustermanager.RemoveCluster(&v1alpha1.Cluster{
				Name: modelName,
			})
			routemanager.RemoveBaseRoute(modelName)
		}
	}
	if isBackendDeleted(BackendFromImageGenerationBackend(backend)) {
		removeBackendFunc()
		return nil
	}

	clusterCfg, err := r.toRegisterClusterConfig(ctx, backend)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	routeCfg := routemanager.InitDirectModelRoute(modelName)

	mulErrs := &multierror.Error{}
	if clusterCfg != nil {
		err = clustermanager.UpsertAndRegisterCluster(clusterCfg, r.LifeCycle)
		if err != nil {
			log.Log.Error(err, "Failed to upsert ImageGenerationBackend", "cluster", clusterCfg)
			mulErrs = multierror.Append(mulErrs, fmt.Errorf("failed to upsert ImageGenerationBackend %s: %w", backend.GetName(), err))
		}

		if err = routemanager.RegisterBaseRouteWithConfig(routeCfg, r.LifeCycle); err != nil {
			log.Log.Error(err, "Failed to register route", "route", modelName)
			mulErrs = multierror.Append(mulErrs, fmt.Errorf("failed to upsert ImageGenerationBackend %s route: %w", backend.GetName(), err))
		}
	}

	if mulErrs.ErrorOrNil() != nil {
		removeBackendFunc()
	}

	return mulErrs.ErrorOrNil()
}

func (r *ImageGenerationBackendReconciler) reconcileUpstreamHealthy(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) error {
	// todo use model list api ?
	return nil
}

func (r *ImageGenerationBackendReconciler) reconcilePhase(_ context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) {
	reconcileBackendPhase(BackendFromImageGenerationBackend(backend))
}

func (r *ImageGenerationBackendReconciler) getReconciles() []reconcileHandler[*knowaydevv1alpha1.ImageGenerationBackend] {
	rhs := []reconcileHandler[*knowaydevv1alpha1.ImageGenerationBackend]{
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

	return rhs
}

func (r *ImageGenerationBackendReconciler) getDeleteReconciles() []reconcileHandler[*knowaydevv1alpha1.ImageGenerationBackend] {
	rhs := []reconcileHandler[*knowaydevv1alpha1.ImageGenerationBackend]{
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

	return rhs
}

func (r *ImageGenerationBackendReconciler) reconcileConfig(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) error {
	if len(backend.Finalizers) == 0 {
		backend.Finalizers = []string{KnowayFinalzer}
		if err := r.Update(ctx, backend.DeepCopy()); err != nil {
			log.Log.Error(err, "update cluster finalizer error")
			return err
		}
	}

	return nil
}

func (r *ImageGenerationBackendReconciler) reconcileFinalDelete(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) error {
	canDelete := true

	for _, con := range backend.Status.Conditions {
		if strings.Contains(con.Type, deleteCondPrefix) && con.Status == metav1.ConditionFalse {
			canDelete = false
		}
	}

	if !canDelete && !shouldForceDeleteBackend(BackendFromImageGenerationBackend(backend)) {
		return errors.New("have delete condition not ready")
	}

	backend.Finalizers = nil
	if err := r.Update(ctx, backend); err != nil {
		log.Log.Error(err, "update ImageGenerationBackend finalizer error")
		return err
	}

	log.Log.Info("remove ImageGenerationBackend finalizer", "name", backend.GetName())

	return nil
}

func (r *ImageGenerationBackendReconciler) reconcileValidator(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) error {
	if backend.Spec.ModelName != nil && *backend.Spec.ModelName == "" {
		return errors.New("spec.modelName cannot be empty")
	}
	if backend.Spec.Upstream.BaseURL == "" {
		return errors.New("upstream.baseUrl cannot be empty")
	}

	if _, err := url.Parse(backend.Spec.Upstream.BaseURL); err != nil {
		return fmt.Errorf("upstream.baseUrl parse error: %w", err)
	}

	allExistingBackend := &knowaydevv1alpha1.ImageGenerationBackendList{}
	if err := r.Client.List(ctx, allExistingBackend); err != nil {
		return fmt.Errorf("failed to list ImageGenerationBackend resources: %w", err)
	}

	imageGenerationBackendModelName := modelNameOrNamespacedName(backend)

	for _, existing := range allExistingBackend.Items {
		if modelNameOrNamespacedName(existing) == imageGenerationBackendModelName && existing.Name != backend.Name {
			return fmt.Errorf("ImageGenerationBackend name '%s' must be unique globally", imageGenerationBackendModelName)
		}
	}

	// validator cluster filter by new
	clusterCfg, err := r.toRegisterClusterConfig(ctx, backend)
	if err != nil {
		return fmt.Errorf("failed to convert ImageGenerationBackend to cluster config: %w", err)
	}

	_, err = cluster.NewWithConfigs(clusterCfg, nil)
	if err != nil {
		return fmt.Errorf("invalid cluster configuration: %w", err)
	}

	return nil
}

func (r *ImageGenerationBackendReconciler) toUpstreamHeaders(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) ([]*v1alpha1.Upstream_Header, error) {
	if backend == nil {
		return nil, nil
	}

	return headerFromSpec(ctx, r.Client, backend.GetNamespace(), backend.Spec.Upstream.Headers, backend.Spec.Upstream.HeadersFrom)
}

func parseImageGenerationBackendModelParams(modelParams *knowaydevv1alpha1.ImageGenerationModelParams, params map[string]*structpb.Value) error {
	if modelParams == nil {
		return nil
	}
	modelTypes := map[string]interface{}{
		"OpenAI": modelParams.OpenAI,
	}

	for name, model := range modelTypes {
		if !lo.IsNil(model) {
			if err := processStruct(model, params); err != nil {
				return fmt.Errorf("error processing %s params: %w", name, err)
			}
		}
	}

	return nil
}

func toImageGenerationBackendParams(backed *knowaydevv1alpha1.ImageGenerationBackend) (map[string]*structpb.Value, map[string]*structpb.Value, error) {
	var defaultParams, overrideParams map[string]*structpb.Value
	if backed == nil {
		return nil, nil, nil
	}

	defaultParams, overrideParams = make(map[string]*structpb.Value), make(map[string]*structpb.Value)

	if err := parseImageGenerationBackendModelParams(backed.Spec.Upstream.DefaultParams, defaultParams); err != nil {
		return nil, nil, fmt.Errorf("error processing DefaultParams: %w", err)
	}

	if err := parseImageGenerationBackendModelParams(backed.Spec.Upstream.OverrideParams, overrideParams); err != nil {
		return nil, nil, fmt.Errorf("error processing OverrideParams: %w", err)
	}

	return defaultParams, overrideParams, nil
}

func (r *ImageGenerationBackendReconciler) toRegisterClusterConfig(ctx context.Context, backend *knowaydevv1alpha1.ImageGenerationBackend) (*v1alpha1.Cluster, error) {
	if backend == nil {
		return nil, nil
	}

	modelName := modelNameOrNamespacedName(backend)

	hs, err := r.toUpstreamHeaders(ctx, backend)
	if err != nil {
		return nil, err
	}

	defaultParams, overrideParams, err := toImageGenerationBackendParams(backend)
	if err != nil {
		return nil, err
	}

	// filters
	var filters []*v1alpha1.ClusterFilter

	for _, fc := range backend.Spec.Filters {
		switch {
		case fc.Custom != nil:
			// TODO: Implement custom filter
			log.Log.Info("Discovered filter during registration of cluster", "type", "Custom", "cluster", backend.Name, "modelName", modelName)
		default:
			// TODO: Implement unknown filter
			log.Log.Info("Discovered filter during registration of cluster", "type", "Unknown", "cluster", backend.Name, "modelName", modelName)
		}
	}

	// usage
	var sizeFrom *v1alpha1.ClusterMeteringPolicy_SizeFrom
	if backend.Spec.MeteringPolicy != nil && backend.Spec.MeteringPolicy.SizeFrom != nil {
		sizeFrom = MapBackendSizeFromClusterSizeFrom(backend.Spec.MeteringPolicy.SizeFrom)
	}

	return &v1alpha1.Cluster{
		Type:     v1alpha1.ClusterType_IMAGE_GENERATION,
		Name:     modelName,
		Provider: MapBackendProviderToClusterProvider(backend.Spec.Provider),
		Created:  backend.GetCreationTimestamp().Unix(),

		// todo configurable to replace hard config
		LoadBalancePolicy: v1alpha1.LoadBalancePolicy_ROUND_ROBIN,

		Upstream: &v1alpha1.Upstream{
			Url:            backend.Spec.Upstream.BaseURL,
			Headers:        hs,
			Timeout:        backend.Spec.Upstream.Timeout,
			DefaultParams:  defaultParams,
			OverrideParams: overrideParams,
		},

		Filters: filters,
		MeteringPolicy: &v1alpha1.ClusterMeteringPolicy{
			SizeFrom: sizeFrom,
		},
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImageGenerationBackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knowaydevv1alpha1.ImageGenerationBackend{}).
		Named("imagegenerationbackend").
		Complete(r)
}
