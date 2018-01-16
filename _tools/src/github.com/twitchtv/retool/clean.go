package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

type stringSet map[string]struct{}

func (ss stringSet) add(val string) { ss[val] = struct{}{} }
func (ss stringSet) has(val string) bool {
	_, ok := ss[val]
	return ok
}

// compute the set of dependencies for a list of packages. The packages must
// already be present in toolDirPath.
func dependencies(pkgs []string) stringSet {
	deps := stringSet{}

	buildCtx := build.Default
	buildCtx.GOPATH = toolDirPath

	var resolve func(string, []string)
	resolve = func(parent string, pkgs []string) {
		for _, pkg := range pkgs {
			if !strings.Contains(pkg, ".") {
				continue
			}

			p, err := buildCtx.Import(pkg, filepath.Join(toolDirPath, "src", parent), 0)
			if err != nil {
				fatal(fmt.Sprintf("couldn't import package %q", pkg), err)
			}

			if deps.has(p.ImportPath) {
				continue
			}

			deps.add(p.ImportPath)
			resolve(p.ImportPath, p.Imports)
		}
	}

	resolve("", pkgs)

	return deps
}

// Remove unused files and unused packages from toolDirPath.
func clean(pkgs []string) {
	deps := dependencies(pkgs)
	base := filepath.Join(toolDirPath, "src")

	// Resolve any symlinks in the packages to keep, because we're going
	// to walk through the file system, so we need to trim stuff by
	// _filename_.
	for pkgPath := range deps {
		fullPath := filepath.Join(base, pkgPath)
		resolved, err := filepath.EvalSymlinks(fullPath)
		if err != nil {
			fatal(fmt.Sprintf("couldn't eval symlinks in %q", pkgPath), err)
		}
		// Undo the filepath.Join from above
		pkgPath, err = filepath.Rel(base, resolved)
		if err != nil {
			fatal(fmt.Sprintf("couldn't eval symlinks in %q", pkgPath), err)
		}
		deps.add(pkgPath)
	}

	var toDelete []string
	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		// Bubble up errors
		if err != nil {
			return err
		}

		// Skip the root directory
		if base == path {
			return nil
		}

		// Get the package directory
		pkg, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}

		// Delete files in packages that aren't in packages we need to build the
		// tools, and any non-go, non-legal files in packages we *do* need.
		if info.Mode().IsRegular() {
			pkg = filepath.Dir(pkg)
			if !(deps.has(pkg) && keepFile(path)) {
				toDelete = append(toDelete, path)
			}
			return nil
		}

		// If the path is a directory that's specially marked for preservation, keep
		// it and all its contents.
		if info.IsDir() && preserveDirectory(path) {
			return filepath.SkipDir
		}

		// If the folder is a kept package or a parent, don't delete it and keep recursing
		for p := range deps {
			if strings.HasPrefix(p, pkg) {
				return nil
			}
		}

		// Otherwise this is a package that isn't imported at all. Delete it and stop recursing
		toDelete = append(toDelete, path)
		return filepath.SkipDir
	})

	if err != nil {
		fatal("unable to clean _tools", err)
	}

	for _, path := range toDelete {
		err = os.RemoveAll(path)
		if err != nil {
			fatal("unable to remove file or directory", err)
		}
	}
}

func keepFile(filename string) bool {
	if strings.HasSuffix(filename, "_test.go") {
		return false
	}

	switch filepath.Ext(filename) {
	case ".go", ".s", ".c", ".h":
		return true
	}

	if isLegalFile(filename) {
		return true
	}
	return false
}

var commonLegalFilePrefixes = []string{
	"licence", // UK spelling
	"license", // US spelling
	"copying",
	"unlicense",
	"copyright",
	"copyleft",
	"authors",
	"contributors",
	"readme", // often has a license inline
}

var commonLegalFileSubstrings = []string{
	"legal",
	"notice",
	"disclaimer",
	"patent",
	"third-party",
	"thirdparty",
}

func isLegalFile(filename string) bool {
	base := strings.ToLower(filepath.Base(filename))
	for _, p := range commonLegalFilePrefixes {
		if strings.HasPrefix(base, p) {
			return true
		}
	}
	for _, s := range commonLegalFileSubstrings {
		if strings.Contains(base, s) {
			return true
		}
	}
	return false
}

// List of directories that should be completely preserved if they are present.
var preservedDirectories = []string{
	// gometalinter vendors its own linters and relies on this directory's
	// existence. See issue #7.
	filepath.Join("github.com", "alecthomas", "gometalinter", "_linters"),

	// sqlboiler requires template files to exist at runtime. See pull request #25.
	filepath.Join("github.com", "vattle", "sqlboiler", "templates"),
	filepath.Join("github.com", "vattle", "sqlboiler", "templates_test"),
	filepath.Join("github.com", "volatiletech", "sqlboiler", "templates"),
	filepath.Join("github.com", "volatiletech", "sqlboiler", "templates_test"),
}

// checks whether path is in the list of preserved directories.
func preserveDirectory(path string) bool {
	for _, d := range preservedDirectories {
		if strings.HasSuffix(path, d) {
			return true
		}
	}
	return false
}
