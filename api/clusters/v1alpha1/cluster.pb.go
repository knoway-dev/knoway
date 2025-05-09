// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: clusters/v1alpha1/cluster.proto

package v1alpha1

import (
	reflect "reflect"
	sync "sync"

	_struct "github.com/golang/protobuf/ptypes/struct"
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

type ClusterType int32

const (
	ClusterType_CLUSTER_TYPE_UNSPECIFIED ClusterType = 0
	ClusterType_LLM                      ClusterType = 1
	ClusterType_IMAGE_GENERATION         ClusterType = 2
)

// Enum value maps for ClusterType.
var (
	ClusterType_name = map[int32]string{
		0: "CLUSTER_TYPE_UNSPECIFIED",
		1: "LLM",
		2: "IMAGE_GENERATION",
	}
	ClusterType_value = map[string]int32{
		"CLUSTER_TYPE_UNSPECIFIED": 0,
		"LLM":                      1,
		"IMAGE_GENERATION":         2,
	}
)

func (x ClusterType) Enum() *ClusterType {
	p := new(ClusterType)
	*p = x
	return p
}

func (x ClusterType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClusterType) Descriptor() protoreflect.EnumDescriptor {
	return file_clusters_v1alpha1_cluster_proto_enumTypes[1].Descriptor()
}

func (ClusterType) Type() protoreflect.EnumType {
	return &file_clusters_v1alpha1_cluster_proto_enumTypes[1]
}

func (x ClusterType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClusterType.Descriptor instead.
func (ClusterType) EnumDescriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{1}
}

type ClusterProvider int32

const (
	ClusterProvider_CLUSTER_PROVIDER_UNSPECIFIED ClusterProvider = 0
	ClusterProvider_OPEN_AI                      ClusterProvider = 1
	ClusterProvider_VLLM                         ClusterProvider = 2
	ClusterProvider_OLLAMA                       ClusterProvider = 3
)

// Enum value maps for ClusterProvider.
var (
	ClusterProvider_name = map[int32]string{
		0: "CLUSTER_PROVIDER_UNSPECIFIED",
		1: "OPEN_AI",
		2: "VLLM",
		3: "OLLAMA",
	}
	ClusterProvider_value = map[string]int32{
		"CLUSTER_PROVIDER_UNSPECIFIED": 0,
		"OPEN_AI":                      1,
		"VLLM":                         2,
		"OLLAMA":                       3,
	}
)

func (x ClusterProvider) Enum() *ClusterProvider {
	p := new(ClusterProvider)
	*p = x
	return p
}

func (x ClusterProvider) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClusterProvider) Descriptor() protoreflect.EnumDescriptor {
	return file_clusters_v1alpha1_cluster_proto_enumTypes[2].Descriptor()
}

func (ClusterProvider) Type() protoreflect.EnumType {
	return &file_clusters_v1alpha1_cluster_proto_enumTypes[2]
}

func (x ClusterProvider) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClusterProvider.Descriptor instead.
func (ClusterProvider) EnumDescriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{2}
}

type ClusterMeteringPolicy_SizeFrom int32

const (
	ClusterMeteringPolicy_SIZE_FROM_UNSPECIFIED ClusterMeteringPolicy_SizeFrom = 0
	// For image generation, the size of the generated image is determined
	// by the input parameters.
	//
	// For example, even if the output image is 1024x1024, as long as the
	// input parameter specified 256x256, the size of the generated image
	// will be account as 256x256.
	ClusterMeteringPolicy_SIZE_FROM_INPUT ClusterMeteringPolicy_SizeFrom = 1
	// For image generation, the size of the generated image is determined
	// by the output image. This is done by parsing through the actual
	// generated image file header by using Golang's std library to
	// determine the size of the image.
	//
	// For example, no matter what the input specified, if the output image
	// is 1024x1024, the size of the generated image will be account as
	// 1024x1024.
	ClusterMeteringPolicy_SIZE_FROM_OUTPUT ClusterMeteringPolicy_SizeFrom = 2
	// For image generation, the size of the generated image is determined
	// by the greatest size of the input parameters and output image
	// resolution.
	//
	// For example, if the input parameter specified 256x256 and the output
	// image is 1024x1024, the size of the generated image will be account
	// as 1024x1024. On the other hand, if the input parameter specified
	// 1024x1024 and the output image is 256x256, the size of the generated
	// image will be account as 1024x1024.
	ClusterMeteringPolicy_SIZE_FROM_GREATEST ClusterMeteringPolicy_SizeFrom = 3
)

