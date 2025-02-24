package clients

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Clients interface {
	KubeClient() (kubernetes.Interface, error)
	DynamicClient() (dynamic.Interface, error)
}
