package utils

import "fmt"

var quiet = false

func SetQuiet(quietOn bool) {
	quiet = quietOn
}

func QuietPrintln(msg string) {
	if !quiet {
		fmt.Println(msg)
	}
}
