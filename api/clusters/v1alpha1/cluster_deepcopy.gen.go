// Code generated by protoc-gen-deepcopy. DO NOT EDIT.
package v1alpha1

import (
	proto "github.com/golang/protobuf/proto"
)

// DeepCopyInto supports using ClusterFilter within kubernetes types, where deepcopy-gen is used.
func (in *ClusterFilter) DeepCopyInto(out *ClusterFilter) {
	p := proto.Clone(in).(*ClusterFilter)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterFilter. Required by controller-gen.
func (in *ClusterFilter) DeepCopy() *ClusterFilter {
	if in == nil {
		return nil
	}
	out := new(ClusterFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new ClusterFilter. Required by controller-gen.
func (in *ClusterFilter) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}

// DeepCopyInto supports using TLSConfig within kubernetes types, where deepcopy-gen is used.
func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	p := proto.Clone(in).(*TLSConfig)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig. Required by controller-gen.
func (in *TLSConfig) DeepCopy() *TLSConfig {
	if in == nil {
		return nil
	}
	out := new(TLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig. Required by controller-gen.
func (in *TLSConfig) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}

// DeepCopyInto supports using Upstream within kubernetes types, where deepcopy-gen is used.
func (in *Upstream) DeepCopyInto(out *Upstream) {
	p := proto.Clone(in).(*Upstream)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Upstream. Required by controller-gen.
func (in *Upstream) DeepCopy() *Upstream {
	if in == nil {
		return nil
	}
	out := new(Upstream)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new Upstream. Required by controller-gen.
func (in *Upstream) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}

// DeepCopyInto supports using Upstream_Header within kubernetes types, where deepcopy-gen is used.
func (in *Upstream_Header) DeepCopyInto(out *Upstream_Header) {
	p := proto.Clone(in).(*Upstream_Header)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Upstream_Header. Required by controller-gen.
func (in *Upstream_Header) DeepCopy() *Upstream_Header {
	if in == nil {
		return nil
	}
	out := new(Upstream_Header)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new Upstream_Header. Required by controller-gen.
func (in *Upstream_Header) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}

// DeepCopyInto supports using Cluster within kubernetes types, where deepcopy-gen is used.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	p := proto.Clone(in).(*Cluster)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster. Required by controller-gen.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new Cluster. Required by controller-gen.
func (in *Cluster) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}
