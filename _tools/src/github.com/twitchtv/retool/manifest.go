package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const manifestFile = "manifest.json"

type manifest map[string]string

func getManifest() manifest {
	m := manifest{}

	file, err := os.Open(filepath.Join(toolDirPath, manifestFile))
	if err != nil {
		return m
	}
	defer func() {
		_ = file.Close()
	}()

	err = json.NewDecoder(file).Decode(&m)
	if err != nil {
		fatal("Failed to decode manifest", err)
	}
	return m
}

func (m manifest) write() {
	f, err := os.Create(filepath.Join(toolDirPath, manifestFile))
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()

	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return
	}

	_, _ = f.Write(bytes)
}

func (m manifest) outOfDate(ts []*tool) bool {
	// Make a copy to check for elements in ts but not m
	m2 := make(map[string]string)
	for k, v := range m {
		m2[k] = v
	}

	for _, t := range ts {
		if v, ok := m[t.Repository]; !ok || v != t.Commit {
			return true
		}
		delete(m2, t.Repository)
	}

	return len(m2) != 0
}

func (m manifest) replace(ts []*tool) {
	for k := range m {
		delete(m, k)
	}
	for _, t := range ts {
		m[t.Repository] = t.Commit
	}
}
