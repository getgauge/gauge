package main

import (
	"github.com/wsxiaoys/terminal"
	"github.com/wsxiaoys/terminal/color"
)

func main() {
	terminal.Stdout.Color("y").
		Print("Hello world").Nl().
		Reset().
		Colorf("@{kW}Hello world\n")

	color.Print("@rHello world")
}
