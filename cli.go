package main

import (
	"fmt"
	"os"

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
	log.SetLevel(log.WarnLevel)
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

	root, err := os.Getwd()
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

	root, err := os.Getwd()
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
	for _, v := range t.Exceptions {
		exceptions[v] = true
	}

	for pkg, lic := range lics {
		contextLogger := log.WithFields(log.Fields{
			"package": pkg,
			"license": lic.Type,
		})

		switch {
		case whitelist[lic.Type] && !blacklist[lic.Type]:
			contextLogger.Info("Found Approved license")
		case exceptions[pkg]:
			contextLogger.Warn("Found exceptioned package")
		default:
			contextLogger.Error("Found Non-Approved license")
			err = fmt.Errorf("Non-Approved license found")
		}
	}

	return err
}
