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

package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/version"
)

const (
	CGO_ENABLED        = "CGO_ENABLED"
	GOARCH             = "GOARCH"
	GOOS               = "GOOS"
	X86                = "386"
	X86_64             = "amd64"
	darwin             = "darwin"
	linux              = "linux"
	windows            = "windows"
	bin                = "bin"
	newDirPermissions  = 0755
	gauge              = "gauge"
	gaugeScreenshot    = "gauge_screenshot"
	deploy             = "deploy"
	installShellScript = "install.sh"
	CC                 = "CC"
	pkg                = ".pkg"
	packagesBuild      = "packagesbuild"
	nightlyDatelayout  = "2006-01-02"
)

var darwinPackageProject = filepath.Join("build", "install", "macosx", "gauge-pkg.pkgproj")

var gaugeScreenshotLocation = filepath.Join("github.com", "getgauge", "gauge_screenshot")

var deployDir = filepath.Join(deploy, gauge)

func set(envName, envValue string) {
	log.Printf("%s = %s\n", envName, envValue)
	err := os.Setenv(envName, envValue)
	if err != nil {
		panic(err)
	}
}

func runProcess(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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

func signExecutable(exeFilePath string, certFilePath string, certFilePwd string) {
	if getGOOS() == windows {
		if certFilePath != "" && certFilePwd != "" {
			log.Printf("Signing: %s", exeFilePath)
			runProcess("signtool", "sign", "/f", certFilePath, "/p", certFilePwd, exeFilePath)
		} else {
			log.Printf("No certificate file passed. Executable won't be signed.")
		}
	}
}

var buildMetadata string

func getBuildVersion() string {
	if buildMetadata != "" {
		return fmt.Sprintf("%s.%s", version.CurrentGaugeVersion.String(), buildMetadata)
	}
	return version.CurrentGaugeVersion.String()
}

func compileGauge() {
	executablePath := getGaugeExecutablePath(gauge)
	runProcess("go", "build", "-ldflags", "-X github.com/getgauge/gauge/version.BuildMetadata="+buildMetadata, "-o", executablePath)
	compileGaugeScreenshot()
}

func compileGaugeScreenshot() {
	getGaugeScreenshot()
	executablePath := getGaugeExecutablePath(gaugeScreenshot)
	runProcess("go", "build", "-o", executablePath, gaugeScreenshotLocation)
}

func getGaugeScreenshot() {
	runProcess("go", "get", "-u", "-d", gaugeScreenshotLocation)
}

func runTests(coverage bool) {
	if coverage {
		runProcess("go", "test", "-covermode=count", "-coverprofile=count.out")
		if coverage {
			runProcess("go", "tool", "cover", "-html=count.out")
		}
	} else {
		runProcess("go", "test", "./...", "-v")
	}
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
			_, err = common.MirrorDir(src, installDst)
		} else {
			err = common.MirrorFile(src, filepath.Join(installDst, base))
		}
		if err != nil {
			panic(err)
		}
	}
}

func copyGaugeFiles(installPath string) {
	files := make(map[string]string)
	files[getGaugeExecutablePath(gauge)] = bin
	files[getGaugeExecutablePath(gaugeScreenshot)] = bin
	files[filepath.Join("skel", "example.spec")] = filepath.Join("share", gauge, "skel")
	files[filepath.Join("skel", "default.properties")] = filepath.Join("share", gauge, "skel", "env")
	files[filepath.Join("skel", "gauge.properties")] = filepath.Join("share", gauge)
	files[filepath.Join("notice.md")] = filepath.Join("share", gauge)
	files = addInstallScripts(files)
	installFiles(files, installPath)
}

func addInstallScripts(files map[string]string) map[string]string {
	if (getGOOS() == darwin || getGOOS() == linux) && (*distro) {
		files[filepath.Join("build", "install", installShellScript)] = ""
	} else if getGOOS() == windows {
		files[filepath.Join("build", "install", "windows", "plugin-install.bat")] = ""
		files[filepath.Join("build", "install", "windows", "backup_properties_file.bat")] = ""
		files[filepath.Join("build", "install", "windows", "set_timestamp.bat")] = ""
	}
	return files
}

func moveOSBinaryToCurrentOSArchDirectory() {
	destDir := path.Join(bin, fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
	moveBinaryToDirectory(gauge, destDir)
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
	if err := common.MirrorFile(srcFile, destFile); err != nil {
		return err
	}
	return os.Remove(srcFile)
}

func setEnv(envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}
}

type compileFunc func()

var test = flag.Bool("test", false, "Run the test cases")
var coverage = flag.Bool("coverage", false, "Run the test cases and show the coverage")
var install = flag.Bool("install", false, "Install to the specified prefix")
var nightly = flag.Bool("nightly", false, "Add nightly build information")
var gaugeInstallPrefix = flag.String("prefix", "", "Specifies the prefix where gauge files will be installed")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles for all platforms windows, linux, darwin both x86 and x86_64")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")
var distro = flag.Bool("distro", false, "Create gauge distributable")
var skipWindowsDistro = flag.Bool("skip-windows", false, "Skips creation of windows distributable on unix machines while cross platform compilation")
var certFile = flag.String("certFile", "", "Should be passed for signing the windows installer along with the password (certFilePwd)")
var certFilePwd = flag.String("certFilePwd", "", "Password for certificate that will be used to sign the windows installer")

type targetOpts struct {
	lookForChanges bool
	targetFunc     compileFunc
	name           string
	dir            string
}

// Defines all the compile targets
// Each target name is the directory name

