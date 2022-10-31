// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.19.4
// source: gopaxos.proto

package __

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Command struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId  int32  `protobuf:"varint,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	CommandId string `protobuf:"bytes,2,opt,name=commandId,proto3" json:"commandId,omitempty"`
	Operation string `protobuf:"bytes,3,opt,name=operation,proto3" json:"operation,omitempty"`
}

func (x *Command) Reset() {
	*x = Command{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Command) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Command) ProtoMessage() {}

func (x *Command) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Command.ProtoReflect.Descriptor instead.
func (*Command) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{0}
}

func (x *Command) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *Command) GetCommandId() string {
	if x != nil {
		return x.CommandId
	}
	return ""
}

func (x *Command) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

// should not send response to client until the command's slot is decided
type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CommandId string `protobuf:"bytes,1,opt,name=commandId,proto3" json:"commandId,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetCommandId() string {
	if x != nil {
		return x.CommandId
	}
	return ""
}

type Responses struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Valid     bool        `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	Responses []*Response `protobuf:"bytes,2,rep,name=responses,proto3" json:"responses,omitempty"`
}

func (x *Responses) Reset() {
	*x = Responses{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Responses) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Responses) ProtoMessage() {}

func (x *Responses) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Responses.ProtoReflect.Descriptor instead.
func (*Responses) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{2}
}

func (x *Responses) GetValid() bool {
	if x != nil {
		return x.Valid
	}
	return false
}

func (x *Responses) GetResponses() []*Response {
	if x != nil {
		return x.Responses
	}
	return nil
}

type Proposal struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SlotNumber int32    `protobuf:"varint,1,opt,name=slotNumber,proto3" json:"slotNumber,omitempty"`
	Command    *Command `protobuf:"bytes,2,opt,name=command,proto3" json:"command,omitempty"`
}

func (x *Proposal) Reset() {
	*x = Proposal{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Proposal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Proposal) ProtoMessage() {}

func (x *Proposal) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Proposal.ProtoReflect.Descriptor instead.
func (*Proposal) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{3}
}

func (x *Proposal) GetSlotNumber() int32 {
	if x != nil {
		return x.SlotNumber
	}
	return 0
}

func (x *Proposal) GetCommand() *Command {
	if x != nil {
		return x.Command
	}
	return nil
}

// Why we need to send the entire command message?
// because a replica may receive decisions not in its command pool
type Decision struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SlotNumber int32    `protobuf:"varint,1,opt,name=slotNumber,proto3" json:"slotNumber,omitempty"`
	Command    *Command `protobuf:"bytes,2,opt,name=command,proto3" json:"command,omitempty"`
}

func (x *Decision) Reset() {
	*x = Decision{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Decision) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Decision) ProtoMessage() {}

func (x *Decision) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Decision.ProtoReflect.Descriptor instead.
func (*Decision) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{4}
}

func (x *Decision) GetSlotNumber() int32 {
	if x != nil {
		return x.SlotNumber
	}
	return 0
}

func (x *Decision) GetCommand() *Command {
	if x != nil {
		return x.Command
	}
	return nil
}

type Decisions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Valid     bool        `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	Decisions []*Decision `protobuf:"bytes,2,rep,name=decisions,proto3" json:"decisions,omitempty"`
}

func (x *Decisions) Reset() {
	*x = Decisions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Decisions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Decisions) ProtoMessage() {}

func (x *Decisions) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Decisions.ProtoReflect.Descriptor instead.
func (*Decisions) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{5}
}

func (x *Decisions) GetValid() bool {
	if x != nil {
		return x.Valid
	}
	return false
}

func (x *Decisions) GetDecisions() []*Decision {
	if x != nil {
		return x.Decisions
	}
	return nil
}

type BSC struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BallotNumber int32    `protobuf:"varint,1,opt,name=ballotNumber,proto3" json:"ballotNumber,omitempty"`
	SlotNumber   int32    `protobuf:"varint,2,opt,name=slotNumber,proto3" json:"slotNumber,omitempty"`
	Command      *Command `protobuf:"bytes,3,opt,name=command,proto3" json:"command,omitempty"`
}

