// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: route/v1alpha1/route.proto

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

type RouteFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Config *anypb.Any `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *RouteFilter) Reset() {
	*x = RouteFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_v1alpha1_route_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RouteFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouteFilter) ProtoMessage() {}

func (x *RouteFilter) ProtoReflect() protoreflect.Message {
	mi := &file_route_v1alpha1_route_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouteFilter.ProtoReflect.Descriptor instead.
func (*RouteFilter) Descriptor() ([]byte, []int) {
	return file_route_v1alpha1_route_proto_rawDescGZIP(), []int{0}
}

func (x *RouteFilter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RouteFilter) GetConfig() *anypb.Any {
	if x != nil {
		return x.Config
	}
	return nil
}

type StringMatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Match:
	//
	//	*StringMatch_Exact
	//	*StringMatch_Prefix
	Match isStringMatch_Match `protobuf_oneof:"match"`
}

func (x *StringMatch) Reset() {
	*x = StringMatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_v1alpha1_route_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringMatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringMatch) ProtoMessage() {}

func (x *StringMatch) ProtoReflect() protoreflect.Message {
	mi := &file_route_v1alpha1_route_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringMatch.ProtoReflect.Descriptor instead.
func (*StringMatch) Descriptor() ([]byte, []int) {
	return file_route_v1alpha1_route_proto_rawDescGZIP(), []int{1}
}

func (m *StringMatch) GetMatch() isStringMatch_Match {
	if m != nil {
		return m.Match
	}
	return nil
}

func (x *StringMatch) GetExact() string {
	if x, ok := x.GetMatch().(*StringMatch_Exact); ok {
		return x.Exact
	}
	return ""
}

func (x *StringMatch) GetPrefix() string {
	if x, ok := x.GetMatch().(*StringMatch_Prefix); ok {
		return x.Prefix
	}
	return ""
}

type isStringMatch_Match interface {
	isStringMatch_Match()
}

type StringMatch_Exact struct {
	Exact string `protobuf:"bytes,1,opt,name=exact,proto3,oneof"`
}

type StringMatch_Prefix struct {
	Prefix string `protobuf:"bytes,2,opt,name=prefix,proto3,oneof"`
}

func (*StringMatch_Exact) isStringMatch_Match() {}

func (*StringMatch_Prefix) isStringMatch_Match() {}

type Match struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Model   *StringMatch `protobuf:"bytes,1,opt,name=model,proto3" json:"model,omitempty"`
	Message *StringMatch `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Match) Reset() {
	*x = Match{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_v1alpha1_route_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Match) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Match) ProtoMessage() {}

