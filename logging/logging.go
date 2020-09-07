package logging

import (
	"log"

	"github.com/fatih/color"
)

var err = color.New(color.FgRed)
var warning = color.New(color.FgYellow)
var success = color.New(color.FgGreen)

// Err is meant to log error messages
func Err(message ...interface{}) {
	err.Println(message...)
}

// Warning is meant to log warning messages
func Warning(message ...interface{}) {
	warning.Println(message...)
}

// Success is meant to log success messages
func Success(message ...interface{}) {
	success.Println(message...)
}

// Info is meant for informational messages
func Info(message ...interface{}) {
	log.Println(message...)
}
