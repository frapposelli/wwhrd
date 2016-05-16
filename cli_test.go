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
		outputWant string
	}{
		{
			[]string{"list"},
			"\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=github.com/fake/package\n",
		},
		{
			[]string{"ls"},
			"\x1b[34mINFO\x1b[0m[0000] Found License                                 \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=github.com/fake/package\n",
		},
	}
	for _, c := range cases {

		// Change working dir to test dir
		err := os.Chdir(dir)
		assert.NoError(t, err)

		_, err = parser.ParseArgs(c.inArgs)
		assert.NoError(t, err)

		assert.Equal(t, c.outputWant, out.String())
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
		outputWant string
		err        error
	}{
		{
			[]string{"check"},
			"\x1b[34mINFO\x1b[0m[0000] Found Approved license                        \x1b[34mlicense\x1b[0m=FreeBSD \x1b[34mpackage\x1b[0m=github.com/fake/package\n",
			nil,
		},
		{
			[]string{"check", "-f", ".wwhrd-ex.yml"},
			"\x1b[33mWARN\x1b[0m[0000] Found exceptioned package                     \x1b[33mlicense\x1b[0m=FreeBSD \x1b[33mpackage\x1b[0m=github.com/fake/package\n",
			nil,
		},
		{
			[]string{"check", "-f", ".wwhrd-bl.yml"},
			"\x1b[31mERRO\x1b[0m[0000] Found Non-Approved license                    \x1b[31mlicense\x1b[0m=FreeBSD \x1b[31mpackage\x1b[0m=github.com/fake/package\n",
			fmt.Errorf("Non-Approved license found"),
		},
		{
			[]string{"check", "-f", "NONEXISTENT"},
			"",
			fmt.Errorf("Can't read config file: stat NONEXISTENT: no such file or directory"),
		},
		{
			[]string{"check", "-f", ".wwhrd-botched.yml"},
			"",
			fmt.Errorf("Can't read config file: Invalid timestamp: 'whitelist - THISMAKESNOSENSE' at line 1, column 0"),
		},
	}
	for _, c := range cases {

		// Change working dir to test dir
		err := os.Chdir(dir)
		assert.NoError(t, err)

		_, err = parser.ParseArgs(c.inArgs)
		assert.Equal(t, c.err, err)

		assert.Equal(t, c.outputWant, out.String())
		out.Reset()

	}
}
