package unittest

import (
	"log"

	goflags "github.com/jessevdk/go-flags"
	"gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/helpers/flags"
)

func getConfiguredParser() (parser *goflags.Parser) {
	parser = goflags.NewNamedParser("tests", goflags.Default)
	for _, optsGroup := range flags.GetLineOptionsGroups() {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return
}

func (s *OshinkoUnitTestSuite) TestGetLogFile(c *check.C) {
	expectedFile := ""
	observedFile := flags.GetLogFile()
	c.Assert(observedFile, check.Equals, expectedFile)

	expectedFile = "test.log.out"
	args := []string{"--log-file", expectedFile}
	parser := getConfiguredParser()
	parser.ParseArgs(args)
	observedFile = parser.FindOptionByLongName("log-file").Value().(string)
	c.Assert(observedFile, check.Equals, expectedFile)

	args = []string{"-l", expectedFile}
	parser = getConfiguredParser()
	parser.ParseArgs(args)
	observedFile = parser.FindOptionByShortName('l').Value().(string)
	c.Assert(observedFile, check.Equals, expectedFile)
}

func (s *OshinkoUnitTestSuite) TestDebugEnabled(c *check.C) {
	expectedState := false
	observedState := flags.DebugEnabled()
	c.Assert(observedState, check.Equals, expectedState)

	expectedState = true
	args := []string{"--debug"}
	parser := getConfiguredParser()
	parser.ParseArgs(args)
	observedState = parser.FindOptionByLongName("debug").Value().(bool)
	c.Assert(observedState, check.Equals, expectedState)

	args = []string{"-d"}
	parser = getConfiguredParser()
	parser.ParseArgs(args)
	observedState = parser.FindOptionByShortName('d').Value().(bool)
	c.Assert(observedState, check.Equals, expectedState)
}
