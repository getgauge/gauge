// Copyright 2019 ThoughtWorks, Inc.

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import "errors"

func RLimit() (int, error) {
	// TODO: Run some system commands to figure out the max no. of open file descriptors.
	return -1, errors.New("Not implemented for windows")
}
