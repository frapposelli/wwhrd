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
				// the package might be nested inside a larger package, we try to find
				// the license starting from the beginning of the path.
				pak := strings.Split(k, "/")
				var path string
				for x, y := range pak {
					if x < 1 {
						path = filepath.Join(root, "vendor", y)
					} else {
						path = filepath.Join(path, y)
					}
					if l, err := license.NewFromDir(path); err != nil {
						continue
					} else {
						// We found a license in the leftmost package, that's enough for now
						lics[k] = l
						break
					}
				}
				if lics[k] == nil {
					// if our search didn't bear any fruit, ¯\_(ツ)_/¯
					lics[k] = &license.License{
						Type: "unrecognized",
						Text: "unrecognized license",
						File: "",
					}
				}
				continue
			}
			lics[k] = l
		}
	}

	return lics
}
