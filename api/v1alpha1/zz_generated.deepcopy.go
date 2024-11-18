//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LLMBackend) DeepCopyInto(out *LLMBackend) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LLMBackend.
func (in *LLMBackend) DeepCopy() *LLMBackend {
	if in == nil {
		return nil
	}
	out := new(LLMBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LLMBackend) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LLMBackendFilter) DeepCopyInto(out *LLMBackendFilter) {
	*out = *in
	in.Value.DeepCopyInto(&out.Value)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LLMBackendFilter.
func (in *LLMBackendFilter) DeepCopy() *LLMBackendFilter {
	if in == nil {
		return nil
	}
	out := new(LLMBackendFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LLMBackendList) DeepCopyInto(out *LLMBackendList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LLMBackend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LLMBackendList.
func (in *LLMBackendList) DeepCopy() *LLMBackendList {
	if in == nil {
		return nil
	}
	out := new(LLMBackendList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LLMBackendList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LLMBackendSpec) DeepCopyInto(out *LLMBackendSpec) {
	*out = *in
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = make([]LLMBackendFilter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LLMBackendSpec.
func (in *LLMBackendSpec) DeepCopy() *LLMBackendSpec {
	if in == nil {
		return nil
	}
	out := new(LLMBackendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LLMBackendStatus) DeepCopyInto(out *LLMBackendStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LLMBackendStatus.
func (in *LLMBackendStatus) DeepCopy() *LLMBackendStatus {
	if in == nil {
		return nil
	}
	out := new(LLMBackendStatus)
	in.DeepCopyInto(out)
	return out
}
