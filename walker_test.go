package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkImports(t *testing.T) {
	dir, rm := mockGoPackageDir(t, "TestWalkImports")
	defer rm()

	pkgs, err := WalkImports(dir)
	assert.NoError(t, err)

	res := make(map[string]bool)
	res["github.com/fake/package"] = true
	res["github.com/fake/nested/inside/a/package"] = true
	res["root"] = true
	assert.Equal(t, res, pkgs)

	negRes := make(map[string]bool)
	negRes["github.com/this/does/not/exist"] = true
	assert.NotEqual(t, negRes, pkgs)

}

func TestGetLicenses(t *testing.T) {
	dir, rm := mockGoPackageDir(t, "TestGetLicenses")
	defer rm()

	pkgs, err := WalkImports(dir)
	assert.NoError(t, err)
	lics := GetLicenses(dir, pkgs)

	res := make(map[string]string)
	res["github.com/fake/package"] = "BSD-3-Clause"
	res["github.com/fake/nested/inside/a/package"] = "BSD-3-Clause"

	assert.Equal(t, res, lics)
}
