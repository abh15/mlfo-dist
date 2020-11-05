// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package momo

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// OrchestrateClient is the client API for Orchestrate service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrchestrateClient interface {
	Deploy(ctx context.Context, in *Pipeline, opts ...grpc.CallOption) (*Status, error)
}

type orchestrateClient struct {
	cc grpc.ClientConnInterface
}

func NewOrchestrateClient(cc grpc.ClientConnInterface) OrchestrateClient {
	return &orchestrateClient{cc}
}

func (c *orchestrateClient) Deploy(ctx context.Context, in *Pipeline, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/momo.Orchestrate/Deploy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrchestrateServer is the server API for Orchestrate service.
// All implementations must embed UnimplementedOrchestrateServer
// for forward compatibility
type OrchestrateServer interface {
	Deploy(context.Context, *Pipeline) (*Status, error)
	mustEmbedUnimplementedOrchestrateServer()
}

// UnimplementedOrchestrateServer must be embedded to have forward compatible implementations.
type UnimplementedOrchestrateServer struct {
}

func (*UnimplementedOrchestrateServer) Deploy(context.Context, *Pipeline) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Deploy not implemented")
}
func (*UnimplementedOrchestrateServer) mustEmbedUnimplementedOrchestrateServer() {}

func RegisterOrchestrateServer(s *grpc.Server, srv OrchestrateServer) {
	s.RegisterService(&_Orchestrate_serviceDesc, srv)
}

func _Orchestrate_Deploy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Pipeline)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrchestrateServer).Deploy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/momo.Orchestrate/Deploy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrchestrateServer).Deploy(ctx, req.(*Pipeline))
	}
	return interceptor(ctx, in, info, handler)
}

var _Orchestrate_serviceDesc = grpc.ServiceDesc{
	ServiceName: "momo.Orchestrate",
	HandlerType: (*OrchestrateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Deploy",
			Handler:    _Orchestrate_Deploy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "momo/momo.proto",
}
