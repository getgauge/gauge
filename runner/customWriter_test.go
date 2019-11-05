package runner

import (
	"bytes"
	"testing"
	"time"
)

func TestCustomWriterShouldExtractPortNumberFromStdout(t *testing.T) {
	portChan := make(chan string)
	var b bytes.Buffer
	w := customWriter{
		file: &b,
		port: portChan,
	}

	go func() {
		n, err := w.Write([]byte("Listening on port:23454"))
		if n <= 0 || err != nil {
			t.Errorf("failed to write port information")
		}
	}()

	select {
	case port := <-portChan:
		close(portChan)
		if port != "23454" {
			t.Errorf("Expected:%s\nGot     :%s", "23454", port)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("Timed out!! Failed to get port info.")
	}
}
