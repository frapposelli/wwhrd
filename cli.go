package main

import (
	"fmt"
	"os"

	manifest "github.com/FiloSottile/gvt/gbvendor"
	log "github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"

	"strings"

	"encoding/json"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
)

type cliOpts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	List  `command:"list" alias:"ls" description:"List licenses"`
	Check `command:"check" alias:"chk" description:"Check licenses against config file"`
}

type List struct {
	YAML     bool `long:"yaml" description:"outputs the licenses in yaml format"`
	Versions bool `long:"versions" description:"outputs the git revisions in yaml format"`
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

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	pkgs, err := WalkImports(root)
	if err != nil {
		return err
	}
	lics := GetLicenses(root, pkgs)

	var j manifest.Manifest

	if l.Versions {

		f, err := os.Open("vendor/manifest")
		if err != nil {
			return err
		}

		d := json.NewDecoder(f)
		err = d.Decode(&j)
		if err != nil {
			return err
		}

		f.Close()
	}

	for k, v := range lics {
		if l.YAML {

			// We rudely remove the first part of the package name, guessing that it's going to be the hostname
			name := strings.SplitAfterN(k, "/", 2)

			// Replace / and . with underscores in package name and lower the case
			r := strings.NewReplacer("/", "_", ".", "_")
			n := r.Replace(name[1])
			n = strings.ToLower(n)

			var ver string

			if l.Versions {
				d, err := j.GetDependencyForImportpath(k)
				if err != nil {
					ver = "git+unspecified"
				} else {
					ver = "git+" + d.Revision[:12]
				}
			} else {
				ver = "latest"
			}

			// build OSSTP package name
			p := "other:" + n + ":" + ver

			// build maps to marshal yaml
			y := make(map[string]map[string]string)
			y[p] = make(map[string]string)

			y[p]["name"] = fmt.Sprintf("%s", n)
			y[p]["license"] = fmt.Sprintf("%s", v.Type)
			y[p]["repository"] = fmt.Sprintf("%s", "Other")
			y[p]["url"] = fmt.Sprintf("%s", "http://"+k)
			y[p]["version"] = fmt.Sprintf("%s", ver)

			o, err := yaml.Marshal(y)
			if err != nil {
				return err
			}

			// spit it out
			fmt.Printf("%v", string(o))

		} else {
			log.WithFields(log.Fields{
				"package": k,
				"license": v.Type,
			}).Info("Found License")
		}
	}

	return nil
}

func (c *Check) Execute(args []string) error {

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

	// Make a map out of the blacklist
	whitelist := make(map[string]bool)
	for _, v := range t.Whitelist {
		whitelist[v] = true
	}

	// Make a map out of the blacklist
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
