package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
)

type Backend interface {
	GetType() knowaydevv1alpha1.BackendType
	GetObjectObjectMeta() metav1.ObjectMeta
	GetStatus() Statusable[knowaydevv1alpha1.StatusEnum]
	GetModelName() string
}

var _ Backend = (*LLMBackend)(nil)

type LLMBackend struct {
	*knowaydevv1alpha1.LLMBackend
}

func (b *LLMBackend) GetType() knowaydevv1alpha1.BackendType {
	return knowaydevv1alpha1.BackendTypeLLM
}

func (b *LLMBackend) GetObjectObjectMeta() metav1.ObjectMeta {
	return b.LLMBackend.ObjectMeta
}

func (b *LLMBackend) GetStatus() Statusable[knowaydevv1alpha1.StatusEnum] {
	return &LLMBackendStatus{LLMBackendStatus: &b.Status}
}

func (b *LLMBackend) GetModelName() string {
	return modelNameOrNamespacedName(b.LLMBackend)
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

func (b *ImageGenerationBackend) GetType() knowaydevv1alpha1.BackendType {
	return knowaydevv1alpha1.BackendTypeImageGeneration
}

func (b *ImageGenerationBackend) GetObjectObjectMeta() metav1.ObjectMeta {
	return b.ImageGenerationBackend.ObjectMeta
}

func (b *ImageGenerationBackend) GetStatus() Statusable[knowaydevv1alpha1.StatusEnum] {
	return &ImageGenerationBackendStatus{ImageGenerationBackendStatus: &b.Status}
}

func (b *ImageGenerationBackend) GetModelName() string {
	return modelNameOrNamespacedName(b.ImageGenerationBackend)
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

func getBackendFromNamespacedName(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName) (Backend, error) {
	var llmBackend knowaydevv1alpha1.LLMBackend

	err := kubeClient.Get(ctx, namespacedName, &llmBackend)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if err == nil {
		return BackendFromLLMBackend(&llmBackend), nil
	}

	var imageGenerationBackend knowaydevv1alpha1.ImageGenerationBackend

	err = kubeClient.Get(ctx, namespacedName, &imageGenerationBackend)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if err == nil {
		return BackendFromImageGenerationBackend(&imageGenerationBackend), nil
	}

	return nil, nil
}
