// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package reswire

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

// ResourceControllerClient is the client API for ResourceController service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ResourceControllerClient interface {
	Create(ctx context.Context, in *CreateResourceIn, opts ...grpc.CallOption) (*CreateResourceOut, error)
	Delete(ctx context.Context, in *DeleteResourceIn, opts ...grpc.CallOption) (*DeleteResourceOut, error)
	Get(ctx context.Context, in *GetIn, opts ...grpc.CallOption) (*GetOut, error)
	Update(ctx context.Context, in *UpdateIn, opts ...grpc.CallOption) (*UpdateOut, error)
	GetStatus(ctx context.Context, in *GetStatusIn, opts ...grpc.CallOption) (*GetStatusOut, error)
	UpdateStatus(ctx context.Context, in *UpdateStatusIn, opts ...grpc.CallOption) (*UpdateStatusOut, error)
	GetEvents(ctx context.Context, in *GetEventsIn, opts ...grpc.CallOption) (*GetEventsOut, error)
	Log(ctx context.Context, in *LogIn, opts ...grpc.CallOption) (*LogOut, error)
	Watcher(ctx context.Context, opts ...grpc.CallOption) (ResourceController_WatcherClient, error)
	List(ctx context.Context, in *ListIn, opts ...grpc.CallOption) (*ListOut, error)
}

type resourceControllerClient struct {
	cc grpc.ClientConnInterface
}

func NewResourceControllerClient(cc grpc.ClientConnInterface) ResourceControllerClient {
	return &resourceControllerClient{cc}
}

