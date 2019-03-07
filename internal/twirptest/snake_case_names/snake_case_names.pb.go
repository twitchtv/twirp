// Code generated by protoc-gen-go. DO NOT EDIT.
// source: snake_case_names.proto

// Test that protoc-gen-twirp follows the same behavior as protoc-gen-go
// for converting RPCs and message names from snake case to camel case.

package snake_case_names

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type MakeHatArgsV1 struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MakeHatArgsV1) Reset()         { *m = MakeHatArgsV1{} }
func (m *MakeHatArgsV1) String() string { return proto.CompactTextString(m) }
func (*MakeHatArgsV1) ProtoMessage()    {}
func (*MakeHatArgsV1) Descriptor() ([]byte, []int) {
	return fileDescriptor_c768f27eb22a6056, []int{0}
}

func (m *MakeHatArgsV1) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeHatArgsV1.Unmarshal(m, b)
}
func (m *MakeHatArgsV1) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeHatArgsV1.Marshal(b, m, deterministic)
}
func (m *MakeHatArgsV1) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeHatArgsV1.Merge(m, src)
}
func (m *MakeHatArgsV1) XXX_Size() int {
	return xxx_messageInfo_MakeHatArgsV1.Size(m)
}
func (m *MakeHatArgsV1) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeHatArgsV1.DiscardUnknown(m)
}

var xxx_messageInfo_MakeHatArgsV1 proto.InternalMessageInfo

type MakeHatArgsV1_HatV1 struct {
	Size                 int32    `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	Color                string   `protobuf:"bytes,2,opt,name=color,proto3" json:"color,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MakeHatArgsV1_HatV1) Reset()         { *m = MakeHatArgsV1_HatV1{} }
func (m *MakeHatArgsV1_HatV1) String() string { return proto.CompactTextString(m) }
func (*MakeHatArgsV1_HatV1) ProtoMessage()    {}
func (*MakeHatArgsV1_HatV1) Descriptor() ([]byte, []int) {
	return fileDescriptor_c768f27eb22a6056, []int{0, 0}
}

func (m *MakeHatArgsV1_HatV1) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeHatArgsV1_HatV1.Unmarshal(m, b)
}
func (m *MakeHatArgsV1_HatV1) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeHatArgsV1_HatV1.Marshal(b, m, deterministic)
}
func (m *MakeHatArgsV1_HatV1) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeHatArgsV1_HatV1.Merge(m, src)
}
func (m *MakeHatArgsV1_HatV1) XXX_Size() int {
	return xxx_messageInfo_MakeHatArgsV1_HatV1.Size(m)
}
func (m *MakeHatArgsV1_HatV1) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeHatArgsV1_HatV1.DiscardUnknown(m)
}

var xxx_messageInfo_MakeHatArgsV1_HatV1 proto.InternalMessageInfo

func (m *MakeHatArgsV1_HatV1) GetSize() int32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *MakeHatArgsV1_HatV1) GetColor() string {
	if m != nil {
		return m.Color
	}
	return ""
}

