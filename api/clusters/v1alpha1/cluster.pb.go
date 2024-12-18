// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: clusters/v1alpha1/cluster.proto

package v1alpha1

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LoadBalancePolicy int32

const (
	LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED LoadBalancePolicy = 0
	LoadBalancePolicy_ROUND_ROBIN                     LoadBalancePolicy = 1
	LoadBalancePolicy_LEAST_CONNECTION                LoadBalancePolicy = 2
	LoadBalancePolicy_IP_HASH                         LoadBalancePolicy = 3
	// CUSTOM means the load balance policy is defined by the filters.
	LoadBalancePolicy_CUSTOM LoadBalancePolicy = 15
)

// Enum value maps for LoadBalancePolicy.
var (
	LoadBalancePolicy_name = map[int32]string{
		0:  "LOAD_BALANCE_POLICY_UNSPECIFIED",
		1:  "ROUND_ROBIN",
		2:  "LEAST_CONNECTION",
		3:  "IP_HASH",
		15: "CUSTOM",
	}
	LoadBalancePolicy_value = map[string]int32{
		"LOAD_BALANCE_POLICY_UNSPECIFIED": 0,
		"ROUND_ROBIN":                     1,
		"LEAST_CONNECTION":                2,
		"IP_HASH":                         3,
		"CUSTOM":                          15,
	}
)

func (x LoadBalancePolicy) Enum() *LoadBalancePolicy {
	p := new(LoadBalancePolicy)
	*p = x
	return p
}

func (x LoadBalancePolicy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LoadBalancePolicy) Descriptor() protoreflect.EnumDescriptor {
	return file_clusters_v1alpha1_cluster_proto_enumTypes[0].Descriptor()
}

func (LoadBalancePolicy) Type() protoreflect.EnumType {
	return &file_clusters_v1alpha1_cluster_proto_enumTypes[0]
}

func (x LoadBalancePolicy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LoadBalancePolicy.Descriptor instead.
func (LoadBalancePolicy) EnumDescriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{0}
}

type Upstream_Method int32

const (
	Upstream_METHOD_UNSPECIFIED Upstream_Method = 0
	Upstream_GET                Upstream_Method = 1
	Upstream_POST               Upstream_Method = 2
)

// Enum value maps for Upstream_Method.
var (
	Upstream_Method_name = map[int32]string{
		0: "METHOD_UNSPECIFIED",
		1: "GET",
		2: "POST",
	}
	Upstream_Method_value = map[string]int32{
		"METHOD_UNSPECIFIED": 0,
		"GET":                1,
		"POST":               2,
	}
)

func (x Upstream_Method) Enum() *Upstream_Method {
	p := new(Upstream_Method)
	*p = x
	return p
}

func (x Upstream_Method) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Upstream_Method) Descriptor() protoreflect.EnumDescriptor {
	return file_clusters_v1alpha1_cluster_proto_enumTypes[1].Descriptor()
}

func (Upstream_Method) Type() protoreflect.EnumType {
	return &file_clusters_v1alpha1_cluster_proto_enumTypes[1]
}

func (x Upstream_Method) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Upstream_Method.Descriptor instead.
func (Upstream_Method) EnumDescriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{2, 0}
}

type ClusterFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Config *anypb.Any `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *ClusterFilter) Reset() {
	*x = ClusterFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterFilter) ProtoMessage() {}

func (x *ClusterFilter) ProtoReflect() protoreflect.Message {
	mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterFilter.ProtoReflect.Descriptor instead.
func (*ClusterFilter) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{0}
}

func (x *ClusterFilter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ClusterFilter) GetConfig() *anypb.Any {
	if x != nil {
		return x.Config
	}
	return nil
}

type TLSConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TLSConfig) Reset() {
	*x = TLSConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TLSConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TLSConfig) ProtoMessage() {}

func (x *TLSConfig) ProtoReflect() protoreflect.Message {
	mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TLSConfig.ProtoReflect.Descriptor instead.
func (*TLSConfig) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{1}
}

type Upstream struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url     string             `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Method  Upstream_Method    `protobuf:"varint,2,opt,name=method,proto3,enum=knoway.clusters.v1alpha1.Upstream_Method" json:"method,omitempty"`
	Headers []*Upstream_Header `protobuf:"bytes,3,rep,name=headers,proto3" json:"headers,omitempty"`
	Timeout int32              `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
}

func (x *Upstream) Reset() {
	*x = Upstream{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Upstream) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Upstream) ProtoMessage() {}

func (x *Upstream) ProtoReflect() protoreflect.Message {
	mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Upstream.ProtoReflect.Descriptor instead.
func (*Upstream) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{2}
}

func (x *Upstream) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Upstream) GetMethod() Upstream_Method {
	if x != nil {
		return x.Method
	}
	return Upstream_METHOD_UNSPECIFIED
}

func (x *Upstream) GetHeaders() []*Upstream_Header {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Upstream) GetTimeout() int32 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

type Cluster struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name              string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	LoadBalancePolicy LoadBalancePolicy `protobuf:"varint,2,opt,name=loadBalancePolicy,proto3,enum=knoway.clusters.v1alpha1.LoadBalancePolicy" json:"loadBalancePolicy,omitempty"`
	Upstream          *Upstream         `protobuf:"bytes,3,opt,name=upstream,proto3" json:"upstream,omitempty"`
	TlsConfig         *TLSConfig        `protobuf:"bytes,4,opt,name=tlsConfig,proto3" json:"tlsConfig,omitempty"`
	Filters           []*ClusterFilter  `protobuf:"bytes,5,rep,name=filters,proto3" json:"filters,omitempty"`
	Provider          string            `protobuf:"bytes,6,opt,name=provider,proto3" json:"provider,omitempty"`
	Created           int64             `protobuf:"varint,7,opt,name=created,proto3" json:"created,omitempty"`
}

func (x *Cluster) Reset() {
	*x = Cluster{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cluster) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cluster) ProtoMessage() {}

func (x *Cluster) ProtoReflect() protoreflect.Message {
	mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cluster.ProtoReflect.Descriptor instead.
func (*Cluster) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{3}
}

func (x *Cluster) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Cluster) GetLoadBalancePolicy() LoadBalancePolicy {
	if x != nil {
		return x.LoadBalancePolicy
	}
	return LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED
}

func (x *Cluster) GetUpstream() *Upstream {
	if x != nil {
		return x.Upstream
	}
	return nil
}

func (x *Cluster) GetTlsConfig() *TLSConfig {
	if x != nil {
		return x.TlsConfig
	}
	return nil
}

func (x *Cluster) GetFilters() []*ClusterFilter {
	if x != nil {
		return x.Filters
	}
	return nil
}

func (x *Cluster) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *Cluster) GetCreated() int64 {
	if x != nil {
		return x.Created
	}
	return 0
}

type Upstream_Header struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Upstream_Header) Reset() {
	*x = Upstream_Header{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Upstream_Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Upstream_Header) ProtoMessage() {}

func (x *Upstream_Header) ProtoReflect() protoreflect.Message {
	mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Upstream_Header.ProtoReflect.Descriptor instead.
func (*Upstream_Header) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{2, 0}
}

func (x *Upstream_Header) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Upstream_Header) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_clusters_v1alpha1_cluster_proto protoreflect.FileDescriptor

var file_clusters_v1alpha1_cluster_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x18, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x19, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x51, 0x0a, 0x0d, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e,
	0x79, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x0b, 0x0a, 0x09, 0x54, 0x4c, 0x53,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xa5, 0x02, 0x0a, 0x08, 0x55, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x41, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x29, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x43, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x6e, 0x6f, 0x77,
	0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x1a, 0x30, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x33, 0x0a, 0x06, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x12, 0x16, 0x0a, 0x12, 0x4d, 0x45, 0x54, 0x48, 0x4f, 0x44, 0x5f, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x47,
	0x45, 0x54, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x4f, 0x53, 0x54, 0x10, 0x02, 0x22, 0xf4,
	0x02, 0x0a, 0x07, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x59,
	0x0a, 0x11, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x6f, 0x6c,
	0x69, 0x63, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x6b, 0x6e, 0x6f, 0x77,
	0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
	0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x11, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x3e, 0x0a, 0x08, 0x75, 0x70, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6b, 0x6e,
	0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52,
	0x08, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x41, 0x0a, 0x09, 0x74, 0x6c, 0x73,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b,
	0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x4c, 0x53, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x09, 0x74, 0x6c, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x41, 0x0a, 0x07,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e,
	0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x12,
	0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x2a, 0x78, 0x0a, 0x11, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x23, 0x0a, 0x1f, 0x4c, 0x4f,
	0x41, 0x44, 0x5f, 0x42, 0x41, 0x4c, 0x41, 0x4e, 0x43, 0x45, 0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43,
	0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x0f, 0x0a, 0x0b, 0x52, 0x4f, 0x55, 0x4e, 0x44, 0x5f, 0x52, 0x4f, 0x42, 0x49, 0x4e, 0x10, 0x01,
	0x12, 0x14, 0x0a, 0x10, 0x4c, 0x45, 0x41, 0x53, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43,
	0x54, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x50, 0x5f, 0x48, 0x41, 0x53,
	0x48, 0x10, 0x03, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x55, 0x53, 0x54, 0x4f, 0x4d, 0x10, 0x0f, 0x42,
	0x22, 0x5a, 0x20, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_clusters_v1alpha1_cluster_proto_rawDescOnce sync.Once
	file_clusters_v1alpha1_cluster_proto_rawDescData = file_clusters_v1alpha1_cluster_proto_rawDesc
)

func file_clusters_v1alpha1_cluster_proto_rawDescGZIP() []byte {
	file_clusters_v1alpha1_cluster_proto_rawDescOnce.Do(func() {
		file_clusters_v1alpha1_cluster_proto_rawDescData = protoimpl.X.CompressGZIP(file_clusters_v1alpha1_cluster_proto_rawDescData)
	})
	return file_clusters_v1alpha1_cluster_proto_rawDescData
}

var file_clusters_v1alpha1_cluster_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_clusters_v1alpha1_cluster_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_clusters_v1alpha1_cluster_proto_goTypes = []interface{}{
	(LoadBalancePolicy)(0),  // 0: knoway.clusters.v1alpha1.LoadBalancePolicy
	(Upstream_Method)(0),    // 1: knoway.clusters.v1alpha1.Upstream.Method
	(*ClusterFilter)(nil),   // 2: knoway.clusters.v1alpha1.ClusterFilter
	(*TLSConfig)(nil),       // 3: knoway.clusters.v1alpha1.TLSConfig
	(*Upstream)(nil),        // 4: knoway.clusters.v1alpha1.Upstream
	(*Cluster)(nil),         // 5: knoway.clusters.v1alpha1.Cluster
	(*Upstream_Header)(nil), // 6: knoway.clusters.v1alpha1.Upstream.Header
	(*anypb.Any)(nil),       // 7: google.protobuf.Any
}
var file_clusters_v1alpha1_cluster_proto_depIdxs = []int32{
	7, // 0: knoway.clusters.v1alpha1.ClusterFilter.config:type_name -> google.protobuf.Any
	1, // 1: knoway.clusters.v1alpha1.Upstream.method:type_name -> knoway.clusters.v1alpha1.Upstream.Method
	6, // 2: knoway.clusters.v1alpha1.Upstream.headers:type_name -> knoway.clusters.v1alpha1.Upstream.Header
	0, // 3: knoway.clusters.v1alpha1.Cluster.loadBalancePolicy:type_name -> knoway.clusters.v1alpha1.LoadBalancePolicy
	4, // 4: knoway.clusters.v1alpha1.Cluster.upstream:type_name -> knoway.clusters.v1alpha1.Upstream
	3, // 5: knoway.clusters.v1alpha1.Cluster.tlsConfig:type_name -> knoway.clusters.v1alpha1.TLSConfig
	2, // 6: knoway.clusters.v1alpha1.Cluster.filters:type_name -> knoway.clusters.v1alpha1.ClusterFilter
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_clusters_v1alpha1_cluster_proto_init() }
func file_clusters_v1alpha1_cluster_proto_init() {
	if File_clusters_v1alpha1_cluster_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_clusters_v1alpha1_cluster_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterFilter); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_clusters_v1alpha1_cluster_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TLSConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_clusters_v1alpha1_cluster_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Upstream); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_clusters_v1alpha1_cluster_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Cluster); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_clusters_v1alpha1_cluster_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Upstream_Header); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_clusters_v1alpha1_cluster_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_clusters_v1alpha1_cluster_proto_goTypes,
		DependencyIndexes: file_clusters_v1alpha1_cluster_proto_depIdxs,
		EnumInfos:         file_clusters_v1alpha1_cluster_proto_enumTypes,
		MessageInfos:      file_clusters_v1alpha1_cluster_proto_msgTypes,
	}.Build()
	File_clusters_v1alpha1_cluster_proto = out.File
	file_clusters_v1alpha1_cluster_proto_rawDesc = nil
	file_clusters_v1alpha1_cluster_proto_goTypes = nil
	file_clusters_v1alpha1_cluster_proto_depIdxs = nil
}
