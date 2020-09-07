package logging

import (
	"github.com/fatih/color"
)

// Err is meant to log error messages
var Err = color.New(color.FgRed).PrintfFunc()
