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
	deploy             = "deploy"
	installShellScript = "install.sh"
)

var deployDir = filepath.Join(deploy, gauge)

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

func compileGauge() {
	runProcess("go", "build", "-o", getGaugeExecutablePath())
}

func runTests(packageName string, coverage bool) {
	if coverage {
		runProcess("go", "test", "-covermode=count", "-coverprofile=count.out")
		if coverage {
			runProcess("go", "tool", "cover", "-html=count.out")
		}
	} else {
		runProcess("go", "test")
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
			err = mirrorDir(src, installDst)
		} else {
			err = mirrorFile(src, filepath.Join(installDst, base))
		}
		if err != nil {
			panic(err)
		}
	}
}

func copyGaugeFiles(installPath string) {
	files := make(map[string]string)
	files[getGaugeExecutablePath()] = bin
	files[filepath.Join("skel", "hello_world.spec")] = filepath.Join("share", gauge, "skel")
	files[filepath.Join("skel", "default.properties")] = filepath.Join("share", gauge, "skel", "env")
	files[filepath.Join("skel", "gauge.properties")] = filepath.Join("share", gauge)
	files = addInstallScripts(files)
	installFiles(files, installPath)
}

func addInstallScripts(files map[string]string) map[string]string {
	if (getOS() == darwin || getOS() == linux) && (*distroVersion != "") {
		files[filepath.Join("build", "install", installShellScript)] = ""
	} else if getOS() == windows {
		files[filepath.Join("build", "install", "windows", "plugin-install.bat")] = ""
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
	if err := mirrorFile(srcFile, destFile); err != nil {
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
var gaugeInstallPrefix = flag.String("prefix", "", "Specifies the prefix where gauge files will be installed")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles for all platforms windows, linux, darwin both x86 and x86_64")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")
var distroVersion = flag.String("distro", "", "Create gauge distributable for specified version")
var skipWindowsDistro = flag.Bool("skip-windows", false, "Skips creation of windows distributable on unix machines while cross platform compilation")

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
		map[string]string{GOARCH: X86, GOOS: windows, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: windows, CGO_ENABLED: "0"},
	}
)

func main() {
	flag.Parse()
	if *test {
		runTests(gauge, *coverage)
	} else if *install {
		installGauge()
	} else if *distroVersion != "" {
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
	if err := mirrorDir(deployDir, *gaugeInstallPrefix); err != nil {
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

func createDistro() {
	if getOS() == windows {
		if !*skipWindowsDistro {
			createWindowsInstaller()
		}
	} else {
		createZipPackage()
	}
}

func createWindowsInstaller() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", gauge, *distroVersion, getOS(), getArch())
	distroDir, err := filepath.Abs(filepath.Join(deploy, packageName))
	if err != nil {
		panic(err)
	}
	copyGaugeFiles(distroDir)
	runProcess("makensis.exe",
		fmt.Sprintf("/DPRODUCT_VERSION=%s", *distroVersion),
		fmt.Sprintf("/DGAUGE_DISTRIBUTABLES_DIR=%s", distroDir),
		fmt.Sprintf("/DOUTPUT_FILE_NAME=%s.exe", filepath.Join(filepath.Dir(distroDir), packageName)),
		filepath.Join("build", "install", "windows", "gauge-install.nsi"))
	os.RemoveAll(distroDir)
}

func createZipPackage() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", gauge, *distroVersion, getOS(), getArch())
	distroDir := filepath.Join(deploy, packageName)
	copyGaugeFiles(distroDir)
	createZip(deploy, packageName)
	os.RemoveAll(distroDir)
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

func getGaugeExecutablePath() string {
	return filepath.Join(getBinDir(), getExecutableName())
}

func getBinDir() string {
	if *binDir != "" {
		return *binDir
	}
	return filepath.Join(bin, fmt.Sprintf("%s_%s", getGOOS(), getGOARCH()))
}

func getExecutableName() string {
	if getGOOS() == windows {
		return gauge + ".exe"
	}
	return gauge
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

func getArch() string {
	arch := os.Getenv(GOARCH)
	if arch == X86 {
		return "x86"
	}
	return "x86_64"
}

func getOS() string {
	os := os.Getenv(GOOS)
	if os == "" {
		return runtime.GOOS

	}
	return os
}
