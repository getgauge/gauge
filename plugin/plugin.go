/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type pluginScope string

const (
	executionScope          pluginScope = "execution"
	docScope                pluginScope = "documentation"
	pluginConnectionPortEnv             = "plugin_connection_port"
)

type plugin struct {
	mutex            *sync.Mutex
	connection       net.Conn
	gRPCConn         *grpc.ClientConn
	ReporterClient   gauge_messages.ReporterClient
	DocumenterClient gauge_messages.DocumenterClient
	pluginCmd        *exec.Cmd
	descriptor       *PluginDescriptor
	killTimer        *time.Timer
}

func isProcessRunning(p *plugin) bool {
	p.mutex.Lock()
	ps := p.pluginCmd.ProcessState
	p.mutex.Unlock()
	return ps == nil || !ps.Exited()
}

func (p *plugin) killGrpcProcess() error {
	var m *gauge_messages.Empty
	var err error
	if p.ReporterClient != nil {
		m, err = p.ReporterClient.Kill(context.Background(), &gauge_messages.KillProcessRequest{})
	} else if p.DocumenterClient != nil {
		m, err = p.DocumenterClient.Kill(context.Background(), &gauge_messages.KillProcessRequest{})
	}
	if m == nil || err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Unavailable {
			// Ref https://www.grpc.io/docs/guides/error/#general-errors
			// GRPC_STATUS_UNAVAILABLE is thrown when Server is shutting down. Ignore it here.
			return nil
		}
		return err
	}
	if p.gRPCConn == nil && p.pluginCmd == nil {
		return nil
	}
	defer func(gRPCConn *grpc.ClientConn) {
		_ = gRPCConn.Close()
	}(p.gRPCConn)

	if isProcessRunning(p) {
		exited := make(chan bool, 1)
		go func() {
			for {
				if isProcessRunning(p) {
					time.Sleep(100 * time.Millisecond)
				} else {
					exited <- true
					return
				}
			}
		}()

		select {
		case done := <-exited:
			if done {
				logger.Debugf(true, "Runner with PID:%d has exited", p.pluginCmd.Process.Pid)
				return nil
			}
		case <-time.After(config.PluginKillTimeout()):
			logger.Warningf(true, "Killing runner with PID:%d forcefully", p.pluginCmd.Process.Pid)
			return p.pluginCmd.Process.Kill()
		}
	}
	return nil
}

func (p *plugin) kill(wg *sync.WaitGroup) error {
	defer wg.Done()
	if p.gRPCConn != nil && p.ReporterClient != nil {
		return p.killGrpcProcess()
	}
	if isProcessRunning(p) {
		defer func(connection net.Conn) {
			_ = connection.Close()
		}(p.connection)
		p.killTimer = time.NewTimer(config.PluginKillTimeout())
		err := conn.SendProcessKillMessage(p.connection)
		if err != nil {
			logger.Warningf(true, "Error while killing plugin %s : %s ", p.descriptor.Name, err.Error())
		}

		exited := make(chan bool, 1)
		go func() {
			for {
				if isProcessRunning(p) {
					time.Sleep(100 * time.Millisecond)
				} else {
					exited <- true
					return
				}
			}
		}()
		select {
		case <-exited:
			if !p.killTimer.Stop() {
				<-p.killTimer.C
			}
			logger.Debugf(true, "Plugin [%s] with pid [%d] has exited", p.descriptor.Name, p.pluginCmd.Process.Pid)
		case <-p.killTimer.C:
			logger.Warningf(true, "Plugin [%s] with pid [%d] did not exit after %.2f seconds. Forcefully killing it.", p.descriptor.Name, p.pluginCmd.Process.Pid, config.PluginKillTimeout().Seconds())
			err := p.pluginCmd.Process.Kill()
			if err != nil {
				logger.Warningf(true, "Error while killing plugin %s : %s ", p.descriptor.Name, err.Error())
			}
			return err
		}
	}
	return nil
}