// Enum value maps for ClusterMeteringPolicy_SizeFrom.
var (
	ClusterMeteringPolicy_SizeFrom_name = map[int32]string{
		0: "SIZE_FROM_UNSPECIFIED",
		1: "SIZE_FROM_INPUT",
		2: "SIZE_FROM_OUTPUT",
		3: "SIZE_FROM_GREATEST",
	}
	ClusterMeteringPolicy_SizeFrom_value = map[string]int32{
		"SIZE_FROM_UNSPECIFIED": 0,
		"SIZE_FROM_INPUT":       1,
		"SIZE_FROM_OUTPUT":      2,
		"SIZE_FROM_GREATEST":    3,
	}
)

func (x ClusterMeteringPolicy_SizeFrom) Enum() *ClusterMeteringPolicy_SizeFrom {
	p := new(ClusterMeteringPolicy_SizeFrom)
	*p = x
	return p
}

func (x ClusterMeteringPolicy_SizeFrom) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClusterMeteringPolicy_SizeFrom) Descriptor() protoreflect.EnumDescriptor {
	return file_clusters_v1alpha1_cluster_proto_enumTypes[3].Descriptor()
}

func (ClusterMeteringPolicy_SizeFrom) Type() protoreflect.EnumType {
	return &file_clusters_v1alpha1_cluster_proto_enumTypes[3]
}

func (x ClusterMeteringPolicy_SizeFrom) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClusterMeteringPolicy_SizeFrom.Descriptor instead.
func (ClusterMeteringPolicy_SizeFrom) EnumDescriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{3, 0}
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

	Url             string                    `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Headers         []*Upstream_Header        `protobuf:"bytes,3,rep,name=headers,proto3" json:"headers,omitempty"`
	Timeout         int32                     `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
	DefaultParams   map[string]*_struct.Value `protobuf:"bytes,5,rep,name=defaultParams,proto3" json:"defaultParams,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	OverrideParams  map[string]*_struct.Value `protobuf:"bytes,6,rep,name=overrideParams,proto3" json:"overrideParams,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	RemoveParamKeys []string                  `protobuf:"bytes,7,rep,name=removeParamKeys,proto3" json:"removeParamKeys,omitempty"`
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

func (x *Upstream) GetDefaultParams() map[string]*_struct.Value {
	if x != nil {
		return x.DefaultParams
	}
	return nil
}

func (x *Upstream) GetOverrideParams() map[string]*_struct.Value {
	if x != nil {
		return x.OverrideParams
	}
	return nil
}

func (x *Upstream) GetRemoveParamKeys() []string {
	if x != nil {
		return x.RemoveParamKeys
	}
	return nil
}

type ClusterMeteringPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SizeFrom *ClusterMeteringPolicy_SizeFrom `protobuf:"varint,1,opt,name=sizeFrom,proto3,enum=knoway.clusters.v1alpha1.ClusterMeteringPolicy_SizeFrom,oneof" json:"sizeFrom,omitempty"`
}

func (x *ClusterMeteringPolicy) Reset() {
	*x = ClusterMeteringPolicy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterMeteringPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterMeteringPolicy) ProtoMessage() {}

func (x *ClusterMeteringPolicy) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use ClusterMeteringPolicy.ProtoReflect.Descriptor instead.
func (*ClusterMeteringPolicy) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{3}
}

func (x *ClusterMeteringPolicy) GetSizeFrom() ClusterMeteringPolicy_SizeFrom {
	if x != nil && x.SizeFrom != nil {
		return *x.SizeFrom
	}
	return ClusterMeteringPolicy_SIZE_FROM_UNSPECIFIED
}

