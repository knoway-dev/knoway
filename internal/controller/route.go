package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

type ModelRouteStatus struct {
	*knowaydevv1alpha1.ModelRouteStatus
}

func (s *ModelRouteStatus) GetStatus() knowaydevv1alpha1.StatusEnum {
	return s.Status
}

func (s *ModelRouteStatus) SetStatus(status knowaydevv1alpha1.StatusEnum) {
	s.Status = status
}

func (s *ModelRouteStatus) GetConditions() []metav1.Condition {
	return s.Conditions
}

func (s *ModelRouteStatus) SetConditions(conditions []metav1.Condition) {
	s.Conditions = conditions
}

func (s *ModelRouteStatus) GetTargetsStatus() []knowaydevv1alpha1.ModelRouteStatusTarget {
	return s.Targets
}

func (s *ModelRouteStatus) SetTargetsStatus(targets []knowaydevv1alpha1.ModelRouteStatusTarget) {
	s.Targets = targets
}