func (x *BSC) Reset() {
	*x = BSC{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BSC) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BSC) ProtoMessage() {}

func (x *BSC) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BSC.ProtoReflect.Descriptor instead.
func (*BSC) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{6}
}

func (x *BSC) GetBallotNumber() int32 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

func (x *BSC) GetSlotNumber() int32 {
	if x != nil {
		return x.SlotNumber
	}
	return 0
}

func (x *BSC) GetCommand() *Command {
	if x != nil {
		return x.Command
	}
	return nil
}

type P1A struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LeaderId     int32 `protobuf:"varint,1,opt,name=leaderId,proto3" json:"leaderId,omitempty"`
	BallotNumber int32 `protobuf:"varint,2,opt,name=ballotNumber,proto3" json:"ballotNumber,omitempty"`
}

func (x *P1A) Reset() {
	*x = P1A{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *P1A) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*P1A) ProtoMessage() {}

func (x *P1A) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use P1A.ProtoReflect.Descriptor instead.
func (*P1A) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{7}
}

func (x *P1A) GetLeaderId() int32 {
	if x != nil {
		return x.LeaderId
	}
	return 0
}

func (x *P1A) GetBallotNumber() int32 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

type P1B struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AcceptorId   int32  `protobuf:"varint,1,opt,name=acceptorId,proto3" json:"acceptorId,omitempty"`
	BallotNumber int32  `protobuf:"varint,2,opt,name=ballotNumber,proto3" json:"ballotNumber,omitempty"`
	Accepted     []*BSC `protobuf:"bytes,3,rep,name=accepted,proto3" json:"accepted,omitempty"`
}

func (x *P1B) Reset() {
	*x = P1B{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *P1B) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*P1B) ProtoMessage() {}

func (x *P1B) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use P1B.ProtoReflect.Descriptor instead.
func (*P1B) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{8}
}

func (x *P1B) GetAcceptorId() int32 {
	if x != nil {
		return x.AcceptorId
	}
	return 0
}

func (x *P1B) GetBallotNumber() int32 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

func (x *P1B) GetAccepted() []*BSC {
	if x != nil {
		return x.Accepted
	}
	return nil
}

type P2A struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LeaderId int32 `protobuf:"varint,1,opt,name=leaderId,proto3" json:"leaderId,omitempty"`
	Bsc      *BSC  `protobuf:"bytes,2,opt,name=bsc,proto3" json:"bsc,omitempty"`
}

func (x *P2A) Reset() {
	*x = P2A{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *P2A) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*P2A) ProtoMessage() {}

func (x *P2A) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use P2A.ProtoReflect.Descriptor instead.
func (*P2A) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{9}
}

func (x *P2A) GetLeaderId() int32 {
	if x != nil {
		return x.LeaderId
	}
	return 0
}

func (x *P2A) GetBsc() *BSC {
	if x != nil {
		return x.Bsc
	}
	return nil
}

type P2B struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AcceptorId   int32 `protobuf:"varint,1,opt,name=acceptorId,proto3" json:"acceptorId,omitempty"`
	BallotNumber int32 `protobuf:"varint,2,opt,name=ballotNumber,proto3" json:"ballotNumber,omitempty"`
}

func (x *P2B) Reset() {
	*x = P2B{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *P2B) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*P2B) ProtoMessage() {}

func (x *P2B) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use P2B.ProtoReflect.Descriptor instead.
func (*P2B) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{10}
}

func (x *P2B) GetAcceptorId() int32 {
	if x != nil {
		return x.AcceptorId
	}
	return 0
}

func (x *P2B) GetBallotNumber() int32 {
	if x != nil {
		return x.BallotNumber
	}
	return 0
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gopaxos_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_gopaxos_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_gopaxos_proto_rawDescGZIP(), []int{11}
}

func (x *Empty) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

var File_gopaxos_proto protoreflect.FileDescriptor

