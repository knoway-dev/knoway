package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	llmv1alpha1 "knoway.dev/api/v1alpha1"
)

type Statusable[S any] interface {
	GetStatus() S
	SetStatus(status S)
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

type RouteStatusable[S any] interface {
	Statusable[S]

	GetTargetsStatus() []llmv1alpha1.ModelRouteStatusTarget
	SetTargetsStatus(targets []llmv1alpha1.ModelRouteStatusTarget)
}
