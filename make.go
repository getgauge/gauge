package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DEPS_DIR    = "deps"
	BUILD_DIR   = "tmp"
	CGO_ENABLED = "CGO_ENABLED"
)

const (
	htmlPluginId      = "html-report"
	gaugeRubyGemfile  = "gauge-ruby-*.gem"
	dotGauge          = ".gauge"
	plugins           = "plugins"
	GOARCH            = "GOARCH"
	GOOS              = "GOOS"
	X86               = "386"
	X86_64            = "amd64"
	DARWIN            = "darwin"
	LINUX             = "linux"
	WINDOWS           = "windows"
	bin               = "bin"
	newDirPermissions = 0755
	pluginJsonFile    = "plugin.json"
	reportTemplate    = "report-template"
	gauge             = "gauge"
)

var BUILD_DIR_BIN = filepath.Join(BUILD_DIR, bin)
var BUILD_DIR_SRC = filepath.Join(BUILD_DIR, "src")
var BUILD_DIR_PKG = filepath.Join(BUILD_DIR, "pkg")
var platformBinDir = filepath.Join(bin, fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))

var gaugePackages = []string{gauge, "gauge-java", "gauge-ruby"}
var gaugePlugins = []string{htmlPluginId}

func hashDir(dirPath string) string {
	var b bytes.Buffer
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			contents, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			h := sha1.New()
			h.Write(contents)
			b.WriteString(fmt.Sprintf("%x", h.Sum(nil)))
		}
		return nil
	})
	h := sha1.New()
	h.Write(b.Bytes())
	return fmt.Sprintf("%x", h.Sum(nil))
}

func isExecMode(mode os.FileMode) bool {
	return (mode & 0111) != 0
}

func mirrorFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.Mode()&os.ModeType != 0 {
		log.Fatalf("mirrorFile can't deal with non-regular file %s", src)
	}
	dfi, err := os.Stat(dst)
	if err == nil &&
		isExecMode(sfi.Mode()) == isExecMode(dfi.Mode()) &&
		(dfi.Mode()&os.ModeType == 0) &&
		dfi.Size() == sfi.Size() &&
		dfi.ModTime().Unix() == sfi.ModTime().Unix() {
		// Seems to not be modified.
		return nil
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, newDirPermissions); err != nil {
		return err
	}

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	n, err := io.Copy(df, sf)
	if err == nil && n != sfi.Size() {
		err = fmt.Errorf("copied wrong size for %s -> %s: copied %d; want %d", src, dst, n, sfi.Size())
	}
	cerr := df.Close()
	if err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(dst, sfi.Mode())
	}
	if err == nil {
		err = os.Chtimes(dst, sfi.ModTime(), sfi.ModTime())
	}
	return err
}

func mirrorDir(src, dst string) error {
	log.Printf("Copying '%s' -> '%s'\n", src, dst)
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		suffix, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("Failed to find Rel(%q, %q): %v", src, path, err)
		}
		return mirrorFile(path, filepath.Join(dst, suffix))
	})
	return err
}

func createGoPathForBuild() {
	err := os.MkdirAll(BUILD_DIR_SRC, newDirPermissions)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(BUILD_DIR_BIN, newDirPermissions)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(BUILD_DIR_PKG, newDirPermissions)
	if err != nil {
		panic(err)
	}
}

func copyDepsToGoPath() {
	err := mirrorDir(DEPS_DIR, BUILD_DIR_SRC)
	if err != nil {
		panic(err)
	}
}

//Copy gauge packages that are available.
func copyGaugePackagesToGoPath() {
	for _, p := range gaugePackages {
		mirrorDir(p, filepath.Join(BUILD_DIR_SRC, p))
	}
}

func copyGaugePluginsToGoPath() {
	for _, pluginName := range gaugePlugins {
		mirrorDir(filepath.Join("plugins", pluginName), filepath.Join(BUILD_DIR_SRC, pluginName))
	}
}

func setGoPath() {
	absBuildDir, err := filepath.Abs(BUILD_DIR)
	if err != nil {
		panic(err)
	}
	log.Printf("GOPATH = %s\n", absBuildDir)
	err = os.Setenv("GOPATH", absBuildDir)
	if err != nil {
		panic(err)
	}
}

