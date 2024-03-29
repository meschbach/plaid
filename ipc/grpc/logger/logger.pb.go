// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.17.3
// source: ipc/grpc/logger/logger.proto

package logger

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RegisterDrainRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *RegisterDrainRequest) Reset() {
	*x = RegisterDrainRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterDrainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterDrainRequest) ProtoMessage() {}

func (x *RegisterDrainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterDrainRequest.ProtoReflect.Descriptor instead.
func (*RegisterDrainRequest) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterDrainRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type RegisterDrainReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DrainID       int64 `protobuf:"varint,1,opt,name=drainID,proto3" json:"drainID,omitempty"`
	InitialOffset int64 `protobuf:"varint,2,opt,name=initialOffset,proto3" json:"initialOffset,omitempty"`
}

func (x *RegisterDrainReply) Reset() {
	*x = RegisterDrainReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterDrainReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterDrainReply) ProtoMessage() {}

func (x *RegisterDrainReply) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterDrainReply.ProtoReflect.Descriptor instead.
func (*RegisterDrainReply) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterDrainReply) GetDrainID() int64 {
	if x != nil {
		return x.DrainID
	}
	return 0
}

func (x *RegisterDrainReply) GetInitialOffset() int64 {
	if x != nil {
		return x.InitialOffset
	}
	return 0
}

type ReadDrainRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DrainID int64 `protobuf:"varint,1,opt,name=drainID,proto3" json:"drainID,omitempty"`
	Offset  int64 `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Count   int32 `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *ReadDrainRequest) Reset() {
	*x = ReadDrainRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadDrainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadDrainRequest) ProtoMessage() {}

func (x *ReadDrainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadDrainRequest.ProtoReflect.Descriptor instead.
func (*ReadDrainRequest) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{2}
}

func (x *ReadDrainRequest) GetDrainID() int64 {
	if x != nil {
		return x.DrainID
	}
	return 0
}

func (x *ReadDrainRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *ReadDrainRequest) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type ReadDrainReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entries         []*ReadDrainReply_LogEvent `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
	BeginningOffset int64                      `protobuf:"varint,3,opt,name=beginningOffset,proto3" json:"beginningOffset,omitempty"`
	NextOffset      int64                      `protobuf:"varint,4,opt,name=nextOffset,proto3" json:"nextOffset,omitempty"`
}

func (x *ReadDrainReply) Reset() {
	*x = ReadDrainReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadDrainReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadDrainReply) ProtoMessage() {}

func (x *ReadDrainReply) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadDrainReply.ProtoReflect.Descriptor instead.
func (*ReadDrainReply) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{3}
}

func (x *ReadDrainReply) GetEntries() []*ReadDrainReply_LogEvent {
	if x != nil {
		return x.Entries
	}
	return nil
}

func (x *ReadDrainReply) GetBeginningOffset() int64 {
	if x != nil {
		return x.BeginningOffset
	}
	return 0
}

func (x *ReadDrainReply) GetNextOffset() int64 {
	if x != nil {
		return x.NextOffset
	}
	return 0
}

type CloseDrainRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DrainID int64 `protobuf:"varint,1,opt,name=drainID,proto3" json:"drainID,omitempty"`
}

func (x *CloseDrainRequest) Reset() {
	*x = CloseDrainRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseDrainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseDrainRequest) ProtoMessage() {}

func (x *CloseDrainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseDrainRequest.ProtoReflect.Descriptor instead.
func (*CloseDrainRequest) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{4}
}

func (x *CloseDrainRequest) GetDrainID() int64 {
	if x != nil {
		return x.DrainID
	}
	return 0
}

type CloseDrainReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DrainID int64 `protobuf:"varint,1,opt,name=drainID,proto3" json:"drainID,omitempty"`
}

func (x *CloseDrainReply) Reset() {
	*x = CloseDrainReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseDrainReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseDrainReply) ProtoMessage() {}

func (x *CloseDrainReply) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseDrainReply.ProtoReflect.Descriptor instead.
func (*CloseDrainReply) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{5}
}

func (x *CloseDrainReply) GetDrainID() int64 {
	if x != nil {
		return x.DrainID
	}
	return 0
}

type WatchDrainRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DrainID *int64 `protobuf:"varint,1,opt,name=drainID,proto3,oneof" json:"drainID,omitempty"`
	Close   *bool  `protobuf:"varint,2,opt,name=close,proto3,oneof" json:"close,omitempty"`
}

func (x *WatchDrainRequest) Reset() {
	*x = WatchDrainRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WatchDrainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WatchDrainRequest) ProtoMessage() {}

func (x *WatchDrainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WatchDrainRequest.ProtoReflect.Descriptor instead.
func (*WatchDrainRequest) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{6}
}

func (x *WatchDrainRequest) GetDrainID() int64 {
	if x != nil && x.DrainID != nil {
		return *x.DrainID
	}
	return 0
}

func (x *WatchDrainRequest) GetClose() bool {
	if x != nil && x.Close != nil {
		return *x.Close
	}
	return false
}

type WatchDrainEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Offset int64 `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
}

