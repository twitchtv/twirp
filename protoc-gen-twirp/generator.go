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
	filesHandled   int
	currentPackage string // Go name of current package we're working on

	reg *typemap.Registry

	// Map to record whether we've built each package
	pkgs          map[string]string
	pkgNamesInUse map[string]bool

	// Map to record whether we've generated Stream types for request and response
	// types.
	streamTypes map[string]bool

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
		streamTypes:         make(map[string]bool),
		fileToGoPackageName: make(map[*descriptor.FileDescriptorProto]string),
		output:              bytes.NewBuffer(nil),
	}

	return t
}

func (t *twirp) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	t.genFiles = gen.FilesToGenerate(in)

	// Collect information on types.
	t.reg = typemap.New(in.ProtoFile)

	// Register names of packages that we import.
	t.registerPackageName("bufio")
	t.registerPackageName("binary")

	t.registerPackageName("bytes")
	t.registerPackageName("strings")
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

	// Next, we need to pick names for all the files that are dependencies.
	for _, f := range in.ProtoFile {
		if fileDescSliceContains(t.genFiles, f) {
			// This is a file we are generating. It gets the shared package name.
			t.fileToGoPackageName[f] = t.genPkgName
		} else {
			// This is a dependency. Use its package name.
			name := f.GetPackage()
			if name == "" {
				name = stringutils.BaseName(f.GetName())
			}
			name = stringutils.CleanIdentifier(name)
			t.fileToGoPackageName[f] = name
			t.registerPackageName(name)
		}
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

	// For each service, generate client stubs and server
	for i, service := range file.Service {
		t.generateService(file, service, i)

		// For streaming RPCs, we need to define Stream types. This won't do
		// anything if the service has no streaming RPCs.
		t.generateStreamTypesForService(service)
	}

	// Util functions only generated once per package
	if t.filesHandled == 0 {
		t.generateUtils()

		if t.genFilesIncludeStreams() {
			t.generateStreamUtils()
		}
	}

	t.generateFileDescriptor(file)

	resp.Name = proto.String(goFileName(file))
	resp.Content = proto.String(t.formattedOutput())
	t.output.Reset()

	t.filesHandled++
	return resp
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
	t.P(`import `, t.pkgs["bytes"], ` "bytes"`)
	t.P(`import `, t.pkgs["strings"], ` "strings"`)
	t.P(`import `, t.pkgs["context"], ` "context"`)
	t.P(`import `, t.pkgs["fmt"], ` "fmt"`)
	t.P(`import `, t.pkgs["ioutil"], ` "io/ioutil"`)
	t.P(`import `, t.pkgs["http"], ` "net/http"`)
	t.P()
	t.P(`import `, t.pkgs["jsonpb"], ` "github.com/golang/protobuf/jsonpb"`)
	t.P(`import `, t.pkgs["proto"], ` "github.com/golang/protobuf/proto"`)
	t.P(`import `, t.pkgs["twirp"], ` "github.com/twitchtv/twirp"`)
	t.P(`import `, t.pkgs["ctxsetters"], ` "github.com/twitchtv/twirp/ctxsetters"`)
	t.P()

	// It's legal to import a message and use it as an input or output for a
	// method. Make sure to import the package of any such message. First, dedupe
	// them.
	deps := make(map[string]string) // Map of package name to quoted import path.
	ourImportPath := path.Dir(goFileName(file))
	for _, s := range file.Service {
		for _, m := range s.Method {
			defs := []*typemap.MessageDefinition{
				t.reg.MethodInputDefinition(m),
				t.reg.MethodOutputDefinition(m),
			}
			for _, def := range defs {
				importPath := path.Dir(goFileName(def.File))
				if importPath != ourImportPath {
					pkg := t.goPackageName(def.File)
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
	t.P(`import `, t.pkgs["strconv"], ` "strconv"`)
	t.P(`import `, t.pkgs["json"], ` "encoding/json"`)
	t.P(`import `, t.pkgs["url"], ` "net/url"`)
	if t.genFilesIncludeStreams() {
		t.P(`import `, t.pkgs["bufio"], ` "bufio"`)
		t.P(`import `, t.pkgs["binary"], ` "encoding/binary"`)
	}
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
	t.P(`	// ServiceDescriptor returns gzipped bytes describing the .proto file that`)
	t.P(`	// this service was generated from. Once unzipped, the bytes can be`)
	t.P(`	// unmarshalled as a`)
	t.P(`	// github.com/golang/protobuf/protoc-gen-go/descriptor.FileDescriptorProto.`)
	t.P(` //`)
	t.P(`	// The returned integer is the index of this particular service within that`)
	t.P(`	// FileDescriptorProto's 'Service' slice of ServiceDescriptorProtos. This is a`)
	t.P(`	// low-level field, expected to be used for reflection.`)
	t.P(`	ServiceDescriptor() ([]byte, int)`)
	t.P(`	// ProtocGenTwirpVersion is the semantic version string of the version of`)
	t.P(`	// twirp used to generate this file.`)
	t.P(`	ProtocGenTwirpVersion() string`)
	t.P(`}`)
	t.P()

	t.P(`// WriteError writes an HTTP response with a valid Twirp error format.`)
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
	t.P(`  resp.Header().Set("Content-Type", "application/json") // Error responses are always JSON (instead of protobuf)`)
	t.P(`  resp.WriteHeader(statusCode) // HTTP response status code`)
	t.P(``)
	t.P(`  respBody := marshalErrorToJSON(twerr)`)
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

	t.P(`// urlBase helps ensure that addr specifies a scheme. If it is unparsable`)
	t.P(`// as a URL, it returns addr unchanged.`)
	t.P(`func urlBase(addr string) string {`)
	t.P(`  // If the addr specifies a scheme, use it. If not, default to`)
	t.P(`  // http. If url.Parse fails on it, return it unchanged.`)
	t.P(`  url, err := `, t.pkgs["url"], `.Parse(addr)`)
	t.P(`  if err != nil {`)
	t.P(`    return addr`)
	t.P(`  }`)
	t.P(`  if url.Scheme == "" {`)
	t.P(`    url.Scheme = "http"`)
	t.P(`  }`)
	t.P(`  return url.String()`)
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

	t.P(`// JSON serialization for errors`)
	t.P(`type twerrJSON struct {`)
	t.P("  Code string            `json:\"code\"`")
	t.P("  Msg  string            `json:\"msg\"`")
	t.P("  Meta map[string]string `json:\"meta,omitempty\"`")
	t.P(`}`)
	t.P()
	t.P(`func (tj twerrJSON) toTwirpError() `, t.pkgs["twirp"], `.Error {`)
	t.P(`	errorCode := `, t.pkgs["twirp"], `.ErrorCode(tj.Code)`)
	t.P(`	if !`, t.pkgs["twirp"], `.IsValidErrorCode(errorCode) {`)
	t.P(`		msg := "invalid type returned from server error response: " + tj.Code`)
	t.P(`		return `, t.pkgs["twirp"], `.InternalError(msg)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	twerr := `, t.pkgs["twirp"], `.NewError(errorCode, tj.Msg)`)
	t.P(`	for k, v := range tj.Meta {`)
	t.P(`		twerr = twerr.WithMeta(k, v)`)
	t.P(`	}`)
	t.P(`	return twerr`)
	t.P(`}`)

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
	t.P(`    return clientError("failed to read server error response body", err)`)
	t.P(`  }`)
	t.P(`  var tj twerrJSON`)
	t.P(`  if err := `, t.pkgs["json"], `.Unmarshal(respBodyBytes, &tj); err != nil {`)
	t.P(`    // Invalid JSON response; it must be an error from an intermediary.`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("Error from intermediary with HTTP status code %d %q", statusCode, statusText)`)
	t.P(`    return twirpErrorFromIntermediary(statusCode, msg, string(respBodyBytes))`)
	t.P(`  }`)
	t.P(``)
	t.P(`  return tj.toTwirpError()`)
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
	t.P(`    case 429, 502, 503, 504: // Too Many Requests, Bad Gateway, Service Unavailable, Gateway Timeout`)
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

	t.P(`func isHTTPRedirect(status int) bool {`)
	t.P(`	return status >= 300 && status <= 399`)
	t.P(`}`)

	t.P(`// wrappedError implements the github.com/pkg/errors.Causer interface, allowing errors to be`)
	t.P(`// examined for their root cause.`)
	t.P(`type wrappedError struct {`)
	t.P(`	msg   string`)
	t.P(`	cause error`)
	t.P(`}`)
	t.P()
	t.P(`func wrapErr(err error, msg string) error { return &wrappedError{msg: msg, cause: err} }`)
	t.P(`func (e *wrappedError) Cause() error  { return e.cause }`)
	t.P(`func (e *wrappedError) Error() string { return e.msg + ": " + e.cause.Error() }`)
	t.P()
	t.P(`// clientError adds consistency to errors generated in the client`)
	t.P(`func clientError(desc string, err error) `, t.pkgs["twirp"], `.Error {`)
	t.P(`	return `, t.pkgs["twirp"], `.InternalErrorWith(wrapErr(err, desc))`)
	t.P(`}`)
	t.P()

	t.P(`// badRouteError is used when the twirp server cannot route a request`)
	t.P(`func badRouteError(msg string, method, url string) `, t.pkgs["twirp"], `.Error {`)
	t.P(`	err := `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.BadRoute, msg)`)
	t.P(`	err = err.WithMeta("twirp_invalid_route", method+" "+url)`)
	t.P(`	return err`)
	t.P(`}`)
	t.P()

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

	t.P(`// doProtobufRequest is common code to make a request to the remote twirp service.`)
	t.P(`func doProtobufRequest(ctx `, t.pkgs["context"], `.Context, client HTTPClient, url string, in, out `, t.pkgs["proto"], `.Message) (err error) {`)
	t.P(`  reqBodyBytes, err := `, t.pkgs["proto"], `.Marshal(in)`)
	t.P(`  if err != nil {`)
	t.P(`    return clientError("failed to marshal proto request", err)`)
	t.P(`  }`)
	t.P(`  reqBody := `, t.pkgs["bytes"], `.NewBuffer(reqBodyBytes)`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return clientError("aborted because context was done", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  req, err := newRequest(ctx, url, reqBody, "application/protobuf")`)
	t.P(`  if err != nil {`)
	t.P(`    return clientError("could not build request", err)`)
	t.P(`  }`)
	t.P(`  resp, err := client.Do(req)`)
	t.P(`  if err != nil {`)
	t.P(`    return clientError("failed to do request", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  defer func() {`)
	t.P(`    cerr := resp.Body.Close()`)
	t.P(`    if err == nil && cerr != nil {`)
	t.P(`      err = clientError("failed to close response body", cerr)`)
	t.P(`    }`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return clientError("aborted because context was done", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  if resp.StatusCode != 200 {`)
	t.P(`    return errorFromResponse(resp)`)
	t.P(`  }`)
	t.P()
	t.P(`  respBodyBytes, err := `, t.pkgs["ioutil"], `.ReadAll(resp.Body)`)
	t.P(`  if err != nil {`)
	t.P(`    return clientError("failed to read response body", err)`)
	t.P(`  }`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return clientError("aborted because context was done", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  if err = `, t.pkgs["proto"], `.Unmarshal(respBodyBytes, out); err != nil {`)
	t.P(`    return clientError("failed to unmarshal proto response", err)`)
	t.P(`  }`)
	t.P(`  return nil`)
	t.P(`}`)
	t.P()

	t.P(`// doJSONRequest is common code to make a request to the remote twirp service.`)
	t.P(`func doJSONRequest(ctx `, t.pkgs["context"], `.Context, client HTTPClient, url string, in, out `, t.pkgs["proto"], `.Message) (err error) {`)
	t.P(`  reqBody := `, t.pkgs["bytes"], `.NewBuffer(nil)`)
	t.P(`  marshaler := &`, t.pkgs["jsonpb"], `.Marshaler{OrigName: true}`)
	t.P(`  if err = marshaler.Marshal(reqBody, in); err != nil {`)
	t.P(`    return clientError("failed to marshal json request", err)`)
	t.P(`  }`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return clientError("aborted because context was done", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  req, err := newRequest(ctx, url, reqBody, "application/json")`)
	t.P(`  if err != nil {`)
	t.P(`    return clientError("could not build request", err)`)
	t.P(`  }`)
	t.P(`  resp, err := client.Do(req)`)
	t.P(`  if err != nil {`)
	t.P(`    return clientError("failed to do request", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  defer func() {`)
	t.P(`    cerr := resp.Body.Close()`)
	t.P(`    if err == nil && cerr != nil {`)
	t.P(`      err = clientError("failed to close response body", cerr)`)
	t.P(`    }`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return clientError("aborted because context was done", err)`)
	t.P(`  }`)
	t.P()
	t.P(`  if resp.StatusCode != 200 {`)
	t.P(`    return errorFromResponse(resp)`)
	t.P(`  }`)
	t.P()
	t.P(`  unmarshaler := `, t.pkgs["jsonpb"], `.Unmarshaler{AllowUnknownFields: true}`)
	t.P(`  if err = unmarshaler.Unmarshal(resp.Body, out); err != nil {`)
	t.P(`    return clientError("failed to unmarshal json response", err)`)
	t.P(`  }`)
	t.P(`  if err = ctx.Err(); err != nil {`)
	t.P(`    return clientError("aborted because context was done", err)`)
	t.P(`  }`)
	t.P(`  return nil`)
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
}

func (t *twirp) generateStreamUtils() {
	t.P(`type protoStreamReader struct {`)
	t.P(`	r *`, t.pkgs["bufio"], `.Reader`)
	t.P(`	c `, t.pkgs["io"], `.Closer`)
	t.P(``)
	t.P(`	maxSize int`)
	t.P(`}`)
	t.P(``)
	t.P(`func (r protoStreamReader) Read(msg `, t.pkgs["proto"], `.Message) error {`)
	t.P(`	// Get next field tag.`)
	t.P(`	tag, err := `, t.pkgs["binary"], `.ReadUvarint(r.r)`)
	t.P(`	if err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P(``)
	t.P(`	const (`)
	t.P(`		msgTag     = (1 << 3) | 2`)
	t.P(`		trailerTag = (2 << 3) | 2`)
	t.P(`	)`)
	t.P(``)
	t.P(`	if tag == trailerTag {`)
	t.P(`		// This is a trailer (twirp error), read it and then close the client`)
	t.P(`		defer r.c.Close()`)
	t.P(`		// Read the length delimiter`)
	t.P(`		l, err := binary.ReadUvarint(r.r)`)
	t.P(`		if err != nil {`)
	t.P(`			return clientError("unable to read trailer's length delimiter", err)`)
	t.P(`		}`)
	t.P(`		sb := new(`, t.pkgs["strings"], `.Builder)`)
	t.P(`		sb.Grow(int(l))`)
	t.P(`		_, err = `, t.pkgs["io"], `.Copy(sb, r.r)`)
	t.P(`		if err != nil {`)
	t.P(`			return clientError("unable to read trailer", err)`)
	t.P(`		}`)
	t.P(`		if sb.String() == "EOF" {`)
	t.P(`			return `, t.pkgs["io"], `.EOF`)
	t.P(`		}`)
	t.P(`		var tj twerrJSON`)
	t.P(`		if err = `, t.pkgs["json"], `.Unmarshal([]byte(sb.String()), &tj); err != nil {`)
	t.P(`			return clientError("unable to decode trailer", err)`)
	t.P(`		}`)
	t.P(`		return tj.toTwirpError()`)
	t.P(`	}`)
	t.P(``)
	t.P(`	if tag != msgTag {`)
	t.P(`		return `, t.pkgs["fmt"], `.Errorf("invalid field tag: %v", tag)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	// This is a real message. How long is it?`)
	t.P(`	l, err := `, t.pkgs["binary"], `.ReadUvarint(r.r)`)
	t.P(`	if err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P(`	if int(l) < 0 || int(l) > r.maxSize {`)
	t.P(`		return `, t.pkgs["io"], `.ErrShortBuffer`)
	t.P(`	}`)
	t.P(`	buf := make([]byte, int(l))`)
	t.P()
	t.P(`	// Go ahead and read a message.`)
	t.P(`	if _, err = `, t.pkgs["io"], `.ReadFull(r.r, buf); err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P()
	t.P(`	if err = `, t.pkgs["proto"], `.Unmarshal(buf, msg); err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P(`	return nil`)
	t.P(`}`)
	t.P()

	t.P(`type jsonStreamReader struct {`)
	t.P(`	dec         *`, t.pkgs["json"], `.Decoder`)
	t.P(`	unmarshaler *`, t.pkgs["jsonpb"], `.Unmarshaler`)
	t.P(`	messageStreamDone bool`)
	t.P(`}`)
	t.P()

	t.P(`func newJSONStreamReader(r `, t.pkgs["io"], `.Reader) (*jsonStreamReader, error) {`)
	t.P(`	// stream should start with {"messages":[`)
	t.P(`	dec := `, t.pkgs["json"], `.NewDecoder(r)`)
	t.P(`	t, err := dec.Token()`)
	t.P(`	if err != nil {`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	delim, ok := t.(`, t.pkgs["json"], `.Delim)`)
	t.P(`	if !ok || delim != '{' {`)
	t.P(`		return nil, fmt.Errorf("missing leading { in JSON stream, found %q", t)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	t, err = dec.Token()`)
	t.P(`	if err != nil {`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	key, ok := t.(string)`)
	t.P(`	if !ok || key != "messages" {`)
	t.P(`		return nil, fmt.Errorf("missing \"messages\" key in JSON stream, found %q", t)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	t, err = dec.Token()`)
	t.P(`	if err != nil {`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	delim, ok = t.(`, t.pkgs["json"], `.Delim)`)
	t.P(`	if !ok || delim != '[' {`)
	t.P(`		return nil, fmt.Errorf("missing [ to open messages array in JSON stream, found %q", t)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	return &jsonStreamReader{`)
	t.P(`		dec:         dec,`)
	t.P(`		unmarshaler: &`, t.pkgs["jsonpb"], `.Unmarshaler{AllowUnknownFields: true},`)
	t.P(`	}, nil`)
	t.P(`}`)
	t.P(``)

	t.P(`func (r *jsonStreamReader) Read(msg proto.Message) error {`)
	t.P(`	if !r.messageStreamDone && r.dec.More() {`)
	t.P(`		return r.unmarshaler.UnmarshalNext(r.dec, msg)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	// else, we hit the end of the message stream. finish up the array, and then read the trailer.`)
	t.P(`	r.messageStreamDone = true`)
	t.P(`	t, err := r.dec.Token()`)
	t.P(`	if err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P(`	delim, ok := t.(json.Delim)`)
	t.P(`	if !ok || delim != ']' {`)
	t.P(`		return fmt.Errorf("missing end of message array in JSON stream, found %q", t)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	t, err = r.dec.Token()`)
	t.P(`	if err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P(`	key, ok := t.(string)`)
	t.P(`	if !ok || key != "trailer" {`)
	t.P(`		return fmt.Errorf("missing trailer after messages in JSON stream, found %q", t)`)
	t.P(`	}`)
	t.P(``)
	t.P(`	var tj twerrJSON`)
	t.P(`	err = r.dec.Decode(&tj)`)
	t.P(`	if err != nil {`)
	t.P(`		return err`)
	t.P(`	}`)
	t.P(``)
	t.P(`	if tj.Code == "stream_complete" {`)
	t.P(`		return `, t.pkgs["io"], `.EOF`)
	t.P(`	}`)
	t.P(``)
	t.P(`	return tj.toTwirpError()`)
	t.P(`}`)
	t.P()
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
	servName := serviceName(service)

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
	servName := serviceName(service)

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
		t.P(t.signature(method))
		t.P()
	}
	t.P(`}`)
}

type rpcType string

const (
	unary         rpcType = "unary"
	download      rpcType = "download"
	upload        rpcType = "upload"
	bidirectional rpcType = "bidirectional"
)

func methodRPCType(method *descriptor.MethodDescriptorProto) rpcType {
	var (
		streamInput  = method.GetClientStreaming()
		streamOutput = method.GetServerStreaming()
	)
	switch {
	case !streamInput && !streamOutput:
		return unary
	case streamInput && !streamOutput:
		return upload
	case !streamInput && streamOutput:
		return download
	case streamInput && streamOutput:
		return bidirectional
	}
	return ""
}

func (t *twirp) signature(method *descriptor.MethodDescriptorProto) string {
	switch methodRPCType(method) {
	case unary:
		return t.unarySignature(method)
	case upload:
		return t.uploadSignature(method)
	case download:
		return t.downloadSignature(method)
	case bidirectional:
		return t.bidirectionalSignature(method)
	}
	return ""
}

func (t *twirp) unarySignature(method *descriptor.MethodDescriptorProto) string {
	methName := methodName(method)
	inputType := t.methodInputType(method)
	outputType := t.methodOutputType(method)
	return fmt.Sprintf(`%s(ctx %s.Context, in *%s) (*%s, error)`, methName, t.pkgs["context"], inputType, outputType)
}

func (t *twirp) uploadSignature(method *descriptor.MethodDescriptorProto) string {
	return ""
}

func (t *twirp) downloadSignature(method *descriptor.MethodDescriptorProto) string {
	methName := methodName(method)
	inputType := t.methodInputType(method)
	outputType := t.methodOutputType(method)
	return fmt.Sprintf(`%s(ctx %s.Context, in *%s) (%s, error)`, methName, t.pkgs["context"], inputType, outputType)
}

func (t *twirp) bidirectionalSignature(method *descriptor.MethodDescriptorProto) string {
	return ""
}

// valid names: 'JSON', 'Protobuf'
func (t *twirp) generateClient(name string, file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servName := serviceName(service)
	pathPrefixConst := servName + "PathPrefix"
	structName := unexported(servName) + name + "Client"
	newClientFunc := "New" + servName + name + "Client"

	methCnt := strconv.Itoa(len(service.Method))
	t.P(`type `, structName, ` struct {`)
	t.P(`  client HTTPClient`)
	t.P(`  urls   [`, methCnt, `]string`)
	t.P(`}`)
	t.P()
	t.P(`// `, newClientFunc, ` creates a `, name, ` client that implements the `, servName, ` interface.`)
	t.P(`// It communicates using `, name, ` and can be configured with a custom HTTPClient.`)
	t.P(`func `, newClientFunc, `(addr string, client HTTPClient) `, servName, ` {`)
	t.P(`  prefix := urlBase(addr) + `, pathPrefixConst)
	t.P(`  urls := [`, methCnt, `]string{`)
	for _, method := range service.Method {
		t.P(`    	prefix + "`, methodName(method), `",`)
	}
	t.P(`  }`)
	t.P(`  if httpClient, ok := client.(*`, t.pkgs["http"], `.Client); ok {`)
	t.P(`    return &`, structName, `{`)
	t.P(`      client: withoutRedirects(httpClient),`)
	t.P(`      urls:   urls,`)
	t.P(`    }`)
	t.P(`  }`)
	t.P(`  return &`, structName, `{`)
	t.P(`    client: client,`)
	t.P(`    urls:   urls,`)
	t.P(`  }`)
	t.P(`}`)
	t.P()

	for i, method := range service.Method {
		methName := methodName(method)
		pkgName := pkgName(file)
		outputType := t.goTypeName(method.GetOutputType())

		sig := t.signature(method)
		if sig != "" {
			t.P(`func (c *`, structName, `) `, sig, " {")
			t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithPackageName(ctx, "`, pkgName, `")`)
			t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithServiceName(ctx, "`, servName, `")`)
			t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
			switch methodRPCType(method) {
			case unary:
				t.P(`  out := new(`, outputType, `)`)
				t.P(`  err := do`, name, `Request(ctx, c.client, c.urls[`, strconv.Itoa(i), `], in, out)`)
				t.P(`  return out, err`)
				t.P(`}`)
				t.P()
			case download:
				t.P(`  reqBodyBytes, err := `, t.pkgs["proto"], `.Marshal(in)`)
				t.P(`  if err != nil {`)
				t.P(`	   return nil, clientError("failed to marshal proto request", err)`)
				t.P(`  }`)
				t.P(`  reqBody := `, t.pkgs["bytes"], `.NewBuffer(reqBodyBytes)`)
				t.P(`  if err = ctx.Err(); err != nil {`)
				t.P(`	   return nil, clientError("aborted because context was done", err)`)
				t.P(`  }`)
				t.P(``)

				if name == "Protobuf" {
					t.P(`  req, err := newRequest(ctx, c.urls[`, strconv.Itoa(i), `], reqBody, "application/protobuf")`)
				} else {
					t.P(`  req, err := newRequest(ctx, c.urls[`, strconv.Itoa(i), `], reqBody, "application/json")`)
				}
				t.P(`  if err != nil {`)
				t.P(`	   return nil, clientError("could not build request", err)`)
				t.P(`  }`)
				t.P(`  resp, err := c.client.Do(req)`)
				t.P(`  if err != nil {`)
				t.P(`	  return nil, clientError("failed to do request", err)`)
				t.P(`  }`)
				t.P(`  if err = ctx.Err(); err != nil {`)
				t.P(`    return nil, clientError("aborted because context was done", err)`)
				t.P(`  }`)
				t.P(`  if resp.StatusCode != 200 {`)
				t.P(`    return nil, errorFromResponse(resp)`)
				t.P(`  }`)
				t.P(``)
				if name == "Protobuf" {
					t.P(`  return &proto`, withoutPackageName(outputType), `StreamReader{`)
					t.P(`	  prs: protoStreamReader{`)
					t.P(`		  r:       `, t.pkgs["bufio"], `.NewReader(resp.Body),`)
					t.P(`		  c:       resp.Body,`)
					t.P(`		  maxSize: 1 << 21, // 1GB`)
					t.P(`	  },`)
					t.P(`  }, nil`)
					t.P(`}`)
				} else {
					t.P(`  jrs, err := newJSONStreamReader(resp.Body)`)
					t.P(`  if err != nil {`)
					t.P(`    return nil, err`)
					t.P(`  }`)
					t.P(`  return &json`, withoutPackageName(outputType), `StreamReader{`)
					t.P(`	  jrs: jrs,`)
					t.P(`   c: resp.Body,`)
					t.P(`  }, nil`)
					t.P(`}`)
				}
			default:
				t.P(`  return nil, nil}`)
			}
		}
	}
}

func (t *twirp) generateServer(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servName := serviceName(service)

	// Server implementation.
	servStruct := serviceStruct(service)
	t.P(`type `, servStruct, ` struct {`)
	t.P(`  `, servName)
	t.P(`  hooks     *`, t.pkgs["twirp"], `.ServerHooks`)
	t.P(`}`)
	t.P()

	// Constructor for server implementation
	t.P(`func New`, servName, `Server(svc `, servName, `, hooks *`, t.pkgs["twirp"], `.ServerHooks) TwirpServer {`)
	t.P(`  return &`, servStruct, `{`)
	t.P(`    `, servName, `: svc,`)
	t.P(`    hooks: hooks,`)
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

// pathPrefix returns the base path for all methods handled by a particular
// service. It includes a trailing slash. (for example
// "/twirp/twitch.example.Haberdasher/").
func pathPrefix(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	return fmt.Sprintf("/twirp/%s/", fullServiceName(file, service))
}

// pathFor returns the complete path for requests to a particular method on a
// particular service.
func pathFor(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) string {
	return pathPrefix(file, service) + stringutils.CamelCase(method.GetName())
}

func (t *twirp) generateServerRouting(servStruct string, file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	pkgName := pkgName(file)
	servName := serviceName(service)

	pathPrefixConst := servName + "PathPrefix"
	t.P(`// `, pathPrefixConst, ` is used for all URL paths on a twirp `, servName, ` server.`)
	t.P(`// Requests are always: POST `, pathPrefixConst, `/method`)
	t.P(`// It can be used in an HTTP mux to route twirp requests along with non-twirp requests on other routes.`)
	t.P(`const `, pathPrefixConst, ` = `, strconv.Quote(pathPrefix(file, service)))
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
	t.P(`    err = badRouteError(msg, req.Method, req.URL.Path)`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  switch req.URL.Path {`)
	for _, method := range service.Method {
		path := pathFor(file, service, method)
		methName := "serve" + stringutils.CamelCase(method.GetName())
		t.P(`  case `, strconv.Quote(path), `:`)
		t.P(`    s.`, methName, `(ctx, resp, req)`)
		t.P(`    return`)
	}
	t.P(`  default:`)
	t.P(`    msg := `, t.pkgs["fmt"], `.Sprintf("no handler for path %q", req.URL.Path)`)
	t.P(`    err = badRouteError(msg, req.Method, req.URL.Path)`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`}`)
	t.P()
}

func (t *twirp) generateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	methName := stringutils.CamelCase(method.GetName())
	servStruct := serviceStruct(service)

	t.P(`func (s *`, servStruct, `) serve`, methName, `(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  header := req.Header.Get("Content-Type")`)
	t.P(`  i := strings.Index(header, ";")`)
	t.P(`  if i == -1 {`)
	t.P(`    i = len(header)`)
	t.P(`  }`)
	t.P(`  switch strings.TrimSpace(strings.ToLower(header[:i])) {`)
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
	methName := stringutils.CamelCase(method.GetName())

	t.P(`func (s *`, servStruct, `) serve`, methName, `JSON(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)

	if methodRPCType(method) == unary {

		t.P(`  var err error`)
		t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
		t.P(`  ctx, err = callRequestRouted(ctx, s.hooks)`)
		t.P(`  if err != nil {`)
		t.P(`    s.writeError(ctx, resp, err)`)
		t.P(`    return`)
		t.P(`  }`)
		t.P()
		t.P(`  reqContent := new(`, t.methodInputType(method), `)`)
		t.P(`  unmarshaler := `, t.pkgs["jsonpb"], `.Unmarshaler{AllowUnknownFields: true}`)
		t.P(`  if err = unmarshaler.Unmarshal(req.Body, reqContent); err != nil {`)
		t.P(`    err = wrapErr(err, "failed to parse request json")`)
		t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
		t.P(`    return`)
		t.P(`  }`)
		t.P()
		t.P(`  // Call service method`)
		t.P(`  var respContent *`, t.methodOutputType(method))
		t.P(`  func() {`)
		t.P(`    defer func() {`)
		t.P(`      // In case of a panic, serve a 500 error and then panic.`)
		t.P(`      if r := recover(); r != nil {`)
		t.P(`        s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("Internal service panic"))`)
		t.P(`        panic(r)`)
		t.P(`      }`)
		t.P(`    }()`)
		t.P(`    respContent, err = s.`, methName, `(ctx, reqContent)`)
		t.P(`  }()`)
		t.P()
		t.P(`  if err != nil {`)
		t.P(`    s.writeError(ctx, resp, err)`)
		t.P(`    return`)
		t.P(`  }`)
		t.P(`  if respContent == nil {`)
		t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("received a nil *`, t.methodOutputType(method), ` and nil error while calling `, methName, `. nil responses are not supported"))`)
		t.P(`    return`)
		t.P(`  }`)
		t.P()
		t.P(`  ctx = callResponsePrepared(ctx, s.hooks)`)
		t.P()
		t.P(`  var buf `, t.pkgs["bytes"], `.Buffer`)
		t.P(`  marshaler := &`, t.pkgs["jsonpb"], `.Marshaler{OrigName: true}`)
		t.P(`  if err = marshaler.Marshal(&buf, respContent); err != nil {`)
		t.P(`    err = wrapErr(err, "failed to marshal json response")`)
		t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
		t.P(`    return`)
		t.P(`  }`)
		t.P()
		t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, `, t.pkgs["http"], `.StatusOK)`)
		t.P(`  resp.Header().Set("Content-Type", "application/json")`)
		t.P(`  resp.WriteHeader(`, t.pkgs["http"], `.StatusOK)`)
		t.P()
		t.P(`  respBytes := buf.Bytes()`)
		t.P(`  if n, err := resp.Write(respBytes); err != nil {`)
		t.P(`    msg := fmt.Sprintf("failed to write response, %d of %d bytes written: %s", n, len(respBytes), err.Error())`)
		t.P(`    twerr := `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.Unknown, msg)`)
		t.P(`    callError(ctx, s.hooks, twerr)`)
		t.P(`  }`)
		t.P(`  callResponseSent(ctx, s.hooks)`)
	}
	t.P(`}`)
	t.P()
}

func (t *twirp) generateServerProtobufMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	servStruct := serviceStruct(service)
	methName := stringutils.CamelCase(method.GetName())
	rpcType := methodRPCType(method)
	var respType string
	if rpcType == download || rpcType == bidirectional {
		respType = t.methodOutputType(method)
	} else {
		respType = `*` + t.methodOutputType(method)
	}

	t.P(`func (s *`, servStruct, `) serve`, methName, `Protobuf(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	if rpcType != unary && rpcType != download {
		t.P(`	s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("rpc type \"`, string(rpcType), `\" is not implemented"))`)
		t.P(`}`)
		t.P(``)
		return
	}
	t.P(`	var err error`)
	t.P(`	ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
	t.P(`	ctx, err = callRequestRouted(ctx, s.hooks)`)
	t.P(`	if err != nil {`)
	t.P(`		s.writeError(ctx, resp, err)`)
	t.P(`		return`)
	t.P(`	}`)
	t.P(``)

	// Read the request
	t.P(`	buf, err := `, t.pkgs["ioutil"], `.ReadAll(req.Body)`)
	t.P(`	if err != nil {`)
	t.P(`		err = wrapErr(err, "failed to read request body")`)
	t.P(`		s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`		return`)
	t.P(`	}`)
	t.P(`	reqContent := new(`, t.methodInputType(method), `)`)
	t.P(`	if err = `, t.pkgs["proto"], `.Unmarshal(buf, reqContent); err != nil {`)
	t.P(`		err = wrapErr(err, "failed to parse request proto")`)
	t.P(`		s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`		return`)
	t.P(`	}`)
	t.P(``)

	// Prepare the response
	t.P(`	// Call service method`)
	t.P(`	var respContent `, respType)
	t.P(`	respFlusher, canFlush := resp.(`, t.pkgs["http"], `.Flusher)`)
	t.P(`	func() {`)
	t.P(`		defer func() {`)
	t.P(`			// In case of a panic, serve a 500 error and then panic.`)
	t.P(`			if r := recover(); r != nil {`)
	t.P(`				s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("Internal service panic"))`)
	t.P(`				if canFlush {`)
	t.P(`					respFlusher.Flush()`)
	t.P(`				}`)
	t.P(`				panic(r)`)
	t.P(`			}`)
	t.P(`		}()`)
	t.P(`		respContent, err = s.`, methName, `(ctx, reqContent)`)
	t.P(`	}()`)
	t.P(``)
	t.P(`	if err != nil {`)
	t.P(`		s.writeError(ctx, resp, err)`)
	t.P(`		return`)
	t.P(`	}`)
	t.P(`	if respContent == nil {`)
	t.P(`		s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("received a nil `, t.methodInputType(method), ` and nil error while calling `, methName, `. nil responses are not supported"))`)
	t.P(`		return`)
	t.P(`	}`)
	t.P(``)
	t.P(`	ctx = callResponsePrepared(ctx, s.hooks)`)

	// Send the response
	if rpcType == unary {
		t.P(`	respBytes, err := `, t.pkgs["proto"], `.Marshal(respContent)`)
		t.P(`	if err != nil {`)
		t.P(`		err = wrapErr(err, "failed to marshal proto response")`)
		t.P(`		s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
		t.P(`		return`)
		t.P(`	}`)
		t.P(``)
	}
	t.P(`	ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, `, t.pkgs["http"], `.StatusOK)`)
	t.P(`	resp.Header().Set("Content-Type", "application/protobuf")`)
	t.P(`	resp.WriteHeader(`, t.pkgs["http"], `.StatusOK)`)
	t.P(``)
	switch rpcType {
	case unary:
		t.P(`	if n, err := resp.Write(respBytes); err != nil {`)
		t.P(`		msg := fmt.Sprintf("failed to write response, %d of %d bytes written: %s", n, len(respBytes), err.Error())`)
		t.P(`		twerr := `, t.pkgs["twirp"], `.NewError(`, t.pkgs["twirp"], `.Unknown, msg)`)
		t.P(`		callError(ctx, s.hooks, twerr)`)
		t.P(`	}`)
		t.P(``)
	case download:
		t.P(`	// Prepare trailer`)
		t.P(`	trailer := `, t.pkgs["proto"], `.NewBuffer(nil)`)
		t.P(`	_ = trailer.EncodeVarint((2 << 3) | 2) // field tag`)
		t.P(`	writeTrailer := func(err error) {`)
		t.P(`		if err == `, t.pkgs["io"], `.EOF {`)
		t.P(`			trailer.EncodeStringBytes("EOF")`)
		t.P(`			return`)
		t.P(`		}`)
		t.P(`		// Write trailer as json-encoded twirp err`)
		t.P(`		twerr, ok := err.(twirp.Error)`)
		t.P(`		if !ok {`)
		t.P(`			twerr = twirp.InternalErrorWith(err)`)
		t.P(`		}`)
		t.P(`		statusCode := `, t.pkgs["twirp"], `.ServerHTTPStatusFromErrorCode(twerr.Code())`)
		t.P(`		ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, statusCode)`)
		t.P(`		ctx = callError(ctx, s.hooks, twerr)`)
		t.P(`		if encodeErr := trailer.EncodeStringBytes(string(marshalErrorToJSON(twerr))); encodeErr != nil {`)
		t.P(`			_ = trailer.EncodeStringBytes("{\"code\":\"" + string(`, t.pkgs["twirp"], `.Internal) + "\",\"msg\":\"There was an error but it could not be serialized into JSON\"}") // fallback`)
		t.P(`		}`)
		t.P(`		_, writeErr := resp.Write(trailer.Bytes())`)
		t.P(`		if writeErr != nil {`)
		t.P(`			// Ignored, for the same reason as in the writeError func`)
		t.P(`			_ = writeErr`)
		t.P(`		}`)
		t.P(`		respContent.End(twerr)`)
		t.P(`	}`)
		t.P(``)
		t.P(`	messages := `, t.pkgs["proto"], `.NewBuffer(nil)`)
		t.P(`	for {`)
		t.P(`		msg, err := respContent.Next(ctx)`)
		t.P(`		if err != nil {`)
		t.P(`			writeTrailer(err)`)
		t.P(`			break`)
		t.P(`		}`)
		t.P(``)
		t.P(`		messages.Reset()`)
		t.P(`		_ = messages.EncodeVarint((1 << 3) | 2) // field tag`)
		t.P(`		err = messages.EncodeMessage(msg)`)
		t.P(`		if err != nil {`)
		t.P(`			err = wrapErr(err, "failed to marshal proto message")`)
		t.P(`			writeTrailer(err)`)
		t.P(`			break`)
		t.P(`		}`)
		t.P(``)
		t.P(`		_, err = resp.Write(messages.Bytes())`)
		t.P(`		if err != nil {`)
		t.P(`			err = wrapErr(err, "failed to send proto message")`)
		t.P(`			writeTrailer(err) // likely to fail on write, but try anyway to ensure ctx gets error code set for responseSent hook`)
		t.P(`			break`)
		t.P(`		}`)
		t.P(``)
		t.P(`		if canFlush {`)
		t.P(`			// TODO: Come up with a batching scheme to improve performance under high load`)
		t.P(`			//       and/or provide a hook for the respStream to control flushing the response.`)
		t.P(`			//       Flushing after each message dramatically reduces high-load throughput --`)
		t.P(`			//       difference can be more than 10x based on initial experiments`)
		t.P(`			respFlusher.Flush()`)
		t.P(`		}`)
		t.P(``)
		t.P(`		// TODO: Call a hook that we sent a message in a stream? (combine with flush hook?)`)
		t.P(`	}`)
		t.P(``)
	}

	t.P(`	callResponseSent(ctx, s.hooks)`)
	t.P(`}`)
	t.P()
}

// Do any of the services we are generating include streaming RPCs?
func (t *twirp) genFilesIncludeStreams() bool {
	for _, f := range t.genFiles {
		for _, s := range f.Service {
			for _, m := range s.Method {
				if m.GetClientStreaming() || m.GetServerStreaming() {
					return true
				}
			}
		}
	}
	return false
}

func (t *twirp) generateStreamTypesForService(service *descriptor.ServiceDescriptorProto) {
	for _, method := range service.Method {
		if method.GetClientStreaming() {
			t.generateStreamType(t.goTypeName(method.GetInputType()))
		}
		if method.GetServerStreaming() {
			t.generateStreamType(t.goTypeName(method.GetOutputType()))
		}
	}
}

func (t *twirp) generateStreamType(typeName string) {
	if t.streamTypes[typeName] {
		// already generated
		return
	}

	defer func() {
		t.streamTypes[typeName] = true
	}()

	streamTypeName := withoutPackageName(typeName) + "Stream"
	t.P(`// `, streamTypeName, ` represents a stream of `, typeName, ` messages.`)
	t.P(`type `, streamTypeName, ` interface {`)
	t.P(`  Next(context.Context) (*`, typeName, `, error)`)
	t.P(`  End(error)`)
	t.P(`}`)
	t.P()

	protoReaderTypeName := "proto" + streamTypeName + "Reader"

	t.P(`type `, protoReaderTypeName, ` struct {`)
	t.P(`  prs protoStreamReader`)
	t.P(`}`)
	t.P()
	t.P(`func (r `, protoReaderTypeName, `) Next(context.Context) (*`, typeName, `, error) {`)
	t.P(`  out := new(`, typeName, `)`)
	t.P(`  err := r.prs.Read(out)`)
	t.P(`  if err != nil {`)
	t.P(`    return nil, err`)
	t.P(`  }`)
	t.P(`  return out, nil`)
	t.P(`}`)
	t.P()
	t.P(`func (r `, protoReaderTypeName, `) End(error) { _ = r.prs.c.Close() }`)
	t.P()

	jsonReaderTypeName := "json" + streamTypeName + "Reader"
	t.P(`type `, jsonReaderTypeName, ` struct {`)
	t.P(`	jrs *jsonStreamReader`)
	t.P(`	c   `, t.pkgs["io"], `.Closer`)
	t.P(`}`)
	t.P()
	t.P(`func (r `, jsonReaderTypeName, `) Next(`, t.pkgs["context"], `.Context) (*`, typeName, `, error) {`)
	t.P(`	out := new(`, typeName, `)`)
	t.P(`	err := r.jrs.Read(out)`)
	t.P(`	if err != nil {`)
	t.P(`		return nil, err`)
	t.P(`	}`)
	t.P(`	return out, nil`)
	t.P(`}`)
	t.P()
	t.P(`func (r `, jsonReaderTypeName, `) End(error) { _ = r.c.Close() }`)
	t.P()
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
	w.Write(b)
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
		name += parent.Descriptor.GetName() + "_"
	}
	name += def.Descriptor.GetName()
	return prefix + name
}

func (t *twirp) methodInputType(method *descriptor.MethodDescriptorProto) string {
	name := t.goTypeName(method.GetInputType())
	if method.GetClientStreaming() {
		name = withoutPackageName(name) + "Stream"
	}
	return name
}

func (t *twirp) methodOutputType(method *descriptor.MethodDescriptorProto) string {
	name := t.goTypeName(method.GetOutputType())
	if method.GetServerStreaming() {
		name = withoutPackageName(name) + "Stream"
	}
	return name
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

func fullServiceName(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := stringutils.CamelCase(service.GetName())
	if pkg := pkgName(file); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func pkgName(file *descriptor.FileDescriptorProto) string {
	return file.GetPackage()
}

func serviceName(service *descriptor.ServiceDescriptorProto) string {
	return stringutils.CamelCase(service.GetName())
}

func serviceStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexported(serviceName(service)) + "Server"
}

func methodName(method *descriptor.MethodDescriptorProto) string {
	return stringutils.CamelCase(method.GetName())
}

// Remove a leading package selector (like 'twirp.' in 'twirp.Error') if present.
func withoutPackageName(v string) string {
	idx := strings.Index(v, ".")
	if idx == -1 {
		return v
	}
	return v[idx+1:]
}

func fileDescSliceContains(slice []*descriptor.FileDescriptorProto, f *descriptor.FileDescriptorProto) bool {
	for _, sf := range slice {
		if f == sf {
			return true
		}
	}
	return false
}
