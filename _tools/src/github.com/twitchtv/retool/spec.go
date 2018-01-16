package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
)

// Filename to read/write the spec data.
const specfile = "tools.json"

type spec struct {
	Tools         []*tool
	RetoolVersion *semver.Version
}

// jsonSpec is a helper type to describe the JSON encoding of a spec
type jsonSpec struct {
	Tools         []*tool
	RetoolVersion string
}

func (s *spec) UnmarshalJSON(data []byte) error {
	js := new(jsonSpec)
	if err := json.Unmarshal(data, js); err != nil {
		return err
	}
	if js.RetoolVersion != "" {
		v, err := semver.NewVersion(js.RetoolVersion)
		if err != nil {
			return err
		}
		s.RetoolVersion = v
	}
	s.Tools = js.Tools
	return nil
}

func (s spec) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonSpec{
		Tools:         s.Tools,
		RetoolVersion: s.RetoolVersion.String(),
	})
}

func (s spec) write() error {
	specfilePath := filepath.Join(baseDirPath, specfile)

	f, err := os.Create(specfilePath)
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", specfile, err)
	}
	defer func() {
		_ = f.Close()
	}()

	// s.write() is called when we have successfully added, removed, or upgraded a
	// tool. The success of that operation indicates that we should be comfortable
	// bumping up this version.
	s.RetoolVersion = version

	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal json spec: %s", err)
	}

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("unable to write %s: %s", specfile, err)
	}

	return nil
}

func (s spec) find(t *tool) int {
	for i, tt := range s.Tools {
		if t.Repository == tt.Repository {
			return i
		}
	}
	return -1
}

func (s spec) cleanup() {
	var pkgs []string
	for _, t := range s.Tools {
		pkgs = append(pkgs, t.Repository)
	}
	clean(pkgs)
}

func readPath(path string) (spec, error) {
	file, err := os.Open(path)
	if err != nil {
		return spec{}, fmt.Errorf("unable to open spec file at %s: %s", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	s := new(spec)
	err = json.NewDecoder(file).Decode(s)
	if err != nil {
		return spec{}, err
	}
	return *s, nil
}

func read() (spec, error) {
	specfilePath := filepath.Join(baseDirPath, specfile)
	return readPath(specfilePath)
}

func specExists() bool {
	specfilePath := filepath.Join(baseDirPath, specfile)

	_, err := os.Stat(specfilePath)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		fatal("unable to stat tools.json: %s", err)
	}
	return true
}

func writeBlankSpec() error {
	return spec{
		Tools:         []*tool{},
		RetoolVersion: version,
	}.write()
}
