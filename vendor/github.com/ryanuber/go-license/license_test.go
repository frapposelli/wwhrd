package license

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewLicense(t *testing.T) {
	l := New("MyLicense", "Some license text.")
	if l.Type != "MyLicense" {
		t.Fatalf("bad license type: %s", l.Type)
	}
	if l.Text != "Some license text." {
		t.Fatalf("bad license text: %s", l.Text)
	}
}

func TestNewFromFile(t *testing.T) {
	lf := filepath.Join("fixtures", "licenses", "MIT")

	lh, err := os.Open(lf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer lh.Close()

	licenseText, err := ioutil.ReadAll(lh)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	l, err := NewFromFile(lf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if l.Type != "MIT" {
		t.Fatalf("unexpected license type: %s", l.Type)
	}

	if l.Text != string(licenseText) {
		t.Fatalf("unexpected license text: %s", l.Text)
	}

	if l.File != lf {
		t.Fatalf("unexpected file path: %s", l.File)
	}

	// Fails properly if the file doesn't exist
	if _, err := NewFromFile("/tmp/go-license-nonexistent"); err == nil {
		t.Fatalf("expected error loading non-existent file")
	}

	// Fails properly if license type from file is not guessable
	f, err := ioutil.TempFile("", "go-license")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	f.WriteString("No license data")
	if _, err := NewFromFile(f.Name()); err == nil {
		t.Fatalf("expected error guessing license type from non-license file")
	}
}

func TestNewFromDir(t *testing.T) {
	d, err := ioutil.TempDir("", "go-license")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(d)

	// Fails properly if the directory contains no license files
	if _, err := NewFromDir(d); err == nil {
		t.Fatalf("expected error loading empty directory")
	}

	fPath := filepath.Join(d, "LICENSE")
	f, err := os.Create(fPath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	lh, err := os.Open(filepath.Join("fixtures", "licenses", "MIT"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer lh.Close()

	licenseText, err := ioutil.ReadAll(lh)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := f.Write(licenseText); err != nil {
		t.Fatalf("err: %s", err)
	}

	l, err := NewFromDir(d)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if l.Type != "MIT" {
		t.Fatalf("unexpected license type: %s", l.Type)
	}

	if l.Text != string(licenseText) {
		t.Fatalf("unexpected license text: %s", l.Text)
	}

	if l.File != fPath {
		t.Fatalf("unexpected file path: %s", l.File)
	}

	// Fails properly if the directory does not exist
	if _, err := NewFromDir("go-license-nonexistent"); err == nil {
		t.Fatalf("expected error loading non-existent directory")
	}

	// Fails properly if the directory specified is actually a file
	if _, err := NewFromDir(fPath); err == nil {
		t.Fatalf("expected error loading file as directory")
	}
}

func TestLicenseRecognized(t *testing.T) {
	// Known licenses are recognized
	l := New("MIT", "The MIT License (MIT)")
	if !l.Recognized() {
		t.Fatalf("license was not recognized")
	}

	// Unknown licenses are not recognized
	l = New("None", "No license text")
	if l.Recognized() {
		t.Fatalf("fake license was recognized")
	}
}

func TestLicenseTypes(t *testing.T) {
	for _, ltype := range KnownLicenses {
		file := filepath.Join("fixtures", "licenses", ltype)
		fh, err := os.Open(file)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		lbytes, err := ioutil.ReadAll(fh)
		if err != nil {
			t.Fatalf("err :%s", err)
		}
		ltext := string(lbytes)

		l := New("", ltext)
		if err := l.GuessType(); err != nil {
			t.Fatalf("err: %s", err)
		}
		if l.Type != ltype {
			t.Fatalf("\nexpected: %s\ngot: %s", ltype, l.Type)
		}
	}
}

func TestLicenseTypes_Abbreviated(t *testing.T) {
	// Abbreviated Apache 2.0 license is recognized
	l := New("", "http://www.apache.org/licenses/LICENSE-2.0")
	if err := l.GuessType(); err != nil {
		t.Fatalf("err: %s", err)
	}
	if l.Type != LicenseApache20 {
		t.Fatalf("\nexpected: %s\ngot: %s", LicenseApache20, l.Type)
	}
}

func TestMatchLicenseFile(t *testing.T) {
	// should always return the original test file (mixed case), and
	//  not the license file version (typically upper case)

	licenses := []string{"copying.txt", "COPYING", "License"}
	tests := []struct {
		files []string
		want  []string
	}{
		{[]string{".", "junk", "COPYING"}, []string{"COPYING"}},
		{[]string{"junk", "copy"}, []string{}},
		{[]string{"LICENSE", "foo"}, []string{"LICENSE"}},
		{[]string{"LICENSE.junk", "foo"}, []string{}},
		{[]string{"something", "Copying.txt"}, []string{"Copying.txt"}},
		{[]string{"COPYING", "junk", "Copying.txt"}, []string{"COPYING", "Copying.txt"}},
	}

	for pos, tt := range tests {
		got := matchLicenseFile(licenses, tt.files)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Test %d: expected %v, got %v", pos, tt.want, got)
		}
	}
}

func TestGetLicenseFile(t *testing.T) {
	// should always return the original test file (mixed case), and
	//  not the license file version (typically upper case)

	licenses := []string{"copying.txt", "COPYING", "License"}
	tests := []struct {
		files []string
		want  string
		err   error
	}{
		{[]string{".", "junk", "COPYING"}, "COPYING", nil},                    // 1 match
		{[]string{"junk", "copy"}, "", ErrNoLicenseFile},                      // 0 match
		{[]string{"COPYING", "junk", "Copying.txt"}, "", ErrMultipleLicenses}, // 2 match
	}

	for pos, tt := range tests {
		got, err := getLicenseFile(licenses, tt.files)
		if got != tt.want || !reflect.DeepEqual(err, tt.err) {
			t.Errorf("Test %d: expected %q with error '%v', got %q with '%v'", pos, tt.want, tt.err, got, err)
		}
	}
}
