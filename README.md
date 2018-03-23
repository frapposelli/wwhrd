# WWHRD? (What Would Henry Rollins Do?) [![Build Status](https://travis-ci.org/frapposelli/wwhrd.svg)](https://travis-ci.org/frapposelli/wwhrd) [![codecov](https://codecov.io/gh/frapposelli/wwhrd/branch/master/graph/badge.svg)](https://codecov.io/gh/frapposelli/wwhrd)

![WWHRD?](http://frapposelli.github.io/wwhrd/img/wwhrd.png)

Have Henry Rollins check vendored licenses in your Go project.

Please note that `wwhrd` **only checks** packages stored under `vendor/`.

## Installation

```console
$ go get -u github.com/frapposelli/wwhrd
```

Using [Brew](https://brew.sh) on macOS:

```console
$ brew install frapposelli/tap/wwhrd
```

## Configuration file

Configuration for `wwhrd` is stored in `.wwhrd.yml` at the root of the repo you want to check.

The format is borrowed from [Anderson](https://github.com/xoebus/anderson) and it's 1:1 compatible (just run `wwhrd check -f .anderson.yml`).

```yaml
---
blacklist:
  - GPL-2.0

whitelist:
  - Apache-2.0
  - MIT

exceptions:
  - github.com/jessevdk/go-flags
  - github.com/pmezard/go-difflib/difflib
```

Having a license in the `blacklist` section will fail the check, unless the package is listed under `exceptions`.

`exceptions` can also be listed as wildcards:

```yaml
exceptions:
  - github.com/davecgh/go-spew/spew/...
```

Will make a blanket exception for all the packages under `github.com/davecgh/go-spew/spew`.

Use it in your CI!

```console
$ wwhrd check
INFO[0000] Found Approved license                        license=MIT package=github.com/stretchr/testify/assert
ERRO[0000] Found Non-Approved license                    license=FreeBSD package=github.com/pmezard/go-difflib/difflib
INFO[0000] Found Approved license                        license=MIT package=github.com/ryanuber/go-license
INFO[0000] Found Approved license                        license=Apache-2.0 package=github.com/cloudfoundry-incubator/candiedyaml
WARN[0000] Found exceptioned package                     license=NewBSD package=github.com/jessevdk/go-flags
FATA[0000] Exiting: Non-Approved license found
$ echo $?
1
```

## Usage

```console
$ wwhrd
Usage:
  wwhrd [OPTIONS] <check | list>

What would Henry Rollins do?

Help Options:
  -h, --help  Show this help message

Available commands:
  check  Check licenses against config file (aliases: chk)
  list   List licenses (aliases: ls)
```

## Acknowledgments

WWHRD? graphic by [Mitch Clem](http://mitchclem.tumblr.com/), used with permission, [support him!](https://store.silversprocket.net/collections/mitchclem).