func (x *Match) ProtoReflect() protoreflect.Message {
	mi := &file_route_v1alpha1_route_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Match.ProtoReflect.Descriptor instead.
func (*Match) Descriptor() ([]byte, []int) {
	return file_route_v1alpha1_route_proto_rawDescGZIP(), []int{2}
}

func (x *Match) GetModel() *StringMatch {
	if x != nil {
		return x.Model
	}
	return nil
}

func (x *Match) GetMessage() *StringMatch {
	if x != nil {
		return x.Message
	}
	return nil
}

type Route struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string         `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Matches     []*Match       `protobuf:"bytes,2,rep,name=matches,proto3" json:"matches,omitempty"`
	ClusterName string         `protobuf:"bytes,3,opt,name=clusterName,proto3" json:"clusterName,omitempty"`
	Filters     []*RouteFilter `protobuf:"bytes,4,rep,name=filters,proto3" json:"filters,omitempty"`
}

func (x *Route) Reset() {
	*x = Route{}
	if protoimpl.UnsafeEnabled {
		mi := &file_route_v1alpha1_route_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Route) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Route) ProtoMessage() {}

func (x *Route) ProtoReflect() protoreflect.Message {
	mi := &file_route_v1alpha1_route_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Route.ProtoReflect.Descriptor instead.
func (*Route) Descriptor() ([]byte, []int) {
	return file_route_v1alpha1_route_proto_rawDescGZIP(), []int{3}
}

func (x *Route) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Route) GetMatches() []*Match {
	if x != nil {
		return x.Matches
	}
	return nil
}

func (x *Route) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *Route) GetFilters() []*RouteFilter {
	if x != nil {
		return x.Filters
	}
	return nil
}

var File_route_v1alpha1_route_proto protoreflect.FileDescriptor

var file_route_v1alpha1_route_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6b, 0x6e,
	0x6f, 0x77, 0x61, 0x79, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4f,
	0x0a, 0x0b, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x2c, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22,
	0x48, 0x0a, 0x0b, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x16,
	0x0a, 0x05, 0x65, 0x78, 0x61, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x05, 0x65, 0x78, 0x61, 0x63, 0x74, 0x12, 0x18, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x42, 0x07, 0x0a, 0x05, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x22, 0x7f, 0x0a, 0x05, 0x4d, 0x61, 0x74,
	0x63, 0x68, 0x12, 0x38, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x22, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67,
	0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x3c, 0x0a, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e,
	0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63,
	0x68, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xb3, 0x01, 0x0a, 0x05, 0x52,
	0x6f, 0x75, 0x74, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x36, 0x0a, 0x07, 0x6d, 0x61, 0x74, 0x63,
	0x68, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x6e, 0x6f, 0x77,
	0x61, 0x79, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x07, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73,
	0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x3c, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x72, 0x6f, 0x75,
	0x74, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x6f, 0x75, 0x74,
	0x65, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73,
	0x42, 0x1f, 0x5a, 0x1d, 0x6b, 0x6e, 0x6f, 0x77, 0x61, 0x79, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_route_v1alpha1_route_proto_rawDescOnce sync.Once
	file_route_v1alpha1_route_proto_rawDescData = file_route_v1alpha1_route_proto_rawDesc
)

func file_route_v1alpha1_route_proto_rawDescGZIP() []byte {
	file_route_v1alpha1_route_proto_rawDescOnce.Do(func() {
		file_route_v1alpha1_route_proto_rawDescData = protoimpl.X.CompressGZIP(file_route_v1alpha1_route_proto_rawDescData)
	})
	return file_route_v1alpha1_route_proto_rawDescData
}

var file_route_v1alpha1_route_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_route_v1alpha1_route_proto_goTypes = []interface{}{
	(*RouteFilter)(nil), // 0: knoway.route.v1alpha1.RouteFilter
	(*StringMatch)(nil), // 1: knoway.route.v1alpha1.StringMatch
	(*Match)(nil),       // 2: knoway.route.v1alpha1.Match
	(*Route)(nil),       // 3: knoway.route.v1alpha1.Route
	(*anypb.Any)(nil),   // 4: google.protobuf.Any
}
var file_route_v1alpha1_route_proto_depIdxs = []int32{
	4, // 0: knoway.route.v1alpha1.RouteFilter.config:type_name -> google.protobuf.Any
	1, // 1: knoway.route.v1alpha1.Match.model:type_name -> knoway.route.v1alpha1.StringMatch
	1, // 2: knoway.route.v1alpha1.Match.message:type_name -> knoway.route.v1alpha1.StringMatch
	2, // 3: knoway.route.v1alpha1.Route.matches:type_name -> knoway.route.v1alpha1.Match
	0, // 4: knoway.route.v1alpha1.Route.filters:type_name -> knoway.route.v1alpha1.RouteFilter
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_route_v1alpha1_route_proto_init() }
func file_route_v1alpha1_route_proto_init() {
	if File_route_v1alpha1_route_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_route_v1alpha1_route_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RouteFilter); i {
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
		file_route_v1alpha1_route_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringMatch); i {
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
		file_route_v1alpha1_route_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Match); i {
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
		file_route_v1alpha1_route_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Route); i {
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
	file_route_v1alpha1_route_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*StringMatch_Exact)(nil),
		(*StringMatch_Prefix)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_route_v1alpha1_route_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_route_v1alpha1_route_proto_goTypes,
		DependencyIndexes: file_route_v1alpha1_route_proto_depIdxs,
		MessageInfos:      file_route_v1alpha1_route_proto_msgTypes,
	}.Build()
	File_route_v1alpha1_route_proto = out.File
	file_route_v1alpha1_route_proto_rawDesc = nil
	file_route_v1alpha1_route_proto_goTypes = nil
	file_route_v1alpha1_route_proto_depIdxs = nil
}