// IsPluginInstalled checks if given plugin with specific version is installed or not.
func IsPluginInstalled(pluginName, pluginVersion string) bool {
	pluginsInstallDir, err := common.GetPluginsInstallDir(pluginName)
	if err != nil {
		return false
	}

	thisPluginDir := filepath.Join(pluginsInstallDir, pluginName)
	if !common.DirExists(thisPluginDir) {
		return false
	}

	if pluginVersion != "" {
		return common.FileExists(filepath.Join(thisPluginDir, pluginVersion, common.PluginJSONFile))
	}
	return true
}

func getPluginJSONPath(pluginName, pluginVersion string) (string, error) {
	if !IsPluginInstalled(pluginName, pluginVersion) {
		plugin := strings.TrimSpace(fmt.Sprintf("%s %s", pluginName, pluginVersion))
		return "", fmt.Errorf("Plugin %s is not installed", plugin)
	}

	pluginInstallDir, err := GetInstallDir(pluginName, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(pluginInstallDir, common.PluginJSONFile), nil
}

// GetPluginDescriptor return the information about the plugin including name, id, commands to start etc.
func GetPluginDescriptor(pluginID, pluginVersion string) (*PluginDescriptor, error) {
	pluginJSON, err := getPluginJSONPath(pluginID, pluginVersion)
	if err != nil {
		return nil, err
	}
	return GetPluginDescriptorFromJSON(pluginJSON)
}

func GetPluginDescriptorFromJSON(pluginJSON string) (*PluginDescriptor, error) {
	pluginJSONContents, err := common.ReadFileContents(pluginJSON)
	if err != nil {
		return nil, err
	}
	var pd PluginDescriptor
	if err = json.Unmarshal([]byte(pluginJSONContents), &pd); err != nil {
		return nil, fmt.Errorf("%s: %s", pluginJSON, err.Error())
	}
	pd.pluginPath = filepath.Dir(pluginJSON)

	return &pd, nil
}

func startPlugin(pd *PluginDescriptor, action pluginScope) (*plugin, error) {
	var command []string
	switch runtime.GOOS {
	case "windows":
		command = pd.Command.Windows
	case "darwin":
		command = pd.Command.Darwin
	default:
		command = pd.Command.Linux
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("Platform specific command not specified: %s.", runtime.GOOS)
	}
	if pd.hasCapability(gRPCSupportCapability) {
		return startGRPCPlugin(pd, command)
	}
	return startLegacyPlugin(pd, command)
}

func startGRPCPlugin(pd *PluginDescriptor, command []string) (*plugin, error) {
	portChan := make(chan string)
	writer := &logger.LogWriter{
		Stderr: logger.NewCustomWriter(portChan, os.Stderr, pd.ID, true),
		Stdout: logger.NewCustomWriter(portChan, os.Stdout, pd.ID, false),
	}
	cmd, err := common.ExecuteCommand(command, pd.pluginPath, writer.Stdout, writer.Stderr)
	go func() {
		err = cmd.Wait()
		if err != nil {
			logger.Errorf(true, "Error occurred while waiting for plugin process to finish.\nError : %s", err.Error())
		}
	}()
	if err != nil {
		return nil, err
	}

	var port string
	select {
	case port = <-portChan:
		close(portChan)
	case <-time.After(config.PluginConnectionTimeout()):
		return nil, fmt.Errorf("timed out connecting to %s", pd.ID)
	}
	logger.Debugf(true, "Attempting to connect to grpc server at port: %s", port)
	gRPCConn, err := grpc.NewClient(fmt.Sprintf("%s:%s", "127.0.0.1", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024*1024*1024), grpc.MaxCallRecvMsgSize(1024*1024*1024)))
	if err != nil {
		return nil, err
	}
	plugin := &plugin{
		pluginCmd:  cmd,
		descriptor: pd,
		gRPCConn:   gRPCConn,
		mutex:      &sync.Mutex{},
	}
	if pd.hasScope(docScope) {
		plugin.DocumenterClient = gauge_messages.NewDocumenterClient(gRPCConn)
	} else {
		plugin.ReporterClient = gauge_messages.NewReporterClient(gRPCConn)
	}

	logger.Debugf(true, "Successfully made the connection with plugin with port: %s", port)
	return plugin, nil
}

