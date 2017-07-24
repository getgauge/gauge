// +build !windows

package goterminal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Flush moves the cursor to location where last write started and clears the text written using previous Write.
func (w *Writer) Clear() {
	for i := 0; i < w.lineCount; i++ {
		fmt.Fprintf(w.Out, "%c[%dA", esc, 0) // move the cursor up
		fmt.Fprintf(w.Out, "%c[2K\r", esc)   // clear the line
	}
	w.lineCount = 0
}

// GetTermDimensions returns the width and height of the current terminal
func (w *Writer) GetTermDimensions() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 25
	}
	splits := strings.Split(strings.Trim(string(out), "\n"), " ")
	height, err := strconv.ParseInt(splits[0], 0, 0)
	width, err1 := strconv.ParseInt(splits[1], 0, 0)
	if err != nil || err1 != nil {
		return 80, 25
	}
	return int(width), int(height)
}
