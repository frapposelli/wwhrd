package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCliListSuccess(t *testing.T) {
	parser := newCli()
	var out = &bytes.Buffer{}
	log.SetOutput(out)

	dir, rm := mockGoPackageDir(t, "TestCliListSuccess")
	defer rm()

	cases := []struct {
		inArgs     []string
		outputWant []string
	}{
		{
			[]string{"list"},
			[]string{"\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n"},
		},
		{
			[]string{"ls"},
			[]string{"\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n"},
		},
	}
	for _, c := range cases {

		// Change working dir to test dir
		err := os.Chdir(dir)
		assert.NoError(t, err)

		_, err = parser.ParseArgs(c.inArgs)
		assert.NoError(t, err)

		assert.Contains(t, c.outputWant, out.String())
		out.Reset()

	}
}

func TestCliCheck(t *testing.T) {
	parser := newCli()
	var out = &bytes.Buffer{}
	log.SetOutput(out)

	dir, rm := mockGoPackageDir(t, "TestCliCheck")
	defer rm()

	cases := []struct {
		inArgs     []string
		outputWant []string
		err        []error
	}{
		{
			[]string{"check"},
			[]string{"\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-ex.yml"},
			[]string{"\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-bl.yml"},
			[]string{"\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]error{fmt.Errorf("Non-Approved license found")},
		},
		{
			[]string{"check", "-f", "NONEXISTENT"},
			[]string{""},
			[]error{fmt.Errorf("Can't read config file: stat NONEXISTENT: no such file or directory"), fmt.Errorf("Can't read config file: GetFileAttributesEx NONEXISTENT: The system cannot find the file specified.")},
		},
		{
			[]string{"check", "-f", ".wwhrd-botched.yml"},
			[]string{""},
			[]error{fmt.Errorf("Can't read config file: Invalid timestamp: 'whitelist - THISMAKESNOSENSE' at line 1, column 0")},
		},
	}
	for _, c := range cases {

		// Change working dir to test dir
		err := os.Chdir(dir)
		assert.NoError(t, err)

		_, err = parser.ParseArgs(c.inArgs)
		assert.Contains(t, c.err, err)

		assert.Contains(t, c.outputWant, out.String())
		out.Reset()

	}
}