func (x *WatchDrainEvent) Reset() {
	*x = WatchDrainEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WatchDrainEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WatchDrainEvent) ProtoMessage() {}

func (x *WatchDrainEvent) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WatchDrainEvent.ProtoReflect.Descriptor instead.
func (*WatchDrainEvent) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{7}
}

func (x *WatchDrainEvent) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

type LogOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Kind    string `protobuf:"bytes,1,opt,name=kind,proto3" json:"kind,omitempty"`
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Name    string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Stream  string `protobuf:"bytes,4,opt,name=stream,proto3" json:"stream,omitempty"`
}

func (x *LogOrigin) Reset() {
	*x = LogOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogOrigin) ProtoMessage() {}

func (x *LogOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogOrigin.ProtoReflect.Descriptor instead.
func (*LogOrigin) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{8}
}

func (x *LogOrigin) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *LogOrigin) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *LogOrigin) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *LogOrigin) GetStream() string {
	if x != nil {
		return x.Stream
	}
	return ""
}

type ReadDrainReply_LogEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Offset int64                  `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	When   *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=when,proto3" json:"when,omitempty"`
	Text   string                 `protobuf:"bytes,3,opt,name=text,proto3" json:"text,omitempty"`
	Origin *LogOrigin             `protobuf:"bytes,4,opt,name=origin,proto3" json:"origin,omitempty"`
}

func (x *ReadDrainReply_LogEvent) Reset() {
	*x = ReadDrainReply_LogEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_grpc_logger_logger_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadDrainReply_LogEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadDrainReply_LogEvent) ProtoMessage() {}

func (x *ReadDrainReply_LogEvent) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_grpc_logger_logger_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadDrainReply_LogEvent.ProtoReflect.Descriptor instead.
func (*ReadDrainReply_LogEvent) Descriptor() ([]byte, []int) {
	return file_ipc_grpc_logger_logger_proto_rawDescGZIP(), []int{3, 0}
}

func (x *ReadDrainReply_LogEvent) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *ReadDrainReply_LogEvent) GetWhen() *timestamppb.Timestamp {
	if x != nil {
		return x.When
	}
	return nil
}

func (x *ReadDrainReply_LogEvent) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *ReadDrainReply_LogEvent) GetOrigin() *LogOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

var File_ipc_grpc_logger_logger_proto protoreflect.FileDescriptor

