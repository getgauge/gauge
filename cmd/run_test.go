package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/spf13/cobra"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestHandleRepeatCommand(c *C) {
	path, _ := filepath.Abs("_testData")
	config.ProjectRoot = path
	args := []string{"gauge", "run", "specs"}
	cmd := &cobra.Command{}

	var execFlowFlag = false
	executeCmd = func(cmd *cobra.Command, lastState []string) {
		execFlowFlag = true
	}

	handleRepeatCommand(cmd, args)
	commandWritten := readPrevCmd()
	c.Assert(commandWritten.Command, DeepEquals, args)

	repeat = true
	handleRepeatCommand(cmd, args)
	c.Assert(execFlowFlag, Equals, true)
	commandWritten = readPrevCmd()
	c.Assert(commandWritten.Command, DeepEquals, args)

	repeat = false
	prevFailed = true
	args2 := []string{"something", "else"}
	handleRepeatCommand(cmd, args2)
	commandWritten = readPrevCmd()
	c.Assert(commandWritten.Command, DeepEquals, args)

	os.RemoveAll(path)
}
