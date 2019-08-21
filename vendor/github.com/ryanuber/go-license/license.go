package license

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// Recognized license types
	LicenseMIT       = "MIT"
	LicenseISC       = "ISC"
	LicenseNewBSD    = "NewBSD"
	LicenseFreeBSD   = "FreeBSD"
	LicenseApache20  = "Apache-2.0"
	LicenseMPL20     = "MPL-2.0"
	LicenseGPL20     = "GPL-2.0"
	LicenseGPL30     = "GPL-3.0"
	LicenseLGPL21    = "LGPL-2.1"
	LicenseLGPL30    = "LGPL-3.0"
	LicenseAGPL30    = "AGPL-3.0"
	LicenseCDDL10    = "CDDL-1.0"
	LicenseEPL10     = "EPL-1.0"
	LicenseUnlicense = "Unlicense"
)

var (
	// Various errors
	ErrNoLicenseFile       = errors.New("license: unable to find any license file")
	ErrUnrecognizedLicense = errors.New("license: could not guess license type")
	ErrMultipleLicenses    = errors.New("license: multiple license files found")
)

// A set of reasonable license file names to use when guessing where the
// license may be. Case does not matter.
var DefaultLicenseFiles = []string{
	"license", "license.txt", "license.md",
	"licence", "licence.txt", "licence.md",
	"copying", "copying.txt", "copying.md",
	"unlicense",
}

// A slice of standardized license abbreviations
var KnownLicenses = []string{
	LicenseMIT,
	LicenseISC,
	LicenseNewBSD,
	LicenseFreeBSD,
	LicenseApache20,
	LicenseMPL20,
	LicenseGPL20,
	LicenseGPL30,
	LicenseLGPL21,
	LicenseLGPL30,
	LicenseAGPL30,
	LicenseCDDL10,
	LicenseEPL10,
	LicenseUnlicense,
}

// License describes a software license
type License struct {
	Type string // The type of license in use
	Text string // License text data
	File string // The path to the source file, if any
}

// New creates a new License from explicitly passed license type and data
func New(licenseType, licenseText string) *License {
	l := &License{
		Type: licenseType,
		Text: licenseText,
	}
	return l
}

// NewFromFile will attempt to load a license from a file on disk, and guess the
// type of license based on the bytes read.
func NewFromFile(path string) (*License, error) {
	licenseText, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	l := &License{
		Text: string(licenseText),
		File: path,
	}

	if err := l.GuessType(); err != nil {
		return nil, err
	}

	return l, nil
}

// NewFromDir will search a directory for well-known and accepted license file
// names, and if one is found, read in its content and guess the license type.
func NewFromDir(dir string) (*License, error) {
	l := new(License)

	if err := l.GuessFile(dir); err != nil {
		return nil, err
	}

	return NewFromFile(l.File)
}

// Recognized determines if the license is known to go-license.
func (l *License) Recognized() bool {
	for _, license := range KnownLicenses {
		if license == l.Type {
			return true
		}
	}
	return false
}

// GuessFile searches a given directory (non-recursively) for files with well-
// established names that indicate license content.
func (l *License) GuessFile(dir string) error {
	files, err := readDirectory(dir)
	if err != nil {
		return err
	}
	match, err := getLicenseFile(DefaultLicenseFiles, files)
	if err != nil {
		return err
	}
	l.File = filepath.Join(dir, match)
	return nil
}