func (c *resourceControllerClient) Create(ctx context.Context, in *CreateResourceIn, opts ...grpc.CallOption) (*CreateResourceOut, error) {
	out := new(CreateResourceOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) Delete(ctx context.Context, in *DeleteResourceIn, opts ...grpc.CallOption) (*DeleteResourceOut, error) {
	out := new(DeleteResourceOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) Get(ctx context.Context, in *GetIn, opts ...grpc.CallOption) (*GetOut, error) {
	out := new(GetOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) Update(ctx context.Context, in *UpdateIn, opts ...grpc.CallOption) (*UpdateOut, error) {
	out := new(UpdateOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) GetStatus(ctx context.Context, in *GetStatusIn, opts ...grpc.CallOption) (*GetStatusOut, error) {
	out := new(GetStatusOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) UpdateStatus(ctx context.Context, in *UpdateStatusIn, opts ...grpc.CallOption) (*UpdateStatusOut, error) {
	out := new(UpdateStatusOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/UpdateStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) GetEvents(ctx context.Context, in *GetEventsIn, opts ...grpc.CallOption) (*GetEventsOut, error) {
	out := new(GetEventsOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/GetEvents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) Log(ctx context.Context, in *LogIn, opts ...grpc.CallOption) (*LogOut, error) {
	out := new(LogOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/Log", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceControllerClient) Watcher(ctx context.Context, opts ...grpc.CallOption) (ResourceController_WatcherClient, error) {
	stream, err := c.cc.NewStream(ctx, &ResourceController_ServiceDesc.Streams[0], "/reswire.ResourceController/Watcher", opts...)
	if err != nil {
		return nil, err
	}
	x := &resourceControllerWatcherClient{stream}
	return x, nil
}

type ResourceController_WatcherClient interface {
	Send(*WatcherEventIn) error
	Recv() (*WatcherEventOut, error)
	grpc.ClientStream
}

type resourceControllerWatcherClient struct {
	grpc.ClientStream
}

func (x *resourceControllerWatcherClient) Send(m *WatcherEventIn) error {
	return x.ClientStream.SendMsg(m)
}

func (x *resourceControllerWatcherClient) Recv() (*WatcherEventOut, error) {
	m := new(WatcherEventOut)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *resourceControllerClient) List(ctx context.Context, in *ListIn, opts ...grpc.CallOption) (*ListOut, error) {
	out := new(ListOut)
	err := c.cc.Invoke(ctx, "/reswire.ResourceController/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ResourceControllerServer is the server API for ResourceController service.
// All implementations must embed UnimplementedResourceControllerServer
// for forward compatibility
type ResourceControllerServer interface {
	Create(context.Context, *CreateResourceIn) (*CreateResourceOut, error)
	Delete(context.Context, *DeleteResourceIn) (*DeleteResourceOut, error)
	Get(context.Context, *GetIn) (*GetOut, error)
	Update(context.Context, *UpdateIn) (*UpdateOut, error)
	GetStatus(context.Context, *GetStatusIn) (*GetStatusOut, error)
	UpdateStatus(context.Context, *UpdateStatusIn) (*UpdateStatusOut, error)
	GetEvents(context.Context, *GetEventsIn) (*GetEventsOut, error)
	Log(context.Context, *LogIn) (*LogOut, error)
	Watcher(ResourceController_WatcherServer) error
	List(context.Context, *ListIn) (*ListOut, error)
	mustEmbedUnimplementedResourceControllerServer()
}

// UnimplementedResourceControllerServer must be embedded to have forward compatible implementations.
type UnimplementedResourceControllerServer struct {
}

func (UnimplementedResourceControllerServer) Create(context.Context, *CreateResourceIn) (*CreateResourceOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedResourceControllerServer) Delete(context.Context, *DeleteResourceIn) (*DeleteResourceOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedResourceControllerServer) Get(context.Context, *GetIn) (*GetOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedResourceControllerServer) Update(context.Context, *UpdateIn) (*UpdateOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedResourceControllerServer) GetStatus(context.Context, *GetStatusIn) (*GetStatusOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedResourceControllerServer) UpdateStatus(context.Context, *UpdateStatusIn) (*UpdateStatusOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStatus not implemented")
}
func (UnimplementedResourceControllerServer) GetEvents(context.Context, *GetEventsIn) (*GetEventsOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEvents not implemented")
}
func (UnimplementedResourceControllerServer) Log(context.Context, *LogIn) (*LogOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Log not implemented")
}
func (UnimplementedResourceControllerServer) Watcher(ResourceController_WatcherServer) error {
	return status.Errorf(codes.Unimplemented, "method Watcher not implemented")
}
func (UnimplementedResourceControllerServer) List(context.Context, *ListIn) (*ListOut, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedResourceControllerServer) mustEmbedUnimplementedResourceControllerServer() {}

// UnsafeResourceControllerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ResourceControllerServer will
// result in compilation errors.
type UnsafeResourceControllerServer interface {
	mustEmbedUnimplementedResourceControllerServer()
}

func RegisterResourceControllerServer(s grpc.ServiceRegistrar, srv ResourceControllerServer) {
	s.RegisterService(&ResourceController_ServiceDesc, srv)
}

func _ResourceController_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateResourceIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).Create(ctx, req.(*CreateResourceIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteResourceIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).Delete(ctx, req.(*DeleteResourceIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).Get(ctx, req.(*GetIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).Update(ctx, req.(*UpdateIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/GetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).GetStatus(ctx, req.(*GetStatusIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_UpdateStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateStatusIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).UpdateStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/UpdateStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).UpdateStatus(ctx, req.(*UpdateStatusIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_GetEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEventsIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).GetEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/GetEvents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).GetEvents(ctx, req.(*GetEventsIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_Log_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).Log(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/Log",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).Log(ctx, req.(*LogIn))
	}
	return interceptor(ctx, in, info, handler)
}

func _ResourceController_Watcher_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ResourceControllerServer).Watcher(&resourceControllerWatcherServer{stream})
}

type ResourceController_WatcherServer interface {
	Send(*WatcherEventOut) error
	Recv() (*WatcherEventIn, error)
	grpc.ServerStream
}

type resourceControllerWatcherServer struct {
	grpc.ServerStream
}

func (x *resourceControllerWatcherServer) Send(m *WatcherEventOut) error {
	return x.ServerStream.SendMsg(m)
}

func (x *resourceControllerWatcherServer) Recv() (*WatcherEventIn, error) {
	m := new(WatcherEventIn)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ResourceController_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListIn)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceControllerServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reswire.ResourceController/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceControllerServer).List(ctx, req.(*ListIn))
	}
	return interceptor(ctx, in, info, handler)
}

// ResourceController_ServiceDesc is the grpc.ServiceDesc for ResourceController service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ResourceController_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "reswire.ResourceController",
	HandlerType: (*ResourceControllerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _ResourceController_Create_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ResourceController_Delete_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _ResourceController_Get_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _ResourceController_Update_Handler,
		},
		{
			MethodName: "GetStatus",
			Handler:    _ResourceController_GetStatus_Handler,
		},
		{
			MethodName: "UpdateStatus",
			Handler:    _ResourceController_UpdateStatus_Handler,
		},
		{
			MethodName: "GetEvents",
			Handler:    _ResourceController_GetEvents_Handler,
		},
		{
			MethodName: "Log",
			Handler:    _ResourceController_Log_Handler,
		},
		{
			MethodName: "List",
			Handler:    _ResourceController_List_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Watcher",
			Handler:       _ResourceController_Watcher_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "ipc/grpc/reswire/resources.proto",
}
