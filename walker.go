package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryanuber/go-license"
)

func WalkImports(root string) (map[string]bool, error) {

	pkgs := make(map[string]bool)

	var walkFn = func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, s := range f.Imports {
			p := strings.Replace(s.Path.Value, "\"", "", -1)
			pkgs[p] = true
		}
		return nil
	}

	if err := filepath.Walk(root, walkFn); err != nil {
		return nil, err
	}

	return pkgs, nil
}

func GetLicenses(root string, list map[string]bool) map[string]*license.License {

	lics := make(map[string]*license.License)

	for k := range list {
		fpath := filepath.Join(root, "vendor", k)
		pkg, err := os.Stat(fpath)
		if err != nil {
			continue
		}
		if pkg.IsDir() {
			l, err := license.NewFromDir(fpath)
			if err != nil {
				continue
			}
			lics[k] = l
		}
	}

	return lics
}
