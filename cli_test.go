package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCliListSuccess(t *testing.T) {
	var out = &bytes.Buffer{}
	log.SetOutput(out)

	dir, rm := mockGoPackageDir(t, "TestCliListSuccess")
	defer rm()

	// Change working dir to test dir
	err := os.Chdir(dir)
	assert.NoError(t, err)

	cases := []struct {
		inArgs            []string
		outputWant        []string
		outputWantNoColor []string
	}{
		{
			[]string{"list"},
			[]string{"\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]string{`level=info msg="Found License" license=FreeBSD package="github.com/fake/package"`, `level=info msg="Found License" license=FreeBSD package="github.com/fake/nested/inside/a/package"`},
		},
		{
			[]string{"ls"},
			[]string{"\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]string{`level=info msg="Found License" license=FreeBSD package="github.com/fake/nested/inside/a/package"`, `level=info msg="Found License" license=FreeBSD package="github.com/fake/nested/inside/a/package"`},
		},
	}

	for _, c := range cases {
		_, err = newCli().ParseArgs(c.inArgs)
		assert.NoError(t, err)

		assert.Contains(t, c.outputWant, out.String())
		out.Reset()

		// no color
		_, err = newCli().ParseArgs(append(c.inArgs, "--no-color"))
		assert.NoError(t, err)

		for _, want := range c.outputWantNoColor {
			assert.Contains(t, out.String(), want)
		}
		out.Reset()
	}
}

func TestCliQuiet(t *testing.T) {
	initial := log.GetLevel()
	defer log.SetLevel(initial)

	err := setQuiet()
	assert.NoError(t, err)

	after := log.GetLevel()
	assert.Equal(t, log.ErrorLevel, after)
}

func TestCliCheck(t *testing.T) {
	var out = &bytes.Buffer{}
	log.SetOutput(out)

	dir, rm := mockGoPackageDir(t, "TestCliCheck")
	defer rm()

	// Change working dir to test dir
	err := os.Chdir(dir)
	assert.NoError(t, err)

	cases := []struct {
		inArgs            []string
		outputWant        []string
		outputWantNoColor []string
		err               []error
	}{
		{
			[]string{"check"},
			[]string{"\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]string{`level=info msg="Found Approved license" license=FreeBSD package="github.com/fake/package"`, `level=info msg="Found Approved license" license=FreeBSD package="github.com/fake/nested/inside/a/package"`},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-ex.yml"},
			[]string{"\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]string{`level=warning msg="Found exceptioned package" license=FreeBSD package="github.com/fake/package"`, `level=warning msg="Found exceptioned package" license=FreeBSD package="github.com/fake/nested/inside/a/package"`},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-exwc.yml"},
			[]string{"\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]string{`level=warning msg="Found exceptioned package" license=FreeBSD package="github.com/fake/package"`, `level=warning msg="Found exceptioned package" license=FreeBSD package="github.com/fake/nested/inside/a/package"`},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-bl.yml"},
			[]string{"\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/package\"\n\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n", "\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/nested/inside/a/package\"\n\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=\"github.com/fake/package\"\n"},
			[]string{`level=error msg="Found Non-Approved license" license=FreeBSD package="github.com/fake/package"`, `level=error msg="Found Non-Approved license" license=FreeBSD package="github.com/fake/nested/inside/a/package"`},
			[]error{fmt.Errorf("Non-Approved license found")},
		},
		{
			[]string{"check", "-f", "NONEXISTENT"},
			[]string{""},
			[]string{""},
			[]error{fmt.Errorf("Can't read config file: stat NONEXISTENT: no such file or directory"), fmt.Errorf("Can't read config file: GetFileAttributesEx NONEXISTENT: The system cannot find the file specified.")},
		},
		{
			[]string{"check", "-f", ".wwhrd-botched.yml"},
			[]string{""},
			[]string{""},
			[]error{fmt.Errorf("Can't read config file: Invalid timestamp: 'whitelist - THISMAKESNOSENSE' at line 1, column 0")},
		},
	}

	for _, c := range cases {
		_, err = newCli().ParseArgs(c.inArgs)
		assert.Contains(t, c.err, err)

		assert.Contains(t, c.outputWant, out.String())
		out.Reset()

		// no color
		_, err = newCli().ParseArgs(append(c.inArgs, "--no-color"))
		assert.Contains(t, c.err, err)

		for _, want := range c.outputWantNoColor {
			assert.Contains(t, out.String(), want)
		}
		out.Reset()
	}
}
