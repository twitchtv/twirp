// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.17.1
// source: empty_service.proto

package empty_service

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_empty_service_proto protoreflect.FileDescriptor

var file_empty_service_proto_rawDesc = []byte{
	0x0a, 0x13, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x25, 0x74, 0x77, 0x69, 0x72, 0x70, 0x2e, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x74, 0x77, 0x69, 0x72, 0x70, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x65, 0x6d, 0x70, 0x74, 0x79, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x32, 0x07, 0x0a, 0x05,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x10, 0x5a, 0x0e, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x5f,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_empty_service_proto_goTypes = []interface{}{}
var file_empty_service_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_empty_service_proto_init() }
func file_empty_service_proto_init() {
	if File_empty_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_empty_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_empty_service_proto_goTypes,
		DependencyIndexes: file_empty_service_proto_depIdxs,
	}.Build()
	File_empty_service_proto = out.File
	file_empty_service_proto_rawDesc = nil
	file_empty_service_proto_goTypes = nil
	file_empty_service_proto_depIdxs = nil
}
