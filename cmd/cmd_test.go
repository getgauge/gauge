package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGetSpecsDirFromStdin(t *testing.T) {
	specsInput := "specs/example1.spec\nspecs/example5.spec\nspecs/example2.spec"
	if os.Getenv("TEST_EXITS") == "1" {
		s := getSpecsDir([]string{})
		if len(s) <= 0 {
			t.Error("No specs found in stdin")
		}
		got := strings.Join(s, "\n")
		if got != specsInput {
			t.Fatalf("Expected \"%s\", got \"%s\"", specsInput, got)
		}
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal("unable to get stdinpipe for subprocess")
	}
	go func() {
		defer stdin.Close()
		_, err = io.WriteString(stdin, specsInput)
		if err != nil {
			panic("unable to get stdinpipe for subprocess")
		}
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s process ran with err %v, want exit status 0. Stdout:\n%s", os.Args, err, out)
	}
}
