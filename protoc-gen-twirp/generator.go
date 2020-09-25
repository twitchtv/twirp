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
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"path"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pkg/errors"
	"github.com/twitchtv/twirp/internal/gen"
	"github.com/twitchtv/twirp/internal/gen/stringutils"
	"github.com/twitchtv/twirp/internal/gen/typemap"
)

type twirp struct {
	filesHandled int

	reg *typemap.Registry

	// Map to record whether we've built each package
	pkgs          map[string]string
	pkgNamesInUse map[string]bool

	importPrefix string            // String to prefix to imported package file names.
	importMap    map[string]string // Mapping from .proto file name to import path.

	// Package output:
	sourceRelativePaths bool // instruction on where to write output files

	// Package naming:
	genPkgName          string // Name of the package that we're generating
	fileToGoPackageName map[*descriptor.FileDescriptorProto]string

	// List of files that were inputs to the generator. We need to hold this in
	// the struct so we can write a header for the file that lists its inputs.
	genFiles []*descriptor.FileDescriptorProto

	// Output buffer that holds the bytes we want to write out for a single file.
	// Gets reset after working on a file.
	output *bytes.Buffer
}

func newGenerator() *twirp {
	t := &twirp{
		pkgs:                make(map[string]string),
		pkgNamesInUse:       make(map[string]bool),
		importMap:           make(map[string]string),
		fileToGoPackageName: make(map[*descriptor.FileDescriptorProto]string),
		output:              bytes.NewBuffer(nil),
	}

	return t
}

func (t *twirp) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	params, err := parseCommandLineParams(in.GetParameter())
	if err != nil {
		gen.Fail("could not parse parameters passed to --twirp_out", err.Error())
	}
	t.importPrefix = params.importPrefix
	t.importMap = params.importMap

	t.genFiles = gen.FilesToGenerate(in)

	t.sourceRelativePaths = params.paths == "source_relative"

	// Collect information on types.
	t.reg = typemap.New(in.ProtoFile)

	// Register names of packages that we import.
	t.registerPackageName("bytes")
	t.registerPackageName("strings")
	t.registerPackageName("path")
	t.registerPackageName("ctxsetters")
	t.registerPackageName("context")
	t.registerPackageName("http")
	t.registerPackageName("io")
	t.registerPackageName("ioutil")
	t.registerPackageName("json")
	t.registerPackageName("jsonpb")
	t.registerPackageName("proto")
	t.registerPackageName("strconv")
	t.registerPackageName("twirp")
	t.registerPackageName("url")
	t.registerPackageName("fmt")

	// Time to figure out package names of objects defined in protobuf. First,
	// we'll figure out the name for the package we're generating.
	genPkgName, err := deduceGenPkgName(t.genFiles)
	if err != nil {
		gen.Fail(err.Error())
	}
	t.genPkgName = genPkgName

	// We also need to figure out the fully import path of the package we're
	// generating. It's possible to import proto definitions from different .proto
	// files which will be generated into the same Go package, which we need to
	// detect (and can only detect if files use fully-specified go_package
	// options).
	genPkgImportPath, _, _ := goPackageOption(t.genFiles[0])

	// Next, we need to pick names for all the files that are dependencies.
	for _, f := range in.ProtoFile {
		// Is this is a file we are generating? If yes, it gets the shared package name.
		if fileDescSliceContains(t.genFiles, f) {
			t.fileToGoPackageName[f] = t.genPkgName
			continue
		}

		// Is this is an imported .proto file which has the same fully-specified
		// go_package as the targeted file for generation? If yes, it gets the
		// shared package name too.
		if genPkgImportPath != "" {
			importPath, _, _ := goPackageOption(f)
			if importPath == genPkgImportPath {
				t.fileToGoPackageName[f] = t.genPkgName
				continue
			}
		}

		// This is a dependency from a different go_package. Use its package name.
		name := f.GetPackage()
		if name == "" {
			name = stringutils.BaseName(f.GetName())
		}
		name = stringutils.CleanIdentifier(name)
		alias := t.registerPackageName(name)
		t.fileToGoPackageName[f] = alias
	}

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range t.genFiles {
		respFile := t.generate(f)
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp
}

func (t *twirp) registerPackageName(name string) (alias string) {
	alias = name
	i := 1
	for t.pkgNamesInUse[alias] {
		alias = name + strconv.Itoa(i)
		i++
	}
	t.pkgNamesInUse[alias] = true
	t.pkgs[name] = alias
	return alias
}

// deduceGenPkgName figures out the go package name to use for generated code.
// Will try to use the explicit go_package setting in a file (if set, must be
// consistent in all files). If no files have go_package set, then use the
// protobuf package name (must be consistent in all files)
func deduceGenPkgName(genFiles []*descriptor.FileDescriptorProto) (string, error) {
	var genPkgName string
	for _, f := range genFiles {
		name, explicit := goPackageName(f)
		if explicit {
			name = stringutils.CleanIdentifier(name)
			if genPkgName != "" && genPkgName != name {
				// Make sure they're all set consistently.
				return "", errors.Errorf("files have conflicting go_package settings, must be the same: %q and %q", genPkgName, name)
			}
			genPkgName = name
		}
	}
	if genPkgName != "" {
		return genPkgName, nil
	}

	// If there is no explicit setting, then check the implicit package name
	// (derived from the protobuf package name) of the files and make sure it's
	// consistent.
	for _, f := range genFiles {
		name, _ := goPackageName(f)
		name = stringutils.CleanIdentifier(name)
		if genPkgName != "" && genPkgName != name {
			return "", errors.Errorf("files have conflicting package names, must be the same or overridden with go_package: %q and %q", genPkgName, name)
		}
		genPkgName = name
	}

	// All the files have the same name, so we're good.
	return genPkgName, nil
}

func (t *twirp) generate(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	resp := new(plugin.CodeGeneratorResponse_File)
	if len(file.Service) == 0 {
		return nil
	}

	t.generateFileHeader(file)

	t.generateImports(file)
	if t.filesHandled == 0 {
		t.generateUtilImports()
	}

	t.generateVersionCheck(file)

	// For each service, generate client stubs and server
	for i, service := range file.Service {
		t.generateService(file, service, i)
	}

	// Util functions only generated once per package
	if t.filesHandled == 0 {
		t.generateUtils()
	}

	t.generateFileDescriptor(file)

	resp.Name = proto.String(t.goFileName(file))
	resp.Content = proto.String(t.formattedOutput())
	t.output.Reset()

	t.filesHandled++
	return resp
}

func (t *twirp) generateVersionCheck(file *descriptor.FileDescriptorProto) {
	t.P(`// This is a compile-time assertion to ensure that this generated file`)
	t.P(`// is compatible with the twirp package used in your project.`)
	t.P(`// A compilation error at this line likely means your copy of the`)
	t.P(`// twirp package needs to be updated.`)
	t.P(`const _ = `, t.pkgs["twirp"], `.TwirpPackageIsVersion7`)
}

func (t *twirp) generateFileHeader(file *descriptor.FileDescriptorProto) {
	t.P("// Code generated by protoc-gen-twirp ", gen.Version, ", DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	if t.filesHandled == 0 {
		t.P("/*")
		t.P("Package ", t.genPkgName, " is a generated twirp stub package.")
		t.P("This code was generated with github.com/twitchtv/twirp/protoc-gen-twirp ", gen.Version, ".")
		t.P()
		comment, err := t.reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "* /", -1)
				t.P(line)
			}
			t.P()
		}
		t.P("It is generated from these files:")
		for _, f := range t.genFiles {
			t.P("\t", f.GetName())
		}
		t.P("*/")
	}
	t.P(`package `, t.genPkgName)
	t.P()
}

