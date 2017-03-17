package debug

import (
	"net/http"

	"encoding/json"

	"strconv"

	"strings"

	m "github.com/getgauge/gauge/gauge_messages"
)

var a *api
var err error

type handler func(http.ResponseWriter, *http.Request)

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err = t.Execute(w, infos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(toBytes(err.Error()))
			return
		}
	case http.MethodPost:
		port := r.URL.Query().Get("port")
		if port == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(toBytes("No port provided."))
			return
		}
		a, err = newAPI(localhost, port)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(toBytes(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toBytes("Connected to port: " + port))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write(toBytes("404 Not Found!"))
	}
}

func projectRootRequest(w http.ResponseWriter, r *http.Request) {
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_GetProjectRootRequest, ProjectRootRequest: &m.GetProjectRootRequest{}}, http.MethodGet)
}

func installationRootRequest(w http.ResponseWriter, r *http.Request) {
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_GetInstallationRootRequest, InstallationRootRequest: &m.GetInstallationRootRequest{}}, http.MethodGet)
}

func libPathRequest(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("pluginName")
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_GetLanguagePluginLibPathRequest, LibPathRequest: &m.GetLanguagePluginLibPathRequest{Language: lang}}, http.MethodGet)
}

func allStepsRequest(w http.ResponseWriter, r *http.Request) {
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_GetAllStepsRequest, AllStepsRequest: &m.GetAllStepsRequest{}}, http.MethodGet)
}

func allConceptsRequest(w http.ResponseWriter, r *http.Request) {
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_GetAllConceptsRequest, AllConceptsRequest: &m.GetAllConceptsRequest{}}, http.MethodGet)
}

func specsRequest(w http.ResponseWriter, r *http.Request) {
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_SpecsRequest, SpecsRequest: &m.SpecsRequest{}}, http.MethodGet)
}

func formatSpecsRequest(w http.ResponseWriter, r *http.Request) {
	specs := strings.Split(r.URL.Query().Get("specs"), "\n")
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_FormatSpecsRequest, FormatSpecsRequest: &m.FormatSpecsRequest{Specs: specs}}, http.MethodPost)
}

func refactoringRequest(w http.ResponseWriter, r *http.Request) {
	oldStep := r.URL.Query().Get("old")
	newStep := r.URL.Query().Get("new")
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_PerformRefactoringRequest, PerformRefactoringRequest: &m.PerformRefactoringRequest{OldStep: oldStep, NewStep: newStep}}, http.MethodPost)
}

func stepValueRequest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	step := q.Get("stepText")
	hasInlineTable, err := strconv.ParseBool(q.Get("hasInlineTable"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(toBytes(err.Error()))
		return
	}
	getResponse(w, r, &m.APIMessage{MessageType: m.APIMessage_GetStepValueRequest, StepValueRequest: &m.GetStepValueRequest{StepText: step, HasInlineTable: hasInlineTable}}, http.MethodGet)
}

func getResponse(w http.ResponseWriter, r *http.Request, msg *m.APIMessage, method string) {
	switch r.Method {
	case method:
		if a == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(toBytes("Before making api requests, please connect to a Gauge api process."))
			return
		}
		resp, err := a.getResponse(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(toBytes(err.Error()))
			return
		}
		e := json.NewEncoder(w)
		e.SetEscapeHTML(false)
		e.SetIndent("", "\t")
		err = e.Encode(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(toBytes(err.Error()))
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write(toBytes("404 Not Found!"))
	}
}

func toBytes(s string) []byte {
	return []byte(s)
}
