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
	"flag"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"knoway.dev/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestImageGenerationBackendReconciler_Reconcile(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func(client.Client) client.Client
		request     reconcile.Request
		expectError bool
		validate    func(*testing.T, client.Client)
	}{
		{
			name: "Valid resource reconciled",
			setupClient: func(cl client.Client) client.Client {
				resource := &v1alpha1.ImageGenerationBackend{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-model",
						Namespace: "default",
					},
					Spec: v1alpha1.ImageGenerationBackendSpec{
						ModelName: lo.ToPtr("test-model"),
						Upstream: v1alpha1.ImageGenerationBackendUpstream{
							BaseURL: "xx/v1",
						},
						Filters: nil,
					},
					Status: v1alpha1.ImageGenerationBackendStatus{},
				}
				err := cl.Create(context.Background(), resource)
				if err != nil {
					t.Fatalf("failed to create resource: %v", err)
				}
				return cl
			},
			request: reconcile.Request{
				NamespacedName: client.ObjectKey{
					Namespace: "default",
					Name:      "test-model",
				},
			},
			expectError: false,
			validate: func(t *testing.T, cl client.Client) {
				t.Helper()
				resource := &v1alpha1.ImageGenerationBackend{}
				err := cl.Get(context.Background(), client.ObjectKey{
					Namespace: "default",
					Name:      "test-model",
				}, resource)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copts := zap.Options{
				Development: true,
			}
			copts.BindFlags(flag.CommandLine)
			ctrl.SetLogger(zap.New(zap.UseFlagOptions(&copts)))

			fakeClient := tt.setupClient(NewFakeClientWithStatus())
			reconciler := &ImageGenerationBackendReconciler{
				Client: fakeClient,
			}

			_, err := reconciler.Reconcile(context.TODO(), tt.request)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, fakeClient)
			}
		})
	}
}
