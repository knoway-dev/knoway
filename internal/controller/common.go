package controller

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"knoway.dev/api/clusters/v1alpha1"
	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

func isBackendDeleted(backend Backend) bool {
	return backend.GetObjectObjectMeta().DeletionTimestamp != nil
}

func shouldForceDeleteBackend(backend Backend) bool {
	if backend.GetObjectObjectMeta().DeletionTimestamp == nil {
		return false
	}

	return backend.GetObjectObjectMeta().DeletionTimestamp.Add(graceDeletePeriod).Before(time.Now())
}

func isModelRouteDeleted(modelRoute *knowaydevv1alpha1.ModelRoute) bool {
	return modelRoute.ObjectMeta.GetDeletionTimestamp() != nil
}

func shouldForceDeleteModelRoute(modelRoute *knowaydevv1alpha1.ModelRoute) bool {
	if modelRoute.ObjectMeta.GetDeletionTimestamp() == nil {
		return false
	}

	return modelRoute.ObjectMeta.GetDeletionTimestamp().Add(graceDeletePeriod).Before(time.Now())
}

func modelNameOrNamespacedName[B *knowaydevv1alpha1.LLMBackend | *knowaydevv1alpha1.ImageGenerationBackend | knowaydevv1alpha1.LLMBackend | knowaydevv1alpha1.ImageGenerationBackend](backend B) string {
	switch v := any(backend).(type) {
	case *knowaydevv1alpha1.LLMBackend:
		if lo.IsNil(v) {
			return ""
		}
		if v.Spec.ModelName != nil {
			return *v.Spec.ModelName
		}

		return fmt.Sprintf("%s/%s", v.Namespace, v.Name)
	case knowaydevv1alpha1.LLMBackend:
		if v.Spec.ModelName != nil {
			return *v.Spec.ModelName
		}

		return fmt.Sprintf("%s/%s", v.Namespace, v.Name)
	case *knowaydevv1alpha1.ImageGenerationBackend:
		if lo.IsNil(v) {
			return ""
		}
		if v.Spec.ModelName != nil {
			return *v.Spec.ModelName
		}

		return fmt.Sprintf("%s/%s", v.Namespace, v.Name)
	case knowaydevv1alpha1.ImageGenerationBackend:
		if v.Spec.ModelName != nil {
			return *v.Spec.ModelName
		}

		return fmt.Sprintf("%s/%s", v.Namespace, v.Name)
	default:
		panic("unknown backend type :" + fmt.Sprintf("%T", backend))
	}
}

func statusEqual[S comparable](status1, status2 Statusable[S]) bool {
	if len(status1.GetConditions()) != len(status2.GetConditions()) {
		return false
	}

	for i := range status1.GetConditions() {
		cond1 := status1.GetConditions()[i]
		cond2 := status2.GetConditions()[i]

		if cond1.Type != cond2.Type || cond1.Status != cond2.Status || cond1.Reason != cond2.Reason || cond1.Message != cond2.Message {
			return false
		}
	}

	return status1.GetStatus() == status2.GetStatus()
}

func routeStatusEqual[S comparable](status1, status2 RouteStatusable[S]) bool {
	if !statusEqual(status1, status2) {
		return false
	}

	if len(status1.GetTargetsStatus()) != len(status2.GetTargetsStatus()) {
		return false
	}

	for i := range status1.GetTargetsStatus() {
		target1 := status1.GetTargetsStatus()[i]
		target2 := status2.GetTargetsStatus()[i]

		if target1.Namespace != target2.Namespace || target1.Backend != target2.Backend || target1.ModelName != target2.ModelName || target1.Status != target2.Status {
			return false
		}
	}

	return true
}

func setStatusCondition(backend Backend, typ string, ready bool, message string) {
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

	for i, cond := range backend.GetStatus().GetConditions() {
		if cond.Type == typ {
			index = i
			break
		}
	}
	if index == -1 {
		backend.GetStatus().SetConditions(append(backend.GetStatus().GetConditions(), newCond))
	} else {
		old := backend.GetStatus().GetConditions()[index]
		if old.Status == newCond.Status && old.Message == newCond.Message {
			return
		}

		backend.GetStatus().GetConditions()[index] = newCond
	}
}

func setModelRouteStatusCondition(modelRoute *knowaydevv1alpha1.ModelRoute, typ string, ready bool, message string) {
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

	for i, cond := range modelRoute.Status.Conditions {
		if cond.Type == typ {
			index = i
			break
		}
	}
	if index == -1 {
		modelRoute.Status.Conditions = append(modelRoute.Status.Conditions, newCond)
	} else {
		old := modelRoute.Status.Conditions[index]
		if old.Status == newCond.Status && old.Message == newCond.Message {
			return
		}

		modelRoute.Status.Conditions[index] = newCond
	}
}

