package clients

import (
	"errors"
	"math"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"knoway.dev/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ Clients = (*clients)(nil)

type clients struct {
	Clients
	config *config.Config
}

func (c *clients) KubeClient() (kubernetes.Interface, error) {
	clusterConfig := NewRESTConfig(c.config.KubeConfig)
	return kubernetes.NewForConfig(clusterConfig)
}

func (c *clients) DynamicClient() (dynamic.Interface, error) {
	clusterConfig := NewRESTConfig(c.config.KubeConfig)
	return dynamic.NewForConfig(clusterConfig)
}

var cli Clients

func GetClients() Clients {
	return cli
}

func InitClients(conf *config.Config) {
	cs := &clients{
		config: conf,
	}
	cli = cs
}

func NewRESTConfig(kubeConfig string) *rest.Config {
	var (
		clusterConfig *rest.Config
		err           error
	)
	if kubeConfig != "" {
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		clusterConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatal("Failed to create kubernetes config: ", err)
	}
	clusterConfig.QPS = math.MaxInt64
	clusterConfig.Burst = 1000

	return clusterConfig
}

func FromUnstructured[T any](obj *unstructured.Unstructured) (*T, error) {
	typedObj := new(T)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), typedObj); err != nil {
		return nil, err
	}

	return typedObj, nil
}

func GetGVKOrDie(obj runtime.Object) schema.GroupVersionKind {
	gvk, err := GetGVK(obj)
	if err != nil {
		panic(err)
	}

	return gvk
}

func GetGVK(obj runtime.Object) (schema.GroupVersionKind, error) {
	gvks, _, _ := scheme.Scheme.ObjectKinds(obj)
	if len(gvks) < 1 {
		return schema.GroupVersionKind{}, errors.New("no gvk found")
	}

	return gvks[0], nil
}