func runProcess(command string, workingDirectory string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDirectory
	log.Printf("Execute %v\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func runCommand(command string, arg ...string) (string, error) {
	cmd := exec.Command(command, arg...)
	bytes, err := cmd.Output()
	return strings.TrimSpace(fmt.Sprintf("%s", bytes)), err
}

func compileGoPackage(packageName string) {
	setGoPath()
	runProcess("go", BUILD_DIR, "install", "-v", packageName)
}

func compileGauge() {
	compileGoPackage(gauge)
}

func compileHtmlPlugin() {
	compileGoPackage(htmlPluginId)
}

func compileGaugeJava() {
	compileGoPackage("gauge-java")
	runProcess("ant", "gauge-java", "jar")
}

func compileGaugeRuby() {
	compileGoPackage("gauge-ruby")
}

func runTests(packageName string, coverage bool) {
	setGoPath()
	runProcess("go", BUILD_DIR, "test", "-covermode=count", "-coverprofile=count.out", packageName)
	if coverage {
		runProcess("go", BUILD_DIR, "tool", "cover", "-html=count.out")
	}
}

func copyBinaries() {
	err := os.MkdirAll(bin, newDirPermissions)
	if err != nil {
		panic(err)
	}

	err = mirrorDir(BUILD_DIR_BIN, bin)
	if err != nil {
		panic(err)
	}

	absBin, err := filepath.Abs(bin)
	if err != nil {
		panic(err)
	}
	log.Printf("Binaries are available at: %s\n", absBin)
}

// key will be the source file and value will be the target
func installFiles(files map[string]string, installDir string) {
	for src, dst := range files {
		base := filepath.Base(src)
		installDst := filepath.Join(installDir, dst)
		log.Printf("Install %s -> %s\n", src, installDst)
		stat, err := os.Stat(src)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			err = mirrorDir(src, installDst)
		} else {
			err = mirrorFile(src, filepath.Join(installDst, base))
		}
		if err != nil {
			panic(err)
		}
	}
}

func saveHash(h, dir string) {
	hashFile := getHasHFile(dir)
	err := ioutil.WriteFile(hashFile, []byte(h), 0644)
	if err != nil {
		log.Println(err.Error())
	}
}

func hasChanges(h, dir string) bool {
	hashFile := getHasHFile(dir)
	contents, err := ioutil.ReadFile(hashFile)
	if err != nil {
		return true
	}
	return string(contents) != h
}

func getHasHFile(dir string) string {
	if strings.Contains(dir, "plugins/") {
		return filepath.Join(BUILD_DIR, "."+strings.Trim(dir, "plugins/"))
	} else {
		return filepath.Join(BUILD_DIR, "."+dir)
	}
}

func installGaugeFiles(installPath string) {
	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join(getBinDir(), "gauge.exe")] = bin
	} else {
		files[filepath.Join(getBinDir(), gauge)] = bin
	}
	files[filepath.Join("skel", "hello_world.spec")] = filepath.Join("share", gauge, "skel")
	files[filepath.Join("skel", "default.properties")] = filepath.Join("share", gauge, "skel", "env")
	files[filepath.Join("skel", "gauge.properties")] = filepath.Join("share", gauge)
	installFiles(files, installPath)
}

func installGaugeJavaFiles(installPath string) error {
	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join(getBinDir(), "gauge-java.exe")] = bin
	} else {
		files[filepath.Join(getBinDir(), "gauge-java")] = bin
	}

	javaRunnerProperties, err := getPluginProperties(filepath.Join("gauge-java", "java.json"))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get java runner properties. %s", err))
	}
	javaRunnerRelativePath := filepath.Join(installPath, "java", javaRunnerProperties["version"].(string))

	files[filepath.Join("gauge-java", "java.json")] = ""
	files[filepath.Join("gauge-java", "skel", "StepImplementation.java")] = filepath.Join("skel")
	files[filepath.Join("gauge-java", "skel", "java.properties")] = filepath.Join("skel", "env")
	files[filepath.Join("gauge-java", "libs")] = filepath.Join("libs")
	files[filepath.Join("gauge-java", "build", "jar")] = filepath.Join("libs")
	installFiles(files, javaRunnerRelativePath)
	return nil
}

