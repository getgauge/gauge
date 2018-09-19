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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/version"
)

const (
	CGO_ENABLED       = "CGO_ENABLED"
	GOARCH            = "GOARCH"
	GOOS              = "GOOS"
	X86               = "386"
	X86_64            = "amd64"
	darwin            = "darwin"
	linux             = "linux"
	freebsd           = "freebsd"
	windows           = "windows"
	bin               = "bin"
	gauge             = "gauge"
	deploy            = "deploy"
	CC                = "CC"
	nightlyDatelayout = "2006-01-02"
)

var deployDir = filepath.Join(deploy, gauge)

func runProcess(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if *verbose {
		log.Printf("Execute %v\n", cmd.Args)
	}
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

var buildMetadata string
var commitHash string

func getBuildVersion() string {
	if buildMetadata != "" {
		return fmt.Sprintf("%s.%s", version.CurrentGaugeVersion.String(), buildMetadata)
	}
	return version.CurrentGaugeVersion.String()
}

func compileGauge() {
	executablePath := getGaugeExecutablePath(gauge)
	ldflags := fmt.Sprintf("-X github.com/getgauge/gauge/version.BuildMetadata=%s -X github.com/getgauge/gauge/version.CommitHash=%s", buildMetadata, commitHash)
	args := []string{
		"build",
		fmt.Sprintf("-gcflags=-trimpath=%s", os.Getenv("GOPATH")),
		fmt.Sprintf("-asmflags=-trimpath=%s", os.Getenv("GOPATH")),
		"-ldflags", ldflags, "-o", executablePath,
	}
	runProcess("go", args...)
}

func runTests(coverage bool) {
	if coverage {
		runProcess("go", "test", "-covermode=count", "-coverprofile=count.out")
		if coverage {
			runProcess("go", "tool", "cover", "-html=count.out")
		}
	} else {
		if *verbose {
			runProcess("go", "test", "./...", "-v")
		} else {
			runProcess("go", "test", "./...")
		}
	}
}

// key will be the source file and value will be the target
func installFiles(files map[string]string, installDir string) {
	for src, dst := range files {
		base := filepath.Base(src)
		installDst := filepath.Join(installDir, dst)
		if *verbose {
			log.Printf("Install %s -> %s\n", src, installDst)
		}
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

func copyGaugeBinaries(installPath string) {
	files := make(map[string]string)
	files[getGaugeExecutablePath(gauge)] = ""
	installFiles(files, installPath)
}

func setEnv(envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}
}

var test = flag.Bool("test", false, "Run the test cases")
var coverage = flag.Bool("coverage", false, "Run the test cases and show the coverage")
var install = flag.Bool("install", false, "Install to the specified prefix")
var nightly = flag.Bool("nightly", false, "Add nightly build information")
var gaugeInstallPrefix = flag.String("prefix", "", "Specifies the prefix where gauge files will be installed")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles for all platforms windows, linux, darwin both x86 and x86_64")
var targetLinux = flag.Bool("target-linux", false, "Compiles for linux only, both x86 and x86_64")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")
var distro = flag.Bool("distro", false, "Create gauge distributable")
var verbose = flag.Bool("verbose", false, "Print verbose details")
var skipWindowsDistro = flag.Bool("skip-windows", false, "Skips creation of windows distributable on unix machines while cross platform compilation")

// Defines all the compile targets
// Each target name is the directory name
var (
	platformEnvs = []map[string]string{
		map[string]string{GOARCH: X86, GOOS: darwin, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: darwin, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: linux, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: linux, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: freebsd, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: freebsd, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: windows, CC: "i586-mingw32-gcc", CGO_ENABLED: "1"},
		map[string]string{GOARCH: X86_64, GOOS: windows, CC: "x86_64-w64-mingw32-gcc", CGO_ENABLED: "1"},
	}
	osDistroMap = map[string]distroFunc{windows: createWindowsDistro, linux: createLinuxPackage, freebsd: createLinuxPackage, darwin: createDarwinPackage}
)

func main() {
	flag.Parse()
	commitHash = revParseHead()
	if *nightly {
		buildMetadata = fmt.Sprintf("nightly-%s", time.Now().Format(nightlyDatelayout))
	}
	if *verbose {
		fmt.Println("Build: " + buildMetadata)
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

func revParseHead() string {
	if _, err := os.Stat(".git"); err != nil {
		return ""
	}
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	var hash bytes.Buffer
	cmd.Stdout = &hash
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s", strings.TrimSpace(hash.String()))
}

func filteredPlatforms() []map[string]string {
	filteredPlatformEnvs := platformEnvs[:0]
	for _, x := range platformEnvs {
		if *targetLinux {
			if x[GOOS] == linux {
				filteredPlatformEnvs = append(filteredPlatformEnvs, x)
			}
		} else {
			filteredPlatformEnvs = append(filteredPlatformEnvs, x)
		}
	}
	return filteredPlatformEnvs
}

func crossCompileGauge() {
	for _, platformEnv := range filteredPlatforms() {
		setEnv(platformEnv)
		if *verbose {
			log.Printf("Compiling for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
		}
		compileGauge()
	}
}

func installGauge() {
	updateGaugeInstallPrefix()
	copyGaugeBinaries(deployDir)
	if _, err := common.MirrorDir(filepath.Join(deployDir), filepath.Join(*gaugeInstallPrefix, bin)); err != nil {
		panic(fmt.Sprintf("Could not install gauge : %s", err))
	}
}

func createGaugeDistributables(forAllPlatforms bool) {
	if forAllPlatforms {
		for _, platformEnv := range filteredPlatforms() {
			setEnv(platformEnv)
			if *verbose {
				log.Printf("Creating distro for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
			}
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
	pName := packageName()
	distroDir, err := filepath.Abs(filepath.Join(deploy, pName))
	if err != nil {
		panic(err)
	}
	copyGaugeBinaries(distroDir)
	createZipFromUtil(deploy, pName, pName)
	os.RemoveAll(distroDir)
}

func createDarwinPackage() {
	distroDir := filepath.Join(deploy, gauge)
	copyGaugeBinaries(distroDir)
	if id := os.Getenv("OS_SIGNING_IDENTITY"); id == "" {
		log.Printf("No singning identity found . Executable won't be signed.")
	} else {
		runProcess("codesign", "-s", id, "--force", "--deep", filepath.Join(distroDir, gauge))
	}
	createZipFromUtil(deploy, gauge, packageName())
	os.RemoveAll(distroDir)
}

func createLinuxPackage() {
	distroDir := filepath.Join(deploy, packageName())
	copyGaugeBinaries(distroDir)
	createZipFromUtil(deploy, packageName(), packageName())
	os.RemoveAll(distroDir)
}

func packageName() string {
	return fmt.Sprintf("%s-%s-%s.%s", gauge, getBuildVersion(), getGOOS(), getPackageArchSuffix())
}

func removeUnwatedFiles(dir, currentOS string) error {
	fileList := []string{
		".DS_STORE",
		".localized",
		"$RECYCLE.BIN",
	}
	if currentOS == "windows" {
		fileList = append(fileList, []string{
			"desktop.ini",
			"Thumbs.db",
		}...)
	}
	for _, f := range fileList {
		err := os.RemoveAll(filepath.Join(dir, f))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func createZipFromUtil(dir, zipDir, pkgName string) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	absdir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	currentOS := getGOOS()

	windowsZipScript := filepath.Join(wd, "build", "create_windows_zipfile.ps1")

	err = removeUnwatedFiles(filepath.Join(dir, zipDir), currentOS)

	if err != nil {
		panic(fmt.Sprintf("Failed to cleanup unwanted file(s): %s", err))
	}

	err = os.Chdir(filepath.Join(dir, zipDir))
	if err != nil {
		panic(fmt.Sprintf("Failed to change directory: %s", err))
	}

	zipcmd := "zip"
	zipargs := []string{"-r", filepath.Join("..", pkgName+".zip"), "."}
	if currentOS == "windows" {
		zipcmd = "powershell.exe"
		zipargs = []string{"-noprofile", "-executionpolicy", "bypass", "-file", windowsZipScript, filepath.Join(absdir, zipDir), filepath.Join(absdir, pkgName+".zip")}
	}
	output, err := runCommand(zipcmd, zipargs...)
	if *verbose {
		fmt.Println(output)
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to zip: %s", err))
	}
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
