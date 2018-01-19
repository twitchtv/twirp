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
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var headerCopyright = regexp.MustCompile(`// Copyright \d{4} Twitch Interactive, Inc.  All Rights Reserved.`)

const headerLicense = `
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
`

var generatedCodeMatcher = regexp.MustCompile("// Code generated .* DO NOT EDIT")

func TestSourceCodeLicenseHeaders(t *testing.T) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == "_tools" || path == "vendor" {
			return filepath.SkipDir
		}

		if !strings.HasSuffix(path, ".go") {
			// Skip non-go files.
			return nil
		}

		if strings.HasSuffix(path, ".twirp.go") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		fileBuf := bytes.NewReader(fileBytes)

		if generatedCodeMatcher.MatchReader(fileBuf) {
			// Skip generated files.
			return nil
		}

		_, err = fileBuf.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		if !headerCopyright.MatchReader(fileBuf) {
			t.Errorf("%v is missing licensing header", path)
			return nil
		}

		if !bytes.Contains(fileBytes, []byte(headerLicense)) {
			t.Errorf("%v is missing licensing header", path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("error scanning directory for source code files: %v", err)
	}
}
