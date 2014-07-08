package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DEPS_DIR  = "deps"
	BUILD_DIR = "tmp"
)

const (
	HTML_PLUGIN_ID = "html-report"
)

var BUILD_DIR_BIN = filepath.Join(BUILD_DIR, "bin")
var BUILD_DIR_SRC = filepath.Join(BUILD_DIR, "src")
var BUILD_DIR_PKG = filepath.Join(BUILD_DIR, "pkg")

var gaugePackages = []string{"gauge", "gauge-java", "gauge-ruby"}
var gaugePlugins = []string{HTML_PLUGIN_ID}

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
	if err := os.MkdirAll(dstDir, 0755); err != nil {
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
	err := os.MkdirAll(BUILD_DIR_SRC, 0755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(BUILD_DIR_BIN, 0755)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(BUILD_DIR_PKG, 0755)
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

func copyGaugePackagesToGoPath() {
	for _, p := range gaugePackages {
		err := mirrorDir(p, filepath.Join(BUILD_DIR_SRC, p))
		if err != nil {
			panic(err)
		}
	}
}

func copyGaugePluginsToGoPath() {
	for _, pluginName := range gaugePlugins {
		err := mirrorDir(filepath.Join("plugins", pluginName), filepath.Join(BUILD_DIR_SRC, pluginName))
		if err != nil {
			panic(err)
		}
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
	return fmt.Sprintf("%s", bytes), err
}

func compileGoPackage(packageName string) {
	setGoPath()
	runProcess("go", BUILD_DIR, "install", "-v", packageName)
}

func compileGauge() {
	compileGoPackage("gauge")
}

func compileHtmlPlugin() {
	compileGoPackage(HTML_PLUGIN_ID)
}

func compileGaugeJava() {
	compileGoPackage("gauge-java")
	runProcess("ant", "gauge-java", "jar")
}

func compileGaugeRuby() {
	compileGoPackage("gauge-ruby")
}

func runTests(packageName string) {
	setGoPath()
	runProcess("go", BUILD_DIR, "test", "-covermode=count", "-coverprofile=count.out", packageName)
	runProcess("go", BUILD_DIR, "tool", "cover", "-html=count.out")
}

func copyBinaries() {
	err := os.MkdirAll("bin", 0755)
	if err != nil {
		panic(err)
	}

	err = mirrorDir(BUILD_DIR_BIN, "bin")
	if err != nil {
		panic(err)
	}

	absBin, err := filepath.Abs("bin")
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

func installGaugeFiles() {
	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join("bin", "gauge.exe")] = "bin"
	} else {
		files[filepath.Join("bin", "gauge")] = "bin"
	}
	files[filepath.Join("skel", "hello_world.spec")] = filepath.Join("share", "gauge", "skel")
	files[filepath.Join("skel", "default.properties")] = filepath.Join("share", "gauge", "skel", "env")
	installFiles(files, *installPrefix)
}

func installGaugeJavaFiles() {
	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join("bin", "gauge-java.exe")] = "bin"
	} else {
		files[filepath.Join("bin", "gauge-java")] = "bin"
	}
	files[filepath.Join("gauge-java", "java.json")] = filepath.Join("share", "gauge", "languages")
	files[filepath.Join("gauge-java", "skel", "StepImplementation.java")] = filepath.Join("share", "gauge", "skel", "java")
	files[filepath.Join("gauge-java", "skel", "java.properties")] = filepath.Join("share", "gauge", "skel", "env")
	files[filepath.Join("gauge-java", "libs")] = filepath.Join("lib", "gauge", "java", "libs")
	files[filepath.Join("gauge-java", "build", "jar")] = filepath.Join("lib", "gauge", "java")
	installFiles(files, *installPrefix)
}

func installGaugeRubyFiles() {
	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join("bin", "gauge-ruby.exe")] = "bin"
	} else {
		files[filepath.Join("bin", "gauge-ruby")] = "bin"
	}

	files[filepath.Join("gauge-ruby", "ruby.json")] = filepath.Join("share", "gauge", "languages")
	files[filepath.Join("gauge-ruby", "skel", "step_implementation.rb")] = filepath.Join("share", "gauge", "skel", "ruby")
	files[filepath.Join("gauge-ruby", "skel", "ruby.properties")] = filepath.Join("share", "gauge", "skel", "env")
	installFiles(files, *installPrefix)

	// packaging ruby gems
	runProcess("gem", "gauge-ruby", "build", "gauge-ruby.gemspec")
	installGaugeRubyGem()
}

func installGaugeRubyGem() {
	gemHome := getGemHome()
	if gemHome == "" {
		runProcess("gem", "gauge-ruby", "install", "--user-install", "gauge-ruby-0.0.1.gem")
	} else {
		runProcess("gem", "gauge-ruby", "install", "gauge-ruby-0.0.1.gem", "--install-dir", gemHome)
	}
}

