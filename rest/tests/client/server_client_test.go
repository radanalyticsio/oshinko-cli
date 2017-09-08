package clienttest

import (
	check "gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/version"
	"github.com/radanalyticsio/oshinko-cli/rest/helpers/info"
	"os"
)

func (s *OshinkoRestTestSuite) TestServerInfo(c *check.C) {
	val := os.Getenv("OSHINKO_CLUSTER_IMAGE")
	os.Setenv("OSHINKO_CLUSTER_IMAGE", "")

	resp, _ := s.cli.Server.GetServerInfo(nil)

	expectedName := version.GetAppName()
	expectedVersion := version.GetVersion()
	expectedImage := info.GetSparkImage()

	observedName := resp.Payload.Application.Name
	observedVersion := resp.Payload.Application.Version
	observedImage := resp.Payload.Application.DefaultClusterImage

	c.Assert(*observedName, check.Equals, expectedName)
	c.Assert(*observedVersion, check.Equals, expectedVersion)
	c.Assert(*observedImage, check.Equals, expectedImage)

	os.Setenv("OSHINKO_CLUSTER_IMAGE", "bobby")
	expectedImage = "bobby"
	resp, _ = s.cli.Server.GetServerInfo(nil)
	observedImage = resp.Payload.Application.DefaultClusterImage
	c.Assert(*observedImage, check.Equals, expectedImage)

	os.Setenv("OSHINKO_CLUSTER_IMAGE", val)
}
