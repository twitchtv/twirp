// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.15.5
// source: service_method_same_name.proto

package service_method_same_name

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

type Msg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Msg) Reset() {
	*x = Msg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_method_same_name_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Msg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Msg) ProtoMessage() {}

func (x *Msg) ProtoReflect() protoreflect.Message {
	mi := &file_service_method_same_name_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Msg.ProtoReflect.Descriptor instead.
func (*Msg) Descriptor() ([]byte, []int) {
	return file_service_method_same_name_proto_rawDescGZIP(), []int{0}
}

var File_service_method_same_name_proto protoreflect.FileDescriptor

var file_service_method_same_name_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x5f, 0x73, 0x61, 0x6d, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x05, 0x0a, 0x03, 0x4d, 0x73, 0x67, 0x32, 0x1c, 0x0a, 0x04, 0x45, 0x63, 0x68, 0x6f, 0x12,
	0x14, 0x0a, 0x04, 0x45, 0x63, 0x68, 0x6f, 0x12, 0x04, 0x2e, 0x4d, 0x73, 0x67, 0x1a, 0x04, 0x2e,
	0x4d, 0x73, 0x67, 0x22, 0x00, 0x42, 0x1a, 0x5a, 0x18, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x5f, 0x73, 0x61, 0x6d, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_service_method_same_name_proto_rawDescOnce sync.Once
	file_service_method_same_name_proto_rawDescData = file_service_method_same_name_proto_rawDesc
)

func file_service_method_same_name_proto_rawDescGZIP() []byte {
	file_service_method_same_name_proto_rawDescOnce.Do(func() {
		file_service_method_same_name_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_method_same_name_proto_rawDescData)
	})
	return file_service_method_same_name_proto_rawDescData
}

var file_service_method_same_name_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_service_method_same_name_proto_goTypes = []interface{}{
	(*Msg)(nil), // 0: Msg
}
var file_service_method_same_name_proto_depIdxs = []int32{
	0, // 0: Echo.Echo:input_type -> Msg
	0, // 1: Echo.Echo:output_type -> Msg
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_service_method_same_name_proto_init() }
func file_service_method_same_name_proto_init() {
	if File_service_method_same_name_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_service_method_same_name_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Msg); i {
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
			RawDescriptor: file_service_method_same_name_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_method_same_name_proto_goTypes,
		DependencyIndexes: file_service_method_same_name_proto_depIdxs,
		MessageInfos:      file_service_method_same_name_proto_msgTypes,
	}.Build()
	File_service_method_same_name_proto = out.File
	file_service_method_same_name_proto_rawDesc = nil
	file_service_method_same_name_proto_goTypes = nil
	file_service_method_same_name_proto_depIdxs = nil
}
