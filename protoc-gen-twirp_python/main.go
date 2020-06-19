// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/twitchtv/twirp/internal/gen"
	"github.com/twitchtv/twirp/internal/gen/stringutils"
	"github.com/twitchtv/twirp/internal/gen/typemap"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(gen.Version)
		os.Exit(0)
	}

	g := newGenerator()
	gen.Main(g)
}

func newGenerator() *generator {
	return &generator{output: new(bytes.Buffer)}
}

type generator struct {
	reg    *typemap.Registry
	output *bytes.Buffer
}

func (g *generator) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	genFiles := gen.FilesToGenerate(in)
	g.reg = typemap.New(in.ProtoFile)

	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range genFiles {
		respFile := g.generateFile(f)
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}

	return resp
}

func (g *generator) generateFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	g.P("# Code generated by protoc-gen-twirp_python ", gen.Version, ", DO NOT EDIT.")
	g.P("# source: ", file.GetName())
	g.P()
	g.P(`try:`)
	g.P(`    import httplib`)
	g.P(`    from urllib2 import Request, HTTPError, urlopen`)
	g.P(`except ImportError:`)
	g.P(`    import http.client as httplib`)
	g.P(`    from urllib.request import Request, urlopen`)
	g.P(`    from urllib.error import HTTPError`)
	g.P(`import json`)
	g.P(`from google.protobuf import symbol_database as _symbol_database`)
	g.P(`import sys`)
	g.P()
	g.P(`_sym_db = _symbol_database.Default()`)
	g.P()
	g.P(`class TwirpException(httplib.HTTPException):`)
	g.P(`    def __init__(self, code, message, meta):`)
	g.P(`        self.code = code`)
	g.P(`        self.message = message`)
	g.P(`        self.meta = meta`)
	g.P(`        super(TwirpException, self).__init__(message)`)
	g.P()
	g.P(`    @classmethod`)
	g.P(`    def from_http_err(cls, err):`)
	g.P(`        try:`)
	g.P(`            jsonerr = json.load(err)`)
	g.P(`            code = jsonerr["code"]`)
	g.P(`            msg = jsonerr["msg"]`)
	g.P(`            meta = jsonerr.get("meta")`)
	g.P(`            if meta is None:`)
	g.P(`                meta = {}`)
	g.P(`        except:`)
	g.P(`            code = "internal"`)
	g.P(`            msg = "Error from intermediary with HTTP status code {} {}".format(`)
	g.P(`                err.code, httplib.responses[err.code],`)
	g.P(`            )`)
	g.P(`            meta = {}`)
	g.P(`        return cls(code, msg, meta)`)
	g.P()
	for _, service := range file.Service {
		g.generateProtobufClient(file, service)
	}

	resp := new(plugin.CodeGeneratorResponse_File)
	resp.Name = proto.String(pyFileName(file))
	resp.Content = proto.String(g.output.String())
	g.output.Reset()

	return resp
}

func (g *generator) generateProtobufClient(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	g.P(`class `, clientName(service), `(object):`)
	comments, err := g.reg.ServiceComments(file, service)
	if err == nil && comments.Leading != "" {
		g.P(`    """`)
		g.printComments(comments, `    `)
		g.P(`    """`)
		g.P()
	}
	g.P(`    def __init__(self, server_address):`)
	g.P(`        """Creates a new client for the `, serviceName(service), ` service.`)
	g.P()
	g.P(`        Args:`)
	g.P(`            server_address: The address of the server to send requests to, in`)
	g.P(`                the full protocol://host:port form.`)
	g.P(`        """`)
	g.P(`        if sys.version_info[0] > 2:`)
	g.P(`            self.__target = server_address`)
	g.P(`        else:`)
	g.P(`            self.__target = server_address.encode('ascii')`)
	g.P(`        self.__service_name = `, strconv.Quote(fullServiceName(file, service)))
	g.P()
	g.P(`    def __make_request(self, body, full_method):`)
	g.P(`        req = Request(`)
	g.P(`            url=self.__target + "/twirp" + full_method,`)
	g.P(`            data=body,`)
	g.P(`            headers={"Content-Type": "application/protobuf"},`)
	g.P(`        )`)
	g.P(`        try:`)
	g.P(`            resp = urlopen(req)`)
	g.P(`        except HTTPError as err:`)
	g.P(`            raise TwirpException.from_http_err(err)`)
	g.P(``)
	g.P(`        return resp.read()`)
	g.P()

	for _, method := range service.Method {
		methName := methodName(method)
		inputName := methodInputName(method)

		// Be careful not to write code that overwrites the input parameter.
		for _, x := range []string{"self", "_sym_db", "full_method", "body",
			"serialize", "deserialize", "resp_str"} {
			if inputName == x {
				inputName = inputName + "_"
			}
		}

		g.P(`    def `, methName, `(self, `, inputName, `):`)
		comments, err := g.reg.MethodComments(file, service, method)
		if err == nil && comments.Leading != "" {
			g.P(`        """`)
			g.printComments(comments, `        `)
			g.P(`        """`)
			g.P()
		}
		g.P(`        serialize = _sym_db.GetSymbol(`,
			strconv.Quote(strings.TrimPrefix(method.GetInputType(), ".")), `).SerializeToString`)
		g.P(`        deserialize = _sym_db.GetSymbol(`,
			strconv.Quote(strings.TrimPrefix(method.GetOutputType(), ".")), `).FromString`)
		g.P()
		g.P(`        full_method = "/{}/{}".format(self.__service_name, `, strconv.Quote(method.GetName()), `)`)
		g.P(`        body = serialize(`, inputName, `)`)
		g.P(`        resp_str = self.__make_request(body=body, full_method=full_method)`)
		g.P(`        return deserialize(resp_str)`)
		g.P()
	}
}

func (g *generator) P(args ...string) {
	for _, v := range args {
		g.output.WriteString(v)
	}
	g.output.WriteByte('\n')
}

func (g *generator) printComments(comments typemap.DefinitionComments, prefix string) {
	text := strings.TrimSuffix(comments.Leading, "\n")
	for _, line := range strings.Split(text, "\n") {
		g.P(prefix, strings.TrimPrefix(line, " "))
	}
}

func serviceName(service *descriptor.ServiceDescriptorProto) string {
	return stringutils.CamelCase(service.GetName())
}

func clientName(service *descriptor.ServiceDescriptorProto) string {
	return serviceName(service) + "Client"
}

func fullServiceName(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := service.GetName()
	if pkg := file.GetPackage(); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func methodName(method *descriptor.MethodDescriptorProto) string {
	return stringutils.SnakeCase(method.GetName())
}

// methodInputName returns the basename of the input type of a method in snake
// case.
func methodInputName(meth *descriptor.MethodDescriptorProto) string {
	fullName := meth.GetInputType()
	split := strings.Split(fullName, ".")
	return stringutils.SnakeCase(split[len(split)-1])
}

func pyFileName(f *descriptor.FileDescriptorProto) string {
	name := *f.Name
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	name += "_pb2_twirp.py"
	return name
}
