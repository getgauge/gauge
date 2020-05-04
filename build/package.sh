#!/bin/sh

# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

# Usage:
# ./build/package.sh [--nightly]

go run build/make.go --all-platforms $1

chmod +x bin/**/* && rm -rf deploy

go run build/make.go --distro --all-platforms --skip-windows $1