var file_ipc_grpc_logger_logger_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x69, 0x70, 0x63, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x6c, 0x6f, 0x67, 0x67, 0x65,
	0x72, 0x2f, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2a, 0x0a, 0x14, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x22, 0x54, 0x0a, 0x12, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44,
	0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x72, 0x61,
	0x69, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x64, 0x72, 0x61, 0x69,
	0x6e, 0x49, 0x44, 0x12, 0x24, 0x0a, 0x0d, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x4f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x69, 0x6e, 0x69, 0x74,
	0x69, 0x61, 0x6c, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x22, 0x5a, 0x0a, 0x10, 0x52, 0x65, 0x61,
	0x64, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a,
	0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07,
	0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0xa9, 0x02, 0x0a, 0x0e, 0x52, 0x65, 0x61, 0x64, 0x44, 0x72,
	0x61, 0x69, 0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x39, 0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6c, 0x6f, 0x67, 0x67,
	0x65, 0x72, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x2e, 0x4c, 0x6f, 0x67, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x12, 0x28, 0x0a, 0x0f, 0x62, 0x65, 0x67, 0x69, 0x6e, 0x6e, 0x69, 0x6e, 0x67,
	0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x62, 0x65,
	0x67, 0x69, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x1e, 0x0a,
	0x0a, 0x6e, 0x65, 0x78, 0x74, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0a, 0x6e, 0x65, 0x78, 0x74, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x1a, 0x91, 0x01,
	0x0a, 0x08, 0x4c, 0x6f, 0x67, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73,
	0x65, 0x74, 0x12, 0x2e, 0x0a, 0x04, 0x77, 0x68, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x77, 0x68,
	0x65, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x29, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e,
	0x4c, 0x6f, 0x67, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69,
	0x6e, 0x22, 0x2d, 0x0a, 0x11, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49,
	0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44,
	0x22, 0x2b, 0x0a, 0x0f, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x22, 0x63, 0x0a,
	0x11, 0x57, 0x61, 0x74, 0x63, 0x68, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x1d, 0x0a, 0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x07, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x88, 0x01,
	0x01, 0x12, 0x19, 0x0a, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x48, 0x01, 0x52, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0a, 0x0a, 0x08,
	0x5f, 0x64, 0x72, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x63, 0x6c, 0x6f,
	0x73, 0x65, 0x22, 0x29, 0x0a, 0x0f, 0x57, 0x61, 0x74, 0x63, 0x68, 0x44, 0x72, 0x61, 0x69, 0x6e,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x22, 0x65, 0x0a,
	0x09, 0x4c, 0x6f, 0x67, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69,
	0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x18,
	0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x32, 0x96, 0x02, 0x0a, 0x02, 0x56, 0x31, 0x12, 0x49, 0x0a, 0x0d, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x12, 0x1c, 0x2e, 0x6c,
	0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x72,
	0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x6c, 0x6f, 0x67,
	0x67, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x72, 0x61, 0x69,
	0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x3d, 0x0a, 0x09, 0x52, 0x65, 0x61, 0x64, 0x44, 0x72,
	0x61, 0x69, 0x6e, 0x12, 0x18, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x61,
	0x64, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x44, 0x72, 0x61, 0x69, 0x6e,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x40, 0x0a, 0x0a, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x44, 0x72,
	0x61, 0x69, 0x6e, 0x12, 0x19, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x43, 0x6c, 0x6f,
	0x73, 0x65, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17,
	0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x44, 0x72, 0x61,
	0x69, 0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x44, 0x0a, 0x0a, 0x57, 0x61, 0x74, 0x63, 0x68,
	0x44, 0x72, 0x61, 0x69, 0x6e, 0x12, 0x19, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x57,
	0x61, 0x74, 0x63, 0x68, 0x44, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x17, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x57, 0x61, 0x74, 0x63, 0x68, 0x44,
	0x72, 0x61, 0x69, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x28, 0x01, 0x30, 0x01, 0x42, 0x3a, 0x5a,
	0x38, 0x67, 0x69, 0x74, 0x2e, 0x6d, 0x65, 0x73, 0x63, 0x68, 0x62, 0x61, 0x63, 0x68, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x65, 0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2e,
	0x67, 0x69, 0x74, 0x2f, 0x70, 0x6c, 0x61, 0x69, 0x64, 0x2f, 0x69, 0x70, 0x63, 0x2f, 0x67, 0x72,
	0x70, 0x63, 0x2f, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_ipc_grpc_logger_logger_proto_rawDescOnce sync.Once
	file_ipc_grpc_logger_logger_proto_rawDescData = file_ipc_grpc_logger_logger_proto_rawDesc
)

