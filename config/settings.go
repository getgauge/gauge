package config

import (
	"encoding/json"

	"github.com/sourcegraph/jsonrpc2"
)

var currentSettings GaugeSettings

type FormatConfig struct {
	SkipEmptyLineInsertions bool `json:"skipEmptyLineInsertions"`
}

type GaugeSettings struct {
	Format FormatConfig `json:"formatting"`
}

type Settings struct {
	Gauge GaugeSettings `json:"gauge"`
}

type DidChangeConfigurationParams struct {
	Settings Settings `json:"settings"`
}

func UpdateSettings(request *jsonrpc2.Request) error {
	var params DidChangeConfigurationParams
	if err := json.Unmarshal(*request.Params, &params); err != nil {
		return err
	}
	SetGaugeSettings(params.Settings.Gauge)
	return nil
}

func SetGaugeSettings(gs GaugeSettings) {
	currentSettings = gs
}

func CurrentGaugeSettings() GaugeSettings {
	return currentSettings
}

func SetSkipEmptyLineInsertions(val bool) {
	currentSettings.Format.SkipEmptyLineInsertions = val
}