func (t *twirp) generateImports(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}

	// stdlib imports
	t.P(`import `, t.pkgs["bytes"], ` "bytes"`)
	t.P(`import `, t.pkgs["strings"], ` "strings"`)
	t.P(`import `, t.pkgs["context"], ` "context"`)
	t.P(`import `, t.pkgs["fmt"], ` "fmt"`)
	t.P(`import `, t.pkgs["ioutil"], ` "io/ioutil"`)
	t.P(`import `, t.pkgs["http"], ` "net/http"`)
	t.P(`import `, t.pkgs["strconv"], ` "strconv"`)
	t.P()
	// dependency imports
	t.P(`import `, t.pkgs["jsonpb"], ` "github.com/golang/protobuf/jsonpb"`)
	t.P(`import `, t.pkgs["proto"], ` "github.com/golang/protobuf/proto"`)
	t.P(`import `, t.pkgs["twirp"], ` "github.com/twitchtv/twirp"`)
	t.P(`import `, t.pkgs["ctxsetters"], ` "github.com/twitchtv/twirp/ctxsetters"`)
	t.P()

	// It's legal to import a message and use it as an input or output for a
	// method. Make sure to import the package of any such message. First, dedupe
	// them.
	deps := make(map[string]string) // Map of package name to quoted import path.
	ourImportPath := path.Dir(t.goFileName(file))
	for _, s := range file.Service {
		for _, m := range s.Method {
			defs := []*typemap.MessageDefinition{
				t.reg.MethodInputDefinition(m),
				t.reg.MethodOutputDefinition(m),
			}
			for _, def := range defs {
				// By default, import path is the dirname of the Go filename.
				importPath := path.Dir(t.goFileName(def.File))
				if importPath == ourImportPath {
					continue
				}

				importPathOpt, _ := parseGoPackageOption(def.File.GetOptions().GetGoPackage())
				if importPathOpt != "" {
					importPath = importPathOpt
				}

				if substitution, ok := t.importMap[def.File.GetName()]; ok {
					importPath = substitution
				}
				importPath = t.importPrefix + importPath
				pkg := t.goPackageName(def.File)
				if pkg != t.genPkgName {
					deps[pkg] = strconv.Quote(importPath)
				}
			}
		}
	}
	for pkg, importPath := range deps {
		t.P(`import `, pkg, ` `, importPath)
	}
	if len(deps) > 0 {
		t.P()
	}
}

func (t *twirp) generateUtilImports() {
	t.P("// Imports only used by utility functions:")
	t.P(`import `, t.pkgs["io"], ` "io"`)
	t.P(`import `, t.pkgs["json"], ` "encoding/json"`)
	t.P(`import `, t.pkgs["path"], ` "path"`)
	t.P(`import `, t.pkgs["url"], ` "net/url"`)
}