type Cluster struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name              string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	LoadBalancePolicy LoadBalancePolicy      `protobuf:"varint,2,opt,name=loadBalancePolicy,proto3,enum=knoway.clusters.v1alpha1.LoadBalancePolicy" json:"loadBalancePolicy,omitempty"`
	Upstream          *Upstream              `protobuf:"bytes,3,opt,name=upstream,proto3" json:"upstream,omitempty"`
	TlsConfig         *TLSConfig             `protobuf:"bytes,4,opt,name=tlsConfig,proto3" json:"tlsConfig,omitempty"`
	Filters           []*ClusterFilter       `protobuf:"bytes,5,rep,name=filters,proto3" json:"filters,omitempty"`
	Provider          ClusterProvider        `protobuf:"varint,6,opt,name=provider,proto3,enum=knoway.clusters.v1alpha1.ClusterProvider" json:"provider,omitempty"`
	Created           int64                  `protobuf:"varint,7,opt,name=created,proto3" json:"created,omitempty"`
	Type              ClusterType            `protobuf:"varint,8,opt,name=type,proto3,enum=knoway.clusters.v1alpha1.ClusterType" json:"type,omitempty"`
	MeteringPolicy    *ClusterMeteringPolicy `protobuf:"bytes,9,opt,name=meteringPolicy,proto3" json:"meteringPolicy,omitempty"`
}

func (x *Cluster) Reset() {
	*x = Cluster{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cluster) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cluster) ProtoMessage() {}

