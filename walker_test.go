package main

import (
	"path/filepath"
	"testing"

	"github.com/ryanuber/go-license"
	"github.com/stretchr/testify/assert"
)

//TestWalkImports test walking imports
func TestWalkImports(t *testing.T) {

	dir, rm := mockGoPackageDir(t, "TestWalkImports")
	defer rm()

	pkgs, err := WalkImports(dir)
	assert.NoError(t, err)

	res := make(map[string]bool)
	res["github.com/fake/package"] = true

	assert.Equal(t, res, pkgs)

}

//TestWalkImports test walking imports
func TestGetLicenses(t *testing.T) {

	dir, rm := mockGoPackageDir(t, "TestGetLicenses")
	defer rm()

	pkgs, err := WalkImports(dir)
	assert.NoError(t, err)
	lics := GetLicenses(dir, pkgs)

	res := make(map[string]*license.License)
	var lic license.License
	lic.File = filepath.Join(dir, "vendor/github.com/fake/package", "LICENSE")
	lic.Text = mockLicense
	lic.Type = "FreeBSD"

	res["github.com/fake/package"] = &lic

	assert.Equal(t, res, lics)

}
