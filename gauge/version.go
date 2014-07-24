package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type version struct {
	major int
	minor int
	patch int
}

func parseVersion(versionText string) (*version, error) {
	splits := strings.Split(versionText, DOT)
	if len(splits) != 3 {
		return nil, errors.New("Incorrect number of '.' characters in Version. Version should be of the form 1.5.7")
	}
	major, err := strconv.Atoi(splits[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing major version number %s to integer. %s", splits[0], err.Error()))
	}
	minor, err := strconv.Atoi(splits[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing minor version number %s to integer. %s", splits[0], err.Error()))
	}
	patch, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing patch version number %s to integer. %s", splits[0], err.Error()))
	}

	return &version{major, minor, patch}, nil
}

const (
	MAJOR_VERSION = 0
	MINOR_VERSION = 0
	PATCH_VERSION = 1
	DOT           = "."
)
