package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

type Backend interface {
	GetObjectObjectMeta() metav1.ObjectMeta
	GetStatus() BackendStatus
}

type BackendStatus interface {
	GetStatus() knowaydevv1alpha1.StatusEnum
	SetStatus(status knowaydevv1alpha1.StatusEnum)
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

var _ Backend = (*LLMBackend)(nil)

type LLMBackend struct {
	*knowaydevv1alpha1.LLMBackend
}

func (b *LLMBackend) GetObjectObjectMeta() metav1.ObjectMeta {
	return b.LLMBackend.ObjectMeta
}

func (b *LLMBackend) GetStatus() BackendStatus {
	return &LLMBackendStatus{LLMBackendStatus: &b.Status}
}

func BackendFromLLMBackend(llmBackend *knowaydevv1alpha1.LLMBackend) Backend {
	return &LLMBackend{
		LLMBackend: llmBackend,
	}
}

type LLMBackendStatus struct {
	*knowaydevv1alpha1.LLMBackendStatus
}

func (s *LLMBackendStatus) GetStatus() knowaydevv1alpha1.StatusEnum {
	return s.Status
}

func (s *LLMBackendStatus) SetStatus(status knowaydevv1alpha1.StatusEnum) {
	s.Status = status
}

func (s *LLMBackendStatus) GetConditions() []metav1.Condition {
	return s.Conditions
}

func (s *LLMBackendStatus) SetConditions(conditions []metav1.Condition) {
	s.Conditions = conditions
}

var _ Backend = (*ImageGenerationBackend)(nil)

type ImageGenerationBackend struct {
	*knowaydevv1alpha1.ImageGenerationBackend
}

func (b *ImageGenerationBackend) GetObjectObjectMeta() metav1.ObjectMeta {
	return b.ImageGenerationBackend.ObjectMeta
}

func (b *ImageGenerationBackend) GetStatus() BackendStatus {
	return &ImageGenerationBackendStatus{ImageGenerationBackendStatus: &b.Status}
}

func BackendFromImageGenerationBackend(imageGenerationBackend *knowaydevv1alpha1.ImageGenerationBackend) Backend {
	return &ImageGenerationBackend{
		ImageGenerationBackend: imageGenerationBackend,
	}
}

type ImageGenerationBackendStatus struct {
	*knowaydevv1alpha1.ImageGenerationBackendStatus
}

func (s *ImageGenerationBackendStatus) GetStatus() knowaydevv1alpha1.StatusEnum {
	return s.Status
}

func (s *ImageGenerationBackendStatus) SetStatus(status knowaydevv1alpha1.StatusEnum) {
	s.Status = status
}

func (s *ImageGenerationBackendStatus) GetConditions() []metav1.Condition {
	return s.Conditions
}

func (s *ImageGenerationBackendStatus) SetConditions(conditions []metav1.Condition) {
	s.Conditions = conditions
}
