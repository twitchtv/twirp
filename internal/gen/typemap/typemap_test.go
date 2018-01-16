package typemap

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestPb(t *testing.T) []*descriptor.FileDescriptorProto {
	t.Helper()
	f, err := ioutil.ReadFile(filepath.Join("testdata", "fileset.pb"))
	require.NoError(t, err, "unable to read testdata protobuf file")

	set := new(descriptor.FileDescriptorSet)
	err = proto.Unmarshal(f, set)
	require.NoError(t, err, "unable to unmarshal testdata protobuf file")

	return set.File
}

func protoFile(files []*descriptor.FileDescriptorProto, name string) *descriptor.FileDescriptorProto {
	for _, f := range files {
		if filepath.Base(f.GetName()) == name {
			return f
		}
	}
	return nil
}

func service(f *descriptor.FileDescriptorProto, name string) *descriptor.ServiceDescriptorProto {
	for _, s := range f.Service {
		if s.GetName() == name {
			return s
		}
	}
	return nil
}

func method(s *descriptor.ServiceDescriptorProto, name string) *descriptor.MethodDescriptorProto {
	for _, m := range s.Method {
		if m.GetName() == name {
			return m
		}
	}
	return nil
}

func TestNewRegistry(t *testing.T) {
	files := loadTestPb(t)
	file := protoFile(files, "service.proto")
	service := service(file, "ServiceWithManyMethods")

	reg := New(files)

	comments, err := reg.ServiceComments(file, service)
	require.NoError(t, err, "unable to load service comments")
	assert.Equal(t, " ServiceWithManyMethods leading\n", comments.Leading)

	method1 := method(service, "Method1")
	require.NotNil(t, method1)

	method1Input := reg.MethodInputDefinition(method1)
	require.NotNil(t, method1Input)
	assert.Equal(t, "RootMsg", method1Input.Descriptor.GetName())
}
