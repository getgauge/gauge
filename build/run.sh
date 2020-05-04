# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

option="${1}"
case ${option} in
    test)
        go test ./... -v
        ;;
    build|"")
        go run build/make.go
        ;;
    *)
        echo "`basename ${0}`:usage: [build|test]"
        exit 1
        ;;
esac
