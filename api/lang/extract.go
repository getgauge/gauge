package lang

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/gauge/conceptExtractor"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const template = `# Specification
## scenario

`

type extractConceptInfo struct {
	Uri         lsp.DocumentURI `json:"uri"`
	Range       lsp.Range       `json:"range"`
	ConceptName string          `json:"conceptName"`
	ConceptFile lsp.DocumentURI `json:"conceptFile"`
}

func extractConcept(req *jsonrpc2.Request) (interface{}, error) {
	var params extractConceptInfo
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("Failed to parse request %s", err.Error())
		return nil, err
	}
	specFile := util.ConvertURItoFilePath(params.Uri)
	cptFileName := string(util.ConvertURItoFilePath(params.ConceptFile))
	textInfo := getSelectedTxtInfo(params.Uri, params.Range)
	steps, err := getStepsInRange(params.Uri, string(specFile), textInfo)
	if err != nil {
		return nil, err
	}

	edits, err := conceptExtractor.ExtractConceptWithoutSaving(&gm.Step{Name: params.ConceptName}, steps, cptFileName, textInfo)
	if err != nil {
		logger.APILog.Debugf("Failed to extract concpet. %v", err.Error())
		return nil, err
	}
	return createWorkSpaceEdits(edits), nil
}

func createWorkSpaceEdits(edits map[string]string) lsp.WorkspaceEdit {
	var result = lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{}}
	for file, edit := range edits {
		result.Changes[file] = []lsp.TextEdit{
			lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      0,
						Character: 0,
					},
					End: lsp.Position{
						Line:      len(strings.Split(edit, "\n")),
						Character: 0,
					},
				},
				NewText: edit,
			}}
	}
	return result
}

func getStepsInRange(uri lsp.DocumentURI, file string, info *gm.TextInfo) ([]*gm.Step, error) {
	text := openFilesCache.contentRange(uri, int(info.StartingLineNo), int(info.EndLineNo))
	specText := fmt.Sprintf("%s\n%s", template, strings.Join(text, "\n"))
	spec, res := new(parser.SpecParser).ParseSpecText(specText, file)
	if !res.Ok {
		logger.APILog.Debugf("Can not extract to cencpet.", res.Errors())
		return nil, fmt.Errorf("Can not extract to cencpet. Selected text contains invalid elements.")
	}
	return convertToAPIStep(spec.Steps()), nil
}

func convertToAPIStep(steps []*gauge.Step) []*gm.Step {
	apiSteps := []*gm.Step{}
	for _, step := range steps {
		s := &gm.Step{Name: step.LineText}
		if step.HasInlineTable {
			s.Table = formatter.FormatTable(&step.GetLastArg().Table)
		}
		apiSteps = append(apiSteps, s)
	}
	return apiSteps
}

func getSelectedTxtInfo(uri lsp.DocumentURI, r lsp.Range) *gm.TextInfo {
	return &gm.TextInfo{
		FileName:       string(util.ConvertURItoFilePath(uri)),
		StartingLineNo: int32(r.Start.Line + 1),
		EndLineNo:      int32(r.End.Line + 1),
	}
}
