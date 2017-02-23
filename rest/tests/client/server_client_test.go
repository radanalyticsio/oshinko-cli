package clienttest

import (
	check "gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/version"
)

func (s *OshinkoRestTestSuite) TestServerInfo(c *check.C) {
	resp, _ := s.cli.Server.GetServerInfo(nil)

	expectedName := version.GetAppName()
	expectedVersion := version.GetVersion()

	observedName := resp.Payload.Application.Name
	observedVersion := resp.Payload.Application.Version

	c.Assert(*observedName, check.Equals, expectedName)
	c.Assert(*observedVersion, check.Equals, expectedVersion)
}