func installGaugeRubyFiles(installPath string) error {
	runProcess("gem", "gauge-ruby", "build", "gauge-ruby.gemspec")

	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join(getBinDir(), "gauge-ruby.exe")] = bin
	} else {
		files[filepath.Join(getBinDir(), "gauge-ruby")] = bin
	}

	rubyRunnerProperties, err := getPluginProperties(filepath.Join("gauge-ruby", "ruby.json"))

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get ruby runner properties. %s", err))
	}
	rubyRunnerRelativePath := filepath.Join(installPath, "ruby", rubyRunnerProperties["version"].(string))

	files[filepath.Join("gauge-ruby", "ruby.json")] = ""
	files[filepath.Join("gauge-ruby", "skel", "step_implementation.rb")] = filepath.Join("skel")
	files[filepath.Join("gauge-ruby", "skel", "ruby.properties")] = filepath.Join("skel", "env")
	gemFile := getGemFile("gauge-ruby")
	if gemFile == "" {
		return errors.New(fmt.Sprintf("Could not find .gem file"))
	}

	files[filepath.Join("gauge-ruby", gemFile)] = ""
	installFiles(files, rubyRunnerRelativePath)
	installGaugeRubyGem()
	return nil
}

func getBinDir() string {
	if *binDir == "" {
		return platformBinDir
	}
	return path.Join(bin, *binDir)
}

func getGemFile(dir string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("Could not find .gem file")
		os.Exit(1)
	} else {
		for _, file := range files {
			if filepath.Ext(filepath.Join(dir, file.Name())) == ".gem" {
				return file.Name()
			}
		}
	}
	return ""
}

func installGaugeRubyGem() {
	gemHome := getGemHome()
	if gemHome == "" {
		runProcess("gem", "gauge-ruby", "install", "--user-install", gaugeRubyGemfile)
	} else {
		runProcess("gem", "gauge-ruby", "install", gaugeRubyGemfile, "--install-dir", gemHome)
	}
}

func getGemHome() string {
	gemHome := os.Getenv("GEM_HOME")
	if gemHome != "" {
		return gemHome
	} else {
		return gemHomeFromRvm()
	}
}

func gemHomeFromRvm() string {
	output, err := runCommand("rvm", "gemdir")
	if err == nil {
		return output
	}
	return ""
}

func installPlugins(installPath string) {
	for pluginName, _ := range pluginInstallers {
		installPlugin(pluginName, installPath)
	}
}

func installPlugin(pluginId, installPath string) {
	fmt.Printf("Installing plugin %s\n", pluginId)
	if _, ok := pluginInstallers[pluginId]; !ok {
		panic(fmt.Sprintf("Invalid plugin name => %s", pluginId))
	}
	err := pluginInstallers[pluginId](installPath)
	if err != nil {
		fmt.Printf("Could not install plugin %s : %s\n", pluginId, err)
	} else {
		fmt.Printf("Successfully installed plugin %s\n", pluginId)
	}
}

// Executes the specified target
// It also keeps a hash of all the contents in the target directory and avoid recompilation if contents are not changed
func executeTarget(target *targetOpts, forAllPlatforms bool) {
	if forAllPlatforms {
		for _, platformEnv := range platformEnvs {
			fmt.Printf("Executing target %s for platform envs:%s \n", target, platformEnv)
			setEnv(platformEnv)
			runTarget(target, false)
		}
	} else {
		runTarget(target, true)
	}
}
func moveOSBinaryToCurrentOSArchDirectory(targetName string) {
	destDir := path.Join(bin, fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
	if targetName == "" {
		for _, target := range getAllTargetNames() {
			moveBinaryToDirectory(path.Base(target), destDir)
		}
	} else {
		moveBinaryToDirectory(path.Base(targetName), destDir)
	}
}

func moveBinaryToDirectory(target, destDir string) error {
	if runtime.GOOS == "windows" {
		target = target + ".exe"
	}
	srcFile := path.Join(bin, target)
	destFile := path.Join(destDir, target)
	if err := os.MkdirAll(destDir, newDirPermissions); err != nil {
		return err
	}
	if err := mirrorFile(srcFile, destFile); err != nil {
		return err
	}
	return os.Remove(srcFile)
}

func runTarget(targetOpts *targetOpts, checkChanges bool) {
	if checkChanges && targetOpts.lookForChanges {
		if hasChanges(hashDir(targetOpts.dir), targetOpts.name) {
			targetOpts.targetFunc()
			saveHash(hashDir(targetOpts.dir), targetOpts.name)
		}
	} else {
		targetOpts.targetFunc()
	}
}

func setEnv(envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}
}

