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
	"reflect"
	"strings"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/types/known/anypb"
	structpb2 "google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"knoway.dev/api/clusters/v1alpha1"
	v1alpha12 "knoway.dev/api/filters/v1alpha1"
	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/clusters/manager"
	"knoway.dev/pkg/registry/cluster"
	"knoway.dev/pkg/registry/route"
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
	log.Log.Info("reconcile LLMBackend modelName", "modelName", llmBackend.Spec.Name)

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

	r.reconcilePhase(ctx, llmBackend)

	var after time.Duration
	if llmBackend.Status.Status == knowaydevv1alpha1.Failed {
		after = 30 * time.Second //nolint:mnd
	}

	newBackend := &knowaydevv1alpha1.LLMBackend{}
	if err := r.Get(ctx, req.NamespacedName, newBackend); err != nil {
		log.Log.Error(err, "reconcile LLMBackend", "name", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !reflect.DeepEqual(llmBackend.Status, newBackend.Status) {
		newBackend.Status = llmBackend.Status
		if err := r.Status().Update(ctx, newBackend); err != nil {
			log.Log.Error(err, "update llmBackend status error", "name", llmBackend.GetName())
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *LLMBackendReconciler) reconcileRegister(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	mName := llmBackend.Spec.Name

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

	clusterCfg, err := r.toRegisterClusterConfig(ctx, llmBackend)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	routeCfg := route.InitDirectModelRoute(mName)

	mulErrs := &multierror.Error{}
	if clusterCfg != nil {
		err = cluster.UpsertAndRegisterCluster(clusterCfg, r.LifeCycle)
		if err != nil {
			log.Log.Error(err, "Failed to upsert llmBackend", "cluster", clusterCfg)
			mulErrs = multierror.Append(mulErrs, fmt.Errorf("failed to upsert llmBackend %s: %w", llmBackend.GetName(), err))
		}

		if err = route.RegisterRouteWithConfig(routeCfg); err != nil {
			log.Log.Error(err, "Failed to register route", "route", mName)
			mulErrs = multierror.Append(mulErrs, fmt.Errorf("failed to upsert llmBackend %s route: %w", llmBackend.GetName(), err))
		}
	}

	if mulErrs.ErrorOrNil() != nil {
		removeBackendFunc()
	}

	return mulErrs.ErrorOrNil()
}

func (r *LLMBackendReconciler) reconcileValidator(ctx context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) error {
	if llmBackend.Spec.Name == "" {
		return fmt.Errorf("spec.name cannot be empty")
	}
	if llmBackend.Spec.Upstream.BaseURL == "" {
		return errors.New("upstream.baseUrl cannot be empty")
	}

	_, err := url.Parse(llmBackend.Spec.Upstream.BaseURL)
	if err != nil {
		return fmt.Errorf("upstream.baseUrl parse error: %w", err)
	}

	existingLLMBackendList := &knowaydevv1alpha1.LLMBackendList{}
	err = r.Client.List(ctx, existingLLMBackendList)
	if err != nil {
		return fmt.Errorf("failed to list LLMBackend resources: %w", err)
	}

	for _, existingLLMBackend := range existingLLMBackendList.Items {
		if existingLLMBackend.Spec.Name == llmBackend.Spec.Name && existingLLMBackend.Name != llmBackend.Name {
			return fmt.Errorf("LLMBackend name '%s' must be unique globally", llmBackend.Spec.Name)
		}
	}

	// validator cluster filter by new
	clusterCfg, err := r.toRegisterClusterConfig(ctx, llmBackend)
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

func (r *LLMBackendReconciler) reconcilePhase(_ context.Context, llmBackend *knowaydevv1alpha1.LLMBackend) {
	llmBackend.Status.Status = knowaydevv1alpha1.Healthy
	if isDeleted(llmBackend) {
		llmBackend.Status.Status = knowaydevv1alpha1.Healthy
		return
	}

	for _, cond := range llmBackend.Status.Conditions {
		if cond.Status == metav1.ConditionFalse {
			llmBackend.Status.Status = knowaydevv1alpha1.Failed
			return
		}
	}
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
		return errors.New("have delete condition not ready")
	}

	llmBackend.Finalizers = nil
	if err := r.Update(ctx, llmBackend); err != nil {
		log.Log.Error(err, "update llmBackend finalizer error")
		return err
	}

	log.Log.Info("remove llmBackend finalizer", "name", llmBackend.GetName())

	return nil
}

func (r *LLMBackendReconciler) toUpstreamHeaders(ctx context.Context, backend *knowaydevv1alpha1.LLMBackend) ([]*v1alpha1.Upstream_Header, error) {
	if backend == nil {
		return nil, nil
	}
	hs := make([]*v1alpha1.Upstream_Header, 0)
	for _, h := range backend.Spec.Upstream.Headers {
		if h.Key == "" || h.Value == "" {
			continue
		}

		hs = append(hs, &v1alpha1.Upstream_Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	for _, valueFrom := range backend.Spec.Upstream.HeadersFrom {
		var data map[string]string

		switch valueFrom.RefType {
		case knowaydevv1alpha1.Secret:
			secret := &v1.Secret{}
			err := r.Client.Get(ctx, client.ObjectKey{Namespace: backend.GetNamespace(), Name: valueFrom.RefName}, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to get Secret %s: %w", valueFrom.RefName, err)
			}
			data = secret.StringData
		case knowaydevv1alpha1.ConfigMap:
			configMap := &v1.ConfigMap{}
			err := r.Client.Get(ctx, client.ObjectKey{Namespace: backend.GetNamespace(), Name: valueFrom.RefName}, configMap)
			if err != nil {
				return nil, fmt.Errorf("failed to get ConfigMap %s: %w", valueFrom.RefName, err)
			}
			data = configMap.Data
		default:
			// noting
		}
		for key, value := range data {
			hs = append(hs, &v1alpha1.Upstream_Header{
				Key:   valueFrom.Prefix + key,
				Value: value,
			})
		}
	}

	return hs, nil
}

// TODO: unit test
func processStruct(v interface{}, params map[string]*structpb.Value) error {
	val := reflect.ValueOf(v)

	// Ensure we have a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to struct, got %v", val.Kind())
	}

	// Get the element and type for iteration
	elem := val.Elem()
	typ := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		structField := typ.Field(i)

		// Handle inline struct fields (embedded fields)
		if structField.Anonymous {
			if err := processStruct(field.Addr().Interface(), params); err != nil {
				return err
			}
			continue
		}

		// Extract the JSON tag, skip if there's no tag or it's marked as "-"
		tag := structField.Tag.Get("json")
		if tag == "-" {
			continue
		}

		// Get the actual JSON key
		jsonKey := strings.Split(tag, ",")[0]

		// Handle the field value; if it's a pointer, and it may be a nil pointer, skip it
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		} else if field.Kind() == reflect.Ptr {
			field = field.Elem()
		} else if field.Kind() == reflect.String && field.IsZero() {
			continue
		}

		// Convert field.Interface() to *structpb.Value
		value, err := structpb2.NewValue(field.Interface())
		if err != nil {
			return fmt.Errorf("failed to convert field %s to *structpb.Value: %w", jsonKey, err)
		}

		params[jsonKey] = value
	}

	return nil
}

func parseModelParams(modelParams *knowaydevv1alpha1.ModelParams, params map[string]*structpb.Value) error {
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

func toParams(backed *knowaydevv1alpha1.LLMBackend) (defaultParams, overrideParams map[string]*structpb.Value, err error) {
	if backed == nil {
		return
	}

	defaultParams, overrideParams = make(map[string]*structpb.Value), make(map[string]*structpb.Value)

	if err = parseModelParams(backed.Spec.Upstream.DefaultParams, defaultParams); err != nil {
		return nil, nil, fmt.Errorf("error processing DefaultParams: %w", err)
	}

	if err = parseModelParams(backed.Spec.Upstream.OverrideParams, overrideParams); err != nil {
		return nil, nil, fmt.Errorf("error processing OverrideParams: %w", err)
	}

	return defaultParams, overrideParams, nil
}

func (r *LLMBackendReconciler) toRegisterClusterConfig(ctx context.Context, backend *knowaydevv1alpha1.LLMBackend) (*v1alpha1.Cluster, error) {
	if backend == nil {
		return nil, nil
	}

	mName := backend.Spec.Name
	hs, err := r.toUpstreamHeaders(ctx, backend)
	if err != nil {
		return nil, err
	}

	defaultParams, overrideParams, err := toParams(backend)
	if err != nil {
		return nil, err
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
			Url:            backend.Spec.Upstream.BaseURL,
			Headers:        hs,
			Timeout:        backend.Spec.Upstream.Timeout,
			DefaultParams:  defaultParams,
			OverrideParams: overrideParams,
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