// GuessType will scan license text and attempt to guess what license type it
// describes. It will return the license type on success, or an error if it
// cannot accurately guess the license type.
//
// This method is a hack. It might be more accurate to also scan the entire body
// of license text and compare it using an algorithm like Jaro-Winkler or
// Levenshtein against a generic version. The problem is that some of the
// common licenses, such as GPL-family licenses, are quite large, and running
// these algorithms against them is considerably more expensive and is still not
// completely deterministic on which license is in play. For now, we will just
// scan until we find differentiating strings and call that good-enuf.gov.
func (l *License) GuessType() error {
	newlineRegexp := regexp.MustCompile("(\r\n|\n)")
	spaceRegexp := regexp.MustCompile("\\s{2,}")

	// Lower case everything to make comparison more adaptable
	comp := strings.ToLower(l.Text)

	// Kill the newlines, since it is not clear if the provided license will
	// contain them or not, and either way it does not change the terms of the
	// license, so one is not "more correct" than the other. This just replaces
	// them with spaces. Also replace multiple spaces with a single space to
	// make comparison more simple.
	comp = newlineRegexp.ReplaceAllLiteralString(comp, " ")
	comp = spaceRegexp.ReplaceAllLiteralString(comp, " ")

	switch {
	case scan(comp, "permission is hereby granted, free of charge, to any "+
		"person obtaining a copy of this software"):
		l.Type = LicenseMIT

	case scan(comp, "permission to use, copy, modify, and/or distribute this "+
		"software for any"):
		l.Type = LicenseISC
	// When originally released the license did not include the term "and/or", this was added by ISC in 2007
	case scan(comp, "permission to use, copy, modify, and distribute this "+
		"software for any"):
		l.Type = LicenseISC

	case scan(comp, "apache license version 2.0, january 2004") ||
		scan(comp, "http://www.apache.org/licenses/license-2.0"):
		l.Type = LicenseApache20

	case scan(comp, "gnu general public license version 2, june 1991"):
		l.Type = LicenseGPL20

	case scan(comp, "gnu general public license version 3, 29 june 2007"):
		l.Type = LicenseGPL30

	case scan(comp, "gnu lesser general public license version 2.1, "+
		"february 1999"):
		l.Type = LicenseLGPL21

	case scan(comp, "gnu lesser general public license version 3, "+
		"29 june 2007"):
		l.Type = LicenseLGPL30

	case scan(comp, "gnu affero general public license "+
		"version 3, 19 november 2007"):
		l.Type = LicenseAGPL30

	case scan(comp, "mozilla public license") && scan(comp, "version 2.0"):
		l.Type = LicenseMPL20

	case scan(comp, "redistribution and use in source and binary forms"):
		switch {
		case scan(comp, "neither the name of"):
			l.Type = LicenseNewBSD
		default:
			l.Type = LicenseFreeBSD
		}

	case scan(comp, "common development and distribution license (cddl) "+
		"version 1.0"):
		l.Type = LicenseCDDL10

	case scan(comp, "eclipse public license - v 1.0"):
		l.Type = LicenseEPL10

	case scan(comp, "this is free and unencumbered software released into "+
		"the public domain"):
		l.Type = LicenseUnlicense

	default:
		return ErrUnrecognizedLicense
	}

	return nil
}

// scan is a shortcut function to check for a literal match within a string
// of text. Any text transformation should be done prior to calling this
// function so that it need not be repeated for every check.
func scan(text, match string) bool {
	return strings.Contains(text, match)
}

// returns a []string of files in a directory, or error
func readDirectory(dir string) ([]string, error) {
	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, len(fileinfos))
	for pos, fi := range fileinfos {
		files[pos] = fi.Name()
	}
	return files, nil
}

// returns files that case-insensitive matches any of the license
// files.  This is generic functionality so pulled out into separate
// function for testing
func matchLicenseFile(licenses []string, files []string) []string {
	out := make([]string, 0, 1)
	for _, file := range files {
		for _, license := range licenses {
			if strings.EqualFold(license, file) {
				out = append(out, file)
			}
		}
	}
	return out
}

// returns a single license filename or error
func getLicenseFile(licenses []string, files []string) (string, error) {
	matches := matchLicenseFile(licenses, files)

	switch len(matches) {
	case 0:
		return "", ErrNoLicenseFile
	case 1:
		return matches[0], nil
	default:
		return "", ErrMultipleLicenses
	}
}
