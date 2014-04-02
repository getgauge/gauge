// This file is part of twist
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

const (
	typeScenario     = 0x1
	typeWorkflow     = 0x2
	typePlainText    = 0x4
	typeWorkflowStep = 0x8
)

const (
	scenarioPrefix = "Scenario:"
	workflowPrefix = "Workflow:"
)

// Represents a single token
type token struct {
	kind   int
	lineNo int
	value  string // processed value
	line   string // as appeared to the parser
	args   []string
}

type parser struct {
	scanner      *bufio.Scanner
	lineNo       int
	currentState int // a bitwise field represents all the states parser holds
	tokens       []*token
}

// Contains any parsing error
type syntaxError struct {
	lineNo  int
	colNo   int
	line    string
	message string
}

func (se *syntaxError) Error() string {
	return se.message
}

func (t token) String() string {
	var kind string
	switch t.kind {
	case typeScenario:
		kind = "Scenario"
		break
	case typeWorkflow:
		kind = "Workflow"
		break
	case typeWorkflowStep:
		kind = "WorkflowStep"
		break
	}

	return fmt.Sprintf("kind: %s, lineNo: %d, value: %s, args: %v\n", kind, t.lineNo, t.value, t.args)
}

// Returns the name portion in a terminal symbol.
// Eg: getNameForTerminalSymbol("Scenario: some name") -> "some name"
func getNameForTerminalSymbol(input string) string {
	return strings.Trim(input[strings.Index(input, ":")+1:], " ")
}

// Advances parser to the next line
// Returns empty text and false if the lexer reaches EOF
func (p *parser) nextLine() (string, bool) {
	scanned := p.scanner.Scan()
	if scanned {
		p.lineNo++
		return p.scanner.Text(), true
	}

	if err := p.scanner.Err(); err != nil {
		panic(err)
	}

	return "", false
}

func (p *parser) isInState(flag int) bool {
	return (p.currentState & flag) > 0
}

func (p *parser) exitIfInWorkflow() {
	if p.isInState(typeWorkflow) {
		p.currentState &= ^(typeWorkflow)
	}
}

// We can accept a scenario if we haven't seen one before
func canAcceptScenario(p *parser, t *token) (bool, error) {
	if p.isInState(typeScenario) {
		return false, &syntaxError{
			lineNo:  p.lineNo,
			message: "One file should contain only one scenario definition",
		}
	}

	return true, nil
}

func canAcceptWorkflow(p *parser, t *token) (bool, error) {
	if !p.isInState(typeScenario) {
		return false, &syntaxError{
			lineNo:  p.lineNo,
			message: "A workflow should be inside a scenario",
		}
	}

	if p.isInState(typeWorkflow) {
		return false, &syntaxError{
			lineNo:  p.lineNo,
			message: "Nested workflow is not supported",
		}
	}

	return true, nil
}

func canAcceptToken(p *parser, t *token) (bool, error) {
	switch t.kind {
	case typeScenario:
		return canAcceptScenario(p, t)
	case typeWorkflow:
		return canAcceptWorkflow(p, t)
	case typeWorkflowStep:
		return true, nil
	}

	return false, nil
}

/* Reads the line from left to right, and extracts the arguments
   Arguments are rewrittern as placeholder like {arg0} {arg1} etc
   and argument values will be added to the args array in the resultant token
*/
func makeWokflowStepToken(p *parser, line string) (*token, error) {
	const (
		inDefault = iota
		inQuotes
		inExitingQuotes
		inEscape
	)
	const (
		quotes = '"'
		escape = '\\'
	)

	// Using an empty struct as a value here as I don't care about the value
	// Empty structs won't allocate any space too
	reservedChars := map[rune]struct{}{'{': {}, '}': {}}
	var stepText, argText bytes.Buffer
	var args []string
	curBuffer := func(state int) *bytes.Buffer {
		if state == inQuotes {
			return &argText
		} else {
			return &stepText
		}
	}

	currentState := inDefault
	lastState := -1
	for colNo, element := range line {
		if currentState == inEscape {
			currentState = lastState
		} else if element == escape {
			lastState = currentState
			currentState = inEscape
			continue
		} else if element == quotes {
			if currentState == inDefault {
				currentState = inQuotes
				continue
			} else {
				currentState = inExitingQuotes
			}
		} else if _, ok := reservedChars[element]; ok {
			return nil,
				&syntaxError{lineNo: p.lineNo, line: line, colNo: colNo + 1, message: fmt.Sprintf("'%c' is a reserved character and should be escaped", element)}
		}

		if currentState == inExitingQuotes {
			stepText.WriteString(fmt.Sprintf("{arg%d}", len(args)))
			args = append(args, argText.String())
			argText.Reset()
			currentState = inDefault
		} else {
			curBuffer(currentState).WriteRune(element)
		}
	}

	// If it is a valid step, the state should be default when the control reaches here
	if currentState == inQuotes {
		return nil, &syntaxError{lineNo: p.lineNo, line: line, message: "String not terminated"}
	}

	return &token{kind: typeWorkflowStep, lineNo: p.lineNo, line: line, value: strings.TrimSpace(stepText.String()), args: args}, nil
}

func (p *parser) accept(t *token) error {
	ok, err := canAcceptToken(p, t)
	if err != nil {
		return err
	}

	if ok {
		p.tokens = append(p.tokens, t)
		p.currentState |= t.kind
	}

	return nil
}

func parse(input string) ([]*token, error) {
	p := &parser{
		scanner: bufio.NewScanner(strings.NewReader(input)),
	}

	err := p.run()
	if err != nil {
		return nil, err
	}

	return p.tokens, nil
}

func (p *parser) run() error {
	for line, hasLine := p.nextLine(); hasLine; line, hasLine = p.nextLine() {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, scenarioPrefix) {
			scenarioName := getNameForTerminalSymbol(trimmedLine)
			if len(scenarioName) == 0 {
				return &syntaxError{lineNo: p.lineNo, line: line, message: "Scenario should have a name"}
			}
			token := &token{kind: typeScenario, lineNo: p.lineNo, line: trimmedLine, value: scenarioName}
			err := p.accept(token)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(trimmedLine, workflowPrefix) {
			workflowName := getNameForTerminalSymbol(trimmedLine)
			token := &token{kind: typeWorkflow, lineNo: p.lineNo, line: trimmedLine, value: workflowName}
			err := p.accept(token)
			if err != nil {
				return err
			}
		} else {
			if len(trimmedLine) == 0 {
				// A new empty line breaks the workflow scope
				p.exitIfInWorkflow()
				continue
			}

			// We skip all plain text. A plain text inside a workflow is an executable step
			if !p.isInState(typeWorkflow) {
				continue
			}

			token, err := makeWokflowStepToken(p, trimmedLine)
			if err != nil {
				return err
			}

			err = p.accept(token)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
