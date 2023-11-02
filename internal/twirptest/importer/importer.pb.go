// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.19.4
// source: importer.proto

// Test to make sure that importing other packages doesnt break

package importer

import (
	importable "github.com/twitchtv/twirp/internal/twirptest/importable"
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

var File_importer_proto protoreflect.FileDescriptor

var file_importer_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x69, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x21, 0x74, 0x77, 0x69, 0x72, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2e, 0x74, 0x77, 0x69, 0x72, 0x70, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x69, 0x6d, 0x70, 0x6f, 0x72,
	0x74, 0x65, 0x72, 0x1a, 0x10, 0x69, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0x62, 0x0a, 0x04, 0x53, 0x76, 0x63, 0x32, 0x12, 0x5a, 0x0a,
	0x04, 0x53, 0x65, 0x6e, 0x64, 0x12, 0x28, 0x2e, 0x74, 0x77, 0x69, 0x72, 0x70, 0x2e, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x74, 0x77, 0x69, 0x72, 0x70, 0x74, 0x65, 0x73, 0x74,
	0x2e, 0x69, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x2e, 0x4d, 0x73, 0x67, 0x1a,
	0x28, 0x2e, 0x74, 0x77, 0x69, 0x72, 0x70, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2e, 0x74, 0x77, 0x69, 0x72, 0x70, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x69, 0x6d, 0x70, 0x6f, 0x72,
	0x74, 0x61, 0x62, 0x6c, 0x65, 0x2e, 0x4d, 0x73, 0x67, 0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x77, 0x69, 0x74, 0x63, 0x68, 0x74, 0x76,
	0x2f, 0x74, 0x77, 0x69, 0x72, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f,
	0x74, 0x77, 0x69, 0x72, 0x70, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x69, 0x6d, 0x70, 0x6f, 0x72, 0x74,
	0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_importer_proto_goTypes = []interface{}{
	(*importable.Msg)(nil), // 0: twirp.internal.twirptest.importable.Msg
}
var file_importer_proto_depIdxs = []int32{
	0, // 0: twirp.internal.twirptest.importer.Svc2.Send:input_type -> twirp.internal.twirptest.importable.Msg
	0, // 1: twirp.internal.twirptest.importer.Svc2.Send:output_type -> twirp.internal.twirptest.importable.Msg
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_importer_proto_init() }
func file_importer_proto_init() {
	if File_importer_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_importer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_importer_proto_goTypes,
		DependencyIndexes: file_importer_proto_depIdxs,
	}.Build()
	File_importer_proto = out.File
	file_importer_proto_rawDesc = nil
	file_importer_proto_goTypes = nil
	file_importer_proto_depIdxs = nil
}