func reconcileBackendPhase(backend Backend) {
	backend.GetStatus().SetStatus(knowaydevv1alpha1.Healthy)
	if isBackendDeleted(backend) {
		backend.GetStatus().SetStatus(knowaydevv1alpha1.Healthy)
		return
	}

	for _, cond := range backend.GetStatus().GetConditions() {
		if cond.Status == metav1.ConditionFalse {
			backend.GetStatus().SetStatus(knowaydevv1alpha1.Failed)
			return
		}
	}
}

func reconcileModelRoutePhase(modelRoute *knowaydevv1alpha1.ModelRoute) {
	modelRoute.Status.Status = knowaydevv1alpha1.Healthy
	if isModelRouteDeleted(modelRoute) {
		modelRoute.Status.Status = knowaydevv1alpha1.Healthy
		return
	}

	for _, cond := range modelRoute.Status.Conditions {
		if cond.Status == metav1.ConditionFalse {
			modelRoute.Status.Status = knowaydevv1alpha1.Failed
			return
		}
	}

	for _, target := range modelRoute.Status.Targets {
		if target.Status == knowaydevv1alpha1.Failed {
			modelRoute.Status.Status = knowaydevv1alpha1.Failed
			return
		}
	}
}

func processStruct(v interface{}, params map[string]*structpb.Value) error {
	val := reflect.ValueOf(v)

	// Ensure we have a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to struct, got %v", val.Kind())
	}

	// Get the element and type for iteration
	elem := val.Elem()
	typ := elem.Type()

	for i := range make([]int, elem.NumField()) {
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

		// Handle nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		var fieldValue interface{}
		if field.Kind() == reflect.Ptr {
			// Dereference pointer
			fieldValue = field.Elem().Interface()
		} else {
			fieldValue = field.Interface()
		}

		// Check if you need to convert to float
		if isFloatString := structField.Tag.Get("floatString"); isFloatString == "true" {
			if strVal, ok := fieldValue.(string); ok {
				if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
					fieldValue = floatVal
				} else {
					return fmt.Errorf("failed to convert string to float for field %s: %w", jsonKey, err)
				}
			}
		}

		// Handle nested struct fields
		if reflect.ValueOf(fieldValue).Kind() == reflect.Struct {
			// Get the pointer to the nested struct
			nestedStruct := field
			if field.Kind() != reflect.Ptr {
				nestedStruct = field.Addr()
			}

			nestedParams := make(map[string]*structpb.Value)
			if err := processStruct(nestedStruct.Interface(), nestedParams); err != nil {
				return err
			}

			// Convert nestedParams to *structpb.Struct
			structValue := &structpb.Struct{
				Fields: nestedParams,
			}

			// Convert to *structpb.Value
			value := structpb.NewStructValue(structValue)
			params[jsonKey] = value

			continue
		}

		// Convert fieldValue to *structpb.Value
		value, err := structpb.NewValue(fieldValue)
		if err != nil {
			return fmt.Errorf("failed to convert field %s to *structpb.Value: %w", jsonKey, err)
		}

		params[jsonKey] = value
	}

	return nil
}

func resolveHeaderFrom(ctx context.Context, c client.Client, namespace string, fromSource knowaydevv1alpha1.HeaderFromSource) (map[string]string, error) {
	var data map[string]string

	switch fromSource.RefType {
	case knowaydevv1alpha1.Secret:
		secret := &corev1.Secret{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: fromSource.RefName}, secret); err != nil {
			return nil, fmt.Errorf("failed to get Secret %s: %w", fromSource.RefName, err)
		}
		data = secret.StringData
	case knowaydevv1alpha1.ConfigMap:
		configMap := &corev1.ConfigMap{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: fromSource.RefName}, configMap); err != nil {
			return nil, fmt.Errorf("failed to get ConfigMap %s: %w", fromSource.RefName, err)
		}
		data = configMap.Data
	default:
		// noting
	}

	return data, nil
}

func headerFromSpec(ctx context.Context, c client.Client, namespace string, headers []knowaydevv1alpha1.Header, headersFrom []knowaydevv1alpha1.HeaderFromSource) ([]*v1alpha1.Upstream_Header, error) {
	hs := make([]*v1alpha1.Upstream_Header, 0)

	for _, h := range headers {
		if h.Key == "" || h.Value == "" {
			continue
		}

		hs = append(hs, &v1alpha1.Upstream_Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	for _, valueFrom := range headersFrom {
		data, err := resolveHeaderFrom(ctx, c, namespace, valueFrom)
		if err != nil {
			return nil, err
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