func startLegacyPlugin(pd *PluginDescriptor, command []string) (*plugin, error) {
	writer := logger.NewLogWriter(pd.ID, true, 0)
	cmd, err := common.ExecuteCommand(command, pd.pluginPath, writer.Stdout, writer.Stderr)

	if err != nil {
		return nil, err
	}
	var mutex = &sync.Mutex{}
	go func() {
		pState, _ := cmd.Process.Wait()
		mutex.Lock()
		cmd.ProcessState = pState
		mutex.Unlock()
	}()
	plugin := &plugin{pluginCmd: cmd, descriptor: pd, mutex: mutex}
	return plugin, nil
}

func SetEnvForPlugin(action pluginScope, pd *PluginDescriptor, m *manifest.Manifest, pluginEnvVars map[string]string) error {
	pluginEnvVars[fmt.Sprintf("%s_action", pd.ID)] = string(action)
	pluginEnvVars["test_language"] = m.Language
	return setEnvironmentProperties(pluginEnvVars)
}

func setEnvironmentProperties(properties map[string]string) error {
	for k, v := range properties {
		if err := common.SetEnvVariable(k, v); err != nil {
			return err
		}
	}
	return nil
}

func IsPluginAdded(m *manifest.Manifest, descriptor *PluginDescriptor) bool {
	for _, pluginID := range m.Plugins {
		if pluginID == descriptor.ID {
			return true
		}
	}
	return false
}

func startPluginsForExecution(m *manifest.Manifest) (Handler, []string) {
	var warnings []string
	handler := &GaugePlugins{}
	envProperties := make(map[string]string)

	for _, pluginID := range m.Plugins {
		pd, err := GetPluginDescriptor(pluginID, "")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Unable to start plugin %s. %s. To install, run `gauge install %s`.", pluginID, err.Error(), pluginID))
			continue
		}
		compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport)
		if compatibilityErr != nil {
			warnings = append(warnings, fmt.Sprintf("Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion))
			continue
		}
		if pd.hasScope(executionScope) {
			gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(0, nil)
			if err != nil {
				warnings = append(warnings, err.Error())
				continue
			}
			envProperties[pluginConnectionPortEnv] = strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber())
			prop, err := common.GetGaugeConfigurationFor(common.GaugePropertiesFile)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Unable to read Gauge configuration. %s", err.Error()))
				continue
			}
			envProperties["plugin_kill_timeout"] = prop["plugin_kill_timeout"]
			err = SetEnvForPlugin(executionScope, pd, m, envProperties)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error setting environment for plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}
			logger.Debugf(true, "Starting %s plugin", pd.Name)
			plugin, err := startPlugin(pd, executionScope)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}
			if plugin.gRPCConn != nil {
				handler.addPlugin(pluginID, plugin)
				continue
			}
			pluginConnection, err := gaugeConnectionHandler.AcceptConnection(config.PluginConnectionTimeout(), make(chan error))
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. Failed to connect to plugin. %s", pd.Name, pd.Version, err.Error()))
				err := plugin.pluginCmd.Process.Kill()
				if err != nil {
					logger.Errorf(false, "unable to kill plugin %s: %s", plugin.descriptor.Name, err.Error())
				}
				continue
			}
			logger.Debugf(true, "Established connection to %s plugin", pd.Name)
			plugin.connection = pluginConnection
			handler.addPlugin(pluginID, plugin)
		}

	}
	return handler, warnings
}

