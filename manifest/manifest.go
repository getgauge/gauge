/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
)

type Manifest struct {
	Language       string
	Plugins        []string
	EnvironmentDir string
}

func ProjectManifest() (*Manifest, error) {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.ManifestFile))
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(strings.NewReader(contents))

	var m Manifest
	for {
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("Failed to read Manifest. %s\n", err.Error())
		}
	}

	return &m, nil
}

func (m *Manifest) Save() error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(common.ManifestFile, b, common.NewFilePermissions)
}