func (m *MakeHatArgsV1_HatV1) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type MakeHatArgsV1_SizeV1 struct {
	Inches               int32    `protobuf:"varint,1,opt,name=inches,proto3" json:"inches,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MakeHatArgsV1_SizeV1) Reset()         { *m = MakeHatArgsV1_SizeV1{} }
func (m *MakeHatArgsV1_SizeV1) String() string { return proto.CompactTextString(m) }
func (*MakeHatArgsV1_SizeV1) ProtoMessage()    {}
func (*MakeHatArgsV1_SizeV1) Descriptor() ([]byte, []int) {
	return fileDescriptor_c768f27eb22a6056, []int{0, 1}
}

func (m *MakeHatArgsV1_SizeV1) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeHatArgsV1_SizeV1.Unmarshal(m, b)
}
func (m *MakeHatArgsV1_SizeV1) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeHatArgsV1_SizeV1.Marshal(b, m, deterministic)
}
func (m *MakeHatArgsV1_SizeV1) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeHatArgsV1_SizeV1.Merge(m, src)
}
func (m *MakeHatArgsV1_SizeV1) XXX_Size() int {
	return xxx_messageInfo_MakeHatArgsV1_SizeV1.Size(m)
}
func (m *MakeHatArgsV1_SizeV1) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeHatArgsV1_SizeV1.DiscardUnknown(m)
}

var xxx_messageInfo_MakeHatArgsV1_SizeV1 proto.InternalMessageInfo

func (m *MakeHatArgsV1_SizeV1) GetInches() int32 {
	if m != nil {
		return m.Inches
	}
	return 0
}

func init() {
	proto.RegisterType((*MakeHatArgsV1)(nil), "twirp.internal.twirptest.snake_case_names.MakeHatArgs_v1")
	proto.RegisterType((*MakeHatArgsV1_HatV1)(nil), "twirp.internal.twirptest.snake_case_names.MakeHatArgs_v1.Hat_v1")
	proto.RegisterType((*MakeHatArgsV1_SizeV1)(nil), "twirp.internal.twirptest.snake_case_names.MakeHatArgs_v1.Size_v1")
}

func init() { proto.RegisterFile("snake_case_names.proto", fileDescriptor_c768f27eb22a6056) }

var fileDescriptor_c768f27eb22a6056 = []byte{
	// 224 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2b, 0xce, 0x4b, 0xcc,
	0x4e, 0x8d, 0x4f, 0x4e, 0x2c, 0x4e, 0x8d, 0xcf, 0x4b, 0xcc, 0x4d, 0x2d, 0xd6, 0x2b, 0x28, 0xca,
	0x2f, 0xc9, 0x17, 0xd2, 0x2c, 0x29, 0xcf, 0x2c, 0x2a, 0xd0, 0xcb, 0xcc, 0x2b, 0x49, 0x2d, 0xca,
	0x4b, 0xcc, 0xd1, 0x03, 0x73, 0x4b, 0x52, 0x8b, 0x4b, 0xf4, 0xd0, 0x35, 0x28, 0x55, 0x73, 0xf1,
	0xf9, 0x26, 0x66, 0xa7, 0x7a, 0x24, 0x96, 0x38, 0x16, 0xa5, 0x17, 0xc7, 0x97, 0x19, 0x4a, 0xb9,
	0x71, 0xb1, 0x79, 0x24, 0x96, 0xc4, 0x97, 0x19, 0x0a, 0x09, 0x71, 0xb1, 0x14, 0x67, 0x56, 0xa5,
	0x4a, 0x30, 0x2a, 0x30, 0x6a, 0xb0, 0x06, 0x81, 0xd9, 0x42, 0x22, 0x5c, 0xac, 0xc9, 0xf9, 0x39,
	0xf9, 0x45, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x10, 0x0e, 0x48, 0x25, 0xc8, 0x38, 0x09,
	0x66, 0xb0, 0x20, 0x98, 0x2d, 0xa5, 0xc8, 0xc5, 0x1e, 0x9c, 0x59, 0x95, 0x0a, 0x32, 0x48, 0x8c,
	0x8b, 0x2d, 0x33, 0x2f, 0x39, 0x23, 0xb5, 0x18, 0x6a, 0x14, 0x94, 0x67, 0xb4, 0x90, 0x91, 0x8b,
	0xdb, 0x23, 0x31, 0x29, 0xb5, 0x28, 0x25, 0xb1, 0x38, 0x23, 0xb5, 0x48, 0x68, 0x22, 0x23, 0x17,
	0x17, 0xd4, 0x35, 0x20, 0x6d, 0x8e, 0x7a, 0x44, 0xfb, 0x43, 0x0f, 0xd5, 0x13, 0x7a, 0x50, 0x9b,
	0xa5, 0x1c, 0xc8, 0x37, 0x02, 0xe2, 0x08, 0x27, 0xa1, 0x28, 0x01, 0x74, 0x95, 0x49, 0x6c, 0xe0,
	0x60, 0x36, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xe5, 0x7b, 0xaf, 0xa9, 0x80, 0x01, 0x00, 0x00,
}
