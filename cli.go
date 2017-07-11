package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

type cliOpts struct {
	List  `command:"list" alias:"ls" description:"List licenses"`
	Check `command:"check" alias:"chk" description:"Check licenses against config file"`
}

type List struct {
}

type Check struct {
	File string `short:"f" long:"file" description:"input file" default:".wwhrd.yml"`
}

func newCli() *flags.Parser {
	var opts cliOpts
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

		log.WithFields(log.Fields{
			"package": k,
			"license": v.Type,
		}).Info("Found License")
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

		switch {
		case whitelist[lic.Type] && !blacklist[lic.Type]:
			log.WithFields(log.Fields{
				"package": pkg,
				"license": lic.Type,
			}).Info("Found Approved license")
		case exceptions[pkg]:
			log.WithFields(log.Fields{
				"package": pkg,
				"license": lic.Type,
			}).Warn("Found exceptioned package")
		default:
			log.WithFields(log.Fields{
				"package": pkg,
				"license": lic.Type,
			}).Error("Found Non-Approved license")
			err = fmt.Errorf("Non-Approved license found")
		}
	}

	return err
}
