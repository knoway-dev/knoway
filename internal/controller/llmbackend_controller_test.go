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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"knoway.dev/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestLLMBackendReconciler_Reconcile(t *testing.T) {
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
				resource := &v1alpha1.LLMBackend{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-model",
						Namespace: "default",
					},
					Spec: v1alpha1.LLMBackendSpec{
						ModelName: lo.ToPtr("test-model"),
						Upstream: v1alpha1.BackendUpstream{
							BaseURL: "xx/v1",
						},
						Filters: nil,
					},
					Status: v1alpha1.LLMBackendStatus{},
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
				resource := &v1alpha1.LLMBackend{}
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
			reconciler := &LLMBackendReconciler{
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

func NewFakeClientWithStatus() client.Client {
	return &FakeClientWithStatus{
		Client: fake.NewClientBuilder().WithScheme(createTestScheme()).Build(),
	}
}

type FakeClientWithStatus struct {
	client.Client
}

func (f *FakeClientWithStatus) Status() client.StatusWriter {
	return &FakeStatusWriter{Client: f.Client}
}

type FakeStatusWriter struct {
	client.Client
}

func (f *FakeStatusWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	panic("implement me")
}

func (f *FakeStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return f.Client.Update(ctx, obj)
}

func (f *FakeStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	panic("implement me")
}

func createTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	return scheme
}
