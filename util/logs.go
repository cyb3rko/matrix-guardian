package util

import "fmt"

func Print(message string) {
	fmt.Print(message + "\n")
}

func Printf(message string, a ...interface{}) {
	fmt.Printf(message+"\n", a...)
}
