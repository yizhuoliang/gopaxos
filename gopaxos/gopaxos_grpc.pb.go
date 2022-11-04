// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.4
// source: gopaxos.proto

package __

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

// ClientReplicaClient is the client API for ClientReplica service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientReplicaClient interface {
	Reqeust(ctx context.Context, in *Command, opts ...grpc.CallOption) (*Empty, error)
	Collect(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Responses, error)
}

type clientReplicaClient struct {
	cc grpc.ClientConnInterface
}

func NewClientReplicaClient(cc grpc.ClientConnInterface) ClientReplicaClient {
	return &clientReplicaClient{cc}
}

func (c *clientReplicaClient) Reqeust(ctx context.Context, in *Command, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/gopaxos.ClientReplica/Reqeust", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientReplicaClient) Collect(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Responses, error) {
	out := new(Responses)
	err := c.cc.Invoke(ctx, "/gopaxos.ClientReplica/Collect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientReplicaServer is the server API for ClientReplica service.
// All implementations must embed UnimplementedClientReplicaServer
// for forward compatibility
type ClientReplicaServer interface {
	Reqeust(context.Context, *Command) (*Empty, error)
	Collect(context.Context, *Empty) (*Responses, error)
	mustEmbedUnimplementedClientReplicaServer()
}

// UnimplementedClientReplicaServer must be embedded to have forward compatible implementations.
type UnimplementedClientReplicaServer struct {
}

func (UnimplementedClientReplicaServer) Reqeust(context.Context, *Command) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reqeust not implemented")
}
func (UnimplementedClientReplicaServer) Collect(context.Context, *Empty) (*Responses, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Collect not implemented")
}
func (UnimplementedClientReplicaServer) mustEmbedUnimplementedClientReplicaServer() {}

// UnsafeClientReplicaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientReplicaServer will
// result in compilation errors.
type UnsafeClientReplicaServer interface {
	mustEmbedUnimplementedClientReplicaServer()
}

func RegisterClientReplicaServer(s grpc.ServiceRegistrar, srv ClientReplicaServer) {
	s.RegisterService(&ClientReplica_ServiceDesc, srv)
}

func _ClientReplica_Reqeust_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Command)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientReplicaServer).Reqeust(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gopaxos.ClientReplica/Reqeust",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientReplicaServer).Reqeust(ctx, req.(*Command))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientReplica_Collect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientReplicaServer).Collect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gopaxos.ClientReplica/Collect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientReplicaServer).Collect(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientReplica_ServiceDesc is the grpc.ServiceDesc for ClientReplica service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientReplica_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gopaxos.ClientReplica",
	HandlerType: (*ClientReplicaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Reqeust",
			Handler:    _ClientReplica_Reqeust_Handler,
		},
		{
			MethodName: "Collect",
			Handler:    _ClientReplica_Collect_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gopaxos.proto",
}

// ReplicaLeaderClient is the client API for ReplicaLeader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReplicaLeaderClient interface {
	Propose(ctx context.Context, in *Proposal, opts ...grpc.CallOption) (*Empty, error)
	Collect(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Decisions, error)
}

type replicaLeaderClient struct {
	cc grpc.ClientConnInterface
}

func NewReplicaLeaderClient(cc grpc.ClientConnInterface) ReplicaLeaderClient {
	return &replicaLeaderClient{cc}
}