// Generate utility functions used in Twirp code.
// These should be generated just once per package.
func (t *twirp) generateUtils() {
	t.sectionComment(`Utils`)
	t.P(`// HTTPClient is the interface used by generated clients to send HTTP requests.`)
	t.P(`// It is fulfilled by *(net/http).Client, which is sufficient for most users.`)
	t.P(`// Users can provide their own implementation for special retry policies.`)
	t.P(`// `)
	t.P(`// HTTPClient implementations should not follow redirects. Redirects are`)
	t.P(`// automatically disabled if *(net/http).Client is passed to client`)
	t.P(`// constructors. See the withoutRedirects function in this file for more`)
	t.P(`// details.`)
	t.P(`type HTTPClient interface {`)
	t.P(`	Do(req *`, t.pkgs["http"], `.Request) (*`, t.pkgs["http"], `.Response, error)`)
	t.P(`}`)
	t.P()
	t.P(`// TwirpServer is the interface generated server structs will support: they're`)
	t.P(`// HTTP handlers with additional methods for accessing metadata about the`)
	t.P(`// service. Those accessors are a low-level API for building reflection tools.`)
	t.P(`// Most people can think of TwirpServers as just http.Handlers.`)
	t.P(`type TwirpServer interface {`)
	t.P(`  `, t.pkgs["http"], `.Handler`)
	t.P()
	t.P(`  // ServiceDescriptor returns gzipped bytes describing the .proto file that`)
	t.P(`  // this service was generated from. Once unzipped, the bytes can be`)
	t.P(`  // unmarshalled as a`)
	t.P(`  // github.com/golang/protobuf/protoc-gen-go/descriptor.FileDescriptorProto.`)
	t.P(`  //`)
	t.P(`  // The returned integer is the index of this particular service within that`)
	t.P(`  // FileDescriptorProto's 'Service' slice of ServiceDescriptorProtos. This is a`)
	t.P(`  // low-level field, expected to be used for reflection.`)
	t.P(`  ServiceDescriptor() ([]byte, int)`)
	t.P()
	t.P(`  // ProtocGenTwirpVersion is the semantic version string of the version of`)
	t.P(`  // twirp used to generate this file.`)
	t.P(`  ProtocGenTwirpVersion() string`)
	t.P()
	t.P(`  // PathPrefix returns the HTTP URL path prefix for all methods handled by this`)
	t.P(`  // service. This can be used with an HTTP mux to route Twirp requests.`)
	t.P(`  // The path prefix is in the form: "/<prefix>/<package>.<Service>/"`)
	t.P(`  // that is, everything in a Twirp route except for the <Method> at the end.`)
	t.P(`  PathPrefix() string`)
	t.P(`}`)
	t.P()

	t.P(`// WriteError writes an HTTP response with a valid Twirp error format (code, msg, meta).`)
	t.P(`// Useful outside of the Twirp server (e.g. http middleware), but does not trigger hooks.`)
	t.P(`// If err is not a twirp.Error, it will get wrapped with twirp.InternalErrorWith(err)`)
	t.P(`func WriteError(resp `, t.pkgs["http"], `.ResponseWriter, err error) {`)
	t.P(`  writeError(`, t.pkgs["context"], `.Background(), resp, err, nil)`)
	t.P(`}`)
	t.P()

	t.P(`// writeError writes Twirp errors in the response and triggers hooks.`)
	t.P(`func writeError(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, err error, hooks *`, t.pkgs["twirp"], `.ServerHooks) {`)
	t.P(`  // Non-twirp errors are wrapped as Internal (default)`)
	t.P(`  twerr, ok := err.(`, t.pkgs["twirp"], `.Error)`)
	t.P(`  if !ok {`)
	t.P(`    twerr = `, t.pkgs["twirp"], `.InternalErrorWith(err)`)
	t.P(`  }`)
	t.P(``)
	t.P(`  statusCode := `, t.pkgs["twirp"], `.ServerHTTPStatusFromErrorCode(twerr.Code())`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, statusCode)`)
	t.P(`  ctx = callError(ctx, hooks, twerr)`)
	t.P(``)
	t.P(`  respBody := marshalErrorToJSON(twerr)`)
	t.P()
	t.P(`  resp.Header().Set("Content-Type", "application/json") // Error responses are always JSON`)
	t.P(`  resp.Header().Set("Content-Length", strconv.Itoa(len(respBody)))`)
	t.P(`  resp.WriteHeader(statusCode) // set HTTP status code and send response`)
	t.P(``)
	t.P(`  _, writeErr := resp.Write(respBody)`)
	t.P(`  if writeErr != nil {`)
	t.P(`    // We have three options here. We could log the error, call the Error`)
	t.P(`    // hook, or just silently ignore the error.`)
	t.P(`    //`)
	t.P(`    // Logging is unacceptable because we don't have a user-controlled `)
	t.P(`    // logger; writing out to stderr without permission is too rude.`)
	t.P(`    //`)
	t.P(`    // Calling the Error hook would confuse users: it would mean the Error`)
	t.P(`    // hook got called twice for one request, which is likely to lead to`)
	t.P(`    // duplicated log messages and metrics, no matter how well we document`)
	t.P(`    // the behavior.`)
	t.P(`    //`)
	t.P(`    // Silently ignoring the error is our least-bad option. It's highly`)
	t.P(`    // likely that the connection is broken and the original 'err' says`)
	t.P(`    // so anyway.`)
	t.P(`    _ = writeErr`)
	t.P(`  }`)
	t.P(``)
	t.P(`  callResponseSent(ctx, hooks)`)
	t.P(`}`)
	t.P()

	t.P(`// sanitizeBaseURL parses the the baseURL, and adds the "http" scheme if needed.`)
	t.P(`// If the URL is unparsable, the baseURL is returned unchaged.`)
	t.P(`func sanitizeBaseURL(baseURL string) string {`)
	t.P(`  u, err := `, t.pkgs["url"], `.Parse(baseURL)`)
	t.P(`  if err != nil {`)
	t.P(`    return baseURL // invalid URL will fail later when making requests`)
	t.P(`  }`)
	t.P(`  if u.Scheme == "" {`)
	t.P(`    u.Scheme = "http"`)
	t.P(`  }`)
	t.P(`  return u.String()`)
	t.P(`}`)
	t.P()

	t.P(`// baseServicePath composes the path prefix for the service (without <Method>).`)
	t.P(`// e.g.: baseServicePath("/twirp", "my.pkg", "MyService")`)
	t.P(`//       returns => "/twirp/my.pkg.MyService/"`)
	t.P(`// e.g.: baseServicePath("", "", "MyService")`)
	t.P(`//       returns => "/MyService/"`)
	t.P(`func baseServicePath(prefix, pkg, service string) string {`)
	t.P(`  fullServiceName := service`)
	t.P(`  if pkg != "" {`)
	t.P(`    fullServiceName = pkg + "." + service`)
	t.P(`  }`)
	t.P(`  return path.Join("/", prefix, fullServiceName) + "/"`)
	t.P(`}`)
	t.P()

	t.P(`// parseTwirpPath extracts path components form a valid Twirp route.`)
	t.P(`// Expected format: "[<prefix>]/<package>.<Service>/<Method>"`)
	t.P(`// e.g.: prefix, pkgService, method := parseTwirpPath("/twirp/pkg.Svc/MakeHat")`)
	t.P(`func parseTwirpPath(path string) (string, string, string) {`)
	t.P(`  parts := `, t.pkgs["strings"], `.Split(path, "/")`)
	t.P(`  if len(parts) < 2 {`)
	t.P(`    return "", "", ""`)
	t.P(`  }`)
	t.P(`  method := parts[len(parts)-1]`)
	t.P(`  pkgService := parts[len(parts)-2]`)
	t.P(`  prefix := `, t.pkgs["strings"], `.Join(parts[0:len(parts)-2], "/")`)
	t.P(`  return prefix, pkgService, method`)
	t.P(`}`)
	t.P()

	t.P(`// getCustomHTTPReqHeaders retrieves a copy of any headers that are set in`)
	t.P(`// a context through the twirp.WithHTTPRequestHeaders function.`)
	t.P(`// If there are no headers set, or if they have the wrong type, nil is returned.`)
	t.P(`func getCustomHTTPReqHeaders(ctx `, t.pkgs["context"], `.Context) `, t.pkgs["http"], `.Header {`)
	t.P(`  header, ok := `, t.pkgs["twirp"], `.HTTPRequestHeaders(ctx)`)
	t.P(`  if !ok || header == nil {`)
	t.P(`    return nil`)
	t.P(`  }`)
	t.P(`  copied := make(`, t.pkgs["http"], `.Header)`)
	t.P(`  for k, vv := range header {`)
	t.P(`    if vv == nil {`)
	t.P(`      copied[k] = nil`)
	t.P(`      continue`)
	t.P(`    }`)
	t.P(`    copied[k] = make([]string, len(vv))`)
	t.P(`    copy(copied[k], vv)`)
	t.P(`  }`)
	t.P(`  return copied`)
	t.P(`}`)
	t.P()

	t.P(`// newRequest makes an http.Request from a client, adding common headers.`)
	t.P(`func newRequest(ctx `, t.pkgs["context"], `.Context, url string, reqBody io.Reader, contentType string) (*`, t.pkgs["http"], `.Request, error) {`)
	t.P(`  req, err := `, t.pkgs["http"], `.NewRequest("POST", url, reqBody)`)
	t.P(`  if err != nil {`)
	t.P(`    return nil, err`)
	t.P(`  }`)
	t.P(`  req = req.WithContext(ctx)`)
	t.P(`  if customHeader := getCustomHTTPReqHeaders(ctx); customHeader != nil {`)
	t.P(`    req.Header = customHeader`)
	t.P(`  }`)
	t.P(`  req.Header.Set("Accept", contentType)`)
	t.P(`  req.Header.Set("Content-Type", contentType)`)
	t.P(`  req.Header.Set("Twirp-Version", "`, gen.Version, `")`)
	t.P(`  return req, nil`)
	t.P(`}`)
	t.P()

	t.P(`// JSON serialization for errors`)
	t.P(`type twerrJSON struct {`)
	t.P("  Code string            `json:\"code\"`")
	t.P("  Msg  string            `json:\"msg\"`")
	t.P("  Meta map[string]string `json:\"meta,omitempty\"`")
	t.P(`}`)
	t.P()
	t.P(`// marshalErrorToJSON returns JSON from a twirp.Error, that can be used as HTTP error response body.`)
	t.P(`// If serialization fails, it will use a descriptive Internal error instead.`)
	t.P(`func marshalErrorToJSON(twerr `, t.pkgs["twirp"], `.Error) []byte {`)
	t.P(`  // make sure that msg is not too large`)
	t.P(`  msg := twerr.Msg()`)
	t.P(`  if len(msg) > 1e6 {`)
	t.P(`    msg = msg[:1e6]`)
	t.P(`  }`)
	t.P(``)
	t.P(`  tj := twerrJSON{`)
	t.P(`    Code: string(twerr.Code()),`)
	t.P(`    Msg:  msg,`)
	t.P(`    Meta: twerr.MetaMap(),`)
	t.P(`  }`)
	t.P(``)
	t.P(`  buf, err := `, t.pkgs["json"], `.Marshal(&tj)`)
	t.P(`  if err != nil {`)
	t.P(`    buf = []byte("{\"type\": \"" + `, t.pkgs["twirp"], `.Internal +"\", \"msg\": \"There was an error but it could not be serialized into JSON\"}") // fallback`)
	t.P(`  }`)
	t.P(``)
	t.P(`  return buf`)
	t.P(`}`)
	t.P()

	t.P(`// errorFromResponse builds a twirp.Error from a non-200 HTTP response.`)
	t.P(`// If the response has a valid serialized Twirp error, then it's returned.`)
	t.P(`// If not, the response status code is used to generate a similar twirp`)
	t.P(`// error. See twirpErrorFromIntermediary for more info on intermediary errors.`)
	t.P(`func errorFromResponse(resp *`, t.pkgs["http"], `.Response) `, t.pkgs["twirp"], `.Error {`)
	t.P(`  statusCode := resp.StatusCode`)
	t.P(`  statusText := `, t.pkgs["http"], `.StatusText(statusCode)`)
	t.P(``)
	t.P(`  if isHTTPRedirect(statusCode) {`)
	t.P(`    // Unexpected redirect: it must be an error from an intermediary.`)
	t.P(`    // Twirp clients don't follow redirects automatically, Twirp only handles`)
	t.P(`    // POST requests, redirects should only happen on GET and HEAD requests.`)
	t.P(`    location := resp.Header.Get("Location")`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("unexpected HTTP status code %d %q received, Location=%q", statusCode, statusText, location)`)
	t.P(`    return twirpErrorFromIntermediary(statusCode, msg, location)`)
	t.P(`  }`)
	t.P(``)
	t.P(`  respBodyBytes, err := `, t.pkgs["ioutil"], `.ReadAll(resp.Body)`)
	t.P(`  if err != nil {`)
	t.P(`    return wrapInternal(err, "failed to read server error response body")`)
	t.P(`  }`)
	t.P(``)
	t.P(`  var tj twerrJSON`)
	t.P(`  dec := `, t.pkgs["json"], `.NewDecoder(`, t.pkgs["bytes"], `.NewReader(respBodyBytes))`)
	t.P(`  dec.DisallowUnknownFields()`)
	t.P(`  if err := dec.Decode(&tj); err != nil || tj.Code == "" {`)
	t.P(`    // Invalid JSON response; it must be an error from an intermediary.`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("Error from intermediary with HTTP status code %d %q", statusCode, statusText)`)
	t.P(`    return twirpErrorFromIntermediary(statusCode, msg, string(respBodyBytes))`)
	t.P(`  }`)
	t.P(``)
	t.P(`  errorCode := `, t.pkgs["twirp"], `.ErrorCode(tj.Code)`)
	t.P(`  if !`, t.pkgs["twirp"], `.IsValidErrorCode(errorCode) {`)
	t.P(`    msg := "invalid type returned from server error response: "+tj.Code`)
	t.P(`    return `, t.pkgs["twirp"], `.InternalError(msg).WithMeta("body", string(respBodyBytes))`)
	t.P(`  }`)
	t.P(``)
	t.P(`  twerr := `, t.pkgs["twirp"], `.NewError(errorCode, tj.Msg)`)
	t.P(`  for k, v := range(tj.Meta) {`)
	t.P(`    twerr = twerr.WithMeta(k, v)`)
	t.P(`  }`)
	t.P(`  return twerr`)
	t.P(`}`)
	t.P()

	t.P(`// twirpErrorFromIntermediary maps HTTP errors from non-twirp sources to twirp errors.`)
	t.P(`// The mapping is similar to gRPC: https://github.com/grpc/grpc/blob/master/doc/http-grpc-status-mapping.md.`)
	t.P(`// Returned twirp Errors have some additional metadata for inspection.`)
	t.P(`func twirpErrorFromIntermediary(status int, msg string, bodyOrLocation string) `, t.pkgs["twirp"], `.Error {`)
	t.P(`  var code `, t.pkgs["twirp"], `.ErrorCode`)
	t.P(`  if isHTTPRedirect(status) { // 3xx`)
	t.P(`    code = `, t.pkgs["twirp"], `.Internal`)
	t.P(`  } else {`)
	t.P(`    switch status {`)
	t.P(`    case 400: // Bad Request`)
	t.P(`      code = `, t.pkgs["twirp"], `.Internal`)
	t.P(`    case 401: // Unauthorized`)
	t.P(`      code = `, t.pkgs["twirp"], `.Unauthenticated`)
	t.P(`    case 403: // Forbidden`)
	t.P(`      code = `, t.pkgs["twirp"], `.PermissionDenied`)
	t.P(`    case 404: // Not Found`)
	t.P(`      code = `, t.pkgs["twirp"], `.BadRoute`)
	t.P(`    case 429: // Too Many Requests`)
	t.P(`      code = `, t.pkgs["twirp"], `.ResourceExhausted`)
	t.P(`    case 502, 503, 504: // Bad Gateway, Service Unavailable, Gateway Timeout`)
	t.P(`      code = `, t.pkgs["twirp"], `.Unavailable`)
	t.P(`    default: // All other codes`)
	t.P(`      code = `, t.pkgs["twirp"], `.Unknown`)
	t.P(`    }`)
	t.P(`  }`)
	t.P(``)
	t.P(`  twerr := `, t.pkgs["twirp"], `.NewError(code, msg)`)
	t.P(`  twerr = twerr.WithMeta("http_error_from_intermediary", "true") // to easily know if this error was from intermediary`)
	t.P(`  twerr = twerr.WithMeta("status_code", `, t.pkgs["strconv"], `.Itoa(status))`)
	t.P(`  if isHTTPRedirect(status) {`)
	t.P(`    twerr = twerr.WithMeta("location", bodyOrLocation)`)
	t.P(`  } else {`)
	t.P(`    twerr = twerr.WithMeta("body", bodyOrLocation)`)
	t.P(`  }`)
	t.P(`  return twerr`)
	t.P(`}`)
	t.P()

	t.P(`func isHTTPRedirect(status int) bool {`)
	t.P(`	return status >= 300 && status <= 399`)
	t.P(`}`)
	t.P()

	t.P(`// wrapInternal wraps an error with a prefix as an Internal error.`)
	t.P(`// The original error cause is accessible by github.com/pkg/errors.Cause.`)
	t.P(`func wrapInternal(err error, prefix string) `, t.pkgs["twirp"], `.Error {`)
	t.P(`	return `, t.pkgs["twirp"], `.InternalErrorWith(&wrappedError{prefix: prefix, cause: err})`)
	t.P(`}`)
	t.P(`type wrappedError struct {`)
	t.P(`	prefix string`)
	t.P(`	cause  error`)
	t.P(`}`)
	t.P(`func (e *wrappedError) Error() string { return e.prefix + ": " + e.cause.Error() }`)
	t.P(`func (e *wrappedError) Unwrap() error  { return e.cause } // for go1.13 + errors.Is/As `)
	t.P(`func (e *wrappedError) Cause() error  { return e.cause } // for github.com/pkg/errors`)
	t.P()

	t.P(`// ensurePanicResponses makes sure that rpc methods causing a panic still result in a Twirp Internal`)
	t.P(`// error response (status 500), and error hooks are properly called with the panic wrapped as an error.`)
	t.P(`// The panic is re-raised so it can be handled normally with middleware.`)
	t.P(`func ensurePanicResponses(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, hooks *`, t.pkgs["twirp"], `.ServerHooks) {`)
	t.P(`	if r := recover(); r != nil {`)
	t.P(`		// Wrap the panic as an error so it can be passed to error hooks.`)
	t.P(`		// The original error is accessible from error hooks, but not visible in the response.`)
	t.P(`		err := errFromPanic(r)`)
	t.P(`		twerr := &internalWithCause{msg: "Internal service panic", cause: err}`)
	t.P(`		// Actually write the error`)
	t.P(`		writeError(ctx, resp, twerr, hooks)`)
	t.P(`		// If possible, flush the error to the wire.`)
	t.P(`		f, ok := resp.(`, t.pkgs["http"], `.Flusher)`)
	t.P(`		if ok {`)
	t.P(`			f.Flush()`)
	t.P(`		}`)
	t.P(``)
	t.P(`		panic(r)`)
	t.P(`	}`)
	t.P(`}`)
	t.P(``)
	t.P(`// errFromPanic returns the typed error if the recovered panic is an error, otherwise formats as error.`)
	t.P(`func errFromPanic(p interface{}) error {`)
	t.P(`	if err, ok := p.(error); ok {`)
	t.P(`	  return err`)
	t.P(`	}`)
	t.P(`	return fmt.Errorf("panic: %v", p)`)
	t.P(`}`)
	t.P(``)
	t.P(`// internalWithCause is a Twirp Internal error wrapping an original error cause,`)
	t.P(`// but the original error message is not exposed on Msg(). The original error`)
	t.P(`// can be checked with go1.13+ errors.Is/As, and also by (github.com/pkg/errors).Unwrap`)
	t.P(`type internalWithCause struct {`)
	t.P(`	msg    string`)
	t.P(`	cause  error`)
	t.P(`}`)
	t.P(`func (e *internalWithCause) Unwrap() error  { return e.cause } // for go1.13 + errors.Is/As`)
	t.P(`func (e *internalWithCause) Cause() error  { return e.cause } // for github.com/pkg/errors`)
	t.P(`func (e *internalWithCause) Error() string { return e.msg + ": " + e.cause.Error()}`)
	t.P(`func (e *internalWithCause) Code() `, t.pkgs["twirp"], `.ErrorCode { return `, t.pkgs["twirp"], `.Internal }`)
	t.P(`func (e *internalWithCause) Msg() string                { return e.msg }`)
	t.P(`func (e *internalWithCause) Meta(key string) string     { return "" }`)
	t.P(`func (e *internalWithCause) MetaMap() map[string]string { return nil }`)
	t.P(`func (e *internalWithCause) WithMeta(key string, val string) `, t.pkgs["twirp"], `.Error { return e }`)
	t.P()

	t.P(`// malformedRequestError is used when the twirp server cannot unmarshal a request`)
	t.P(`func malformedRequestError(msg string) `, t.pkgs["twirp"], `.Error {`)
	t.P(`	return `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.Malformed, msg)`)
	t.P(`}`)
	t.P()

	t.P(`// badRouteError is used when the twirp server cannot route a request`)
	t.P(`func badRouteError(msg string, method, url string) `, t.pkgs["twirp"], `.Error {`)
	t.P(`	err := `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.BadRoute, msg)`)
	t.P(`	err = err.WithMeta("twirp_invalid_route", method+" "+url)`)
	t.P(`	return err`)
	t.P(`}`)
	t.P()

	t.P(`// withoutRedirects makes sure that the POST request can not be redirected.`)
	t.P(`// The standard library will, by default, redirect requests (including POSTs) if it gets a 302 or`)
	t.P(`// 303 response, and also 301s in go1.8. It redirects by making a second request, changing the`)
	t.P(`// method to GET and removing the body. This produces very confusing error messages, so instead we`)
	t.P(`// set a redirect policy that always errors. This stops Go from executing the redirect.`)
	t.P(`//`)
	t.P(`// We have to be a little careful in case the user-provided http.Client has its own CheckRedirect`)
	t.P(`// policy - if so, we'll run through that policy first.`)
	t.P(`//`)
	t.P(`// Because this requires modifying the http.Client, we make a new copy of the client and return it.`)
	t.P(`func withoutRedirects(in *`, t.pkgs["http"], `.Client) *`, t.pkgs["http"], `.Client {`)
	t.P(`	copy := *in`)
	t.P(`	copy.CheckRedirect = func(req *`, t.pkgs["http"], `.Request, via []*`, t.pkgs["http"], `.Request) error {`)
	t.P(`		if in.CheckRedirect != nil {`)
	t.P(`			// Run the input's redirect if it exists, in case it has side effects, but ignore any error it`)
	t.P(`			// returns, since we want to use ErrUseLastResponse.`)
	t.P(`			err := in.CheckRedirect(req, via)`)
	t.P(`			_ = err // Silly, but this makes sure generated code passes errcheck -blank, which some people use.`)
	t.P(`		}`)
	t.P(`		return `, t.pkgs["http"], `.ErrUseLastResponse`)
	t.P(`	}`)
	t.P(`	return &copy`)
	t.P(`}`)
	t.P()

	t.P(`// doProtobufRequest makes a Protobuf request to the remote Twirp service.`)
	t.P(`func doProtobufRequest(ctx `, t.pkgs["context"], `.Context, client HTTPClient, hooks *`, t.pkgs["twirp"], `.ClientHooks, url string, in, out `, t.pkgs["proto"], `.Message) (_ `, t.pkgs["context"], `.Context, err error) {`)
	t.P(`  reqBodyBytes, err := `, t.pkgs["proto"], `.Marshal(in)`)
	t.P(`  if err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to marshal proto request")`)
	t.P(`  }`)
	t.P(`  reqBody := `, t.pkgs["bytes"], `.NewBuffer(reqBodyBytes)`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "aborted because context was done")`)
	t.P(`  }`)
	t.P()
	t.P(`  req, err := newRequest(ctx, url, reqBody, "application/protobuf")`)
	t.P(`  if err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "could not build request")`)
	t.P(`  }`)
	t.P(`  ctx, err = callClientRequestPrepared(ctx, hooks, req)`)
	t.P(`	 if err != nil {`)
	t.P(`    return ctx, err`)
	t.P(`  }`)
	t.P()
	t.P(`  req = req.WithContext(ctx)`)
	t.P(`  resp, err := client.Do(req)`)
	t.P(`  if err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to do request")`)
	t.P(`  }`)
	t.P()
	t.P(`  defer func() {`)
	t.P(`    cerr := resp.Body.Close()`)
	t.P(`    if err == nil && cerr != nil {`)
	t.P(`      err = wrapInternal(cerr, "failed to close response body")`)
	t.P(`    }`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "aborted because context was done")`)
	t.P(`  }`)
	t.P()
	t.P(`  if resp.StatusCode != 200 {`)
	t.P(`    return ctx, errorFromResponse(resp)`)
	t.P(`  }`)
	t.P()
	t.P(`  respBodyBytes, err := `, t.pkgs["ioutil"], `.ReadAll(resp.Body)`)
	t.P(`  if err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to read response body")`)
	t.P(`  }`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "aborted because context was done")`)
	t.P(`  }`)
	t.P()
	t.P(`  if err = `, t.pkgs["proto"], `.Unmarshal(respBodyBytes, out); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to unmarshal proto response")`)
	t.P(`  }`)
	t.P(`  return ctx, nil`)
	t.P(`}`)
	t.P()

	t.P(`// doJSONRequest makes a JSON request to the remote Twirp service.`)
	t.P(`func doJSONRequest(ctx `, t.pkgs["context"], `.Context, client HTTPClient, hooks *`, t.pkgs["twirp"], `.ClientHooks, url string, in, out `, t.pkgs["proto"], `.Message) (_ `, t.pkgs["context"], `.Context, err error) {`)
	t.P(`  reqBody := `, t.pkgs["bytes"], `.NewBuffer(nil)`)
	t.P(`  marshaler := &`, t.pkgs["jsonpb"], `.Marshaler{OrigName: true}`)
	t.P(`  if err = marshaler.Marshal(reqBody, in); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to marshal json request")`)
	t.P(`  }`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "aborted because context was done")`)
	t.P(`  }`)
	t.P()
	t.P(`  req, err := newRequest(ctx, url, reqBody, "application/json")`)
	t.P(`  if err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "could not build request")`)
	t.P(`  }`)
	t.P(`  ctx, err = callClientRequestPrepared(ctx, hooks, req)`)
	t.P(`	 if err != nil {`)
	t.P(`    return ctx, err`)
	t.P(`  }`)
	t.P()
	t.P(`  req = req.WithContext(ctx)`)
	t.P(`  resp, err := client.Do(req)`)
	t.P(`  if err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to do request")`)
	t.P(`  }`)
	t.P()
	t.P(`  defer func() {`)
	t.P(`    cerr := resp.Body.Close()`)
	t.P(`    if err == nil && cerr != nil {`)
	t.P(`      err = wrapInternal(cerr, "failed to close response body")`)
	t.P(`    }`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "aborted because context was done")`)
	t.P(`  }`)
	t.P()
	t.P(`  if resp.StatusCode != 200 {`)
	t.P(`    return ctx, errorFromResponse(resp)`)
	t.P(`  }`)
	t.P()
	t.P(`  unmarshaler := `, t.pkgs["jsonpb"], `.Unmarshaler{AllowUnknownFields: true}`)
	t.P(`  if err = unmarshaler.Unmarshal(resp.Body, out); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "failed to unmarshal json response")`)
	t.P(`  }`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return ctx, wrapInternal(err, "aborted because context was done")`)
	t.P(`  }`)
	t.P(`  return ctx, nil`)
	t.P(`}`)
	t.P()

	t.P(`// Call twirp.ServerHooks.RequestReceived if the hook is available`)
	t.P(`func callRequestReceived(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ServerHooks) (`, t.pkgs["context"], `.Context, error) {`)
	t.P(`  if h == nil || h.RequestReceived == nil {`)
	t.P(`    return ctx, nil`)
	t.P(`  }`)
	t.P(`  return h.RequestReceived(ctx)`)
	t.P(`}`)
	t.P()
	t.P(`// Call twirp.ServerHooks.RequestRouted if the hook is available`)
	t.P(`func callRequestRouted(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ServerHooks) (`, t.pkgs["context"], `.Context, error) {`)
	t.P(`  if h == nil || h.RequestRouted == nil {`)
	t.P(`    return ctx, nil`)
	t.P(`  }`)
	t.P(`  return h.RequestRouted(ctx)`)
	t.P(`}`)
	t.P()
	t.P(`// Call twirp.ServerHooks.ResponsePrepared if the hook is available`)
	t.P(`func callResponsePrepared(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ServerHooks) `, t.pkgs["context"], `.Context {`)
	t.P(`  if h == nil || h.ResponsePrepared == nil {`)
	t.P(`    return ctx`)
	t.P(`  }`)
	t.P(`  return h.ResponsePrepared(ctx)`)
	t.P(`}`)
	t.P()
	t.P(`// Call twirp.ServerHooks.ResponseSent if the hook is available`)
	t.P(`func callResponseSent(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ServerHooks) {`)
	t.P(`  if h == nil || h.ResponseSent == nil {`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  h.ResponseSent(ctx)`)
	t.P(`}`)
	t.P()
	t.P(`// Call twirp.ServerHooks.Error if the hook is available`)
	t.P(`func callError(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ServerHooks, err `, t.pkgs["twirp"], `.Error) `, t.pkgs["context"], `.Context {`)
	t.P(`  if h == nil || h.Error == nil {`)
	t.P(`    return ctx`)
	t.P(`  }`)
	t.P(`  return h.Error(ctx, err)`)
	t.P(`}`)
	t.P()

	t.generateClientHooks()
}