var file_gopaxos_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x22, 0x61, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12,
	0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x49, 0x64, 0x12, 0x1c, 0x0a,
	0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x28, 0x0a, 0x08, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x61,
	0x6e, 0x64, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x49, 0x64, 0x22, 0x52, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x12, 0x2f, 0x0a, 0x09, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x67, 0x6f,
	0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x09,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x22, 0x56, 0x0a, 0x08, 0x50, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x61, 0x6c, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d,
	0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x73, 0x6c, 0x6f, 0x74, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73,
	0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x22, 0x56, 0x0a, 0x08, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1e, 0x0a,
	0x0a, 0x73, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x73, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a, 0x0a,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22, 0x52, 0x0a, 0x09, 0x44, 0x65, 0x63,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x12, 0x2f, 0x0a, 0x09,
	0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x11, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x52, 0x09, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x75, 0x0a,
	0x03, 0x42, 0x53, 0x43, 0x12, 0x22, 0x0a, 0x0c, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75,
	0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x62, 0x61, 0x6c, 0x6c,
	0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x6c, 0x6f, 0x74,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x73, 0x6c,
	0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x67, 0x6f, 0x70, 0x61,
	0x78, 0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x07, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x22, 0x45, 0x0a, 0x03, 0x50, 0x31, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x6c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x62, 0x61, 0x6c, 0x6c, 0x6f,
	0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x62,
	0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x73, 0x0a, 0x03, 0x50,
	0x31, 0x62, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x6f, 0x72, 0x49, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x6f, 0x72,
	0x49, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x28, 0x0a, 0x08, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74,
	0x65, 0x64, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78,
	0x6f, 0x73, 0x2e, 0x42, 0x53, 0x43, 0x52, 0x08, 0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64,
	0x22, 0x41, 0x0a, 0x03, 0x50, 0x32, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x03, 0x62, 0x73, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0c, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x42, 0x53, 0x43, 0x52, 0x03,
	0x62, 0x73, 0x63, 0x22, 0x49, 0x0a, 0x03, 0x50, 0x32, 0x62, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x6f, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a,
	0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x62, 0x61,
	0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0c, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x21,
	0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x32, 0x6f, 0x0a, 0x0d, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x12, 0x2d, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x65, 0x75, 0x73, 0x74, 0x12, 0x10, 0x2e,
	0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x1a,
	0x0e, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x12, 0x2f, 0x0a, 0x07, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x12, 0x0e, 0x2e, 0x67,
	0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x12, 0x2e, 0x67,
	0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73,
	0x22, 0x00, 0x32, 0x70, 0x0a, 0x0d, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x4c, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x12, 0x2e, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x12, 0x11,
	0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61,
	0x6c, 0x1a, 0x0e, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x00, 0x12, 0x2f, 0x0a, 0x07, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x12, 0x0e,
	0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x12,
	0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x22, 0x00, 0x32, 0x66, 0x0a, 0x0e, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x41, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x6f, 0x72, 0x12, 0x28, 0x0a, 0x08, 0x53, 0x63, 0x6f, 0x75, 0x74, 0x69,
	0x6e, 0x67, 0x12, 0x0c, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x31, 0x61,
	0x1a, 0x0c, 0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x31, 0x62, 0x22, 0x00,
	0x12, 0x2a, 0x0a, 0x0a, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x0c,
	0x2e, 0x67, 0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x32, 0x61, 0x1a, 0x0c, 0x2e, 0x67,
	0x6f, 0x70, 0x61, 0x78, 0x6f, 0x73, 0x2e, 0x50, 0x32, 0x62, 0x22, 0x00, 0x42, 0x04, 0x5a, 0x02,
	0x2e, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gopaxos_proto_rawDescOnce sync.Once
	file_gopaxos_proto_rawDescData = file_gopaxos_proto_rawDesc
)

func file_gopaxos_proto_rawDescGZIP() []byte {
	file_gopaxos_proto_rawDescOnce.Do(func() {
		file_gopaxos_proto_rawDescData = protoimpl.X.CompressGZIP(file_gopaxos_proto_rawDescData)
	})
	return file_gopaxos_proto_rawDescData
}