func getGemHome() string {
	gemHome := os.Getenv("GEM_HOME")
	if gemHome == "" {
		return gemHomeFromRvm()
	}
	return ""
}

func gemHomeFromRvm() string {
	output, err := runCommand("abc", "gemdir")
	if err == nil {
		return output
	}
	return ""
}

func installPlugins() {
	for pluginName, pluginInstaller := range pluginInstallers {
		fmt.Printf("Installing plugin %s\n", pluginName)
		err := pluginInstaller(*installPrefix)
		if err != nil {
			fmt.Printf("Could not install plugin %s : %s\n", pluginName, err)
		} else {
			fmt.Printf("Successfully installed plugin %s\n", pluginName)
		}
	}
}

// Executes the specified target
// It also keeps a hash of all the contents in the target directory and avoid recompilation if contents are not changed
func executeTarget(target string) {
	opts, ok := targets[target]
	if !ok {
		log.Fatalf("Unknown target: %s\n", target)
	}

	if opts.lookForChanges {
		if hasChanges(hashDir(target), target) {
			opts.targetFunc()
			saveHash(hashDir(target), target)
		}
	} else {
		opts.targetFunc()
	}
}

type compileFunc func()

var test = flag.Bool("test", false, "Run the test cases")
var install = flag.Bool("install", false, "Install to the specified prefix")
var installPrefix = flag.String("prefix", "", "Specifies the prefix where files will be installed")
var compileTarget = flag.String("target", "", "Specifies the target to be executed")

type targetOpts struct {
	lookForChanges bool
	targetFunc     compileFunc
}

// Defines all the compile targets
// Each target name is the directory name

var (
	targets = map[string]*targetOpts{
		"gauge":               &targetOpts{lookForChanges: true, targetFunc: compileGauge},
		"gauge-java":          &targetOpts{lookForChanges: true, targetFunc: compileGaugeJava},
		"gauge-ruby":          &targetOpts{lookForChanges: true, targetFunc: compileGaugeRuby},
		"plugins/html-report": &targetOpts{lookForChanges: true, targetFunc: compileHtmlPlugin},
	}
)

var (
	pluginInstallers = map[string]func(string) error{
		HTML_PLUGIN_ID: installHtmlPlugin,
	}
)

func installHtmlPlugin(installPath string) error {
	pluginSrcBasePath := filepath.Join("plugins", HTML_PLUGIN_ID)
	pluginPropertiesJson, err := ioutil.ReadFile(filepath.Join(pluginSrcBasePath, "plugin.json"))
	if err != nil {
		fmt.Printf("Could not read plugin.json: %s\n", err)
		return err
	}
	var pluginJson interface{}
	if err = json.Unmarshal([]byte(pluginPropertiesJson), &pluginJson); err != nil {
		fmt.Printf("Could not read plugin.json: %s\n", err)
		return err
	}
	pluginProperties := pluginJson.(map[string]interface{})
	pluginRelativePath := filepath.Join("share", "gauge", "plugins", pluginProperties["id"].(string), pluginProperties["version"].(string))
	pluginInstallPath := filepath.Join(installPath, pluginRelativePath)
	err = os.MkdirAll(pluginInstallPath, 0755)
	if err != nil {
		fmt.Printf("Could not create %s: %s\n", pluginInstallPath, err)
		return err
	}

	files := make(map[string]string)
	if runtime.GOOS == "windows" {
		files[filepath.Join("bin", HTML_PLUGIN_ID+".exe")] = pluginRelativePath
	} else {
		files[filepath.Join("bin", HTML_PLUGIN_ID)] = pluginRelativePath
	}
	files[filepath.Join(pluginSrcBasePath, "plugin.json")] = pluginRelativePath
	files[filepath.Join(pluginSrcBasePath, "report-template")] = filepath.Join(pluginRelativePath, "report-template")
	installFiles(files, installPath)
	return nil
}

func main() {
	flag.Parse()
	createGoPathForBuild()
	copyDepsToGoPath()
	copyGaugePackagesToGoPath()
	copyGaugePluginsToGoPath()

	if *test {
		runTests("gauge")
	} else if *install {
		if *installPrefix == "" {
			if runtime.GOOS == "windows" {
				*installPrefix = os.Getenv("PROGRAMFILES")
				if *installPrefix == "" {
					panic(fmt.Errorf("Failed to find programfiles"))
				}
				*installPrefix = filepath.Join(*installPrefix, "gauge")
			} else {
				*installPrefix = "/usr/local"
			}
		}
		installGaugeFiles()
		installGaugeJavaFiles()
		installGaugeRubyFiles()
		installPlugins()
	} else {
		if *compileTarget == "" {
			for target, _ := range targets {
				executeTarget(target)
			}
		} else {
			executeTarget(*compileTarget)
		}
		copyBinaries()
	}
}
