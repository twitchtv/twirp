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
	"testing"

	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

func TestParseGoPackageOption(t *testing.T) {
	testcase := func(in, wantImport, wantPkg string) func(*testing.T) {
		return func(t *testing.T) {
			haveImport, havePkg := parseGoPackageOption(in)
			if haveImport != wantImport {
				t.Errorf("wrong importPath, have=%q want=%q", haveImport, wantImport)
			}
			if havePkg != wantPkg {
				t.Errorf("wrong packageName, have=%q want=%q", havePkg, wantPkg)
			}
		}
	}

	t.Run("empty string", testcase("", "", ""))
	t.Run("bare package", testcase("foo", "", "foo"))
	t.Run("full import", testcase("github.com/example/foo", "github.com/example/foo", "foo"))
	t.Run("full import with override",
		testcase("github.com/example/foo;bar", "github.com/example/foo", "bar"))
	t.Run("non dotted import with package", testcase("foo;bar", "foo", "bar"))
}

func TestGoPackageOption(t *testing.T) {
	testcase := func(in, wantImport, wantPkg string, wantOK bool) func(*testing.T) {
		return func(t *testing.T) {
			haveImport, havePkg, haveOK := goPackageOption(&descriptor.FileDescriptorProto{
				Options: &descriptor.FileOptions{
					GoPackage: &in,
				},
			})
			if wantOK != haveOK {
				t.Errorf("wrong ok, have=%t want=%t", haveOK, wantOK)
			}
			if haveImport != wantImport {
				t.Errorf("wrong importPath, have=%q want=%q", haveImport, wantImport)
			}
			if havePkg != wantPkg {
				t.Errorf("wrong packageName, have=%q want=%q", havePkg, wantPkg)
			}
		}
	}

	t.Run("empty string", testcase("", "", "", false))
	t.Run("bare package", testcase("foo", "", "foo", true))
	t.Run("full import", testcase("github.com/example/foo", "github.com/example/foo", "foo", true))
	t.Run("full import with override",
		testcase("github.com/example/foo;bar", "github.com/example/foo", "bar", true))
	t.Run("non dotted import with package", testcase("foo;bar", "foo", "bar", true))
}

func TestGoFileName(t *testing.T) {
	testcase := func(srcrelpaths bool, modprefix, fname, gopkg, wantName string) func(t2 *testing.T) {
		return func(t *testing.T) {
			f := &descriptor.FileDescriptorProto{
				Name: &fname,
				Options: &descriptor.FileOptions{
					GoPackage: &gopkg,
				},
			}

			tw := &twirp{
				sourceRelativePaths: srcrelpaths,
				modulePrefix:        modprefix,
			}

			if name := tw.goFileName(f); name != wantName {
				t.Errorf("wrong goFileName, have=%q want=%q", name, wantName)
			}
		}
	}

	t.Run("paths=source_relative",
		testcase(true, "",
			"rpc/v1/service.proto", "example.com/module/package/rpc/v1",
			"rpc/v1/service.twirp.go"))

	t.Run("paths=import,module=example.com/module/package",
		testcase(false, "example.com/module/package",
			"rpc/v1/service.proto", "example.com/module/package/rpc/v1",
			"rpc/v1/service.twirp.go"))

	t.Run("paths=import,module=example.com/module/package/",
		testcase(false, "example.com/module/package/",
			"rpc/v1/service.proto", "example.com/module/package/rpc/v1",
			"rpc/v1/service.twirp.go"))
}