func GenerateDoc(pluginName string, specDirs []string, startAPIFunc func([]string) int) {
	pd, err := GetPluginDescriptor(pluginName, "")
	if err != nil {
		logger.Fatalf(true, "Error starting plugin %s. Failed to get plugin.json. %s. To install, run `gauge install %s`.", pluginName, err.Error(), pluginName)
	}
	if err := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport); err != nil {
		logger.Fatalf(true, "Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion)
	}
	if !pd.hasScope(docScope) {
		logger.Fatalf(true, "Invalid plugin name: %s, this plugin cannot generate documentation.", pd.Name)
	}
	var sources []string
	for _, src := range specDirs {
		path, _ := filepath.Abs(src)
		sources = append(sources, path)
	}
	_ = os.Setenv("GAUGE_SPEC_DIRS", strings.Join(sources, "||"))
	_ = os.Setenv("GAUGE_PROJECT_ROOT", config.ProjectRoot)
	if pd.hasCapability(gRPCSupportCapability) {
		p, err := startPlugin(pd, docScope)
		if err != nil {
			logger.Fatalf(true, " %s %s. %s", pd.Name, pd.Version, err.Error())
		}
		_, err = p.DocumenterClient.GenerateDocs(context.Background(), getSpecDetails(specDirs))
		grpcErr := p.killGrpcProcess()
		if grpcErr != nil {
			logger.Errorf(false, "Unable to kill plugin %s : %s", p.descriptor.Name, grpcErr.Error())
		}
		if err != nil {
			logger.Fatalf(true, "Failed to generate docs. %s", err.Error())
		}
	} else {
		port := startAPIFunc(specDirs)
		err := os.Setenv(common.APIPortEnvVariableName, strconv.Itoa(port))
		if err != nil {
			logger.Fatalf(true, "Failed to set env GAUGE_API_PORT. %s", err.Error())
		}
		p, err := startPlugin(pd, docScope)
		if err != nil {
			logger.Fatalf(true, " %s %s. %s", pd.Name, pd.Version, err.Error())
		}
		for isProcessRunning(p) {
		}
	}
}

func (p *plugin) invokeService(m *gauge_messages.Message) error {
	ctx := context.Background()
	var err error
	switch m.GetMessageType() {
	case gauge_messages.Message_SuiteExecutionResult:
		_, err = p.ReporterClient.NotifySuiteResult(ctx, m.GetSuiteExecutionResult())
	case gauge_messages.Message_ExecutionStarting:
		_, err = p.ReporterClient.NotifyExecutionStarting(ctx, m.GetExecutionStartingRequest())
	case gauge_messages.Message_ExecutionEnding:
		_, err = p.ReporterClient.NotifyExecutionEnding(ctx, m.GetExecutionEndingRequest())
	case gauge_messages.Message_SpecExecutionEnding:
		_, err = p.ReporterClient.NotifySpecExecutionEnding(ctx, m.GetSpecExecutionEndingRequest())
	case gauge_messages.Message_SpecExecutionStarting:
		_, err = p.ReporterClient.NotifySpecExecutionStarting(ctx, m.GetSpecExecutionStartingRequest())
	case gauge_messages.Message_ScenarioExecutionEnding:
		_, err = p.ReporterClient.NotifyScenarioExecutionEnding(ctx, m.GetScenarioExecutionEndingRequest())
	case gauge_messages.Message_ScenarioExecutionStarting:
		_, err = p.ReporterClient.NotifyScenarioExecutionStarting(ctx, m.GetScenarioExecutionStartingRequest())
	case gauge_messages.Message_StepExecutionEnding:
		_, err = p.ReporterClient.NotifyStepExecutionEnding(ctx, m.GetStepExecutionEndingRequest())
	case gauge_messages.Message_StepExecutionStarting:
		_, err = p.ReporterClient.NotifyStepExecutionStarting(ctx, m.GetStepExecutionStartingRequest())
	case gauge_messages.Message_ConceptExecutionEnding:
		_, err = p.ReporterClient.NotifyConceptExecutionEnding(ctx, m.GetConceptExecutionEndingRequest())
	case gauge_messages.Message_ConceptExecutionStarting:
		_, err = p.ReporterClient.NotifyConceptExecutionStarting(ctx, m.GetConceptExecutionStartingRequest())
	}
	return err
}