var (
	platformEnvs = []map[string]string{
		map[string]string{GOARCH: X86, GOOS: darwin, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: darwin, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: linux, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: linux, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: windows, CC: "i586-mingw32-gcc", CGO_ENABLED: "1"},
		map[string]string{GOARCH: X86_64, GOOS: windows, CC: "x86_64-w64-mingw32-gcc", CGO_ENABLED: "1"},
	}
	osDistroMap = map[string]distroFunc{windows: createWindowsDistro, linux: createLinuxPackage, darwin: createDarwinPackage}
)

func main() {
	flag.Parse()
	if *nightly {
		buildMetadata = fmt.Sprintf("nightly-%s", time.Now().Format(nightlyDatelayout))
	}
	if *test {
		runTests(*coverage)
	} else if *install {
		installGauge()
	} else if *distro {
		createGaugeDistributables(*allPlatforms)
	} else {
		if *allPlatforms {
			crossCompileGauge()
		} else {
			compileGauge()
		}
	}
}

func crossCompileGauge() {
	for _, platformEnv := range platformEnvs {
		setEnv(platformEnv)
		log.Printf("Compiling for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
		compileGauge()
	}
}

func installGauge() {
	updateGaugeInstallPrefix()
	copyGaugeFiles(deployDir)
	if _, err := common.MirrorDir(deployDir, *gaugeInstallPrefix); err != nil {
		panic(fmt.Sprintf("Could not install gauge : %s", err))
	}
}

func createGaugeDistributables(forAllPlatforms bool) {
	if forAllPlatforms {
		for _, platformEnv := range platformEnvs {
			setEnv(platformEnv)
			log.Printf("Creating distro for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
			createDistro()
		}
	} else {
		createDistro()
	}
}

type distroFunc func()

func createDistro() {
	osDistroMap[getGOOS()]()
}

func createWindowsDistro() {
	if !*skipWindowsDistro {
		createWindowsInstaller()
	}
}

func createWindowsInstaller() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", gauge, getBuildVersion(), getGOOS(), getPackageArchSuffix())
	distroDir, err := filepath.Abs(filepath.Join(deploy, packageName))
	installerFileName := filepath.Join(filepath.Dir(distroDir), packageName)
	if err != nil {
		panic(err)
	}
	copyGaugeFiles(distroDir)
	runProcess("makensis.exe",
		fmt.Sprintf("/DPRODUCT_VERSION=%s", getBuildVersion()),
		fmt.Sprintf("/DGAUGE_DISTRIBUTABLES_DIR=%s", distroDir),
		fmt.Sprintf("/DOUTPUT_FILE_NAME=%s.exe", installerFileName),
		filepath.Join("build", "install", "windows", "gauge-install.nsi"))
	os.RemoveAll(distroDir)
	signExecutable(installerFileName+".exe", *certFile, *certFilePwd)
}

func createDarwinPackage() {
	distroDir := filepath.Join(deploy, gauge)
	copyGaugeFiles(distroDir)
	runProcess(packagesBuild, "-v", darwinPackageProject)
	runProcess("mv", filepath.Join(deploy, gauge+pkg), filepath.Join(deploy, fmt.Sprintf("%s-%s-%s.%s%s", gauge, getBuildVersion(), getGOOS(), getPackageArchSuffix(), pkg)))
	os.RemoveAll(distroDir)
}

func createLinuxPackage() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", gauge, getBuildVersion(), getGOOS(), getPackageArchSuffix())
	distroDir := filepath.Join(deploy, packageName)
	copyGaugeFiles(distroDir)
	createZipFromUtil(deploy, packageName)
	os.RemoveAll(distroDir)
}

func createZipFromUtil(dir, name string) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(filepath.Join(dir, name))
	output, err := runCommand("zip", "-r", filepath.Join("..", name+".zip"), ".")
	fmt.Println(output)
	if err != nil {
		panic(fmt.Sprintf("Failed to zip: %s", err))
	}
	os.Chdir(wd)
}

func createZip(dir, packageName string) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(dir)

	zipFileName := packageName + ".zip"
	newfile, err := os.Create(zipFileName)
	if err != nil {
		panic(err)
	}
	defer newfile.Close()
	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	filepath.Walk(packageName, func(path string, info os.FileInfo, err error) error {
		infoHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			panic(err)
		}
		infoHeader.Name = strings.Replace(path, fmt.Sprintf("%s%c", packageName, filepath.Separator), "", 1)
		if info.IsDir() {
			return nil
		}
		writer, err := zipWriter.CreateHeader(infoHeader)
		if err != nil {
			panic(err)
		}
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		if err != nil {
			panic(err)
		}
		return nil
	})
	log.Printf("Created zip: ", zipFileName)
	os.Chdir(wd)
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

func getUserHome() string {
	return os.Getenv("HOME")
}

func getGaugeExecutablePath(file string) string {
	return filepath.Join(getBinDir(), getExecutableName(file))
}

func getBinDir() string {
	if *binDir != "" {
		return *binDir
	}
	return filepath.Join(bin, fmt.Sprintf("%s_%s", getGOOS(), getGOARCH()))
}

func getExecutableName(file string) string {
	if getGOOS() == windows {
		return file + ".exe"
	}
	return file
}

func getGOARCH() string {
	goArch := os.Getenv(GOARCH)
	if goArch == "" {
		goArch = runtime.GOARCH
	}
	return goArch
}

func getGOOS() string {
	goOS := os.Getenv(GOOS)
	if goOS == "" {
		goOS = runtime.GOOS
	}
	return goOS
}

func getPackageArchSuffix() string {
	if strings.HasSuffix(*binDir, "386") {
		return "x86"
	}

	if strings.HasSuffix(*binDir, "amd64") {
		return "x86_64"
	}

	if arch := getGOARCH(); arch == X86 {
		return "x86"
	}
	return "x86_64"
}
