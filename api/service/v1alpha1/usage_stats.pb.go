// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: service/v1alpha1/usage_stats.proto

package v1alpha1

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type UsageReportRequest_Mode int32

const (
	UsageReportRequest_MODE_UNSPECIFIED UsageReportRequest_Mode = 0
	// The PER_REQUEST mode means that each time a request is received,
	// the usage of the request will be included.
	// If the server fails to process, statistical data may be lost.
	UsageReportRequest_PER_REQUEST UsageReportRequest_Mode = 1
)

// Enum value maps for UsageReportRequest_Mode.
var (
	UsageReportRequest_Mode_name = map[int32]string{
		0: "MODE_UNSPECIFIED",
		1: "PER_REQUEST",
	}
	UsageReportRequest_Mode_value = map[string]int32{
		"MODE_UNSPECIFIED": 0,
		"PER_REQUEST":      1,
	}
)

func (x UsageReportRequest_Mode) Enum() *UsageReportRequest_Mode {
	p := new(UsageReportRequest_Mode)
	*p = x
	return p
}

func (x UsageReportRequest_Mode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UsageReportRequest_Mode) Descriptor() protoreflect.EnumDescriptor {
	return file_service_v1alpha1_usage_stats_proto_enumTypes[0].Descriptor()
}

func (UsageReportRequest_Mode) Type() protoreflect.EnumType {
	return &file_service_v1alpha1_usage_stats_proto_enumTypes[0]
}

func (x UsageReportRequest_Mode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UsageReportRequest_Mode.Descriptor instead.
func (UsageReportRequest_Mode) EnumDescriptor() ([]byte, []int) {
	return file_service_v1alpha1_usage_stats_proto_rawDescGZIP(), []int{0, 0}
}

type UsageReportRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApiKeyId string `protobuf:"bytes,1,opt,name=api_key_id,json=apiKeyId,proto3" json:"api_key_id,omitempty"`
	// user_model_name The name of the model that the user is using, such as
	// "kebe/mnist".
	UserModelName string `protobuf:"bytes,2,opt,name=user_model_name,json=userModelName,proto3" json:"user_model_name,omitempty"`
	// upstream_model_name The name of the model that the gateway send the
	// request to, such as "kebe-mnist".
	UpstreamModelName string                    `protobuf:"bytes,3,opt,name=upstream_model_name,json=upstreamModelName,proto3" json:"upstream_model_name,omitempty"`
	Usage             *UsageReportRequest_Usage `protobuf:"bytes,4,opt,name=usage,proto3" json:"usage,omitempty"`
	Mode              UsageReportRequest_Mode   `protobuf:"varint,5,opt,name=mode,proto3,enum=knoway.service.v1alpha1.UsageReportRequest_Mode" json:"mode,omitempty"`
}

func (x *UsageReportRequest) Reset() {
	*x = UsageReportRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_v1alpha1_usage_stats_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UsageReportRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsageReportRequest) ProtoMessage() {}

