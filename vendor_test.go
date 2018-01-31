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

package twirp

import (
	"go/build"
	"os"
	"strings"
	"testing"
)

func TestNoExternalDeps(t *testing.T) {
	// Twirp commits its vendor directory so that 'go get' works for its main
	// packages, but vendoring dependencies of the 'twirp' package could cause
	// problems for users.
	//
	// The simplest way to make things safe is to have no non-stdlib dependencies
	// in the twirp package.

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unable to get current working directory: %v", err)
	}

	// Gather all imports of the current package recursively
	allPkgs := make(map[string]bool)

	var walkImports func(string)
	walkImports = func(pkgName string) {
		if allPkgs[pkgName] {
			// already visited
			return
		}
		allPkgs[pkgName] = true

		pkg, err := build.Default.Import(pkgName, wd, 0)
		if err != nil {
			t.Fatalf("unable to import package %s: %s", pkgName, err)
		}
		for _, imported := range pkg.Imports {
			// Standard libary packages don't have a '.' in them.
			if !strings.Contains(imported, ".") {
				continue
			}
			// This is a non-stdlib package. It's okay if it's a twirp package - as
			// long as it doesn't have any external deps itself.
			if strings.HasPrefix(imported, "github.com/twitchtv/twirp") {
				walkImports(imported)
			} else {
				t.Errorf("imported external dependency: %v, imported by %v", imported, pkgName)
			}
		}
	}
	walkImports("github.com/twitchtv/twirp")
}
