// Code generated by protoc-gen-go. DO NOT EDIT.
// source: no_package_name.proto

package no_package_name

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

type Msg struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Msg) Reset()         { *m = Msg{} }
func (m *Msg) String() string { return proto.CompactTextString(m) }
func (*Msg) ProtoMessage()    {}
func (*Msg) Descriptor() ([]byte, []int) {
	return fileDescriptor_542f9b76ea7dac2e, []int{0}
}

func (m *Msg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Msg.Unmarshal(m, b)
}
func (m *Msg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Msg.Marshal(b, m, deterministic)
}
func (m *Msg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Msg.Merge(m, src)
}
func (m *Msg) XXX_Size() int {
	return xxx_messageInfo_Msg.Size(m)
}
func (m *Msg) XXX_DiscardUnknown() {
	xxx_messageInfo_Msg.DiscardUnknown(m)
}

var xxx_messageInfo_Msg proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Msg)(nil), "Msg")
}

func init() { proto.RegisterFile("no_package_name.proto", fileDescriptor_542f9b76ea7dac2e) }

var fileDescriptor_542f9b76ea7dac2e = []byte{
	// 89 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xcd, 0xcb, 0x8f, 0x2f,
	0x48, 0x4c, 0xce, 0x4e, 0x4c, 0x4f, 0x8d, 0xcf, 0x4b, 0xcc, 0x4d, 0xd5, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x57, 0x62, 0xe5, 0x62, 0xf6, 0x2d, 0x4e, 0x37, 0x92, 0xe4, 0x62, 0x0e, 0x2e, 0x4b, 0x16,
	0x12, 0xe2, 0x62, 0x09, 0x4e, 0xcd, 0x4b, 0x11, 0x62, 0xd1, 0xf3, 0x2d, 0x4e, 0x97, 0x02, 0x93,
	0x4e, 0x82, 0x51, 0xfc, 0x68, 0x5a, 0x93, 0xd8, 0xc0, 0x7a, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff,
	0xff, 0xbc, 0x5e, 0x8a, 0x09, 0x54, 0x00, 0x00, 0x00,
}