var file_gopaxos_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_gopaxos_proto_goTypes = []interface{}{
	(*Command)(nil),   // 0: gopaxos.Command
	(*Response)(nil),  // 1: gopaxos.Response
	(*Responses)(nil), // 2: gopaxos.Responses
	(*Proposal)(nil),  // 3: gopaxos.Proposal
	(*Decision)(nil),  // 4: gopaxos.Decision
	(*Decisions)(nil), // 5: gopaxos.Decisions
	(*BSC)(nil),       // 6: gopaxos.BSC
	(*P1A)(nil),       // 7: gopaxos.P1a
	(*P1B)(nil),       // 8: gopaxos.P1b
	(*P2A)(nil),       // 9: gopaxos.P2a
	(*P2B)(nil),       // 10: gopaxos.P2b
	(*Empty)(nil),     // 11: gopaxos.Empty
}
var file_gopaxos_proto_depIdxs = []int32{
	1,  // 0: gopaxos.Responses.responses:type_name -> gopaxos.Response
	0,  // 1: gopaxos.Proposal.command:type_name -> gopaxos.Command
	0,  // 2: gopaxos.Decision.command:type_name -> gopaxos.Command
	4,  // 3: gopaxos.Decisions.decisions:type_name -> gopaxos.Decision
	0,  // 4: gopaxos.BSC.command:type_name -> gopaxos.Command
	6,  // 5: gopaxos.P1b.accepted:type_name -> gopaxos.BSC
	6,  // 6: gopaxos.P2a.bsc:type_name -> gopaxos.BSC
	0,  // 7: gopaxos.ClientReplica.Reqeust:input_type -> gopaxos.Command
	11, // 8: gopaxos.ClientReplica.Collect:input_type -> gopaxos.Empty
	3,  // 9: gopaxos.ReplicaLeader.Propose:input_type -> gopaxos.Proposal
	11, // 10: gopaxos.ReplicaLeader.Collect:input_type -> gopaxos.Empty
	7,  // 11: gopaxos.LeaderAcceptor.Scouting:input_type -> gopaxos.P1a
	9,  // 12: gopaxos.LeaderAcceptor.Commanding:input_type -> gopaxos.P2a
	11, // 13: gopaxos.ClientReplica.Reqeust:output_type -> gopaxos.Empty
	2,  // 14: gopaxos.ClientReplica.Collect:output_type -> gopaxos.Responses
	11, // 15: gopaxos.ReplicaLeader.Propose:output_type -> gopaxos.Empty
	5,  // 16: gopaxos.ReplicaLeader.Collect:output_type -> gopaxos.Decisions
	8,  // 17: gopaxos.LeaderAcceptor.Scouting:output_type -> gopaxos.P1b
	10, // 18: gopaxos.LeaderAcceptor.Commanding:output_type -> gopaxos.P2b
	13, // [13:19] is the sub-list for method output_type
	7,  // [7:13] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_gopaxos_proto_init() }
func file_gopaxos_proto_init() {
	if File_gopaxos_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gopaxos_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Command); i {
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
		file_gopaxos_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
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
		file_gopaxos_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Responses); i {
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
		file_gopaxos_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Proposal); i {
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
		file_gopaxos_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Decision); i {
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
		file_gopaxos_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Decisions); i {
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
		file_gopaxos_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BSC); i {
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
		file_gopaxos_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*P1A); i {
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
		file_gopaxos_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*P1B); i {
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
		file_gopaxos_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*P2A); i {
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
		file_gopaxos_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*P2B); i {
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
		file_gopaxos_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
			RawDescriptor: file_gopaxos_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   3,
		},
		GoTypes:           file_gopaxos_proto_goTypes,
		DependencyIndexes: file_gopaxos_proto_depIdxs,
		MessageInfos:      file_gopaxos_proto_msgTypes,
	}.Build()
	File_gopaxos_proto = out.File
	file_gopaxos_proto_rawDesc = nil
	file_gopaxos_proto_goTypes = nil
	file_gopaxos_proto_depIdxs = nil
}
