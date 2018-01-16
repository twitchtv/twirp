package gen

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Generator interface {
	Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse
}

func Main(g Generator) {
	req := readGenRequest(os.Stdin)
	resp := g.Generate(req)
	writeResponse(os.Stdout, resp)
}

func FilesToGenerate(req *plugin.CodeGeneratorRequest) []*descriptor.FileDescriptorProto {
	genFiles := make([]*descriptor.FileDescriptorProto, 0)
Outer:
	for _, name := range req.FileToGenerate {
		for _, f := range req.ProtoFile {
			if f.GetName() == name {
				genFiles = append(genFiles, f)
				continue Outer
			}
		}
		Fail("could not find file named", name)
	}

	return genFiles
}

func readGenRequest(r io.Reader) *plugin.CodeGeneratorRequest {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		Error(err, "reading input")
	}

	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		Error(err, "parsing input proto")
	}

	if len(req.FileToGenerate) == 0 {
		Fail("no files to generate")
	}

	return req
}

func writeResponse(w io.Writer, resp *plugin.CodeGeneratorResponse) {
	data, err := proto.Marshal(resp)
	if err != nil {
		Error(err, "marshaling response")
	}
	_, err = w.Write(data)
	if err != nil {
		Error(err, "writing response")
	}
}
