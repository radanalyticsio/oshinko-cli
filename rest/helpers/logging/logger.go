// Package logging provides primitives for configuring and creating logs.
package logging

import (
	"errors"
	"log"
	"os"

	"github.com/radanalyticsio/oshinko-cli/rest/helpers/flags"
)

var logger *log.Logger
var logfile *os.File

func getStderrLogger() *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

// GetLogger returns the active logging object for creating new messages.
func GetLogger() *log.Logger {
	if logger == nil {
		logger = getStderrLogger()
	}
	return logger
}

// SetLoggerFile changes the logging destination to a file. If a logger
// is already active, or the file cannot be opened, then the stderr logger
// will be used and an error will be returned.
func SetLoggerFile(filename string) (err error) {
	if logger == nil {
		var fp *os.File
		fp, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger = getStderrLogger()
		} else {
			logger = log.New(fp, "", log.LstdFlags)
		}
	} else {
		err = errors.New("Logger already allocated")
	}
	return
}

// Debug prints a message to the log if debug mode is enabled, it accepts
// arguments similar to log.Println.
func Debug(a ...interface{}) {
	if flags.DebugEnabled() {
		GetLogger().Println(a...)
	}
}
