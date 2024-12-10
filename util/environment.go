package util

import (
	"os"
	"strings"
)

func GetEnv(name string, trim bool, lowercase bool) string {
	value := os.Getenv(name)
	if trim {
		value = strings.TrimSpace(value)
	}
	if lowercase {
		value = strings.ToLower(value)
	}
	return value
}
