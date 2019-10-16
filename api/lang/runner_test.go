package lang

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type MockContext struct {
}

func (ctx *MockContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (ctx *MockContext) Done() <-chan struct{} {
	return nil
}

func (ctx *MockContext) Err() error {
	return nil
}

func (ctx *MockContext) Value(key interface{}) interface{} {
	return nil
}

type MockConn struct {
	method  string
	params  interface{}
	options []jsonrpc2.CallOption
}

func (conn *MockConn) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error {
	return nil
}

func (conn *MockConn) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	conn.method = method
	conn.params = params
	conn.options = opt
	return nil
}

func (conn *MockConn) Close() error {
	return nil
}

func TestRunnerCompatibilityWarning(t *testing.T) {
	conn := &MockConn{}
	ctx := &MockContext{}
	expectedMethod := "window/showMessage"
	expectedParams := lsp.ShowMessageParams{
		Type:    lsp.Warning,
		Message: "Current gauge language runner is not compatible with gauge LSP. Some of the editing feature will not work as expected",
	}
	err := informRunnerCompatibility(ctx, conn)
	if err != nil {
		t.Error(err.Error())
	}
	if conn.method != expectedMethod {
		t.Errorf("Expected: %s\nGot: %s", expectedMethod, conn.method)
	}
	if !reflect.DeepEqual(conn.params, expectedParams) {
		t.Errorf("\nExpected: %v\nGot: %v", expectedParams, conn.params)
	}
}

func TestRunnerCompatibilityWarningWhenRunnerSupportLSP(t *testing.T) {
	lRunner.lspID = "foo"
	conn := &MockConn{}
	ctx := &MockContext{}
	err := informRunnerCompatibility(ctx, conn)
	if err != nil {
		t.Error(err)
	}
	if conn.method != "" {
		t.Errorf("\nExpected: %s\nGot: %s", "", conn.method)
	}
}
