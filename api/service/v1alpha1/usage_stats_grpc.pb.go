// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: service/v1alpha1/usage_stats.proto

package v1alpha1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// UsageStatsServiceClient is the client API for UsageStatsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsageStatsServiceClient interface {
	UsageReport(ctx context.Context, in *UsageReportRequest, opts ...grpc.CallOption) (*UsageReportResponse, error)
}

type usageStatsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUsageStatsServiceClient(cc grpc.ClientConnInterface) UsageStatsServiceClient {
	return &usageStatsServiceClient{cc}
}

func (c *usageStatsServiceClient) UsageReport(ctx context.Context, in *UsageReportRequest, opts ...grpc.CallOption) (*UsageReportResponse, error) {
	out := new(UsageReportResponse)
	err := c.cc.Invoke(ctx, "/knoway.service.v1alpha1.UsageStatsService/UsageReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsageStatsServiceServer is the server API for UsageStatsService service.
// All implementations must embed UnimplementedUsageStatsServiceServer
// for forward compatibility
type UsageStatsServiceServer interface {
	UsageReport(context.Context, *UsageReportRequest) (*UsageReportResponse, error)
	mustEmbedUnimplementedUsageStatsServiceServer()
}

// UnimplementedUsageStatsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUsageStatsServiceServer struct {
}

func (UnimplementedUsageStatsServiceServer) UsageReport(context.Context, *UsageReportRequest) (*UsageReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UsageReport not implemented")
}
func (UnimplementedUsageStatsServiceServer) mustEmbedUnimplementedUsageStatsServiceServer() {}

// UnsafeUsageStatsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UsageStatsServiceServer will
// result in compilation errors.
type UnsafeUsageStatsServiceServer interface {
	mustEmbedUnimplementedUsageStatsServiceServer()
}

func RegisterUsageStatsServiceServer(s grpc.ServiceRegistrar, srv UsageStatsServiceServer) {
	s.RegisterService(&UsageStatsService_ServiceDesc, srv)
}

func _UsageStatsService_UsageReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UsageReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageStatsServiceServer).UsageReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/knoway.service.v1alpha1.UsageStatsService/UsageReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageStatsServiceServer).UsageReport(ctx, req.(*UsageReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UsageStatsService_ServiceDesc is the grpc.ServiceDesc for UsageStatsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UsageStatsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "knoway.service.v1alpha1.UsageStatsService",
	HandlerType: (*UsageStatsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UsageReport",
			Handler:    _UsageStatsService_UsageReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/v1alpha1/usage_stats.proto",
}
