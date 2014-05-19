// The colors package provide a simple way to bring colorful charcaters to terminal interface.
//
// This example will output the text with a Blue foreground and a Black background
//      color.Println("@{bK}Example Text")
//
// This one will output the text with a red foreground
//      color.Println("@rExample Text")
//
// This one will escape the @
//      color.Println("@@")
//
// Full color syntax code
//      @{rgbcmykwRGBCMYKW}  foreground/background color
//        r/R:  Red
//        g/G:  Green
//        b/B:  Blue
//        c/C:  Cyan
//        m/M:  Magenta
//        y/Y:  Yellow
//        k/K:  Black
//        w/W:  White
//      @{|}  Reset format style
//      @{!./_} Bold / Dim / Italic / Underline
//      @{^&} Blink / Fast blink
//      @{?} Reverse the foreground and background color
//      @{-} Hide the text
// Note some of the functions are not widely supported, like "Fast blink" and "Italic".
package color

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
)

const (
	EscapeChar = '@'       // Escape character for color syntax
	ResetCode  = "\033[0m" // Short for reset to default style
)

// Mapping from character to concrete escape code.
var codeMap = map[int]int{
	'|': 0,
	'!': 1,
	'.': 2,
	'/': 3,
	'_': 4,
	'^': 5,
	'&': 6,
	'?': 7,
	'-': 8,

	'k': 30,
	'r': 31,
	'g': 32,
	'y': 33,
	'b': 34,
	'm': 35,
	'c': 36,
	'w': 37,
	'd': 39,

	'K': 40,
	'R': 41,
	'G': 42,
	'Y': 43,
	'B': 44,
	'M': 45,
	'C': 46,
	'W': 47,
	'D': 49,
}

// Compile color syntax string like "rG" to escape code.
func Colorize(x string) string {
	attr := 0
	fg := 39
	bg := 49

	for _, key := range x {
		c, ok := codeMap[int(key)]
		switch {
		case !ok:
			log.Printf("Wrong color syntax: %c", key)
		case 0 <= c && c <= 8:
			attr = c
		case 30 <= c && c <= 37:
			fg = c
		case 40 <= c && c <= 47:
			bg = c
		}
	}
	return fmt.Sprintf("\033[%d;%d;%dm", attr, fg, bg)
}

// Handle state after meeting one '@'
func compileColorSyntax(input, output *bytes.Buffer) {
	i, _, err := input.ReadRune()
	if err != nil {
		// EOF got
		log.Print("Parse failed on color syntax")
		return
	}

	switch i {
	default:
		output.WriteString(Colorize(string(i)))
	case '{':
		color := bytes.NewBufferString("")
		for {
			i, _, err := input.ReadRune()
			if err != nil {
				log.Print("Parse failed on color syntax")
				break
			}
			if i == '}' {
				break
			}
			color.WriteRune(i)
		}
		output.WriteString(Colorize(color.String()))
	case EscapeChar:
		output.WriteRune(EscapeChar)
	}
}

// Compile the string and replace color syntax with concrete escape code.
func compile(x string) string {
	if x == "" {
		return ""
	}

	input := bytes.NewBufferString(x)
	output := bytes.NewBufferString("")

	for {
		i, _, err := input.ReadRune()
		if err != nil {
			break
		}
		switch i {
		default:
			output.WriteRune(i)
		case EscapeChar:
			compileColorSyntax(input, output)
		}
	}
	return output.String()
}

// Compile multiple values, only do compiling on string type.
func compileValues(a *[]interface{}) {
	for i, x := range *a {
		if str, ok := x.(string); ok {
			(*a)[i] = compile(str)
		}
	}
}

// Similar to fmt.Print, will reset the color at the end.
func Print(a ...interface{}) (int, error) {
	a = append(a, ResetCode)
	compileValues(&a)
	return fmt.Print(a...)
}

// Similar to fmt.Println, will reset the color at the end.
func Println(a ...interface{}) (int, error) {
	a = append(a, ResetCode)
	compileValues(&a)
	return fmt.Println(a...)
}

// Similar to fmt.Printf, will reset the color at the end.
func Printf(format string, a ...interface{}) (int, error) {
	format += ResetCode
	format = compile(format)
	return fmt.Printf(format, a...)
}

// Similar to fmt.Fprint, will reset the color at the end.
func Fprint(w io.Writer, a ...interface{}) (int, error) {
	a = append(a, ResetCode)
	compileValues(&a)
	return fmt.Fprint(w, a...)
}

// Similar to fmt.Fprintln, will reset the color at the end.
func Fprintln(w io.Writer, a ...interface{}) (int, error) {
	a = append(a, ResetCode)
	compileValues(&a)
	return fmt.Fprintln(w, a...)
}

// Similar to fmt.Fprintf, will reset the color at the end.
func Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	format += ResetCode
	format = compile(format)
	return fmt.Fprintf(w, format, a...)
}

// Similar to fmt.Sprint, will reset the color at the end.
func Sprint(a ...interface{}) string {
	a = append(a, ResetCode)
	compileValues(&a)
	return fmt.Sprint(a...)
}

// Similar to fmt.Sprintf, will reset the color at the end.
func Sprintf(format string, a ...interface{}) string {
	format += ResetCode
	format = compile(format)
	return fmt.Sprintf(format, a...)
}

// Similar to fmt.Errorf, will reset the color at the end.
func Errorf(format string, a ...interface{}) error {
	return errors.New(Sprintf(format, a...))
}
