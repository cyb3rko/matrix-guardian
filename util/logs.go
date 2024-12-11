package util

import "fmt"

func Printf(message string, a ...interface{}) {
	fmt.Printf(message+"\n", a...)
}