type compileFunc func()

var test = flag.Bool("test", false, "Run the test cases")
var coverage = flag.Bool("test-coverage", false, "Run the test cases and show the coverage")
var install = flag.Bool("install", false, "Install to the specified prefix")
var gaugeInstallPrefix = flag.String("prefix", "", "Specifies the prefix where gauge files will be installed")
var pluginInstallPrefix = flag.String("plugin-prefix", "", "Specifies the prefix where gauge plugins will be installed")
var target = flag.String("target", "", "Specifies the target to be compiled or installed. Can be gauge, default plugin or language plugin")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles for all platforms windows, linux, darwin both x86 and x86_64")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")
var gaugeOnly = flag.Bool(gauge, false, "Compiles or Installs only gauge. Skips language installation")
var allPlugins = flag.Bool("plugins", false, "Compiles or installs only Plugins including all language runners")
var defaultPlugins = flag.Bool("default-plugins", false, "Compiles or installs only default Plugins. This does not include languages")

type targetOpts struct {
	lookForChanges bool
	targetFunc     compileFunc
	name           string
	dir            string
}

// Defines all the compile targets
// Each target name is the directory name

var (
	gaugeCompileTarget    = &targetOpts{lookForChanges: true, targetFunc: compileGauge, name: gauge, dir: gauge}
	languagePluginTargets = map[string]*targetOpts{
		"gauge-java": &targetOpts{lookForChanges: true, targetFunc: compileGaugeJava, name: "gauge-java", dir: "gauge-java"},
		"gauge-ruby": &targetOpts{lookForChanges: true, targetFunc: compileGaugeRuby, name: "gauge-ruby", dir: "gauge-ruby"},
	}
	defaultPluginTargets = map[string]*targetOpts{
		"html-report": &targetOpts{lookForChanges: true, targetFunc: compileHtmlPlugin, name: htmlPluginId, dir: "plugins/html-report"},
	}

	platformEnvs = []map[string]string{
		map[string]string{GOARCH: X86, GOOS: DARWIN, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: DARWIN, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: LINUX, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: LINUX, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: WINDOWS, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: WINDOWS, CGO_ENABLED: "0"},
	}
)

var (
	pluginInstallers = map[string]func(string) error{
		htmlPluginId: installHtmlPlugin,
		"java":       installGaugeJavaFiles,
		"ruby":       installGaugeRubyFiles,
	}
)

func installHtmlPlugin(installPath string) error {
	pluginSrcBasePath := filepath.Join("plugins", htmlPluginId)

	pluginProperties, err := getPluginProperties(filepath.Join(pluginSrcBasePath, pluginJsonFile))
	if err != nil {
		return err
	}
	pluginRelativePath := filepath.Join(pluginProperties["id"].(string), pluginProperties["version"].(string))
	pluginInstallPath := filepath.Join(installPath, pluginRelativePath)
	err = os.MkdirAll(pluginInstallPath, newDirPermissions)
	if err != nil {
		fmt.Printf("Could not create %s: %s\n", pluginInstallPath, err)
		return err
	}

	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join(getBinDir(), htmlPluginId+".exe")] = filepath.Join(pluginRelativePath, bin)
	} else {
		files[filepath.Join(getBinDir(), htmlPluginId)] = filepath.Join(pluginRelativePath, bin)
	}
	files[filepath.Join(pluginSrcBasePath, pluginJsonFile)] = pluginRelativePath
	files[filepath.Join(pluginSrcBasePath, reportTemplate)] = filepath.Join(pluginRelativePath, reportTemplate)
	installFiles(files, installPath)
	return nil
}

func getPluginProperties(jsonPropertiesFile string) (map[string]interface{}, error) {
	pluginPropertiesJson, err := ioutil.ReadFile(jsonPropertiesFile)
	if err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
		return nil, err
	}
	var pluginJson interface{}
	if err = json.Unmarshal([]byte(pluginPropertiesJson), &pluginJson); err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
		return nil, err
	}
	return pluginJson.(map[string]interface{}), nil
}

