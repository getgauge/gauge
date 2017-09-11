package goterminal

import (
	"bytes"
	"io"
	"sync"
)

const esc = 27

// termWidth marks the boundary of UI, beyond which text written should go to next line.
var termWidth int

// Writer represents the IO writer which updates the UI and holds the buffer.
type Writer struct {
	Out       io.Writer
	Buf       bytes.Buffer
	lineCount int
	mtx       sync.Mutex
}

// New returns a new instance of the Writer. It initializes the terminal width and buffer.
func New(out io.Writer) *Writer {
	writer := &Writer{Out: out}
	if termWidth == 0 {
		termWidth, _ = writer.GetTermDimensions()
	}
	return writer
}

// Reset resets the Writer.
func (w *Writer) Reset() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.Buf.Reset()
	w.lineCount = 0
}

// Print writes the buffer contents to Out and resets the buffer.
// It stores the number of lines to go up the Writer in the Writer.lineCount.
func (w *Writer) Print() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	// do nothing if buffer is empty
	if len(w.Buf.Bytes()) == 0 {
		return nil
	}

	var currentLine bytes.Buffer
	for _, b := range w.Buf.Bytes() {
		if b == '\n' {
			w.lineCount++
			currentLine.Reset()
		} else {
			currentLine.Write([]byte{b})
			if currentLine.Len() > termWidth {
				w.lineCount++
				currentLine.Reset()
			}
		}
	}
	_, err := w.Out.Write(w.Buf.Bytes())
	w.Buf.Reset()
	return err
}

func (w *Writer) Write(b []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.Buf.Write(b)
}
