# geodbtools

[![Go Report Card](https://goreportcard.com/badge/anexia-it/geodbtools)](https://goreportcard.com/report/anexia-it/geodbtools)
[![Build Status](https://travis-ci.org/anexia-it/geodbtools.svg?branch=master)](https://travis-ci.org/anexia-it/geodbtools)
[![codecov](https://codecov.io/gh/anexia-it/geodbtools/branch/master/graph/badge.svg)](https://codecov.io/gh/anexia-it/geodbtools)
[![GoDoc](https://godoc.org/github.com/anexia-it/geodbtools?status.svg)](https://godoc.org/github.com/anexia-it/geodbtools)


`geodbtools` is a swiss army knife for working with GeoIP databases.

## geodbtool CLI

`geodbtool` is the CLI application that is provided along with the `geodbtools` library.

### Features

* database lookups (`lookup` command)
* database information (`info` command)
* database type conversion (`convert` command)

### Installation

You can either grab pre-compiled binaries from the [releases](https://github.com/anexia-it/geodbtools/releases) page on Github, or install from source:

```
go get github.com/anexia-it/geodbtools/cmd/geodbtool
```

### Usage

The tool's help command provides you with a list of available commands:

```
geodbtool help
```

## geodbtools library

The library part of this repository provides functionality for working with
GeoIP databases.

### Implementation status

- [ ] MaxMind DAT format support
  - [x] Country databases
    - [x] Read
	- [x] Write
  - [ ] City databases
  - [ ] AS number databases

- [ ] MaxMind MMDB format support
  - [ ] Country databases
    - [x] Read (via https://github.com/oschwald/maxminddb-golang)
    - [ ] Write
  - [ ] City databases
  - [ ] AS number databases
  
- [ ] MaxMind legacy CSV format support
- [ ] MaxMind GeoIP2 CSV format support

## Contributing

Contributions are always welcome, in every possible way.
We are always happy to receive bug reports and pull requests.

## License

`geodbtools` is free software, licensed under the MIT license.

