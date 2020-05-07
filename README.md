hep
===

[![GitHub release](https://img.shields.io/github/release/go-hep/hep.svg)](https://github.com/go-hep/hep/releases)
[![Build Status](https://travis-ci.org/go-hep/hep.svg?branch=master)](https://travis-ci.org/go-hep/hep)
[![Build status](https://ci.appveyor.com/api/projects/status/qnnp26vv2c71f560?svg=true)](https://ci.appveyor.com/project/sbinet/hep)
[![CI](https://github.com/go-hep/hep/workflows/CI/badge.svg)](https://github.com/go-hep/hep/actions)
[![codecov](https://codecov.io/gh/go-hep/hep/branch/master/graph/badge.svg)](https://codecov.io/gh/go-hep/hep)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-hep/hep)](https://goreportcard.com/report/github.com/go-hep/hep)
[![GoDoc](https://godoc.org/go-hep.org/x/hep?status.svg)](https://godoc.org/go-hep.org/x/hep)
[![License](https://img.shields.io/badge/License-BSD--3-blue.svg)](https://go-hep.org/license)
[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.597940.svg)](https://doi.org/10.5281/zenodo.597940)
[![JOSS Paper](http://joss.theoj.org/papers/0b007c81073186f7c61f95ea26ad7971/status.svg)](http://joss.theoj.org/papers/0b007c81073186f7c61f95ea26ad7971)
[![Binder](https://mybinder.org/badge.svg)](https://mybinder.org/v2/gh/go-hep/binder/master)
[![Twitter Follow](https://img.shields.io/twitter/follow/go_hep.svg?style=social&label=Follow)](https://twitter.com/go_hep)
[![TODOs](https://badgen.net/https/api.tickgit.com/badgen/github.com/go-hep/hep)](https://www.tickgit.com/browse?repo=github.com/go-hep/hep)

`hep` is a set of libraries and tools to perform High Energy Physics analyses with ease and [Go](https://golang.org)

See [go-hep.org](https://go-hep.org) for more informations.




## License

`hep` is released under the `BSD-3` license.

## Documentation

Documentation for `hep` is served by [GoDoc](https://godoc.org/go-hep.org/x/hep).

## Contributing

Guidelines for contributing to [go-hep](https://go-hep.org) are available here:
 [go-hep.org/contributing](https://go-hep.org/contributing)
 
### Contributors

This project exists thanks to all the people who contribute. 

# Motivations

Writing analyses in HEP involves many steps and one needs a few tools to
successfully carry out such an endeavour.
But - at minima - one needs to be able to read (and possibly write) ROOT files
to be able to interoperate with the rest of the HEP community or to insert
one's work into an already existing analysis pipeline.

Go-HEP provides this necessary interoperability layer, in the Go programming
language.
This allows physicists to leverage the great concurrency primitives of Go,
together with the surrounding tooling and software engineering ecosystem of Go,
to implement physics analyses.

## Content

Go-HEP currently sports the following packages:

- [go-hep.org/x/hep/brio](https://go-hep.org/x/hep/brio): a toolkit to generate serialization code
- [go-hep.org/x/hep/fads](https://go-hep.org/x/hep/fads): a fast detector simulation toolkit
- [go-hep.org/x/hep/fastjet](https://go-hep.org/x/hep/fastjet): a jet clustering algorithms package (WIP)
- [go-hep.org/x/hep/fit](https://go-hep.org/x/hep/fit): a fitting function toolkit (WIP)
- [go-hep.org/x/hep/fmom](https://go-hep.org/x/hep/fmom): a 4-vectors library
- [go-hep.org/x/hep/fwk](https://go-hep.org/x/hep/fwk): a concurrency-enabled framework
- [go-hep.org/x/hep/groot](https://go-hep.org/x/hep/groot): a pure [Go](https://golang.org) package for [ROOT](https://root.cern.ch) I/O (WIP)
- [go-hep.org/x/hep/hbook](https://go-hep.org/x/hep/hbook): histograms and n-tuples (WIP)
- [go-hep.org/x/hep/hplot](https://go-hep.org/x/hep/hplot): interactive plotting (WIP)
- [go-hep.org/x/hep/hepmc](https://go-hep.org/x/hep/hepmc): `HepMC` in pure [Go](https://golang.org) (EDM + I/O)
- [go-hep.org/x/hep/hepevt](https://go-hep.org/x/hep/hepevt): `HEPEVT` bindings
- [go-hep.org/x/hep/heppdt](https://go-hep.org/x/hep/heppdt): `HEP` particle data table
- [go-hep.org/x/hep/lcio](https://go-hep.org/x/hep/lcio): read/write support for `LCIO` event data model
- [go-hep.org/x/hep/lhef](https://go-hep.org/x/hep/lhef): Les Houches Event File format
- [go-hep.org/x/hep/rio](https://go-hep.org/x/hep/rio): `go-hep` record oriented I/O
- [go-hep.org/x/hep/sio](https://go-hep.org/x/hep/sio): basic, low-level, serial I/O used by `LCIO`
- [go-hep.org/x/hep/slha](https://go-hep.org/x/hep/slha): `SUSY` Les Houches Accord I/O
- [go-hep.org/x/hep/xrootd](https://go-hep.org/x/hep/xrootd): [XRootD](http://xrootd.org) client in pure [Go](https://golang.org)

## Installation

Go-HEP packages are installable via the `go get` command:

```sh
$ go get go-hep.org/x/hep/fads
```

Just select the package you are interested in and `go get` will take care of fetching, building and installing it, as well as its dependencies, recursively.
