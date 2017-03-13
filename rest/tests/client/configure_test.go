package clienttest

/*
INFO(elmiko) This file exists as a helper to ensure that the tests in this
directory are hooked into the go testing infrastructure. The Test function
declaration needs to be included once in a package for tests, hence the
existence of this file.
*/

import (
	"fmt"
	"log"
	"testing"

	loads "github.com/go-openapi/loads"
	httptransport "github.com/go-openapi/runtime/client"
	strfmt "github.com/go-openapi/strfmt"
	flags "github.com/jessevdk/go-flags"
	check "gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/client"
	"github.com/radanalyticsio/oshinko-cli/rest/restapi"
	"github.com/radanalyticsio/oshinko-cli/rest/restapi/operations"
)

// Test connects gocheck to the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

// OshinkoRestTestSuite is the basic object for all tests
type OshinkoRestTestSuite struct {
	server *restapi.Server
	cli    *client.OshinkoRest
}

var _ = check.Suite(&OshinkoRestTestSuite{})

// SetUpSuite is run once before the entire test suite
func (s *OshinkoRestTestSuite) SetUpSuite(c *check.C) {
	swaggerSpec, _ := loads.Analyzed(restapi.SwaggerJSON, "")

	api := operations.NewOshinkoRestAPI(swaggerSpec)
	server := restapi.NewServer(api)

	parser := flags.NewNamedParser("oshinko-rest-client-test", flags.Default)
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}
	args := []string{"--loglevel=10"}
	parser.ParseArgs(args)

	server.ConfigureAPI()

	server.Host = "127.0.0.1"
	server.EnabledListeners = []string{"http"}

	s.server = server

	server.Listen()
	go server.Serve()
}

// SetUpTest is run once before each test
func (s *OshinkoRestTestSuite) SetUpTest(c *check.C) {
	transport := httptransport.New(fmt.Sprintf("%s:%d", s.server.Host, s.server.Port), "/", []string{"http"})
	formats := strfmt.Default
	s.cli = client.New(transport, formats)
}

// TearDowSuite is run once after all tests have finished
func (s *OshinkoRestTestSuite) TearDownSuite(c *check.C) {
	s.server.Shutdown()
}
