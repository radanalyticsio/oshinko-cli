package tests

import (
	"fmt"
	"testing"

	loads "github.com/go-openapi/loads"
	httptransport "github.com/go-openapi/runtime/client"
	strfmt "github.com/go-openapi/strfmt"
	check "gopkg.in/check.v1"

	"github.com/redhatanalytics/oshinko-rest/client"
	"github.com/redhatanalytics/oshinko-rest/restapi"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations"
	"github.com/redhatanalytics/oshinko-rest/version"
)

// Connect gocheck to the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type ServerTestSuite struct {
	server *restapi.Server
}

var _ = check.Suite(&ServerTestSuite{})

func (s *ServerTestSuite) SetUpSuite(c *check.C) {
	swaggerSpec, _ := loads.Analyzed(restapi.SwaggerJSON, "")

	api := operations.NewOshinkoRestAPI(swaggerSpec)
	server := restapi.NewServer(api)

	server.ConfigureAPI()

	server.Host = "127.0.0.1"

	s.server = server

	server.Listen()
	go server.Serve()
}

func (s *ServerTestSuite) TestServerInfo(c *check.C) {
	transport := httptransport.New(fmt.Sprintf("%s:%d", s.server.Host, s.server.Port), "/", []string{"http"})
	formats := strfmt.Default
	cli := client.New(transport, formats)

	resp, _ := cli.Server.GetServerInfo(nil)

	expectedName := version.GetAppName()
	expectedVersion := version.GetVersion()

	observedName := resp.Payload.Application.Name
	observedVersion := resp.Payload.Application.Version

	c.Assert(*observedName, check.Equals, expectedName)
	c.Assert(*observedVersion, check.Equals, expectedVersion)
}
