package terminal

import (
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"log"
	"os"
)

type TerminalWriter struct {
	io.Writer
}

var (
	Stdout = &TerminalWriter{os.Stdout}
	Stderr = &TerminalWriter{os.Stderr}
)

func (w *TerminalWriter) checkOutput(s string) {
	if _, err := io.WriteString(w, s); err != nil {
		log.Fatal("Write to %v failed.", w)
	}
}

func (w *TerminalWriter) Color(syntax string) *TerminalWriter {
	escapeCode := color.Colorize(syntax)
	w.checkOutput(escapeCode)
	return w
}

func (w *TerminalWriter) Reset() *TerminalWriter {
	w.checkOutput(color.ResetCode)
	return w
}

func (w *TerminalWriter) Print(a ...interface{}) *TerminalWriter {
	fmt.Fprint(w, a...)
	return w
}

func (w *TerminalWriter) Nl(a ...interface{}) *TerminalWriter {
	length := 1
	if len(a) > 0 {
		length = a[0].(int)
	}
	for i := 0; i < length; i++ {
		w.checkOutput("\n")
	}
	return w
}

func (w *TerminalWriter) Colorf(format string, a ...interface{}) *TerminalWriter {
	w.checkOutput(color.Sprintf(format, a...))
	return w
}

func (w *TerminalWriter) Clear() *TerminalWriter {
	w.checkOutput("\033[2J")
	return w
}

func (w *TerminalWriter) ClearLine() *TerminalWriter {
	w.checkOutput("\033[2K")
	return w
}

func (w *TerminalWriter) Move(x, y int) *TerminalWriter {
	w.checkOutput(fmt.Sprintf("\033[%d;%dH", x, y))
	return w
}

func (w *TerminalWriter) Up(x int) *TerminalWriter {
	w.checkOutput(fmt.Sprintf("\033[%dA", x))
	return w
}

func (w *TerminalWriter) Down(x int) *TerminalWriter {
	w.checkOutput(fmt.Sprintf("\033[%dB", x))
	return w
}

func (w *TerminalWriter) Right(x int) *TerminalWriter {
	w.checkOutput(fmt.Sprintf("\033[%dC", x))
	return w
}

func (w *TerminalWriter) Left(x int) *TerminalWriter {
	w.checkOutput(fmt.Sprintf("\033[%dD", x))
	return w
}
