package projectInit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

const (
	specsDirName      = "specs"
	skelFileName      = "example.spec"
	envDefaultDirName = "default"
	metadataFileName  = "metadata.json"
)

var defaultPlugins = []string{"html-report"}

type templateMetadata struct {
	Name           string
	Description    string
	Version        string
	PostInstallCmd string
}

func initializeTemplate(templateName string) error {
	tempDir := common.GetTempDir()
	defer util.Remove(tempDir)
	unzippedTemplate, err := util.DownloadAndUnzip(getTemplateURL(templateName), tempDir)
	if err != nil {
		return err
	}

	wd := config.ProjectRoot

	logger.Info("Copying Gauge template %s to current directory ...", templateName)
	filesAdded, err := common.MirrorDir(filepath.Join(unzippedTemplate, templateName), wd)
	if err != nil {
		return fmt.Errorf("Failed to copy Gauge template: %s", err.Error())
	}

	metadataFile := filepath.Join(wd, metadataFileName)
	metadataContents, err := common.ReadFileContents(metadataFile)
	if err != nil {
		return fmt.Errorf("Failed to read file contents of %s: %s", metadataFile, err.Error())
	}

	metadata := &templateMetadata{}
	err = json.Unmarshal([]byte(metadataContents), metadata)
	if err != nil {
		return err
	}

	if metadata.PostInstallCmd != "" {
		command := strings.Fields(metadata.PostInstallCmd)
		cmd, err := common.ExecuteSystemCommand(command, wd, os.Stdout, os.Stderr)
		cmd.Wait()
		if err != nil {
			for _, file := range filesAdded {
				pathSegments := strings.Split(file, string(filepath.Separator))
				util.Remove(filepath.Join(wd, pathSegments[0]))
			}
			return fmt.Errorf("Failed to run post install commands: %s", err.Error())
		}
	}

	util.Remove(metadataFile)
	return nil
}

func getTemplateURL(templateName string) string {
	return config.GaugeTemplatesUrl() + "/" + templateName + ".zip"
}

func getTemplateLangauge(templateName string) string {
	return strings.Split(templateName, "_")[0]
}

// InitializeProject initializes a Gauge project with specified template
func InitializeProject(templateName string) {
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Failed to find working directory. %s", err.Error())
	}
	config.ProjectRoot = wd

	exists, _ := common.UrlExists(getTemplateURL(templateName))

	if exists {
		err = initializeTemplate(templateName)
	} else {
		err = createProjectTemplate(templateName)
	}

	if err != nil {
		logger.Fatalf("Failed to initialize project. %s", err.Error())
	}
	logger.Info("Successfully initialized the project. Run specifications with \"gauge specs/\"\n")

	language := getTemplateLangauge(templateName)
	if !install.IsCompatiblePluginInstalled(language, true) {
		logger.Info("Compatible langauge plugin %s is not installed. Installing plugin...", language)

		install.HandleInstallResult(install.InstallPlugin(language, ""), language, true)
	}
}

// ListTemplates lists all the Gauge templates available in GaugeTemplatesURL
func ListTemplates() {
	templatesURL := config.GaugeTemplatesUrl()
	_, err := common.UrlExists(templatesURL)
	if err != nil {
		logger.Fatalf("Gauge templates URL is not reachable: %s", err.Error())
	}
	tempDir := common.GetTempDir()
	defer util.Remove(tempDir)
	templatesPage, err := util.Download(templatesURL, tempDir, "", true)
	if err != nil {
		util.Remove(tempDir)
		logger.Fatalf("Error occurred while downloading templates list: %s", err.Error())
	}

	templatePageContents, err := common.ReadFileContents(templatesPage)
	if err != nil {
		util.Remove(tempDir)
		logger.Fatalf("Failed to read contents of file %s: %s", templatesPage, err.Error())
	}
	templates := getTemplateNames(templatePageContents)
	for _, template := range templates {
		logger.Info(template)
	}
	logger.Info("\nRun `gauge --init <template_name>` to create a new Gauge project.")
}

func getTemplateNames(text string) []string {
	templateRegex, _ := regexp.Compile(`rel="nofollow">([a-z_0-9]*).zip</a></pre>`)
	submatches := templateRegex.FindAllStringSubmatch(text, -1)
	var matches []string
	for _, match := range submatches {
		matches = append(matches, match[1])
	}
	// add other templates
	matches = append(matches, "csharp")
	matches = append(matches, "ruby")

	sort.Strings(matches)
	return matches
}

func showMessage(action, filename string) {
	logger.Info(" %s  %s", action, filename)
}

func createProjectTemplate(language string) error {
	err := runner.ExecuteInitHookForRunner(language)
	if err != nil {
		return err
	}
	// Create the project manifest
	showMessage("create", common.ManifestFile)
	if common.FileExists(common.ManifestFile) {
		showMessage("skip", common.ManifestFile)
	}
	manifest := &manifest.Manifest{Language: language, Plugins: defaultPlugins}
	if err := manifest.Save(); err != nil {
		return err
	}

	// creating the spec directory
	showMessage("create", specsDirName)
	if !common.DirExists(specsDirName) {
		err := os.Mkdir(specsDirName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specsDirName, err.Error()))
		}
	} else {
		showMessage("skip", specsDirName)
	}

	// Copying the skeleton file
	skelFile, err := common.GetSkeletonFilePath(skelFileName)
	if err != nil {
		return err
	}
	specFile := filepath.Join(specsDirName, skelFileName)
	showMessage("create", specFile)
	if common.FileExists(specFile) {
		showMessage("skip", specFile)
	} else {
		err = common.CopyFile(skelFile, specFile)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specFile, err.Error()))
		}
	}

	// Creating the env directory
	showMessage("create", common.EnvDirectoryName)
	if !common.DirExists(common.EnvDirectoryName) {
		err = os.Mkdir(common.EnvDirectoryName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", common.EnvDirectoryName, err.Error()))
		}
	}
	defaultEnv := filepath.Join(common.EnvDirectoryName, envDefaultDirName)
	showMessage("create", defaultEnv)
	if !common.DirExists(defaultEnv) {
		err = os.Mkdir(defaultEnv, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultEnv, err.Error()))
		}
	}
	defaultJSON, err := common.GetSkeletonFilePath(filepath.Join(common.EnvDirectoryName, common.DefaultEnvFileName))
	if err != nil {
		return err
	}
	defaultJSONDest := filepath.Join(defaultEnv, common.DefaultEnvFileName)
	showMessage("create", defaultJSONDest)
	err = common.CopyFile(defaultJSON, defaultJSONDest)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultJSONDest, err.Error()))
	}
	return nil
}
