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

func isDeleted(backend Backend) bool {
	return backend.GetObjectObjectMeta().DeletionTimestamp != nil
}

func shouldForceDelete(backend Backend) bool {
	if backend.GetObjectObjectMeta().DeletionTimestamp == nil {
		return false
	}

	return backend.GetObjectObjectMeta().DeletionTimestamp.Add(graceDeletePeriod).Before(time.Now())
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

func reconcilePhase(backend Backend) {
	backend.GetStatus().SetStatus(knowaydevv1alpha1.Healthy)
	if isDeleted(backend) {
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
