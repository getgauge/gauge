package cmd

import (
	"bytes"
	"fmt"
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
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	cmd.Stdin = strings.NewReader(specsInput)
	err := cmd.Run()
	if err != nil {
		t.Fatalf("%s process ran with err %v, want exit status 0. Stdout:\n%s", os.Args, err, stdout.Bytes())
	}
}
