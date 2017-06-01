package debug

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"testing"
)

type writer struct {
	t *testing.T
	c string
	h int
}

func (w *writer) Header() http.Header {
	return nil
}

func (w *writer) Write(b []byte) (int, error) {
	w.c += string(b)
	return len(b), nil
}

func (w *writer) WriteHeader(i int) {
	w.h = i
}

func TestHandleWithGetRequest(test *testing.T) {
	w := &writer{t: test}
	r := &http.Request{Method: http.MethodGet}
	infos = []processInfo{{Port: "1234", Cwd: "CWD", Pid: 1234}}
	t, _ = template.New("html").Parse(html)

	handle(w, r)

	var tpl bytes.Buffer
	t.Execute(&tpl, infos)
	expectedMessage := tpl.String()
	if w.c != expectedMessage || w.h != 0 {
		test.Errorf("Test failed: expected { message: %s, header: %d}, actual { message: %s, header: %d}", expectedMessage, 0, w.c, w.h)
	}
}

func TestHandleWithPostRequestWithNoPort(test *testing.T) {
	w := &writer{t: test}
	r := &http.Request{Method: http.MethodPost, URL: &url.URL{RawQuery: ""}}

	handle(w, r)

	expectedMessage := "No port provided."
	if w.c != expectedMessage || w.h != http.StatusBadRequest {
		test.Errorf("Test failed: expected { message: %s, header: %d}, actual { message: %s, header: %d}", expectedMessage, http.StatusBadRequest, w.c, w.h)
	}
}

func TestHandleWithPostRequestWithWrongPort(test *testing.T) {
	w := &writer{t: test}
	r := &http.Request{Method: http.MethodPost, URL: &url.URL{RawQuery: "port=1567777"}}

	handle(w, r)

	expectedMessage := "dial tcp: address 1567777: invalid port"
	if w.c != expectedMessage || w.h != http.StatusInternalServerError {
		test.Errorf("Test failed: expected { message: %s, header: %d}, actual { message: %s, header: %d}", expectedMessage, http.StatusInternalServerError, w.c, w.h)
	}
}

func TestHandleWithInvalidRequestMethod(test *testing.T) {
	w := &writer{t: test}
	r := &http.Request{Method: http.MethodPut}

	handle(w, r)

	expectedMessage := "404 Not Found!"
	if w.c != expectedMessage || w.h != http.StatusNotFound {
		test.Errorf("Test failed: expected { message: %s, header: %d}, actual { message: %s, header: %d}", expectedMessage, http.StatusNotFound, w.c, w.h)
	}
}

func TestGetResponseWithInvalidRequestMethod(test *testing.T) {
	w := &writer{t: test}
	r := &http.Request{Method: http.MethodPut}

	getResponse(w, r, nil, http.MethodPost)

	expectedMessage := "404 Not Found!"
	if w.c != expectedMessage || w.h != http.StatusNotFound {
		test.Errorf("Test failed: expected { message: %s, header: %d}, actual { message: %s, header: %d}", expectedMessage, http.StatusNotFound, w.c, w.h)
	}

}

func TestGetResponseWithValidRequestMethod(test *testing.T) {
	w := &writer{t: test}
	r := &http.Request{Method: http.MethodPost}
	a = nil

	getResponse(w, r, nil, http.MethodPost)

	expectedMessage := "Before making api requests, please connect to a Gauge api process."
	if w.c != expectedMessage || w.h != http.StatusBadRequest {
		test.Errorf("Test failed: expected { message: %s, header: %d}, actual { message: %s, header: %d}", expectedMessage, http.StatusBadRequest, w.c, w.h)
	}

}
