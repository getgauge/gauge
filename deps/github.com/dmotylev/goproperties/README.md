# Goproperties [![Build Status](https://travis-ci.org/dmotylev/goproperties.png)](https://travis-ci.org/dmotylev/goproperties)

Package implements read operations of **[.properties](http://en.wikipedia.org/wiki/.properties)** source.


# Documentation

The Goproperties API reference is available on [GoDoc](http://godoc.org/github.com/dmotylev/goproperties).

# Installation

Install Goproperties using the `go get` command:

	go get -u github.com/dmotylev/goproperties

# Usage

Example:

```go
package main

import "github.com/dmotylev/goproperties"

func main() {
	p, _ := properties.Load("credentials")
	username := p.String("username","demo")
	password := p.String("password","demo")

	// ... use given credentials

	_, _ = username, password
}
```

Look at [properties_test.go](https://github.com/dmotylev/goproperties/blob/master/properties_test.go) for more usage hints.


# Dependencies

## Main functionality

The _main_ part depends on the Go distribution only.


## Test functionality

The _test_ part depends on **[gocheck](http://labix.org/gocheck)** package.

Use following `go get` command to install _test_ dependencies:

	go get -u launchpad.net/gocheck


# License

For the license see [LICENSE](https://github.com/dmotylev/goproperties/blob/master/LICENSE).
