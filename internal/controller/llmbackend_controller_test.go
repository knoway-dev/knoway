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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knowaydevv1alpha1 "knoway.dev/api/v1alpha1"
	"knoway.dev/pkg/bootkit"
)

var _ = Describe("LLMBackend Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		llmbackend := &knowaydevv1alpha1.LLMBackend{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind LLMBackend")
			err := k8sClient.Get(ctx, typeNamespacedName, llmbackend)
			if err != nil && errors.IsNotFound(err) {
				resource := &knowaydevv1alpha1.LLMBackend{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &knowaydevv1alpha1.LLMBackend{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance LLMBackend")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &LLMBackendReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				LifeCycle: bootkit.NewEmptyLifeCycle(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

func TestToParams(t *testing.T) {
	newPrtInt := func(i int) *int {
		return &i
	}

	// Define some test cases for different backend configurations
	tests := []struct {
		name         string
		backend      *knowaydevv1alpha1.LLMBackend
		wantDefault  map[string]string
		wantOverride map[string]string
		wantErr      bool
	}{
		{
			name: "Valid OpenAI and SystemParams",
			backend: &knowaydevv1alpha1.LLMBackend{
				Spec: knowaydevv1alpha1.LLMBackendSpec{
					Upstream: knowaydevv1alpha1.BackendUpstream{
						OverrideParams: &knowaydevv1alpha1.ModelParams{
							OpenAI: &knowaydevv1alpha1.OpenAIParam{
								MaxTokens: newPrtInt(50),
							},
						},
						DefaultParams: &knowaydevv1alpha1.ModelParams{
							OpenAI: &knowaydevv1alpha1.OpenAIParam{
								CommonParams: knowaydevv1alpha1.CommonParams{
									Model: "gpt-4",
								},
							},
						},
					},
				},
			},
			wantDefault: map[string]string{
				"model": "gpt-4",
			},
			wantOverride: map[string]string{
				"max_tokens": "50",
			},
			wantErr: false,
		},
		{
			name: "Valid LLama Params",
			backend: &knowaydevv1alpha1.LLMBackend{
				Spec: knowaydevv1alpha1.LLMBackendSpec{
					Upstream: knowaydevv1alpha1.BackendUpstream{
						OverrideParams: &knowaydevv1alpha1.ModelParams{
							LLama: &knowaydevv1alpha1.LLamaParam{
								MaxLength: newPrtInt(40),
							},
						},

						DefaultParams: &knowaydevv1alpha1.ModelParams{
							LLama: &knowaydevv1alpha1.LLamaParam{
								CommonParams: knowaydevv1alpha1.CommonParams{
									Model: "llama-2",
								},
							},
						},
					},
				},
			},
			wantDefault: map[string]string{
				"model": "llama-2",
			},
			wantOverride: map[string]string{
				"max_length": "40",
			},
			wantErr: false,
		},
		{
			name: "Custom Params in UserParams",
			backend: &knowaydevv1alpha1.LLMBackend{
				Spec: knowaydevv1alpha1.LLMBackendSpec{
					Upstream: knowaydevv1alpha1.BackendUpstream{
						OverrideParams: &knowaydevv1alpha1.ModelParams{
							Custom: &runtime.RawExtension{
								Raw: json.RawMessage(`{"custom_param1": "value1", "custom_param2": "42", "custom_object": { "o1": "v1", "o2": "v2"} }`),
							},
						},
						DefaultParams: &knowaydevv1alpha1.ModelParams{
							OpenAI: &knowaydevv1alpha1.OpenAIParam{
								CommonParams: knowaydevv1alpha1.CommonParams{
									Model: "gpt-4",
								},
							},
						},
					},
				},
			},

			wantDefault: map[string]string{
				"model": "gpt-4",
			},
			wantOverride: map[string]string{
				"custom_param1": "value1",
				"custom_param2": "42",                    // Convert integer to string
				"custom_object": `{"o1":"v1","o2":"v2"}`, // Convert nested map to JSON string
			},
			wantErr: false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, gotO, err := toParams(tt.backend)
			if (err != nil) != tt.wantErr {
				t.Errorf("toParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantDefault, gotD)
			assert.Equal(t, tt.wantOverride, gotO)
		})
	}
}