// P forwards to g.gen.P, which prints output.
func (t *twirp) P(args ...string) {
	for _, v := range args {
		t.output.WriteString(v)
	}
	t.output.WriteByte('\n')
}

// Big header comments to makes it easier to visually parse a generated file.
func (t *twirp) sectionComment(sectionTitle string) {
	t.P()
	t.P(`// `, strings.Repeat("=", len(sectionTitle)))
	t.P(`// `, sectionTitle)
	t.P(`// `, strings.Repeat("=", len(sectionTitle)))
	t.P()
}

func (t *twirp) generateService(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, index int) {
	servName := serviceNameCamelCased(service)

	t.sectionComment(servName + ` Interface`)
	t.generateTwirpInterface(file, service)

	t.sectionComment(servName + ` Protobuf Client`)
	t.generateClient("Protobuf", file, service)

	t.sectionComment(servName + ` JSON Client`)
	t.generateClient("JSON", file, service)

	// Server
	t.sectionComment(servName + ` Server Handler`)
	t.generateServer(file, service)
}

func (t *twirp) generateTwirpInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servName := serviceNameCamelCased(service)

	comments, err := t.reg.ServiceComments(file, service)
	if err == nil {
		t.printComments(comments)
	}
	t.P(`type `, servName, ` interface {`)
	for _, method := range service.Method {
		comments, err = t.reg.MethodComments(file, service, method)
		if err == nil {
			t.printComments(comments)
		}
		t.P(t.generateSignature(method))
		t.P()
	}
	t.P(`}`)
}

