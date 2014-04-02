package main

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
)

func printSuccess(text ...string) {
	ct.ChangeColor(ct.Green, false, ct.None, false)
	for _, value := range text {
		fmt.Println(value)
	}
	ct.ResetColor()
}

func printFailure(text ...string) {
	ct.ChangeColor(ct.Red, false, ct.None, false)
	for _, value := range text {
		fmt.Println(value)
	}
	ct.ResetColor()
}