func (p *plugin) sendMessage(message *gauge_messages.Message) error {
	if p.gRPCConn != nil {
		return p.invokeService(message)
	}
	messageID := common.GetUniqueID()
	message.MessageId = messageID
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	err = conn.Write(p.connection, messageBytes)
	if err != nil {
		return fmt.Errorf("[Warning] Failed to send message to plugin: %s  %s", p.descriptor.ID, err.Error())
	}
	return nil
}

func StartPlugins(m *manifest.Manifest) Handler {
	pluginHandler, warnings := startPluginsForExecution(m)
	logger.HandleWarningMessages(true, warnings)
	return pluginHandler
}

func PluginsWithoutScope() (infos []pluginInfo.PluginInfo) {
	if plugins, err := pluginInfo.GetAllInstalledPluginsWithVersion(); err == nil {
		for _, p := range plugins {
			pd, err := GetPluginDescriptor(p.Name, p.Version.String())
			if err == nil && !pd.hasAnyScope() {
				infos = append(infos, p)
			}
		}
	}
	return
}

// GetInstallDir returns the install directory of given plugin and a given version.
func GetInstallDir(pluginName, v string) (string, error) {
	allPluginsInstallDir, err := common.GetPluginsInstallDir(pluginName)
	if err != nil {
		return "", err
	}
	pluginDir := filepath.Join(allPluginsInstallDir, pluginName)
	if v != "" {
		pluginDir = filepath.Join(pluginDir, v)
	} else {
		latestPlugin, err := pluginInfo.GetLatestInstalledPlugin(pluginDir)
		if err != nil {
			return "", err
		}
		pluginDir = latestPlugin.Path
	}
	return pluginDir, nil
}

func GetLanguageJSONFilePath(language string) (string, error) {
	languageInstallDir, err := GetInstallDir(language, "")
	if err != nil {
		return "", err
	}
	languageJSON := filepath.Join(languageInstallDir, fmt.Sprintf("%s.json", language))
	if !common.FileExists(languageJSON) {
		return "", fmt.Errorf("Failed to find the implementation for: %s. %s does not exist.", language, languageJSON)
	}

	return languageJSON, nil
}

func IsLanguagePlugin(plugin string) bool {
	if _, err := GetLanguageJSONFilePath(plugin); err != nil {
		return false
	}
	return true
}

func QueryParams() string {
	return fmt.Sprintf("?l=%s&p=%s&o=%s&a=%s", language(), plugins(), runtime.GOOS, runtime.GOARCH)
}

func language() string {
	if config.ProjectRoot == "" {
		return ""
	}
	m, err := manifest.ProjectManifest()
	if err != nil {
		return ""
	}
	return m.Language
}

func plugins() string {
	pluginInfos, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		return ""
	}
	var plugins []string
	for _, p := range pluginInfos {
		plugins = append(plugins, p.Name)
	}
	return strings.Join(plugins, ",")
}

func getSpecDetails(specDirs []string) *gauge_messages.SpecDetails {
	sig := &infoGatherer.SpecInfoGatherer{SpecDirs: specDirs}
	sig.Init()
	specDetails := make([]*gauge_messages.SpecDetails_SpecDetail, 0)
	for _, d := range sig.GetAvailableSpecDetails(specDirs) {
		detail := &gauge_messages.SpecDetails_SpecDetail{}
		if d.HasSpec() {
			detail.Spec = gauge.ConvertToProtoSpec(d.Spec)
		}
		for _, e := range d.Errs {
			detail.ParseErrors = append(detail.ParseErrors, &gauge_messages.Error{Type: gauge_messages.Error_PARSE_ERROR, Filename: e.FileName, Message: e.Message, LineNumber: int32(e.LineNo)})
		}
		specDetails = append(specDetails, detail)
	}
	return &gauge_messages.SpecDetails{
		Details: specDetails,
	}
}
