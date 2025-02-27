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
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	filtersv1alpha1 "knoway.dev/api/filters/v1alpha1"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	llmv1alpha1 "knoway.dev/api/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/registry/route"
	"knoway.dev/pkg/route/manager"
)

// ModelRouteReconciler reconciles a ModelRoute object
type ModelRouteReconciler struct {
	client.Client

	Scheme    *runtime.Scheme
	LifeCycle bootkit.LifeCycle
}

// +kubebuilder:rbac:groups=llm.knoway.dev,resources=modelroutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llm.knoway.dev,resources=modelroutes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llm.knoway.dev,resources=modelroutes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ModelRoute object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *ModelRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	modelRoute := &llmv1alpha1.ModelRoute{}
	if err := r.Get(ctx, req.NamespacedName, modelRoute); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Log.Info("reconcile ModelRoute", "name", modelRoute.GetName(), "namespace", modelRoute.GetNamespace())

	rrs := r.getReconciles()
	if modelRoute.GetObjectMeta().GetDeletionTimestamp() != nil {
		rrs = r.getDeleteReconciles()
	}

	modelRoute.Status.Conditions = nil

	for _, rr := range rrs {
		typ := rr.typ

		err := rr.reconciler(ctx, modelRoute)
		if err != nil {
			if isModelRouteDeleted(modelRoute) &&
				shouldForceDeleteModelRoute(modelRoute) {
				continue
			}

			log.Log.Error(err, "ModelRoute reconcile error", "name", modelRoute.Name, "type", typ)
			setModelRouteStatusCondition(modelRoute, typ, false, err.Error())

			break
		} else {
			setModelRouteStatusCondition(modelRoute, typ, true, "")
		}
	}

	r.reconcilePhase(ctx, modelRoute)

	var after time.Duration
	if modelRoute.Status.Status == llmv1alpha1.Failed {
		after = 30 * time.Second //nolint:mnd
	}

	newModelRoute := &llmv1alpha1.ModelRoute{}
	if err := r.Get(ctx, req.NamespacedName, newModelRoute); err != nil {
		log.Log.Error(err, "reconcile ModelRoute", "name", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !routeStatusEqual(&ModelRouteStatus{ModelRouteStatus: &modelRoute.Status}, &ModelRouteStatus{ModelRouteStatus: &newModelRoute.Status}) {
		newModelRoute.Status = modelRoute.Status
		if err := r.Status().Update(ctx, newModelRoute); err != nil {
			log.Log.Error(err, "update ModelRoute status error", "name", modelRoute.GetName())
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *ModelRouteReconciler) mapCRDTargetsToBackends(ctx context.Context, targets []*routev1alpha1.RouteTarget) (map[string]Backend, error) {
	backends := make(map[string]Backend)

	for _, target := range targets {
		nsName := types.NamespacedName{
			Namespace: target.GetDestination().GetNamespace(),
			Name:      target.GetDestination().GetBackend(),
		}

		backend, err := getBackendFromNamespacedName(ctx, r.Client, nsName)
		if err != nil {
			return make(map[string]Backend), err
		}
		if backend == nil {
			backends[nsName.String()] = nil
			continue
		}

		backends[nsName.String()] = backend
	}

	return backends, nil
}

func (r *ModelRouteReconciler) reconcileRegister(ctx context.Context, modelRoute *llmv1alpha1.ModelRoute) error {
	modelName := modelRoute.Spec.ModelName

	removeBackendFunc := func() {
		if modelName != "" {
			route.RemoveMatchRoute(modelName)
		}
	}
	if isModelRouteDeleted(modelRoute) {
		removeBackendFunc()
		return nil
	}

	crdTargets := r.getModelRouteTargets(modelRoute)

	mBackends, err := r.mapCRDTargetsToBackends(ctx, crdTargets)
	if err != nil {
		return err
	}

	routeConfig := r.toRegisterRouteConfig(ctx, modelRoute, mBackends)

	mulErrs := &multierror.Error{}
	if routeConfig != nil {
		if err := route.RegisterMatchRouteWithConfig(routeConfig); err != nil {
			log.Log.Error(err, "Failed to register route", "route", modelName)
			mulErrs = multierror.Append(mulErrs, fmt.Errorf("failed to upsert ModelRoute %s route: %w", modelRoute.GetName(), err))
		}
	}

	if mulErrs.ErrorOrNil() != nil {
		removeBackendFunc()
	}

	return mulErrs.ErrorOrNil()
}

func (r *ModelRouteReconciler) reconcileDestinationHealthy(ctx context.Context, modelRoute *llmv1alpha1.ModelRoute) error {
	crdTargets := r.getModelRouteTargets(modelRoute)

	mBackends, err := r.mapCRDTargetsToBackends(ctx, crdTargets)
	if err != nil {
		return err
	}

	targetsStatus := make([]llmv1alpha1.ModelRouteStatusTarget, 0, len(mBackends))

	for _, target := range crdTargets {
		tns := target.GetDestination().GetNamespace()
		if tns == "" {
			tns = modelRoute.GetNamespace()
		}

		nsName := types.NamespacedName{
			Namespace: tns,
			Name:      target.GetDestination().GetBackend(),
		}

		backend, ok := mBackends[nsName.String()]
		if !ok || lo.IsNil(backend) {
			targetsStatus = append(targetsStatus, llmv1alpha1.ModelRouteStatusTarget{
				Namespace: nsName.Namespace,
				Backend:   nsName.Name,
				ModelName: "",
				Status:    llmv1alpha1.Failed,
			})

			continue
		}

		targetsStatus = append(targetsStatus, llmv1alpha1.ModelRouteStatusTarget{
			Namespace: nsName.Namespace,
			Backend:   nsName.Name,
			ModelName: backend.GetModelName(),
			Status:    backend.GetStatus().GetStatus(),
		})
	}

	modelRoute.Status.Targets = targetsStatus

	return nil
}

func (r *ModelRouteReconciler) reconcilePhase(_ context.Context, modelRoute *llmv1alpha1.ModelRoute) {
	reconcileModelRoutePhase(modelRoute)
}

func (r *ModelRouteReconciler) getReconciles() []reconcileHandler[*llmv1alpha1.ModelRoute] {
	rhs := []reconcileHandler[*llmv1alpha1.ModelRoute]{
		{
			typ:        condConfig,
			reconciler: r.reconcileConfig,
		},
		{
			typ:        condValidator,
			reconciler: r.reconcileValidator,
		},
		{
			typ:        condDestinationHealthy,
			reconciler: r.reconcileDestinationHealthy,
		},
		{
			typ:        condRegister,
			reconciler: r.reconcileRegister,
		},
	}

	return rhs
}

func (r *ModelRouteReconciler) getDeleteReconciles() []reconcileHandler[*llmv1alpha1.ModelRoute] {
	rhs := []reconcileHandler[*llmv1alpha1.ModelRoute]{
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

func (r *ModelRouteReconciler) reconcileConfig(ctx context.Context, backend *llmv1alpha1.ModelRoute) error {
	if len(backend.Finalizers) == 0 {
		backend.Finalizers = []string{KnowayFinalzer}
		if err := r.Update(ctx, backend.DeepCopy()); err != nil {
			log.Log.Error(err, "update cluster finalizer error")
			return err
		}
	}

	return nil
}

func (r *ModelRouteReconciler) reconcileFinalDelete(ctx context.Context, modelRoute *llmv1alpha1.ModelRoute) error {
	canDelete := true

	for _, con := range modelRoute.Status.Conditions {
		if strings.Contains(con.Type, deleteCondPrefix) && con.Status == metav1.ConditionFalse {
			canDelete = false
		}
	}

	if !canDelete && !shouldForceDeleteModelRoute(modelRoute) {
		return errors.New("have delete condition not ready")
	}

	modelRoute.Finalizers = nil
	if err := r.Update(ctx, modelRoute); err != nil {
		log.Log.Error(err, "update ModelRoute finalizer error")
		return err
	}

	log.Log.Info("remove ModelRoute finalizer", "name", modelRoute.GetName())

	return nil
}

func (r *ModelRouteReconciler) reconcileValidator(ctx context.Context, modelRoute *llmv1alpha1.ModelRoute) error {
	if modelRoute.Spec.ModelName == "" {
		return errors.New("spec.modelName cannot be empty")
	}
	if modelRoute.Spec.Route != nil && len(modelRoute.Spec.Route.Targets) != 0 {
		for index, target := range modelRoute.Spec.Route.Targets {
			if target.Destination.Backend == "" {
				return fmt.Errorf("spec.route.targets[%d].destination.backend cannot be empty", index)
			}
			if target.Destination.Weight != nil && *target.Destination.Weight < 0 {
				return fmt.Errorf("spec.route.targets[%d].destination.weight cannot be less than 0", index)
			}
		}

		if !(lo.EveryBy(modelRoute.Spec.Route.Targets, func(target llmv1alpha1.ModelRouteRouteTarget) bool {
			return target.Destination.Weight == nil
		}) || lo.EveryBy(modelRoute.Spec.Route.Targets, func(target llmv1alpha1.ModelRouteRouteTarget) bool {
			return target.Destination.Weight != nil && *target.Destination.Weight >= 0
		})) {
			return errors.New("spec.route.targets.[].destination.weight must be either all set or all unset")
		}
	}
	if modelRoute.Spec.Fallback != nil {
		if modelRoute.Spec.Fallback.PostDelay != nil && *modelRoute.Spec.Fallback.PostDelay < 0 {
			return errors.New("spec.fallback.postDelay must be greater than or equal to 0")
		}
		if modelRoute.Spec.Fallback.PreDelay != nil && *modelRoute.Spec.Fallback.PreDelay < 0 {
			return errors.New("spec.fallback.preDelay must be greater than or equal to 0")
		}
		if modelRoute.Spec.Fallback.MaxRetires != nil && *modelRoute.Spec.Fallback.MaxRetires <= 0 {
			return errors.New("spec.fallback.maxRetries must be greater than 0")
		}
	}

	allExistingBackend := &llmv1alpha1.ModelRouteList{}
	if err := r.Client.List(ctx, allExistingBackend); err != nil {
		return fmt.Errorf("failed to list ModelRoute resources: %w", err)
	}

	for _, existing := range allExistingBackend.Items {
		if existing.Spec.ModelName == modelRoute.Spec.ModelName && existing.Name != modelRoute.Name {
			return fmt.Errorf("ModelRoute modelName and name '%s' must be unique globally", modelRoute.Spec.ModelName)
		}
	}

	crdTargets := r.getModelRouteTargets(modelRoute)

	mBackends, err := r.mapCRDTargetsToBackends(ctx, crdTargets)
	if err != nil {
		return err
	}

	// validator cluster filter by new
	routeConfig := r.toRegisterRouteConfig(ctx, modelRoute, mBackends)

	_, err = manager.NewWithConfig(routeConfig)
	if err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}

	return nil
}

func (r *ModelRouteReconciler) getModelRouteTargets(modelRoute *llmv1alpha1.ModelRoute) []*routev1alpha1.RouteTarget {
	if modelRoute.Spec.Route != nil && len(modelRoute.Spec.Route.Targets) != 0 {
		areAllWeightsUnset := true
		targets := make([]*routev1alpha1.RouteTarget, 0, len(modelRoute.Spec.Route.Targets))

		for _, target := range modelRoute.Spec.Route.Targets {
			var weight *int32
			if target.Destination.Weight != nil {
				areAllWeightsUnset = false
				weight = lo.ToPtr(int32(lo.FromPtr(target.Destination.Weight)))
			}

			tns := target.Destination.Namespace
			if tns == "" {
				tns = modelRoute.GetNamespace()
			}
			targets = append(targets, &routev1alpha1.RouteTarget{
				Destination: &routev1alpha1.RouteDestination{
					Namespace: tns,
					Backend:   target.Destination.Backend,
					Weight:    weight,
				},
			})
		}

		if areAllWeightsUnset {
			lo.ForEach(targets, func(target *routev1alpha1.RouteTarget, _ int) {
				target.Destination.Weight = lo.ToPtr(int32(1))
			})
		}

		return targets
	}

	return make([]*routev1alpha1.RouteTarget, 0)
}

func (r *ModelRouteReconciler) mapModelRouteTargetsToBackends(targets []*routev1alpha1.RouteTarget, mBackends map[string]Backend) []*routev1alpha1.RouteTarget {
	backends := make([]*routev1alpha1.RouteTarget, 0, len(targets))

	for _, target := range targets {
		nsName := types.NamespacedName{
			Namespace: target.GetDestination().GetNamespace(),
			Name:      target.GetDestination().GetBackend(),
		}

		backend, ok := mBackends[nsName.String()]
		if !ok || lo.IsNil(backend) {
			continue
		}

		backends = append(backends, &routev1alpha1.RouteTarget{
			Destination: &routev1alpha1.RouteDestination{
				Namespace: target.GetDestination().GetNamespace(),
				Backend:   target.GetDestination().GetBackend(),
				Cluster:   backend.GetModelName(),
				Weight:    target.GetDestination().Weight, //nolint:protogetter
			},
		})
	}

	return backends
}

func (r *ModelRouteReconciler) buildRateLimitPolicies(rateLimits []*llmv1alpha1.RateLimitRule) []*filtersv1alpha1.RateLimitPolicy {
	if len(rateLimits) == 0 {
		return nil
	}

	res := make([]*filtersv1alpha1.RateLimitPolicy, 0)

	for _, rateLimit := range rateLimits {
		var pMatch *filtersv1alpha1.StringMatch
		if rateLimit.Match != nil {
			if rateLimit.Match.Exact != "" {
				pMatch = &filtersv1alpha1.StringMatch{
					Match: &filtersv1alpha1.StringMatch_Exact{Exact: rateLimit.Match.Exact},
				}
			} else if rateLimit.Match.Prefix != "" {
				pMatch = &filtersv1alpha1.StringMatch{
					Match: &filtersv1alpha1.StringMatch_Prefix{Prefix: rateLimit.Match.Exact},
				}
			}
		}

		res = append(res, &filtersv1alpha1.RateLimitPolicy{
			BasedOn:  MapCRDRateLimitBaseOnConfigRateLimitBaseOn(rateLimit.BasedOn),
			Limit:    int32(rateLimit.Limit),
			Duration: durationpb.New(time.Duration(rateLimit.Duration) * time.Second),
			Match:    pMatch,
		})
	}

	return res
}

func (r *ModelRouteReconciler) toRegisterRouteConfig(_ context.Context, modelRoute *llmv1alpha1.ModelRoute, mBackends map[string]Backend) *routev1alpha1.Route {
	if modelRoute == nil {
		return nil
	}

	modelName := modelRoute.Spec.ModelName

	loadBalancePolicy := routev1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED
	if modelRoute.Spec.Route != nil {
		loadBalancePolicy = MapCRDLoadBalancePolicyModelConfigLoadBalancePolicy(modelRoute.Spec.Route.LoadBalancePolicy)
	}

	var filters []*routev1alpha1.RouteFilter
	if modelRoute.Spec.RateLimit != nil {
		filters = make([]*routev1alpha1.RouteFilter, 0)

		filters = append(filters, &routev1alpha1.RouteFilter{
			Name: "route-rate-limits",
			Config: lo.Must(anypb.New(&filtersv1alpha1.RateLimitConfig{
				Policies: r.buildRateLimitPolicies(modelRoute.Spec.RateLimit.Rules),
			})),
		})
	}

	var fallback *routev1alpha1.RouteFallback
	if modelRoute.Spec.Fallback != nil {
		fallback = &routev1alpha1.RouteFallback{}

		if modelRoute.Spec.Fallback.PreDelay != nil {
			fallback.PreDelay = durationpb.New(time.Duration(*modelRoute.Spec.Fallback.PreDelay) * time.Second)
		}
		if modelRoute.Spec.Fallback.PostDelay != nil {
			fallback.PostDelay = durationpb.New(time.Duration(*modelRoute.Spec.Fallback.PostDelay) * time.Second)
		}
		if modelRoute.Spec.Fallback.MaxRetires != nil {
			fallback.MaxRetries = modelRoute.Spec.Fallback.MaxRetires
		}
	}

	return &routev1alpha1.Route{
		Name: modelName,
		Matches: []*routev1alpha1.Match{
			{
				Model: &routev1alpha1.StringMatch{
					Match: &routev1alpha1.StringMatch_Exact{
						Exact: modelName,
					},
				},
			},
		},
		LoadBalancePolicy: loadBalancePolicy,
		Targets:           r.mapModelRouteTargetsToBackends(r.getModelRouteTargets(modelRoute), mBackends),
		Filters:           filters,
		Fallback:          fallback,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llmv1alpha1.ModelRoute{}).
		Named("modelroute").
		Complete(r)
}
