package util

import "fmt"

func Print(message string) {
	fmt.Println(message)
}

func Printf(message string, a ...interface{}) {
	fmt.Printf(message+"\n", a...)
}
