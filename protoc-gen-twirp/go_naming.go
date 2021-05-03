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
	"path"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/twitchtv/twirp/internal/gen/stringutils"
)

// goPackageOption interprets the file's go_package option.
// If there is no go_package, it returns ("", "", false).
// If there's a simple name, it returns ("", pkg, true).
// If the option implies an import path, it returns (impPath, pkg, true).
func goPackageOption(f *descriptor.FileDescriptorProto) (impPath, pkg string, ok bool) {
	pkg = f.GetOptions().GetGoPackage()
	if pkg == "" {
		return "", "", false
	}
	if bits := strings.Split(pkg, ";"); len(bits) == 2 {
		return bits[0], bits[1], true
	}
	// The presence of a slash implies there's an import path.
	slash := strings.LastIndex(pkg, "/")
	if slash < 0 {
		return "", pkg, true
	}
	impPath, pkg = pkg, pkg[slash+1:]
	// A semicolon-delimited suffix overrides the package name.
	sc := strings.IndexByte(impPath, ';')
	if sc < 0 {
		return impPath, pkg, true
	}
	impPath, pkg = impPath[:sc], impPath[sc+1:]
	return impPath, pkg, true
}

// goPackageName returns the Go package name to use in the generated Go file.
// The result explicitly reports whether the name came from an option go_package
// statement. If explicit is false, the name was derived from the protocol
// buffer's package statement or the input file name.
func goPackageName(f *descriptor.FileDescriptorProto) (name string, explicit bool) {
	// Does the file have a "go_package" option?
	if _, pkg, ok := goPackageOption(f); ok {
		return pkg, true
	}

	// Does the file have a package clause?
	if pkg := f.GetPackage(); pkg != "" {
		return pkg, false
	}
	// Use the file base name.
	return stringutils.BaseName(f.GetName()), false
}

// goFileName returns the output name for the generated Go file.
func (t *twirp) goFileName(f *descriptor.FileDescriptorProto) string {
	name := *f.Name // proto file name
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)] // remove extension
	}
	name += ".twirp.go" // add twirp extension

	// with paths=source_relative, the directory is the same as the proto file
	if t.sourceRelativePaths {
		return name
	}
	// otherwise, the directory is taken from the option go_package
	if impPath, _, ok := goPackageOption(f); ok && impPath != "" {
		if t.modulePrefix != "" {
			impPath = strings.TrimPrefix(impPath, t.modulePrefix)
		}

		// Replace the existing dirname with the import path from go_package
		_, name = path.Split(name)
		name = path.Join(impPath, name)
		return name
	}

	return name
}

func parseGoPackageOption(v string) (importPath, packageName string) {
	// Allowed formats:
	// option go_package = "foo";
	// option go_package = "github.com/example/foo";
	// option go_package = "github.com/example/foo;bar";

	semicolonPos := strings.Index(v, ";")
	if semicolonPos > -1 {
		importPath = v[:semicolonPos]
		packageName = v[semicolonPos+1:]
		return
	}

	if strings.Contains(v, "/") {
		importPath = v
		_, packageName = path.Split(v)
		return
	}

	importPath = ""
	packageName = v
	return
}
