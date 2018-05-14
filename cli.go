package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

type cliOpts struct {
	List        `command:"list" alias:"ls" description:"List licenses"`
	Check       `command:"check" alias:"chk" description:"Check licenses against config file"`
	VersionFlag func() error `long:"version" short:"v" description:"Show CLI version"`

	Quiet func() error `short:"q" long:"quiet" description:"quiet mode, do not log accepted packages"`
}

type List struct {
}

type Check struct {
	File string `short:"f" long:"file" description:"input file" default:".wwhrd.yml"`
}

const VersionHelp flags.ErrorType = 1961

var (
	version = "dev"
	commit  = "1961213"
	date    = "1961-02-13T20:06:35Z"
)

func setQuiet() error {
	log.SetLevel(log.ErrorLevel)
	return nil
}

func newCli() *flags.Parser {
	opts := cliOpts{
		VersionFlag: func() error {
			return &flags.Error{
				Type:    VersionHelp,
				Message: fmt.Sprintf("version %s\ncommit %s\ndate %s\n", version, commit, date),
			}
		},
		Quiet: setQuiet,
	}
	parser := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	parser.LongDescription = "What would Henry Rollins do?"

	return parser
}

func (l *List) Execute(args []string) error {

	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	root, err := rootDir()
	if err != nil {
		return err
	}

	pkgs, err := WalkImports(root)
	if err != nil {
		return err
	}
	lics := GetLicenses(root, pkgs)

	for k, v := range lics {
		if v.Recognized() {
			log.WithFields(log.Fields{
				"package": k,
				"license": v.Type,
			}).Info("Found License")
		} else {
			log.WithFields(log.Fields{
				"package": k,
			}).Warning("Did not find recognized license!")
		}
	}

	return nil
}

func (c *Check) Execute(args []string) error {

	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	t, err := ReadConfig(c.File)
	if err != nil {
		err = fmt.Errorf("Can't read config file: %s", err)
		return err
	}

	root, err := rootDir()
	if err != nil {
		return err
	}

	pkgs, err := WalkImports(root)
	if err != nil {
		return err
	}
	lics := GetLicenses(root, pkgs)

	// Make a map out of the blacklist
	blacklist := make(map[string]bool)
	for _, v := range t.Blacklist {
		blacklist[v] = true
	}

	// Make a map out of the whitelist
	whitelist := make(map[string]bool)
	for _, v := range t.Whitelist {
		whitelist[v] = true
	}

	// Make a map out of the exceptions list
	exceptions := make(map[string]bool)
	exceptionsWildcard := make(map[string]bool)
	for _, v := range t.Exceptions {
		if strings.HasSuffix(v, "/...") {
			exceptionsWildcard[strings.TrimRight(v, "/...")] = true
		} else {
			exceptions[v] = true
		}
	}

PackageList:
	for pkg, lic := range lics {
		contextLogger := log.WithFields(log.Fields{
			"package": pkg,
			"license": lic.Type,
		})

		// License is whitelisted and not specified in blacklist
		if whitelist[lic.Type] && !blacklist[lic.Type] {
			contextLogger.Info("Found Approved license")
			continue PackageList
		}

		// if we have exceptions wildcards, let's run through them
		if len(exceptionsWildcard) > 0 {
			for wc, _ := range exceptionsWildcard {
				if strings.HasPrefix(pkg, wc) {
					// we have a match
					contextLogger.Warn("Found exceptioned package")
					continue PackageList
				}
			}
		}

		// match single-package exceptions
		if _, exists := exceptions[pkg]; exists {
			contextLogger.Warn("Found exceptioned package")
			continue PackageList
		}

		// no matches, it's a non-approved license
		contextLogger.Error("Found Non-Approved license")
		err = fmt.Errorf("Non-Approved license found")

	}

	return err
}

func rootDir() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	info, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		root, err = os.Readlink(root)
		if err != nil {
			return "", err
		}
	}
	return root, nil
}
