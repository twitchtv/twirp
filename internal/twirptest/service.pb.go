// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

package twirptest

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

type Hat struct {
	Size                 int32    `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	Color                string   `protobuf:"bytes,2,opt,name=color,proto3" json:"color,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Hat) Reset()         { *m = Hat{} }
func (m *Hat) String() string { return proto.CompactTextString(m) }
func (*Hat) ProtoMessage()    {}
func (*Hat) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{0}
}

func (m *Hat) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Hat.Unmarshal(m, b)
}
func (m *Hat) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Hat.Marshal(b, m, deterministic)
}
func (m *Hat) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Hat.Merge(m, src)
}
func (m *Hat) XXX_Size() int {
	return xxx_messageInfo_Hat.Size(m)
}
func (m *Hat) XXX_DiscardUnknown() {
	xxx_messageInfo_Hat.DiscardUnknown(m)
}

var xxx_messageInfo_Hat proto.InternalMessageInfo

func (m *Hat) GetSize() int32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *Hat) GetColor() string {
	if m != nil {
		return m.Color
	}
	return ""
}

func (m *Hat) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type Size struct {
	Inches               int32    `protobuf:"varint,1,opt,name=inches,proto3" json:"inches,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Size) Reset()         { *m = Size{} }
func (m *Size) String() string { return proto.CompactTextString(m) }
func (*Size) ProtoMessage()    {}
func (*Size) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{1}
}

func (m *Size) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Size.Unmarshal(m, b)
}
func (m *Size) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Size.Marshal(b, m, deterministic)
}
func (m *Size) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Size.Merge(m, src)
}
func (m *Size) XXX_Size() int {
	return xxx_messageInfo_Size.Size(m)
}
func (m *Size) XXX_DiscardUnknown() {
	xxx_messageInfo_Size.DiscardUnknown(m)
}

var xxx_messageInfo_Size proto.InternalMessageInfo

func (m *Size) GetInches() int32 {
	if m != nil {
		return m.Inches
	}
	return 0
}

func init() {
	proto.RegisterType((*Hat)(nil), "twirp.internal.twirptest.Hat")
	proto.RegisterType((*Size)(nil), "twirp.internal.twirptest.Size")
}

func init() { proto.RegisterFile("service.proto", fileDescriptor_a0b84a42fa06f626) }

var fileDescriptor_a0b84a42fa06f626 = []byte{
	// 186 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0x28, 0x29, 0xcf, 0x2c, 0x2a,
	0xd0, 0xcb, 0xcc, 0x2b, 0x49, 0x2d, 0xca, 0x4b, 0xcc, 0xd1, 0x03, 0x73, 0x4b, 0x52, 0x8b, 0x4b,
	0x94, 0x9c, 0xb9, 0x98, 0x3d, 0x12, 0x4b, 0x84, 0x84, 0xb8, 0x58, 0x8a, 0x33, 0xab, 0x52, 0x25,
	0x18, 0x15, 0x18, 0x35, 0x58, 0x83, 0xc0, 0x6c, 0x21, 0x11, 0x2e, 0xd6, 0xe4, 0xfc, 0x9c, 0xfc,
	0x22, 0x09, 0x26, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x08, 0x07, 0xa4, 0x32, 0x2f, 0x31, 0x37, 0x55,
	0x82, 0x19, 0x2c, 0x08, 0x66, 0x2b, 0xc9, 0x71, 0xb1, 0x04, 0x83, 0x74, 0x88, 0x71, 0xb1, 0x65,
	0xe6, 0x25, 0x67, 0xa4, 0x16, 0x43, 0xcd, 0x81, 0xf2, 0x8c, 0xc2, 0xb9, 0xb8, 0x3d, 0x12, 0x93,
	0x52, 0x8b, 0x52, 0x12, 0x8b, 0x33, 0x52, 0x8b, 0x84, 0x3c, 0xb8, 0xd8, 0x7d, 0x13, 0xb3, 0x53,
	0x41, 0xf6, 0xca, 0xe9, 0xe1, 0x72, 0x99, 0x1e, 0xc8, 0x44, 0x29, 0x59, 0xdc, 0xf2, 0x1e, 0x89,
	0x25, 0x4e, 0xdc, 0x51, 0x9c, 0x70, 0x81, 0x24, 0x36, 0xb0, 0x5f, 0x8d, 0x01, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x97, 0xbc, 0x6e, 0xa2, 0xfc, 0x00, 0x00, 0x00,
}
