/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
	"google.golang.org/grpc"
)

type mockResultClient struct {
	invoked bool
}

func (client *mockResultClient) NotifySuiteResult(c context.Context, r *gm.SuiteExecutionResult, opts ...grpc.CallOption) (*gm.Empty, error) {
	client.invoked = true
	return nil, nil
}

func (client *mockResultClient) NotifyExecutionStarting(c context.Context, r *gm.ExecutionStartingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyExecutionEnding(c context.Context, r *gm.ExecutionEndingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifySpecExecutionStarting(c context.Context, r *gm.SpecExecutionStartingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifySpecExecutionEnding(c context.Context, r *gm.SpecExecutionEndingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyScenarioExecutionStarting(c context.Context, r *gm.ScenarioExecutionStartingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyScenarioExecutionEnding(c context.Context, r *gm.ScenarioExecutionEndingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyStepExecutionStarting(c context.Context, r *gm.StepExecutionStartingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyStepExecutionEnding(c context.Context, r *gm.StepExecutionEndingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyConceptExecutionStarting(c context.Context, r *gm.ConceptExecutionStartingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) NotifyConceptExecutionEnding(c context.Context, r *gm.ConceptExecutionEndingRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func (client *mockResultClient) Kill(c context.Context, r *gm.KillProcessRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func TestGetPluginDescriptorFromJSON(t *testing.T) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	pd, err := GetPluginDescriptorFromJSON(filepath.Join(path, "_test.json"))

	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	t.Run("ID", func(t *testing.T) {
		if pd.ID != "html-report" {
			t.Errorf("expected %s, got %s", "html-report", pd.ID)
		}
	})
	t.Run("Version", func(t *testing.T) {
		if pd.Version != "1.1.0" {
			t.Errorf("expected %s, got %s", "1.1.0", pd.Version)
		}
	})
	t.Run("Name", func(t *testing.T) {
		if pd.Name != "Html Report" {
			t.Errorf("expected %s, got %s", "Html Report", pd.Name)
		}
	})
	t.Run("Description", func(t *testing.T) {
		if pd.Description != "Html reporting plugin" {
			t.Errorf("expected %s, got %s", "Html reporting plugin", pd.Name)
		}
	})
	t.Run("Path", func(t *testing.T) {
		if pd.pluginPath != path {
			t.Errorf("expected %s, got %s", path, pd.pluginPath)
		}
	})
	t.Run("Min Version", func(t *testing.T) {
		if pd.GaugeVersionSupport.Minimum != "0.2.0" {
			t.Errorf("expected %s, got %s", "0.2.0", pd.GaugeVersionSupport.Minimum)
		}
	})
	t.Run("Max Version", func(t *testing.T) {
		if pd.GaugeVersionSupport.Maximum != "0.4.0" {
			t.Errorf("expected %s, got %s", "0.4.0", pd.GaugeVersionSupport.Maximum)
		}
	})
	t.Run("Scope", func(t *testing.T) {
		if !reflect.DeepEqual(pd.Scope, []string{"Execution"}) {
			t.Errorf("expected %s, got %s", []string{"Execution"}, pd.Scope)
		}
	})
	htmlCommand := []string{"bin/html-report"}
	t.Run("Windows Command", func(t *testing.T) {
		if !reflect.DeepEqual(pd.Command.Windows, htmlCommand) {
			t.Errorf("expected %s, got %s", htmlCommand, pd.Command.Windows)
		}
	})
	t.Run("Darwin Command", func(t *testing.T) {
		if !reflect.DeepEqual(pd.Command.Darwin, htmlCommand) {
			t.Errorf("expected %s, got %s", htmlCommand, pd.Command.Darwin)
		}
	})
	t.Run("Linux Command", func(t *testing.T) {
		if !reflect.DeepEqual(pd.Command.Linux, htmlCommand) {
			t.Errorf("expected %s, got %s", htmlCommand, pd.Command.Linux)
		}
	})
}

func TestGetPluginDescriptorFromNonExistingJSON(t *testing.T) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)
	JSONPath := filepath.Join(path, "_test1.json")
	_, err := GetPluginDescriptorFromJSON(JSONPath)

	expected := fmt.Errorf("File %s doesn't exist.", JSONPath)
	if err.Error() != expected.Error() {
		t.Errorf("expected %s, got %s", expected, err)
	}
}

func TestGetLanguageQueryParamWhenProjectRootNotSet(t *testing.T) {
	config.ProjectRoot = ""

	l := language()

	if l != "" {
		t.Errorf("expected empty language, got %s", l)
	}
}

func TestGetLanguageQueryParam(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "sample"))
	config.ProjectRoot = path

	l := language()

	if l != "java" {
		t.Errorf("expected java, got %s", l)
	}
}

func TestGetPluginsWithoutScope(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join("_testdata"))
	_ = os.Setenv(common.GaugeHome, path)

	got := PluginsWithoutScope()

	want := []pluginInfo.PluginInfo{
		{
			Name:    "noscope",
			Version: &version.Version{Major: 1, Minor: 0, Patch: 0},
			Path:    filepath.Join(path, "plugins", "noscope", "1.0.0"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Failed GetPluginWithoutScope.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestSendMessageShouldUseGRPCConnectionIfAvailable(t *testing.T) {
	c := &mockResultClient{}
	p := &plugin{
		gRPCConn:       &grpc.ClientConn{},
		ReporterClient: c,
	}

	e := p.sendMessage(&gm.Message{MessageType: gm.Message_SuiteExecutionResult, SuiteExecutionResult: &gm.SuiteExecutionResult{}})

	if e != nil {
		t.Errorf("Expected error to be nil. Got : %v", e)
	}

	if !c.invoked {
		t.Errorf("Expected grpc client to be invoked")
	}
}