func (x *UsageReportRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_v1alpha1_usage_stats_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsageReportRequest.ProtoReflect.Descriptor instead.
func (*UsageReportRequest) Descriptor() ([]byte, []int) {
	return file_service_v1alpha1_usage_stats_proto_rawDescGZIP(), []int{0}
}

func (x *UsageReportRequest) GetApiKeyId() string {
	if x != nil {
		return x.ApiKeyId
	}
	return ""
}

func (x *UsageReportRequest) GetUserModelName() string {
	if x != nil {
		return x.UserModelName
	}
	return ""
}

func (x *UsageReportRequest) GetUpstreamModelName() string {
	if x != nil {
		return x.UpstreamModelName
	}
	return ""
}

func (x *UsageReportRequest) GetUsage() *UsageReportRequest_Usage {
	if x != nil {
		return x.Usage
	}
	return nil
}

func (x *UsageReportRequest) GetMode() UsageReportRequest_Mode {
	if x != nil {
		return x.Mode
	}
	return UsageReportRequest_MODE_UNSPECIFIED
}

type UsageReportResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// accepted required: If it is true, it means that the report is successful.
	Accepted bool `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
}

func (x *UsageReportResponse) Reset() {
	*x = UsageReportResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_v1alpha1_usage_stats_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UsageReportResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsageReportResponse) ProtoMessage() {}

func (x *UsageReportResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_v1alpha1_usage_stats_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsageReportResponse.ProtoReflect.Descriptor instead.
func (*UsageReportResponse) Descriptor() ([]byte, []int) {
	return file_service_v1alpha1_usage_stats_proto_rawDescGZIP(), []int{1}
}

func (x *UsageReportResponse) GetAccepted() bool {
	if x != nil {
		return x.Accepted
	}
	return false
}

type UsageReportRequest_Usage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InputTokens  int32 `protobuf:"varint,1,opt,name=input_tokens,json=inputTokens,proto3" json:"input_tokens,omitempty"`
	OutputTokens int32 `protobuf:"varint,2,opt,name=output_tokens,json=outputTokens,proto3" json:"output_tokens,omitempty"`
}

func (x *UsageReportRequest_Usage) Reset() {
	*x = UsageReportRequest_Usage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_v1alpha1_usage_stats_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UsageReportRequest_Usage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsageReportRequest_Usage) ProtoMessage() {}

func (x *UsageReportRequest_Usage) ProtoReflect() protoreflect.Message {
	mi := &file_service_v1alpha1_usage_stats_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsageReportRequest_Usage.ProtoReflect.Descriptor instead.
func (*UsageReportRequest_Usage) Descriptor() ([]byte, []int) {
	return file_service_v1alpha1_usage_stats_proto_rawDescGZIP(), []int{0, 0}
}

func (x *UsageReportRequest_Usage) GetInputTokens() int32 {
	if x != nil {
		return x.InputTokens
	}
	return 0
}

func (x *UsageReportRequest_Usage) GetOutputTokens() int32 {
	if x != nil {
		return x.OutputTokens
	}
	return 0
}

var File_service_v1alpha1_usage_stats_proto protoreflect.FileDescriptor

var file_service_v1alpha1_usage_stats_proto_rawDesc = []byte{
	0x0a, 0x22, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2f, 0x75, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22, 0x99, 0x03,
	0x0a, 0x12, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x0a, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x70, 0x69, 0x4b, 0x65, 0x79,
	0x49, 0x64, 0x12, 0x26, 0x0a, 0x0f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x75, 0x73, 0x65,
	0x72, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x13, 0x75, 0x70,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x47, 0x0a, 0x05, 0x75, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x6b, 0x6e, 0x6f, 0x77,
	0x61, 0x79, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x05, 0x75, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x44, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x30, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x73, 0x61, 0x67,
	0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4d,
	0x6f, 0x64, 0x65, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x1a, 0x4f, 0x0a, 0x05, 0x55, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x5f, 0x74, 0x6f, 0x6b, 0x65,
	0x6e, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x5f,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x6f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x22, 0x2d, 0x0a, 0x04, 0x4d, 0x6f,
	0x64, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x50, 0x45, 0x52, 0x5f,
	0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x01, 0x22, 0x31, 0x0a, 0x13, 0x55, 0x73, 0x61,
	0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x08, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x32, 0x7f, 0x0a, 0x11,
	0x55, 0x73, 0x61, 0x67, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x6a, 0x0a, 0x0b, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x12, 0x2b, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e,
	0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x55, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x21, 0x5a,
	0x1f, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_service_v1alpha1_usage_stats_proto_rawDescOnce sync.Once
	file_service_v1alpha1_usage_stats_proto_rawDescData = file_service_v1alpha1_usage_stats_proto_rawDesc
)

func file_service_v1alpha1_usage_stats_proto_rawDescGZIP() []byte {
	file_service_v1alpha1_usage_stats_proto_rawDescOnce.Do(func() {
		file_service_v1alpha1_usage_stats_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_v1alpha1_usage_stats_proto_rawDescData)
	})
	return file_service_v1alpha1_usage_stats_proto_rawDescData
}

var file_service_v1alpha1_usage_stats_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_service_v1alpha1_usage_stats_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_service_v1alpha1_usage_stats_proto_goTypes = []interface{}{
	(UsageReportRequest_Mode)(0),     // 0: knoway.service.v1alpha1.UsageReportRequest.Mode
	(*UsageReportRequest)(nil),       // 1: knoway.service.v1alpha1.UsageReportRequest
	(*UsageReportResponse)(nil),      // 2: knoway.service.v1alpha1.UsageReportResponse
	(*UsageReportRequest_Usage)(nil), // 3: knoway.service.v1alpha1.UsageReportRequest.Usage
}
var file_service_v1alpha1_usage_stats_proto_depIdxs = []int32{
	3, // 0: knoway.service.v1alpha1.UsageReportRequest.usage:type_name -> knoway.service.v1alpha1.UsageReportRequest.Usage
	0, // 1: knoway.service.v1alpha1.UsageReportRequest.mode:type_name -> knoway.service.v1alpha1.UsageReportRequest.Mode
	1, // 2: knoway.service.v1alpha1.UsageStatsService.UsageReport:input_type -> knoway.service.v1alpha1.UsageReportRequest
	2, // 3: knoway.service.v1alpha1.UsageStatsService.UsageReport:output_type -> knoway.service.v1alpha1.UsageReportResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_service_v1alpha1_usage_stats_proto_init() }
func file_service_v1alpha1_usage_stats_proto_init() {
	if File_service_v1alpha1_usage_stats_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_service_v1alpha1_usage_stats_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UsageReportRequest); i {
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
		file_service_v1alpha1_usage_stats_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UsageReportResponse); i {
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
		file_service_v1alpha1_usage_stats_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UsageReportRequest_Usage); i {
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
			RawDescriptor: file_service_v1alpha1_usage_stats_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_v1alpha1_usage_stats_proto_goTypes,
		DependencyIndexes: file_service_v1alpha1_usage_stats_proto_depIdxs,
		EnumInfos:         file_service_v1alpha1_usage_stats_proto_enumTypes,
		MessageInfos:      file_service_v1alpha1_usage_stats_proto_msgTypes,
	}.Build()
	File_service_v1alpha1_usage_stats_proto = out.File
	file_service_v1alpha1_usage_stats_proto_rawDesc = nil
	file_service_v1alpha1_usage_stats_proto_goTypes = nil
	file_service_v1alpha1_usage_stats_proto_depIdxs = nil
}