func file_ipc_grpc_logger_logger_proto_rawDescGZIP() []byte {
	file_ipc_grpc_logger_logger_proto_rawDescOnce.Do(func() {
		file_ipc_grpc_logger_logger_proto_rawDescData = protoimpl.X.CompressGZIP(file_ipc_grpc_logger_logger_proto_rawDescData)
	})
	return file_ipc_grpc_logger_logger_proto_rawDescData
}

var file_ipc_grpc_logger_logger_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_ipc_grpc_logger_logger_proto_goTypes = []interface{}{
	(*RegisterDrainRequest)(nil),    // 0: logger.RegisterDrainRequest
	(*RegisterDrainReply)(nil),      // 1: logger.RegisterDrainReply
	(*ReadDrainRequest)(nil),        // 2: logger.ReadDrainRequest
	(*ReadDrainReply)(nil),          // 3: logger.ReadDrainReply
	(*CloseDrainRequest)(nil),       // 4: logger.CloseDrainRequest
	(*CloseDrainReply)(nil),         // 5: logger.CloseDrainReply
	(*WatchDrainRequest)(nil),       // 6: logger.WatchDrainRequest
	(*WatchDrainEvent)(nil),         // 7: logger.WatchDrainEvent
	(*LogOrigin)(nil),               // 8: logger.LogOrigin
	(*ReadDrainReply_LogEvent)(nil), // 9: logger.ReadDrainReply.LogEvent
	(*timestamppb.Timestamp)(nil),   // 10: google.protobuf.Timestamp
}
var file_ipc_grpc_logger_logger_proto_depIdxs = []int32{
	9,  // 0: logger.ReadDrainReply.entries:type_name -> logger.ReadDrainReply.LogEvent
	10, // 1: logger.ReadDrainReply.LogEvent.when:type_name -> google.protobuf.Timestamp
	8,  // 2: logger.ReadDrainReply.LogEvent.origin:type_name -> logger.LogOrigin
	0,  // 3: logger.V1.RegisterDrain:input_type -> logger.RegisterDrainRequest
	2,  // 4: logger.V1.ReadDrain:input_type -> logger.ReadDrainRequest
	4,  // 5: logger.V1.CloseDrain:input_type -> logger.CloseDrainRequest
	6,  // 6: logger.V1.WatchDrain:input_type -> logger.WatchDrainRequest
	1,  // 7: logger.V1.RegisterDrain:output_type -> logger.RegisterDrainReply
	3,  // 8: logger.V1.ReadDrain:output_type -> logger.ReadDrainReply
	5,  // 9: logger.V1.CloseDrain:output_type -> logger.CloseDrainReply
	7,  // 10: logger.V1.WatchDrain:output_type -> logger.WatchDrainEvent
	7,  // [7:11] is the sub-list for method output_type
	3,  // [3:7] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_ipc_grpc_logger_logger_proto_init() }
func file_ipc_grpc_logger_logger_proto_init() {
	if File_ipc_grpc_logger_logger_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ipc_grpc_logger_logger_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterDrainRequest); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterDrainReply); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadDrainRequest); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadDrainReply); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseDrainRequest); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseDrainReply); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WatchDrainRequest); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WatchDrainEvent); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogOrigin); i {
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
		file_ipc_grpc_logger_logger_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadDrainReply_LogEvent); i {
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
	file_ipc_grpc_logger_logger_proto_msgTypes[6].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ipc_grpc_logger_logger_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ipc_grpc_logger_logger_proto_goTypes,
		DependencyIndexes: file_ipc_grpc_logger_logger_proto_depIdxs,
		MessageInfos:      file_ipc_grpc_logger_logger_proto_msgTypes,
	}.Build()
	File_ipc_grpc_logger_logger_proto = out.File
	file_ipc_grpc_logger_logger_proto_rawDesc = nil
	file_ipc_grpc_logger_logger_proto_goTypes = nil
	file_ipc_grpc_logger_logger_proto_depIdxs = nil
}
