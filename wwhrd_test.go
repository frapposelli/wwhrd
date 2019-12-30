package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
)

func mockGoPackageDir(t *testing.T, prefix string) (dir string, rm func()) {

	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "vendor/github.com/fake/package"), 0755); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "vendor/github.com/faux/package"), 0755); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "vendor/github.com/fake/nested/inside/a/package"), 0755); err != nil {
		log.Fatal(err)
	}

	files := []struct {
		name    string
		content []byte
	}{
		{"mockpkg.go", []byte(mockGo)},
		{".wwhrd.yml", []byte(mockConf)},
		{".wwhrd-bl.yml", []byte(mockConfBL)},
		{".wwhrd-ex.yml", []byte(mockConfEX)},
		{".wwhrd-exwc.yml", []byte(mockConfEXWC)},
		{".wwhrd-botched.yml", []byte(mockConfBotched)},
		{filepath.Join("vendor/github.com/fake/package", "mockpkg.go"), []byte(mockVendor)},
		{filepath.Join("vendor/github.com/fake/package", "LICENSE"), []byte(mockLicense)}, // American English spelling
		{filepath.Join("vendor/github.com/faux/package", "mockpkg.go"), []byte(mockVendor)},
		{filepath.Join("vendor/github.com/faux/package", "LICENCE"), []byte(mockLicense)}, // British English spelling
		{filepath.Join("vendor/github.com/fake/nested", "LICENSE"), []byte(mockLicense)},
		{filepath.Join("vendor/github.com/fake/nested/inside/a/package", "mockpkg.go"), []byte(mockVendor)},
	}

	for _, c := range files {
		tmpfn := filepath.Join(dir, c.name)
		if err := ioutil.WriteFile(tmpfn, c.content, 0666); err != nil {
			log.Fatal(err)
		}
	}

	return dir, func() {
		defer os.RemoveAll(dir)
	}

}

//TestKillCmdParseErrors test for cli arguments and flags
func TestCliCommandsErrors(t *testing.T) {
	parser := newCli()

	cases := []struct {
		inArgs  []string
		errWant error
	}{
		{
			[]string{"check", "-f"}, &flags.Error{Type: flags.ErrExpectedArgument},
		},
	}
	for _, c := range cases {

		// Parse the flags
		_, errGot := parser.ParseArgs(c.inArgs)
		assert.IsType(t, c.errWant, errGot, "type mismatch")
		if _, ok := errGot.(*flags.Error); ok {
			typ := errGot.(*flags.Error).Type
			assert.Equal(t, typ, c.errWant.(*flags.Error).Type)
		}
	}
}

var mockConf = `---
whitelist:
  - BSD-3-Clause
`

var mockConfBL = `---
blacklist:
  - BSD-3-Clause
`

var mockConfEX = `---
exceptions:
  - github.com/fake/package
  - github.com/fake/nested/inside/a/package
`

var mockConfEXWC = `---
exceptions:
  - github.com/fake/...
`

var mockConfBotched = `---
whitelist
- THISMAKESNOSENSE
`

var mockGo = `package main
import (
	"github.com/fake/package"
	"github.com/fake/nested/inside/a/package"
)
func main() {}
`

var mockVendor = `package main
func main() {}
`

var mockLicense = `Copyright (c) 2016, Fabio Rapposelli
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

		Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
		Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.
		The names of its contributors may not be used to endorse or promote
products derived from this software without specific prior written
permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
`
