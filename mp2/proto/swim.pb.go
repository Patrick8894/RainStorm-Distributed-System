// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: proto/swim.proto

package proto

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

type SWIMMessage_MessageType int32

const (
	SWIMMessage_DIRECT_PING   SWIMMessage_MessageType = 0
	SWIMMessage_INDIRECT_PING SWIMMessage_MessageType = 1 // relay
	SWIMMessage_PONG          SWIMMessage_MessageType = 2 // Take indirect ACK messages
	SWIMMessage_JOIN          SWIMMessage_MessageType = 3
)

// Enum value maps for SWIMMessage_MessageType.
var (
	SWIMMessage_MessageType_name = map[int32]string{
		0: "DIRECT_PING",
		1: "INDIRECT_PING",
		2: "PONG",
		3: "JOIN",
	}
	SWIMMessage_MessageType_value = map[string]int32{
		"DIRECT_PING":   0,
		"INDIRECT_PING": 1,
		"PONG":          2,
		"JOIN":          3,
	}
)

func (x SWIMMessage_MessageType) Enum() *SWIMMessage_MessageType {
	p := new(SWIMMessage_MessageType)
	*p = x
	return p
}

func (x SWIMMessage_MessageType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SWIMMessage_MessageType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_swim_proto_enumTypes[0].Descriptor()
}

func (SWIMMessage_MessageType) Type() protoreflect.EnumType {
	return &file_proto_swim_proto_enumTypes[0]
}

func (x SWIMMessage_MessageType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SWIMMessage_MessageType.Descriptor instead.
func (SWIMMessage_MessageType) EnumDescriptor() ([]byte, []int) {
	return file_proto_swim_proto_rawDescGZIP(), []int{0, 0}
}

// SWIMMessage defines the types of messages exchanged between nodes
type SWIMMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type       SWIMMessage_MessageType `protobuf:"varint,1,opt,name=type,proto3,enum=swim.SWIMMessage_MessageType" json:"type,omitempty"`
	Sender     string                  `protobuf:"bytes,2,opt,name=sender,proto3" json:"sender,omitempty"`
	Target     string                  `protobuf:"bytes,3,opt,name=target,proto3" json:"target,omitempty"`
	Membership []*MembershipInfo       `protobuf:"bytes,10,rep,name=membership,proto3" json:"membership,omitempty"` // For gossiping membership info
}

func (x *SWIMMessage) Reset() {
	*x = SWIMMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_swim_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SWIMMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SWIMMessage) ProtoMessage() {}

func (x *SWIMMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_swim_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SWIMMessage.ProtoReflect.Descriptor instead.
func (*SWIMMessage) Descriptor() ([]byte, []int) {
	return file_proto_swim_proto_rawDescGZIP(), []int{0}
}

func (x *SWIMMessage) GetType() SWIMMessage_MessageType {
	if x != nil {
		return x.Type
	}
	return SWIMMessage_DIRECT_PING
}

func (x *SWIMMessage) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *SWIMMessage) GetTarget() string {
	if x != nil {
		return x.Target
	}
	return ""
}

func (x *SWIMMessage) GetMembership() []*MembershipInfo {
	if x != nil {
		return x.Membership
	}
	return nil
}

// MembershipInfo stores information about a node in the cluster
type MembershipInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MemberID string `protobuf:"bytes,1,opt,name=MemberID,proto3" json:"MemberID,omitempty"` // IP+PORT+TIMESTAMP
	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`     // Can be "Alive", "Suspected", or "Failed"
}

func (x *MembershipInfo) Reset() {
	*x = MembershipInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_swim_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MembershipInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MembershipInfo) ProtoMessage() {}

func (x *MembershipInfo) ProtoReflect() protoreflect.Message {
	mi := &file_proto_swim_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MembershipInfo.ProtoReflect.Descriptor instead.
func (*MembershipInfo) Descriptor() ([]byte, []int) {
	return file_proto_swim_proto_rawDescGZIP(), []int{1}
}

func (x *MembershipInfo) GetMemberID() string {
	if x != nil {
		return x.MemberID
	}
	return ""
}

func (x *MembershipInfo) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

var File_proto_swim_proto protoreflect.FileDescriptor

var file_proto_swim_proto_rawDesc = []byte{
	0x0a, 0x10, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x77, 0x69, 0x6d, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x04, 0x73, 0x77, 0x69, 0x6d, 0x22, 0xed, 0x01, 0x0a, 0x0b, 0x53, 0x57, 0x49,
	0x4d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x31, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d, 0x2e, 0x73, 0x77, 0x69, 0x6d, 0x2e, 0x53, 0x57,
	0x49, 0x4d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73,
	0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x12, 0x34, 0x0a, 0x0a, 0x6d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x73, 0x77, 0x69, 0x6d, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69,
	0x70, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0a, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69,
	0x70, 0x22, 0x45, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x0f, 0x0a, 0x0b, 0x44, 0x49, 0x52, 0x45, 0x43, 0x54, 0x5f, 0x50, 0x49, 0x4e, 0x47, 0x10,
	0x00, 0x12, 0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x44, 0x49, 0x52, 0x45, 0x43, 0x54, 0x5f, 0x50, 0x49,
	0x4e, 0x47, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x4f, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x08,
	0x0a, 0x04, 0x4a, 0x4f, 0x49, 0x4e, 0x10, 0x03, 0x22, 0x44, 0x0a, 0x0e, 0x4d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x4d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x0b,
	0x5a, 0x09, 0x6d, 0x70, 0x32, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_proto_swim_proto_rawDescOnce sync.Once
	file_proto_swim_proto_rawDescData = file_proto_swim_proto_rawDesc
)

func file_proto_swim_proto_rawDescGZIP() []byte {
	file_proto_swim_proto_rawDescOnce.Do(func() {
		file_proto_swim_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_swim_proto_rawDescData)
	})
	return file_proto_swim_proto_rawDescData
}

var file_proto_swim_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_swim_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_swim_proto_goTypes = []any{
	(SWIMMessage_MessageType)(0), // 0: swim.SWIMMessage.MessageType
	(*SWIMMessage)(nil),          // 1: swim.SWIMMessage
	(*MembershipInfo)(nil),       // 2: swim.MembershipInfo
}
var file_proto_swim_proto_depIdxs = []int32{
	0, // 0: swim.SWIMMessage.type:type_name -> swim.SWIMMessage.MessageType
	2, // 1: swim.SWIMMessage.membership:type_name -> swim.MembershipInfo
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_swim_proto_init() }
func file_proto_swim_proto_init() {
	if File_proto_swim_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_swim_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SWIMMessage); i {
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
		file_proto_swim_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*MembershipInfo); i {
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
			RawDescriptor: file_proto_swim_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_swim_proto_goTypes,
		DependencyIndexes: file_proto_swim_proto_depIdxs,
		EnumInfos:         file_proto_swim_proto_enumTypes,
		MessageInfos:      file_proto_swim_proto_msgTypes,
	}.Build()
	File_proto_swim_proto = out.File
	file_proto_swim_proto_rawDesc = nil
	file_proto_swim_proto_goTypes = nil
	file_proto_swim_proto_depIdxs = nil
}