func main() {
	flag.Parse()

	createGoPathForBuild()
	copyDepsToGoPath()
	copyGaugePackagesToGoPath()
	copyGaugePluginsToGoPath()

	if *test {
		runTests(gauge, false)
	} else if *coverage {
		runTests(gauge, true)
	} else if *install {
		installGaugeComponents()
	} else {
		compileGaugeComponents()
	}
}

func compileGaugeComponents() {
	if *gaugeOnly {
		executeTarget(gaugeCompileTarget, *allPlatforms)
	} else if *allPlugins {
		compilePlugins()
	} else if *defaultPlugins {
		compileDefaultPlugins()
	} else if *target == "" {
		compileAllTargets()
	} else {
		executeTarget(getTargetOpts(*target), *allPlatforms)
	}
	copyBinaries()
	moveOSBinaryToCurrentOSArchDirectory(*target)
}

func installGaugeComponents() {
	updateGaugeInstallPrefix()
	updatePluginInstallPrefix()

	if *gaugeOnly {
		installGaugeFiles(*gaugeInstallPrefix)
	} else if *allPlugins {
		installPlugins(*pluginInstallPrefix)
	} else if *defaultPlugins {
		installDefaultPlugins(*pluginInstallPrefix)
	} else if *target != "" {
		installPlugin(*target, *pluginInstallPrefix)
	} else {
		installGaugeFiles(*gaugeInstallPrefix)
		installPlugins(*pluginInstallPrefix)
	}
}

func updateGaugeInstallPrefix() {
	if *gaugeInstallPrefix == "" {
		if runtime.GOOS == "windows" {
			*gaugeInstallPrefix = os.Getenv("PROGRAMFILES")
			if *gaugeInstallPrefix == "" {
				panic(fmt.Errorf("Failed to find programfiles"))
			}
			*gaugeInstallPrefix = filepath.Join(*gaugeInstallPrefix, gauge)
		} else {
			*gaugeInstallPrefix = "/usr/local"
		}
	}
}

func installDefaultPlugins(pluginInstallPrefix string) {
	for defaultPluginTarget, _ := range defaultPluginTargets {
		installPlugin(defaultPluginTarget, pluginInstallPrefix)
	}
}

func updatePluginInstallPrefix() {
	if *pluginInstallPrefix == "" {
		if runtime.GOOS == "windows" {
			*pluginInstallPrefix = os.Getenv("APPDATA")
			if *pluginInstallPrefix == "" {
				panic(fmt.Errorf("Failed to find AppData directory"))
			}
			*pluginInstallPrefix = filepath.Join(*pluginInstallPrefix, gauge, plugins)
		} else {
			userHome := getUserHome()
			if userHome == "" {
				panic(fmt.Errorf("Failed to find User Home directory"))
			}
			*pluginInstallPrefix = filepath.Join(userHome, dotGauge, plugins)
		}
	}
}

func compileAllTargets() {
	executeTarget(gaugeCompileTarget, *allPlatforms)
	for _, targetOpt := range languagePluginTargets {
		executeTarget(targetOpt, *allPlatforms)
	}
	for _, targetOpt := range defaultPluginTargets {
		executeTarget(targetOpt, *allPlatforms)
	}
}

func getTargetOpts(target string) *targetOpts {
	if target == gauge {
		return gaugeCompileTarget
	}
	if opt, ok := languagePluginTargets[target]; ok {
		return opt
	}
	if opt, ok := defaultPluginTargets[target]; ok {
		return opt
	}
	panic(fmt.Sprintf("Invalid target specified: %s", target))

}

func compilePlugins() {
	for _, targetOpt := range languagePluginTargets {
		executeTarget(targetOpt, *allPlatforms)
	}
	compileDefaultPlugins()
}

func compileDefaultPlugins() {
	for _, targetOpt := range defaultPluginTargets {
		executeTarget(targetOpt, *allPlatforms)
	}
}

func getAllTargetNames() []string {
	targetNames := make([]string, 0)
	targetNames = append(targetNames, gaugeCompileTarget.name)
	for _, targetOpt := range defaultPluginTargets {
		targetNames = append(targetNames, targetOpt.name)
	}
	for _, targetOpt := range languagePluginTargets {
		targetNames = append(targetNames, targetOpt.name)
	}
	return targetNames
}

func getUserHome() string {
	return os.Getenv("HOME")
}
