// Package flags provides helper functions for accessing command line flags.
//
// It also serves as a central point for creating new flags and providing
// helpers for interacting with the available information.
package flags

import (
	"github.com/go-openapi/swag"
)

var optionsGroups *[]swag.CommandLineOptionsGroup

// oshinkoOptions are the command line flags, these are formatted according
// to the documentation at https://github.com/jessevdk/go-flags
// New flags should be added to this structure, with helper functions added
// in this package.
type oshinkoOptions struct {
	Info       bool   `long:"info" description:"log version information and exit"`
	LogFile    string `short:"l" long:"log-file" description:"the file to write logs into, defaults to stdout"`
	DebugState bool   `short:"d" long:"debug" description:"enable debug mode in the server"`
	LogLevel   string `long:"loglevel" description:"set the log level (0-10)" default:"0"`
}

// GetLineOptionsGroups returns the CommandLineOptionsGroup structure that
// can be used to configure the command line flags for the rest server.
func GetLineOptionsGroups() []swag.CommandLineOptionsGroup {
	if optionsGroups == nil {
		newOptionsGroups := []swag.CommandLineOptionsGroup{
			{
				ShortDescription: "Oshinko REST server options",
				Options:          &oshinkoOptions{},
			},
		}
		optionsGroups = &newOptionsGroups
	}
	return *optionsGroups
}

// GetLogFile returns the log filename specified on the command line or an
// empty string in the case that no file is specified.
func GetLogFile() string {
	retval := ""
	if optionsGroups != nil {
		for _, optsGroup := range *optionsGroups {
			opts, ok := optsGroup.Options.(*oshinkoOptions)
			if ok == true {
				if opts.LogFile != "" {
					retval = opts.LogFile
				}
			}
		}
	}
	return retval
}

// DebugEnabled returns true if the debug flag has been invoked on the
// command line, otherwise false.
func DebugEnabled() bool {
	retval := false
	if optionsGroups != nil {
		for _, optsGroup := range *optionsGroups {
			opts, ok := optsGroup.Options.(*oshinkoOptions)
			if ok == true {
				retval = opts.DebugState
			}
		}
	}
	return retval
}

// InfoEnabled returns true if the info flag has been invoked on the
// command line, otherwise false.
func InfoEnabled() bool {
	retval := false
	if optionsGroups != nil {
		for _, optsGroup := range *optionsGroups {
			opts, ok := optsGroup.Options.(*oshinkoOptions)
			if ok == true {
				retval = opts.Info
			}
		}
	}
	return retval
}

// Return the requested log level
func LogLevel() string {
	retval := "0"
	if optionsGroups != nil {
		for _, optsGroup := range *optionsGroups {
			opts, ok := optsGroup.Options.(*oshinkoOptions)
			if ok == true {
				retval = opts.LogLevel
			}
		}
	}
	return retval
}
