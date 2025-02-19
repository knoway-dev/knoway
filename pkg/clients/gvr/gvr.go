package gvr

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"knoway.dev/pkg/clients"
)

func gvkToGVR(gvk schema.GroupVersionKind) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
}

func From[T client.Object](o T) schema.GroupVersionResource {
	return gvkToGVR(clients.GetGVKOrDie(o))
}
