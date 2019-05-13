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

import "testing"

func TestParseGoPackageOption(t *testing.T) {
	testcase := func(in, wantImport, wantPkg string) func(*testing.T) {
		return func(t *testing.T) {
			in := ""
			wantImport, wantPkg := "", ""
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
}
