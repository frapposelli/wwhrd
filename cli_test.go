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
		outputWantNoColor []string
	}{
		{
			[]string{"list"},
			[]string{`level=info msg="Found License" license=BSD-3-Clause package="github.com/fake/package"`, `level=info msg="Found License" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`},
		},
		{
			[]string{"ls"},
			[]string{`level=info msg="Found License" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`, `level=info msg="Found License" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`},
		},
	}

	for _, c := range cases {
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
		outputWantNoColor []string
		err               []error
	}{
		{
			[]string{"check"},
			[]string{`level=info msg="Found Approved license" license=BSD-3-Clause package="github.com/fake/package"`, `level=info msg="Found Approved license" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-ex.yml"},
			[]string{`level=warning msg="Found exceptioned package" license=BSD-3-Clause package="github.com/fake/package"`, `level=warning msg="Found exceptioned package" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-exwc.yml"},
			[]string{`level=warning msg="Found exceptioned package" license=BSD-3-Clause package="github.com/fake/package"`, `level=warning msg="Found exceptioned package" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`},
			[]error{nil},
		},
		{
			[]string{"check", "-f", ".wwhrd-bl.yml"},
			[]string{`level=error msg="Found Non-Approved license" license=BSD-3-Clause package="github.com/fake/package"`, `level=error msg="Found Non-Approved license" license=BSD-3-Clause package="github.com/fake/nested/inside/a/package"`},
			[]error{fmt.Errorf("Non-Approved license found")},
		},
		{
			[]string{"check", "-f", "NONEXISTENT"},
			[]string{""},
			[]error{
				fmt.Errorf("can't read config file: stat NONEXISTENT: no such file or directory"),
				fmt.Errorf("can't read config file: GetFileAttributesEx NONEXISTENT: The system cannot find the file specified."),
				fmt.Errorf("can't read config file: CreateFile NONEXISTENT: The system cannot find the file specified."),
				fmt.Errorf("can't read config file: FindFirstFile NONEXISTENT: The system cannot find the file specified."),
			},
		},
		{
			[]string{"check", "-f", ".wwhrd-botched.yml"},
			[]string{""},
			[]error{fmt.Errorf("can't read config file: Invalid timestamp: 'whitelist - THISMAKESNOSENSE' at line 1, column 0")},
		},
	}

	for _, c := range cases {
		_, err = newCli().ParseArgs(append(c.inArgs, "--no-color"))
		assert.Contains(t, c.err, err)

		for _, want := range c.outputWantNoColor {
			assert.Contains(t, out.String(), want)
		}
		out.Reset()
	}
}
