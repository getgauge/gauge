// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package skel

var DefaultProperties = `# default.properties
# properties set here will be available to the test execution as environment variables

# sample_key = sample_value

# The path to the gauge reports directory. Should be either relative to the project directory or an absolute path
gauge_reports_dir = reports

# Set as false if gauge reports should not be overwritten on each execution. A new time-stamped directory will be created on each execution.
overwrite_reports = true

# Set to false to disable screenshots on failure in reports.
screenshot_on_failure = true

# The path to the gauge logs directory. Should be either relative to the project directory or an absolute path
logs_directory = logs

# Set to true to use multithreading for parallel execution
enable_multithreading = false

# Possible values for this property are 'suite', 'spec' or 'scenario'.
# 'scenario' clears the objects after the execution of each scenario, new objects are created for next execution.
gauge_clear_state_level = scenario

# The path the gauge specifications directory. Takes a comma separated list of specification files/directories.
gauge_specs_dir = specs

# The default delimiter used read csv files.
csv_delimiter = ,
`
var ExampleSpec = `# Specification Heading

This is an executable specification file. This file follows markdown syntax.
Every heading in this file denotes a scenario. Every bulleted point denotes a step.

To execute this specification, run

    gauge specs


* Vowels in English language are "aeiou".

## Vowel counts in single word

tags: single word

* The word "gauge" has "3" vowels.


## Vowel counts in multiple word

This is the second scenario in this specification

Here's a step that takes a table

* Almost all words have vowels
     |Word  |Vowel Count|
     |------|-----------|
     |Gauge |3          |
     |Mingle|2          |
     |Snap  |1          |
     |GoCD  |1          |
     |Rhythm|0          |
`
var Notice = `
| Dependency Name | Copyright Information | Description |	Repo URL | License Type	| License URL |	Forked from |
|-----------------|-----------------------|-------------|----------|--------------|-------------|-------------|
|Goproperties|Copyright (c) 2012 The Goproperties Authors.|Simple library for reading .properties (java properties) files for Go	|github.com/dmotylev/goproperties	|BSD Styled|	https://raw.githubusercontent.com/dmotylev/goproperties/master/LICENSE|
|Gauge Common|	Copyright 2015 ThoughtWorks, Inc|	|	github.com/getgauge/common|	GPLv3	|||
|mflag	|Copyright (c) 2014 The Docker & Go Authors.	|mflag (aka multiple-flag) implements command-line flag parsing.It's a hacky fork of the official golang package(http://golang.org/pkg/flag/)||BSD Styled	|https://raw.githubusercontent.com/getgauge/mflag/master/LICENSE	|https://github.com/docker/docker/tree/master/pkg/mflag|
|terminal|	Copyright (c) 2013 Meng Zhang	|Colorful terminal output for Golang|github.com/wsxiaoys/terminal	|BSD Styled|https://raw.githubusercontent.com/wsxiaoys/terminal/master/LICENSE||
|gocheck| Copyright (c) 2010-2013 Gustavo Niemeyer <gustavo@niemeyer.net>	|Rich testing for the Go language	|gopkg.in/check.v1	|Simplified BSD	|https://raw.githubusercontent.com/go-check/check/v1/LICENSE||
|protobuf	|Copyright 2010 The Go Authors.	|Go support for Google's protocol buffers	|https://github.com/golang/protobuf	|BSD Styled	|https://raw.githubusercontent.com/golang/protobuf/master/LICENSE|
|grpc-go|Copyright 2014 gRPC authors|The Go language implementation of gRPC. HTTP/2 based RPC|https://google.golang.org/grpc|Apache 2.0|http://www.apache.org/licenses/LICENSE-2.0||
|go-logging|Copyright (c) 2013 Örjan Persson.|Golang logging library|https://github.com/op/go-logging|BSD 3-clause|https://github.com/op/go-logging/blob/master/LICENSE||
|lumberjack|Copyright (c) 2014 Nate Finch |lumberjack is a rolling logger for Go|github.com/natefinch/lumberjack|MIT|https://github.com/natefinch/lumberjack/blob/v2.0/LICENSE||
|go-ogle-analytics|Copyright © 2015 dev@jpillora.com, Google Inc.|Monitor your Go (golang) servers with Google Analytics|github.com/jpillora/go-ogle-analytics|MIT|https://github.com/jpillora/go-ogle-analytics/blob/master/LICENSE||
|fsnotify|Copyright (c) 2012 The Go Authors. All rights reserved.Copyright (c) 2012 fsnotify Authors. All rights reserved.|Cross-platform file system notifications for Go. https://fsnotify.org|github.com/fsnotify/fsnotify|BSD 3-clause|https://github.com/fsnotify/fsnotify/blob/master/LICENSE||
|go-colortext|Copyright (c) 2016, David Deng|Change the color of console text.|github.com/daviddengcn/go-colortext|BSD 3-clause|https://github.com/daviddengcn/go-colortext/blob/master/LICENSE||
|goterminal|Copyright (c) 2015 Apoorva M|A cross-platform Go-library for updating progress in terminal.|github.com/apoorvam/goterminal|MIT|https://github.com/apoorvam/goterminal/blob/master/LICENSE||
|go.uuid|Copyright (C) 2013-2016 by Maxim Bublis <b@codemonkey.ru>|UUID package for Go|github.com/satori/go.uuid|MIT|https://github.com/satori/go.uuid/blob/master/LICENSE||
|go-langserver|Copyright (c) 2016 Sourcegraph|Go language server to add Go support to editors and other tools that use the Language Server Protocol (LSP)|github.com/sourcegraph/go-langserver|MIT|https://github.com/sourcegraph/go-langserver/blob/master/LICENSE||
|jsonrpc2|Copyright (c) 2016 Sourcegraph Inc|Package jsonrpc2 provides a client and server implementation of JSON-RPC 2.0|github.com/sourcegraph/jsonrpc2|MIT|https://github.com/sourcegraph/jsonrpc2/blob/master/LICENSE||
|cobra|Copyright © 2015 Steve Francia <spf@spf13.com>.|A Commander for modern Go CLI interactions|github.com/spf13/cobra|Apache 2.0|https://github.com/spf13/cobra/blob/master/LICENSE.txt||
|pflag|Copyright (c) 2012 Alex Ogier. All rights reserved.Copyright (c) 2012 The Go Authors. All rights reserved.|Drop-in replacement for Go's flag package, implementing POSIX/GNU-style --flags.|github.com/spf13/pflag|BSD 3-clause|https://github.com/spf13/pflag/blob/master/LICENSE|https://github.com/ogier/pflag|
`

var Gitignore = `# Gauge - metadata dir
.gauge

# Gauge - log files dir
logs

# Gauge - reports generated by reporting plugins
reports

`