func (x *Cluster) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Cluster.ProtoReflect.Descriptor instead.
func (*Cluster) Descriptor() ([]byte, []int) {
	return file_clusters_v1alpha1_cluster_proto_rawDescGZIP(), []int{4}
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

func (x *Cluster) GetProvider() ClusterProvider {
	if x != nil {
		return x.Provider
	}
	return ClusterProvider_CLUSTER_PROVIDER_UNSPECIFIED
}

func (x *Cluster) GetCreated() int64 {
	if x != nil {
		return x.Created
	}
	return 0
}

func (x *Cluster) GetType() ClusterType {
	if x != nil {
		return x.Type
	}
	return ClusterType_CLUSTER_TYPE_UNSPECIFIED
}

func (x *Cluster) GetMeteringPolicy() *ClusterMeteringPolicy {
	if x != nil {
		return x.MeteringPolicy
	}
	return nil
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
		mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Upstream_Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Upstream_Header) ProtoMessage() {}

func (x *Upstream_Header) ProtoReflect() protoreflect.Message {
	mi := &file_clusters_v1alpha1_cluster_proto_msgTypes[5]
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
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x51, 0x0a, 0x0d, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x46,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52,
	0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x0b, 0x0a, 0x09, 0x54, 0x4c, 0x53, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x22, 0xc9, 0x04, 0x0a, 0x08, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x75, 0x72, 0x6c, 0x12, 0x43, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52,
	0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65,
	0x6f, 0x75, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x12, 0x5b, 0x0a, 0x0d, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x6b, 0x6e, 0x6f, 0x77,
	0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x44, 0x65,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x0d, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12,
	0x5e, 0x0a, 0x0e, 0x6f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x36, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79,
	0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x4f, 0x76, 0x65, 0x72,
	0x72, 0x69, 0x64, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x0e, 0x6f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12,
	0x28, 0x0a, 0x0f, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x4b, 0x65,
	0x79, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65,
	0x50, 0x61, 0x72, 0x61, 0x6d, 0x4b, 0x65, 0x79, 0x73, 0x1a, 0x30, 0x0a, 0x06, 0x48, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x58, 0x0a, 0x12, 0x44,
	0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x59, 0x0a, 0x13, 0x4f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64,
	0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0xe9, 0x01, 0x0a, 0x15, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x74, 0x65,
	0x72, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x59, 0x0a, 0x08, 0x73, 0x69,
	0x7a, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x38, 0x2e, 0x6b,
	0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4d,
	0x65, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x2e, 0x53, 0x69,
	0x7a, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x48, 0x00, 0x52, 0x08, 0x73, 0x69, 0x7a, 0x65, 0x46, 0x72,
	0x6f, 0x6d, 0x88, 0x01, 0x01, 0x22, 0x68, 0x0a, 0x08, 0x53, 0x69, 0x7a, 0x65, 0x46, 0x72, 0x6f,
	0x6d, 0x12, 0x19, 0x0a, 0x15, 0x53, 0x49, 0x5a, 0x45, 0x5f, 0x46, 0x52, 0x4f, 0x4d, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0f,
	0x53, 0x49, 0x5a, 0x45, 0x5f, 0x46, 0x52, 0x4f, 0x4d, 0x5f, 0x49, 0x4e, 0x50, 0x55, 0x54, 0x10,
	0x01, 0x12, 0x14, 0x0a, 0x10, 0x53, 0x49, 0x5a, 0x45, 0x5f, 0x46, 0x52, 0x4f, 0x4d, 0x5f, 0x4f,
	0x55, 0x54, 0x50, 0x55, 0x54, 0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x49, 0x5a, 0x45, 0x5f,
	0x46, 0x52, 0x4f, 0x4d, 0x5f, 0x47, 0x52, 0x45, 0x41, 0x54, 0x45, 0x53, 0x54, 0x10, 0x03, 0x42,
	0x0b, 0x0a, 0x09, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x22, 0xb3, 0x04, 0x0a,
	0x07, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x59, 0x0a, 0x11,
	0x6c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79,
	0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x52, 0x11, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63,
	0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x3e, 0x0a, 0x08, 0x75, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6b, 0x6e, 0x6f, 0x77,
	0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x08, 0x75,
	0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x41, 0x0a, 0x09, 0x74, 0x6c, 0x73, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x6e, 0x6f,
	0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x4c, 0x53, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x09, 0x74, 0x6c, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x41, 0x0a, 0x07, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6b, 0x6e,
	0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x46, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x12, 0x45, 0x0a,
	0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x29, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x39,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x6b,
	0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x57, 0x0a, 0x0e, 0x6d, 0x65, 0x74,
	0x65, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2f, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x52, 0x0e, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x2a, 0x78, 0x0a, 0x11, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63,
	0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x23, 0x0a, 0x1f, 0x4c, 0x4f, 0x41, 0x44, 0x5f,
	0x42, 0x41, 0x4c, 0x41, 0x4e, 0x43, 0x45, 0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b,
	0x52, 0x4f, 0x55, 0x4e, 0x44, 0x5f, 0x52, 0x4f, 0x42, 0x49, 0x4e, 0x10, 0x01, 0x12, 0x14, 0x0a,
	0x10, 0x4c, 0x45, 0x41, 0x53, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x49, 0x4f,
	0x4e, 0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x50, 0x5f, 0x48, 0x41, 0x53, 0x48, 0x10, 0x03,
	0x12, 0x0a, 0x0a, 0x06, 0x43, 0x55, 0x53, 0x54, 0x4f, 0x4d, 0x10, 0x0f, 0x2a, 0x4a, 0x0a, 0x0b,
	0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x18, 0x43,
	0x4c, 0x55, 0x53, 0x54, 0x45, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x4c, 0x4c, 0x4d,
	0x10, 0x01, 0x12, 0x14, 0x0a, 0x10, 0x49, 0x4d, 0x41, 0x47, 0x45, 0x5f, 0x47, 0x45, 0x4e, 0x45,
	0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x2a, 0x56, 0x0a, 0x0f, 0x43, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x12, 0x20, 0x0a, 0x1c, 0x43,
	0x4c, 0x55, 0x53, 0x54, 0x45, 0x52, 0x5f, 0x50, 0x52, 0x4f, 0x56, 0x49, 0x44, 0x45, 0x52, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a,
	0x07, 0x4f, 0x50, 0x45, 0x4e, 0x5f, 0x41, 0x49, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x56, 0x4c,
	0x4c, 0x4d, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x4f, 0x4c, 0x4c, 0x41, 0x4d, 0x41, 0x10, 0x03,
	0x42, 0x22, 0x5a, 0x20, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_clusters_v1alpha1_cluster_proto_enumTypes = make([]protoimpl.EnumInfo, 4)
var file_clusters_v1alpha1_cluster_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_clusters_v1alpha1_cluster_proto_goTypes = []interface{}{
	(LoadBalancePolicy)(0),              // 0: knoway.clusters.v1alpha1.LoadBalancePolicy
	(ClusterType)(0),                    // 1: knoway.clusters.v1alpha1.ClusterType
	(ClusterProvider)(0),                // 2: knoway.clusters.v1alpha1.ClusterProvider
	(ClusterMeteringPolicy_SizeFrom)(0), // 3: knoway.clusters.v1alpha1.ClusterMeteringPolicy.SizeFrom
	(*ClusterFilter)(nil),               // 4: knoway.clusters.v1alpha1.ClusterFilter
	(*TLSConfig)(nil),                   // 5: knoway.clusters.v1alpha1.TLSConfig
	(*Upstream)(nil),                    // 6: knoway.clusters.v1alpha1.Upstream
	(*ClusterMeteringPolicy)(nil),       // 7: knoway.clusters.v1alpha1.ClusterMeteringPolicy
	(*Cluster)(nil),                     // 8: knoway.clusters.v1alpha1.Cluster
	(*Upstream_Header)(nil),             // 9: knoway.clusters.v1alpha1.Upstream.Header
	nil,                                 // 10: knoway.clusters.v1alpha1.Upstream.DefaultParamsEntry
	nil,                                 // 11: knoway.clusters.v1alpha1.Upstream.OverrideParamsEntry
	(*anypb.Any)(nil),                   // 12: google.protobuf.Any
	(*_struct.Value)(nil),               // 13: google.protobuf.Value
}
var file_clusters_v1alpha1_cluster_proto_depIdxs = []int32{
	12, // 0: knoway.clusters.v1alpha1.ClusterFilter.config:type_name -> google.protobuf.Any
	9,  // 1: knoway.clusters.v1alpha1.Upstream.headers:type_name -> knoway.clusters.v1alpha1.Upstream.Header
	10, // 2: knoway.clusters.v1alpha1.Upstream.defaultParams:type_name -> knoway.clusters.v1alpha1.Upstream.DefaultParamsEntry
	11, // 3: knoway.clusters.v1alpha1.Upstream.overrideParams:type_name -> knoway.clusters.v1alpha1.Upstream.OverrideParamsEntry
	3,  // 4: knoway.clusters.v1alpha1.ClusterMeteringPolicy.sizeFrom:type_name -> knoway.clusters.v1alpha1.ClusterMeteringPolicy.SizeFrom
	0,  // 5: knoway.clusters.v1alpha1.Cluster.loadBalancePolicy:type_name -> knoway.clusters.v1alpha1.LoadBalancePolicy
	6,  // 6: knoway.clusters.v1alpha1.Cluster.upstream:type_name -> knoway.clusters.v1alpha1.Upstream
	5,  // 7: knoway.clusters.v1alpha1.Cluster.tlsConfig:type_name -> knoway.clusters.v1alpha1.TLSConfig
	4,  // 8: knoway.clusters.v1alpha1.Cluster.filters:type_name -> knoway.clusters.v1alpha1.ClusterFilter
	2,  // 9: knoway.clusters.v1alpha1.Cluster.provider:type_name -> knoway.clusters.v1alpha1.ClusterProvider
	1,  // 10: knoway.clusters.v1alpha1.Cluster.type:type_name -> knoway.clusters.v1alpha1.ClusterType
	7,  // 11: knoway.clusters.v1alpha1.Cluster.meteringPolicy:type_name -> knoway.clusters.v1alpha1.ClusterMeteringPolicy
	13, // 12: knoway.clusters.v1alpha1.Upstream.DefaultParamsEntry.value:type_name -> google.protobuf.Value
	13, // 13: knoway.clusters.v1alpha1.Upstream.OverrideParamsEntry.value:type_name -> google.protobuf.Value
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
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
			switch v := v.(*ClusterMeteringPolicy); i {
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
		file_clusters_v1alpha1_cluster_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
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
	file_clusters_v1alpha1_cluster_proto_msgTypes[3].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_clusters_v1alpha1_cluster_proto_rawDesc,
			NumEnums:      4,
			NumMessages:   8,
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