func (c *replicaLeaderClient) Propose(ctx context.Context, in *Proposal, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/gopaxos.ReplicaLeader/Propose", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicaLeaderClient) Collect(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Decisions, error) {
	out := new(Decisions)
	err := c.cc.Invoke(ctx, "/gopaxos.ReplicaLeader/Collect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicaLeaderServer is the server API for ReplicaLeader service.
// All implementations must embed UnimplementedReplicaLeaderServer
// for forward compatibility
type ReplicaLeaderServer interface {
	Propose(context.Context, *Proposal) (*Empty, error)
	Collect(context.Context, *Empty) (*Decisions, error)
	mustEmbedUnimplementedReplicaLeaderServer()
}

// UnimplementedReplicaLeaderServer must be embedded to have forward compatible implementations.
type UnimplementedReplicaLeaderServer struct {
}

func (UnimplementedReplicaLeaderServer) Propose(context.Context, *Proposal) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Propose not implemented")
}
func (UnimplementedReplicaLeaderServer) Collect(context.Context, *Empty) (*Decisions, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Collect not implemented")
}
func (UnimplementedReplicaLeaderServer) mustEmbedUnimplementedReplicaLeaderServer() {}

// UnsafeReplicaLeaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplicaLeaderServer will
// result in compilation errors.
type UnsafeReplicaLeaderServer interface {
	mustEmbedUnimplementedReplicaLeaderServer()
}

func RegisterReplicaLeaderServer(s grpc.ServiceRegistrar, srv ReplicaLeaderServer) {
	s.RegisterService(&ReplicaLeader_ServiceDesc, srv)
}

func _ReplicaLeader_Propose_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Proposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaLeaderServer).Propose(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gopaxos.ReplicaLeader/Propose",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaLeaderServer).Propose(ctx, req.(*Proposal))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplicaLeader_Collect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicaLeaderServer).Collect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gopaxos.ReplicaLeader/Collect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicaLeaderServer).Collect(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// ReplicaLeader_ServiceDesc is the grpc.ServiceDesc for ReplicaLeader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReplicaLeader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gopaxos.ReplicaLeader",
	HandlerType: (*ReplicaLeaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Propose",
			Handler:    _ReplicaLeader_Propose_Handler,
		},
		{
			MethodName: "Collect",
			Handler:    _ReplicaLeader_Collect_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gopaxos.proto",
}

// LeaderAcceptorClient is the client API for LeaderAcceptor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LeaderAcceptorClient interface {
	Scouting(ctx context.Context, in *P1A, opts ...grpc.CallOption) (*P1B, error)
	Commanding(ctx context.Context, in *P2A, opts ...grpc.CallOption) (*P2B, error)
}

type leaderAcceptorClient struct {
	cc grpc.ClientConnInterface
}

func NewLeaderAcceptorClient(cc grpc.ClientConnInterface) LeaderAcceptorClient {
	return &leaderAcceptorClient{cc}
}

func (c *leaderAcceptorClient) Scouting(ctx context.Context, in *P1A, opts ...grpc.CallOption) (*P1B, error) {
	out := new(P1B)
	err := c.cc.Invoke(ctx, "/gopaxos.LeaderAcceptor/Scouting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *leaderAcceptorClient) Commanding(ctx context.Context, in *P2A, opts ...grpc.CallOption) (*P2B, error) {
	out := new(P2B)
	err := c.cc.Invoke(ctx, "/gopaxos.LeaderAcceptor/Commanding", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LeaderAcceptorServer is the server API for LeaderAcceptor service.
// All implementations must embed UnimplementedLeaderAcceptorServer
// for forward compatibility
type LeaderAcceptorServer interface {
	Scouting(context.Context, *P1A) (*P1B, error)
	Commanding(context.Context, *P2A) (*P2B, error)
	mustEmbedUnimplementedLeaderAcceptorServer()
}

// UnimplementedLeaderAcceptorServer must be embedded to have forward compatible implementations.
type UnimplementedLeaderAcceptorServer struct {
}

func (UnimplementedLeaderAcceptorServer) Scouting(context.Context, *P1A) (*P1B, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Scouting not implemented")
}
func (UnimplementedLeaderAcceptorServer) Commanding(context.Context, *P2A) (*P2B, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commanding not implemented")
}
func (UnimplementedLeaderAcceptorServer) mustEmbedUnimplementedLeaderAcceptorServer() {}

// UnsafeLeaderAcceptorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LeaderAcceptorServer will
// result in compilation errors.
type UnsafeLeaderAcceptorServer interface {
	mustEmbedUnimplementedLeaderAcceptorServer()
}

func RegisterLeaderAcceptorServer(s grpc.ServiceRegistrar, srv LeaderAcceptorServer) {
	s.RegisterService(&LeaderAcceptor_ServiceDesc, srv)
}

func _LeaderAcceptor_Scouting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(P1A)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LeaderAcceptorServer).Scouting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gopaxos.LeaderAcceptor/Scouting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LeaderAcceptorServer).Scouting(ctx, req.(*P1A))
	}
	return interceptor(ctx, in, info, handler)
}

func _LeaderAcceptor_Commanding_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(P2A)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LeaderAcceptorServer).Commanding(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gopaxos.LeaderAcceptor/Commanding",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LeaderAcceptorServer).Commanding(ctx, req.(*P2A))
	}
	return interceptor(ctx, in, info, handler)
}

// LeaderAcceptor_ServiceDesc is the grpc.ServiceDesc for LeaderAcceptor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LeaderAcceptor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gopaxos.LeaderAcceptor",
	HandlerType: (*LeaderAcceptorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Scouting",
			Handler:    _LeaderAcceptor_Scouting_Handler,
		},
		{
			MethodName: "Commanding",
			Handler:    _LeaderAcceptor_Commanding_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gopaxos.proto",
}
