package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/gauge"
	"github.com/spf13/cobra"
)

func CreateStepCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run-step [spec-file:line-number]",
		Short:   "Run a single step",
		Long:    `Run a single step.`,
		Example: `gauge run-step specs/example.spec:14`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.SetProjectRoot(args); err != nil {
				return fmt.Errorf("failed to set project root: %w", err)
			}
			return executeStep(cmd)
		},
		DisableAutoGenTag: true,
	}
	return cmd
}

func executeStep(cmd *cobra.Command) error {
	loadEnvAndReinitLogger(cmd)
	ensureScreenshotsDir()

	args := cmd.Flags().Args()

	if err := validateStepArgs(args); err != nil {
		return fmt.Errorf("invalid step arguments: %w", err)
	}

	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument in format 'specfile:linenumber', got %d", len(args))
	}

	parts := strings.Split(args[0], ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Expected 'specfile:linenumber', got '%s'", args[0])
	}

	specFile := parts[0]
	lineNo, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid line number '%s': %w", parts[1], err)
	}

	argsString := strings.Join(os.Args[1:], " ")

	spec := getSpecDir(specFile)
	rerun.SaveSingleStepState(argsString, spec)

	if !skipCommandSave {
		rerun.WritePrevArgs(os.Args)
	}

	installMissingPlugins(installPlugins, false)

	step := &gauge.Step{
		Value:     strings.Join(args, " "),
		LineText:  strings.Join(args, " "),
		LineNo:    lineNo,
		IsConcept: false,
	}

	exitCode := execution.ExecuteStep(step)
	if failSafe && exitCode != execution.ParseFailed {
		exitCode = 0
	}

	os.Exit(exitCode)
	return nil
}

func init() {
	stepCmd := CreateStepCommand()
	GaugeCmd.AddCommand(stepCmd)
}

func validateStepArgs(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Please provide a spec file and line number in format: specs/filename.spec:line-number")
	}

	// get the spec directory from commons
	matched, err := regexp.MatchString(`^specs/[^:]+\.spec:\d+$`, args[0])
	if err != nil {
		return fmt.Errorf("Error validating format")
	}
	if !matched {
		return fmt.Errorf("Invalid format. Expected specs/filename.spec:line-number")
	}

	return nil
}

func getStepFromSpecFile(specFile string, lineNo int) (*gauge.Step, error) {
	// Read the spec file
	content, err := os.ReadFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	// Parse the file to locate the step at the given line number
	lines := strings.Split(string(content), "\n")
	if lineNo <= 0 || lineNo > len(lines) {
		return nil, fmt.Errorf("line number %d is out of range (file has %d lines)", lineNo, len(lines))
	}

	// Get the step text (line numbers in spec files are 1-based)
	stepText := strings.TrimSpace(lines[lineNo-1])

	// Create and return the step
	return &gauge.Step{
		Value:     stepText,
		LineText:  stepText,
		LineNo:    lineNo,
		IsConcept: false,
		FileName:  specFile,
	}, nil
}