func (t *twirp) generateSignature(method *descriptor.MethodDescriptorProto) string {
	methName := methodNameCamelCased(method)
	inputType := t.goTypeName(method.GetInputType())
	outputType := t.goTypeName(method.GetOutputType())
	return fmt.Sprintf(`	%s(%s.Context, *%s) (*%s, error)`, methName, t.pkgs["context"], inputType, outputType)
}

// valid names: 'JSON', 'Protobuf'
func (t *twirp) generateClient(name string, file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servPkg := pkgName(file)
	servName := serviceNameCamelCased(service)
	structName := unexported(servName) + name + "Client"
	newClientFunc := "New" + servName + name + "Client"
	servNameLit := serviceNameLiteral(service)
	servNameCc := servName

	methCnt := strconv.Itoa(len(service.Method))
	t.P(`type `, structName, ` struct {`)
	t.P(`  client HTTPClient`)
	t.P(`  urls  [`, methCnt, `]string`)
	t.P(`  interceptor `, t.pkgs["twirp"], `.Interceptor`)
	t.P(`  opts `, t.pkgs["twirp"], `.ClientOptions`)
	t.P(`}`)
	t.P()

	t.P(`// `, newClientFunc, ` creates a `, name, ` client that implements the `, servName, ` interface.`)
	t.P(`// It communicates using `, name, ` and can be configured with a custom HTTPClient.`)
	t.P(`func `, newClientFunc, `(baseURL string, client HTTPClient, opts ...`, t.pkgs["twirp"], `.ClientOption) `, servName, ` {`)
	t.P(`  if c, ok := client.(*`, t.pkgs["http"], `.Client); ok {`)
	t.P(`    client = withoutRedirects(c)`)
	t.P(`  }`)
	t.P()
	t.P(`  clientOpts := `, t.pkgs["twirp"], `.ClientOptions{}`)
	t.P(`  for _, o := range opts {`)
	t.P(`    o(&clientOpts)`)
	t.P(`  }`)
	t.P()
	if len(service.Method) > 0 {
		t.P(`  // Build method URLs: <baseURL>[<prefix>]/<package>.<Service>/<Method>`)
		t.P(`  serviceURL := sanitizeBaseURL(baseURL)`)
		if servNameLit == servNameCc {
			t.P(`  serviceURL += baseServicePath(clientOpts.PathPrefix(), "`, servPkg, `", "`, servNameCc, `")`)
		} else { // proto service name is not CamelCased, then it needs to check client option to decide if needs to change case
			t.P(`  if clientOpts.LiteralURLs {`)
			t.P(`    serviceURL += baseServicePath(clientOpts.PathPrefix(), "`, servPkg, `", "`, servNameLit, `")`)
			t.P(`  } else {`)
			t.P(`    serviceURL += baseServicePath(clientOpts.PathPrefix(), "`, servPkg, `", "`, servNameCc, `")`)
			t.P(`  }`)
		}
	}
	t.P(`  urls := [`, methCnt, `]string{`)
	for _, method := range service.Method {
		t.P(`    serviceURL + "`, methodNameCamelCased(method), `",`)
	}
	t.P(`  }`)

	allMethodsCamelCased := true
	for _, method := range service.Method {
		methNameLit := methodNameLiteral(method)
		methNameCc := methodNameCamelCased(method)
		if methNameCc != methNameLit {
			allMethodsCamelCased = false
			break
		}
	}
	if !allMethodsCamelCased {
		t.P(`  if clientOpts.LiteralURLs {`)
		t.P(`    urls = [`, methCnt, `]string{`)
		for _, method := range service.Method {
			t.P(`    serviceURL + "`, methodNameLiteral(method), `",`)
		}
		t.P(`    }`)
		t.P(`  }`)
	}

	t.P()
	t.P(`  return &`, structName, `{`)
	t.P(`    client: client,`)
	t.P(`    urls:   urls,`)
	t.P(`    interceptor: `, t.pkgs["twirp"], `.ChainInterceptors(clientOpts.Interceptors...),`)
	t.P(`    opts: clientOpts,`)
	t.P(`  }`)
	t.P(`}`)
	t.P()

	for i, method := range service.Method {
		methName := methodNameCamelCased(method)
		pkgName := pkgName(file)
		inputType := t.goTypeName(method.GetInputType())
		outputType := t.goTypeName(method.GetOutputType())
		t.P(`func (c *`, structName, `) `, methName, `(ctx `, t.pkgs["context"], `.Context, in *`, inputType, `) (*`, outputType, `, error) {`)
		t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithPackageName(ctx, "`, pkgName, `")`)
		t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithServiceName(ctx, "`, servName, `")`)
		t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
		t.P(`  caller := c.call`, methName)
		t.P(`  if c.interceptor != nil {`)
		t.generateClientInterceptorCaller(method)
		t.P(`  }`)
		t.P(`  return caller(ctx, in)`)
		t.P(`}`)
		t.P()
		t.P(`func (c *`, structName, `) call`, methName, `(ctx `, t.pkgs["context"], `.Context, in *`, inputType, `) (*`, outputType, `, error) {`)
		t.P(`  out := new(`, outputType, `)`)
		t.P(`  ctx, err := do`, name, `Request(ctx, c.client, c.opts.Hooks, c.urls[`, strconv.Itoa(i), `], in, out)`)
		t.P(`  if err != nil {`)
		t.P(`    twerr, ok := err.(`, t.pkgs["twirp"], `.Error)`)
		t.P(`    if !ok {`)
		t.P(`      twerr = `, t.pkgs["twirp"], `.InternalErrorWith(err)`)
		t.P(`    }`)
		t.P(`    callClientError(ctx, c.opts.Hooks, twerr)`)
		t.P(`    return nil, err`)
		t.P(`  }`)
		t.P()
		t.P(`  callClientResponseReceived(ctx, c.opts.Hooks)`)
		t.P()
		t.P(`  return out, nil`)
		t.P(`}`)
		t.P()
	}
}

