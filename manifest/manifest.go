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

package manifest

import (
	"encoding/json"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Manifest struct {
	Language string
	Plugins  []string
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
	return ioutil.WriteFile(common.ManifestFile, b, common.NewFilePermissions)
}