func (t *twirp) generateClientHooks() {
	t.P(`func callClientResponseReceived(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ClientHooks) {`)
	t.P(`  if h == nil || h.ResponseReceived == nil {`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  h.ResponseReceived(ctx)`)
	t.P(`}`)
	t.P()
	t.P(`func callClientRequestPrepared(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ClientHooks, req *`, t.pkgs["http"], `.Request) (`, t.pkgs["context"], `.Context, error) {`)
	t.P(`  if h == nil || h.RequestPrepared == nil {`)
	t.P(`    return ctx, nil`)
	t.P(`  }`)
	t.P(`  return h.RequestPrepared(ctx, req)`)
	t.P(`}`)
	t.P()
	t.P(`func callClientError(ctx `, t.pkgs["context"], `.Context, h *`, t.pkgs["twirp"], `.ClientHooks, err `, t.pkgs["twirp"], `.Error) {`)
	t.P(`  if h == nil || h.Error == nil {`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  h.Error(ctx, err)`)
	t.P(`}`)
}

func (t *twirp) generateServer(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servName := serviceNameCamelCased(service)

	// Server implementation.
	servStruct := serviceStruct(service)
	t.P(`type `, servStruct, ` struct {`)
	t.P(`  `, servName)
	t.P(`  interceptor `, t.pkgs["twirp"], `.Interceptor`)
	t.P(`  hooks     *`, t.pkgs["twirp"], `.ServerHooks`)
	t.P(`  pathPrefix string // prefix for routing`)
	t.P(`  jsonSkipDefaults bool // do not include unpopulated fields (default values) in the response`)
	t.P(`}`)
	t.P()

	// Constructor for server implementation
	t.P(`// New`, servName, `Server builds a TwirpServer that can be used as an http.Handler to handle`)
	t.P(`// HTTP requests that are routed to the right method in the provided svc implementation.`)
	t.P(`// The opts are twirp.ServerOption modifiers, for example twirp.WithServerHooks(hooks).`)
	t.P(`func New`, servName, `Server(svc `, servName, `, opts ...interface{}) TwirpServer {`)
	t.P(`  serverOpts := `, t.pkgs["twirp"], `.ServerOptions{}`)
	t.P(`  for _, opt := range opts {`)
	t.P(`    switch o := opt.(type) {`)
	t.P(`    case `, t.pkgs["twirp"], `.ServerOption:`)
	t.P(`      o(&serverOpts)`)
	t.P(`    case *`, t.pkgs["twirp"], `.ServerHooks: // backwards compatibility, allow to specify hooks as an argument`)
	t.P(`      twirp.WithServerHooks(o)(&serverOpts)`)
	t.P(`    case nil: // backwards compatibility, allow nil value for the argument`)
	t.P(`      continue`)
	t.P(`    default:`)
	t.P(`      panic(`, t.pkgs["fmt"], `.Sprintf("Invalid option type %T on New`, servName, `Server", o))`)
	t.P(`    }`)
	t.P(`  }`)
	t.P()
	t.P(`  return &`, servStruct, `{`)
	t.P(`    `, servName, `: svc,`)
	t.P(`    pathPrefix: serverOpts.PathPrefix(),`)
	t.P(`    interceptor: `, t.pkgs["twirp"], `.ChainInterceptors(serverOpts.Interceptors...),`)
	t.P(`    hooks: serverOpts.Hooks,`)
	t.P(`    jsonSkipDefaults: serverOpts.JSONSkipDefaults,`)
	t.P(`  }`)
	t.P(`}`)
	t.P()

	// Write Errors
	t.P(`// writeError writes an HTTP response with a valid Twirp error format, and triggers hooks.`)
	t.P(`// If err is not a twirp.Error, it will get wrapped with twirp.InternalErrorWith(err)`)
	t.P(`func (s *`, servStruct, `) writeError(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, err error) {`)
	t.P(`  writeError(ctx, resp, err, s.hooks)`)
	t.P(`}`)
	t.P()

	// Routing.
	t.generateServerRouting(servStruct, file, service)

	// Methods.
	for _, method := range service.Method {
		t.generateServerMethod(service, method)
	}

	t.generateServiceMetadataAccessors(file, service)
}

func (t *twirp) generateServerRouting(servStruct string, file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	pkgName := pkgName(file)
	servName := serviceNameCamelCased(service)
	pkgServNameLit := pkgServiceNameLiteral(file, service)
	pkgServNameCc := pkgServiceNameCamelCased(file, service)

	t.P(`// `, servName, `PathPrefix is a convenience constant that could used to identify URL paths.`)
	t.P(`// Should be used with caution, it only matches routes generated by Twirp Go clients,`)
	t.P(`// that add a "/twirp" prefix by default, and use CamelCase service and method names.`)
	t.P(`// More info: https://twitchtv.github.io/twirp/docs/routing.html`)
	t.P(`const `, servName, `PathPrefix = "/twirp/`, pkgServNameCc, `/"`)
	t.P()

	t.P(`func (s *`, servStruct, `) ServeHTTP(resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  ctx := req.Context()`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithPackageName(ctx, "`, pkgName, `")`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithServiceName(ctx, "`, servName, `")`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithResponseWriter(ctx, resp)`)
	t.P()
	t.P(`  var err error`)
	t.P(`  ctx, err = callRequestReceived(ctx, s.hooks)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  if req.Method != "POST" {`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("unsupported method %q (only POST is allowed)", req.Method)`)
	t.P(`    s.writeError(ctx, resp, badRouteError(msg, req.Method, req.URL.Path))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  // Verify path format: [<prefix>]/<package>.<Service>/<Method>`)
	t.P(`  prefix, pkgService, method := parseTwirpPath(req.URL.Path)`)
	if pkgServNameLit == pkgServNameCc {
		t.P(`  if pkgService != `, strconv.Quote(pkgServNameLit), ` {`)
	} else { // proto service name is not CamelCased, but need to support CamelCased routes for Go clients (https://github.com/twitchtv/twirp/pull/257)
		t.P(`  if pkgService != `, strconv.Quote(pkgServNameLit), ` && pkgService != `, strconv.Quote(pkgServNameCc), ` {`)
	}
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("no handler for path %q", req.URL.Path)`)
	t.P(`    s.writeError(ctx, resp, badRouteError(msg, req.Method, req.URL.Path))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  if prefix != s.pathPrefix {`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("invalid path prefix %q, expected %q, on path %q", prefix, s.pathPrefix, req.URL.Path)`)
	t.P(`    s.writeError(ctx, resp, badRouteError(msg, req.Method, req.URL.Path))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  switch method {`)
	for _, method := range service.Method {
		methNameLit := methodNameLiteral(method)
		methNameCc := methodNameCamelCased(method)

		if methNameCc == methNameLit {
			t.P(`  case `, strconv.Quote(methNameLit), `:`)
		} else { // proto method name is not CamelCased, but need to support CamelCased routes for Go clients (https://github.com/twitchtv/twirp/pull/257)
			t.P(`  case `, strconv.Quote(methNameLit), `, `, strconv.Quote(methNameCc), `:`)
		}
		t.P(`    s.serve`, methNameCc, `(ctx, resp, req)`)
		t.P(`    return`)
	}
	t.P(`  default:`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("no handler for path %q", req.URL.Path)`)
	t.P(`    s.writeError(ctx, resp, badRouteError(msg, req.Method, req.URL.Path))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`}`)
	t.P()
}

func (t *twirp) generateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	methName := methodNameCamelCased(method)
	servStruct := serviceStruct(service)
	t.P(`func (s *`, servStruct, `) serve`, methName, `(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  header := req.Header.Get("Content-Type")`)
	t.P(`  i := `, t.pkgs["strings"], `.Index(header, ";")`)
	t.P(`  if i == -1 {`)
	t.P(`    i = len(header)`)
	t.P(`  }`)
	t.P(`  switch `, t.pkgs["strings"], `.TrimSpace(`, t.pkgs["strings"], `.ToLower(header[:i])) {`)
	t.P(`  case "application/json":`)
	t.P(`    s.serve`, methName, `JSON(ctx, resp, req)`)
	t.P(`  case "application/protobuf":`)
	t.P(`    s.serve`, methName, `Protobuf(ctx, resp, req)`)
	t.P(`  default:`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("unexpected Content-Type: %q", req.Header.Get("Content-Type"))`)
	t.P(`    twerr := badRouteError(msg, req.Method, req.URL.Path)`)
	t.P(`    s.writeError(ctx, resp, twerr)`)
	t.P(`  }`)
	t.P(`}`)
	t.P()
	t.generateServerJSONMethod(service, method)
	t.generateServerProtobufMethod(service, method)
}

func (t *twirp) generateServerJSONMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	servStruct := serviceStruct(service)
	methName := methodNameCamelCased(method)
	servName := serviceNameCamelCased(service)
	t.P(`func (s *`, servStruct, `) serve`, methName, `JSON(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  var err error`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
	t.P(`  ctx, err = callRequestRouted(ctx, s.hooks)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  reqContent := new(`, t.goTypeName(method.GetInputType()), `)`)
	t.P(`  unmarshaler := `, t.pkgs["jsonpb"], `.Unmarshaler{AllowUnknownFields: true}`)
	t.P(`  if err = unmarshaler.Unmarshal(req.Body, reqContent); err != nil {`)
	t.P(`    s.writeError(ctx, resp, malformedRequestError("the json request could not be decoded"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  handler := s.`, servName, `.`, methName)
	t.P(`  if s.interceptor != nil {`)
	t.generateServerInterceptorHandler(service, method)
	t.P(`  }`)
	t.P()
	t.P(`  // Call service method`)
	t.P(`  var respContent *`, t.goTypeName(method.GetOutputType()))
	t.P(`  func() {`)
	t.P(`    defer ensurePanicResponses(ctx, resp, s.hooks)`)
	t.P(`    respContent, err = handler(ctx, reqContent)`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  if respContent == nil {`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("received a nil *`, t.goTypeName(method.GetOutputType()), ` and nil error while calling `, methName, `. nil responses are not supported"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = callResponsePrepared(ctx, s.hooks)`)
	t.P()
	t.P(`  var buf `, t.pkgs["bytes"], `.Buffer`)
	t.P(`  marshaler := &`, t.pkgs["jsonpb"], `.Marshaler{OrigName: true, EmitDefaults: !s.jsonSkipDefaults}`)
	t.P(`  if err = marshaler.Marshal(&buf, respContent); err != nil {`)
	t.P(`    s.writeError(ctx, resp, wrapInternal(err, "failed to marshal json response"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, `, t.pkgs["http"], `.StatusOK)`)
	t.P(`  respBytes := buf.Bytes()`)
	t.P(`  resp.Header().Set("Content-Type", "application/json")`)
	t.P(`  resp.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))`)
	t.P(`  resp.WriteHeader(`, t.pkgs["http"], `.StatusOK)`)
	t.P()
	t.P(`  if n, err := resp.Write(respBytes); err != nil {`)
	t.P(`    msg := fmt.Sprintf("failed to write response, %d of %d bytes written: %s", n, len(respBytes), err.Error())`)
	t.P(`    twerr := `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.Unknown, msg)`)
	t.P(`    ctx = callError(ctx, s.hooks, twerr)`)
	t.P(`  }`)
	t.P(`  callResponseSent(ctx, s.hooks)`)
	t.P(`}`)
	t.P()
}

func (t *twirp) generateServerProtobufMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	servStruct := serviceStruct(service)
	methName := methodNameCamelCased(method)
	servName := serviceNameCamelCased(service)
	t.P(`func (s *`, servStruct, `) serve`, methName, `Protobuf(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  var err error`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
	t.P(`  ctx, err = callRequestRouted(ctx, s.hooks)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  buf, err := `, t.pkgs["ioutil"], `.ReadAll(req.Body)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, wrapInternal(err, "failed to read request body"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  reqContent := new(`, t.goTypeName(method.GetInputType()), `)`)
	t.P(`  if err = `, t.pkgs["proto"], `.Unmarshal(buf, reqContent); err != nil {`)
	t.P(`    s.writeError(ctx, resp, malformedRequestError("the protobuf request could not be decoded"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  handler := s.`, servName, `.`, methName)
	t.P(`  if s.interceptor != nil {`)
	t.generateServerInterceptorHandler(service, method)
	t.P(`  }`)
	t.P()
	t.P(`  // Call service method`)
	t.P(`  var respContent *`, t.goTypeName(method.GetOutputType()))
	t.P(`  func() {`)
	t.P(`    defer ensurePanicResponses(ctx, resp, s.hooks)`)
	t.P(`    respContent, err = handler(ctx, reqContent)`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  if respContent == nil {`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("received a nil *`, t.goTypeName(method.GetOutputType()), ` and nil error while calling `, methName, `. nil responses are not supported"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = callResponsePrepared(ctx, s.hooks)`)
	t.P()
	t.P(`  respBytes, err := `, t.pkgs["proto"], `.Marshal(respContent)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, wrapInternal(err, "failed to marshal proto response"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, `, t.pkgs["http"], `.StatusOK)`)
	t.P(`  resp.Header().Set("Content-Type", "application/protobuf")`)
	t.P(`  resp.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))`)
	t.P(`  resp.WriteHeader(`, t.pkgs["http"], `.StatusOK)`)
	t.P(`  if n, err := resp.Write(respBytes); err != nil {`)
	t.P(`    msg := fmt.Sprintf("failed to write response, %d of %d bytes written: %s", n, len(respBytes), err.Error())`)
	t.P(`    twerr := `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.Unknown, msg)`)
	t.P(`    ctx = callError(ctx, s.hooks, twerr)`)
	t.P(`  }`)
	t.P(`  callResponseSent(ctx, s.hooks)`)
	t.P(`}`)
	t.P()
}

func (t *twirp) generateClientInterceptorCaller(method *descriptor.MethodDescriptorProto) {
	methName := methodNameCamelCased(method)
	t.generateInterceptorFunc("c", "caller", "c.call"+methName, method)
}

func (t *twirp) generateServerInterceptorHandler(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	methName := methodNameCamelCased(method)
	servName := serviceNameCamelCased(service)
	t.generateInterceptorFunc("s", "handler", "s."+servName+"."+methName, method)
}

func (t *twirp) generateInterceptorFunc(
	receiverName string,
	varName string,
	delegateFuncName string,
	method *descriptor.MethodDescriptorProto,
) {
	inputType := t.goTypeName(method.GetInputType())
	outputType := t.goTypeName(method.GetOutputType())
	t.P(`    `, varName, ` = func(ctx `, t.pkgs["context"], `.Context, req *`, inputType, `) (*`, outputType, `, error) {`)
	t.P(`      resp, err := `, receiverName, `.interceptor(`)
	t.P(`        func(ctx `, t.pkgs["context"], ` .Context, req interface{}) (interface{}, error) {`)
	t.P(`          typedReq, ok := req.(*`, inputType, `)`)
	t.P(`          if !ok {`)
	t.P(`            return nil, `, t.pkgs["twirp"], `.InternalError("failed type assertion req.(*`, inputType, `) when calling interceptor")`)
	t.P(`          }`)
	t.P(`          return `, delegateFuncName, `(ctx, typedReq)`)
	t.P(`        },`)
	t.P(`      )(ctx, req)`)
	t.P(`      if resp != nil {`)
	t.P(`        typedResp, ok := resp.(*`, outputType, `)`)
	t.P(`        if !ok {`)
	t.P(`          return nil, `, t.pkgs["twirp"], `.InternalError("failed type assertion resp.(*`, outputType, `) when calling interceptor")`)
	t.P(`        }`)
	t.P(`        return typedResp, err`)
	t.P(`      }`)
	t.P(`      return nil, err`)
	t.P(`    }`)
}

// serviceMetadataVarName is the variable name used in generated code to refer
// to the compressed bytes of this descriptor. It is not exported, so it is only
// valid inside the generated package.
//
// protoc-gen-go writes its own version of this file, but so does
// protoc-gen-gogo - with a different name! Twirp aims to be compatible with
// both; the simplest way forward is to write the file descriptor again as
// another variable that we control.
func (t *twirp) serviceMetadataVarName() string {
	return fmt.Sprintf("twirpFileDescriptor%d", t.filesHandled)
}

func (t *twirp) generateServiceMetadataAccessors(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servStruct := serviceStruct(service)
	servPkg := pkgName(file)

	index := 0
	for i, s := range file.Service {
		if s.GetName() == service.GetName() {
			index = i
		}
	}
	t.P(`func (s *`, servStruct, `) ServiceDescriptor() ([]byte, int) {`)
	t.P(`  return `, t.serviceMetadataVarName(), `, `, strconv.Itoa(index))
	t.P(`}`)
	t.P()
	t.P(`func (s *`, servStruct, `) ProtocGenTwirpVersion() (string) {`)
	t.P(`  return `, strconv.Quote(gen.Version))
	t.P(`}`)
	t.P()
	t.P(`// PathPrefix returns the base service path, in the form: "/<prefix>/<package>.<Service>/"`)
	t.P(`// that is everything in a Twirp route except for the <Method>. This can be used for routing,`)
	t.P(`// for example to identify the requests that are targeted to this service in a mux.`)
	t.P(`func (s *`, servStruct, `) PathPrefix() (string) {`)
	servName := serviceNameCamelCased(service) // it should be serviceNameLiteral(service), but needs to use CamelCase routes for backwards compatibility
	t.P(`  return baseServicePath(s.pathPrefix, "`, servPkg, `", "`, servName, `") `)
	t.P(`}`)
}

func (t *twirp) generateFileDescriptor(file *descriptor.FileDescriptorProto) {
	// Copied straight of of protoc-gen-go, which trims out comments.
	pb := proto.Clone(file).(*descriptor.FileDescriptorProto)
	pb.SourceCodeInfo = nil

	b, err := proto.Marshal(pb)
	if err != nil {
		gen.Fail(err.Error())
	}

	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, _ = w.Write(b)
	w.Close()
	b = buf.Bytes()

	v := t.serviceMetadataVarName()
	t.P()
	t.P("var ", v, " = []byte{")
	t.P("	// ", fmt.Sprintf("%d", len(b)), " bytes of a gzipped FileDescriptorProto")
	for len(b) > 0 {
		n := 16
		if n > len(b) {
			n = len(b)
		}

		s := ""
		for _, c := range b[:n] {
			s += fmt.Sprintf("0x%02x,", c)
		}
		t.P(`	`, s)

		b = b[n:]
	}
	t.P("}")
}

func (t *twirp) printComments(comments typemap.DefinitionComments) bool {
	text := strings.TrimSuffix(comments.Leading, "\n")
	if len(strings.TrimSpace(text)) == 0 {
		return false
	}
	split := strings.Split(text, "\n")
	for _, line := range split {
		t.P("// ", strings.TrimPrefix(line, " "))
	}
	return len(split) > 0
}

// Given a protobuf name for a Message, return the Go name we will use for that
// type, including its package prefix.
func (t *twirp) goTypeName(protoName string) string {
	def := t.reg.MessageDefinition(protoName)
	if def == nil {
		gen.Fail("could not find message for", protoName)
	}

	var prefix string
	if pkg := t.goPackageName(def.File); pkg != t.genPkgName {
		prefix = pkg + "."
	}

	var name string
	for _, parent := range def.Lineage() {
		name += stringutils.CamelCase(parent.Descriptor.GetName()) + "_"
	}
	name += stringutils.CamelCase(def.Descriptor.GetName())
	return prefix + name
}

func (t *twirp) goPackageName(file *descriptor.FileDescriptorProto) string {
	return t.fileToGoPackageName[file]
}

func (t *twirp) formattedOutput() string {
	// Reformat generated code.
	fset := token.NewFileSet()
	raw := t.output.Bytes()
	ast, err := parser.ParseFile(fset, "", raw, parser.ParseComments)
	if err != nil {
		// Print out the bad code with line numbers.
		// This should never happen in practice, but it can while changing generated code,
		// so consider this a debugging aid.
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(raw))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		gen.Fail("bad Go source code was generated:", err.Error(), "\n"+src.String())
	}

	out := bytes.NewBuffer(nil)
	err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(out, fset, ast)
	if err != nil {
		gen.Fail("generated Go source code could not be reformatted:", err.Error())
	}

	return out.String()
}

func unexported(s string) string { return strings.ToLower(s[:1]) + s[1:] }

func pkgName(file *descriptor.FileDescriptorProto) string {
	return file.GetPackage()
}

func serviceNameCamelCased(service *descriptor.ServiceDescriptorProto) string {
	return stringutils.CamelCase(service.GetName())
}

func serviceNameLiteral(service *descriptor.ServiceDescriptorProto) string {
	return service.GetName()
}

func pkgServiceNameCamelCased(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := serviceNameCamelCased(service)
	if pkg := pkgName(file); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func pkgServiceNameLiteral(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := serviceNameLiteral(service)
	if pkg := pkgName(file); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func serviceStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexported(serviceNameCamelCased(service)) + "Server"
}

func methodNameCamelCased(method *descriptor.MethodDescriptorProto) string {
	return stringutils.CamelCase(method.GetName())
}

func methodNameLiteral(method *descriptor.MethodDescriptorProto) string {
	return method.GetName()
}

func fileDescSliceContains(slice []*descriptor.FileDescriptorProto, f *descriptor.FileDescriptorProto) bool {
	for _, sf := range slice {
		if f == sf {
			return true
		}
	}
	return false
}
