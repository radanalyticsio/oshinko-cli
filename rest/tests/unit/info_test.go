package unittest

import (
	"os"

	"gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/helpers/info"
	"github.com/radanalyticsio/oshinko-cli/rest/version"
)

func (s *OshinkoUnitTestSuite) TestInAPod(c *check.C) {
	os.Setenv("OSHINKO_REST_POD_NAME", "")
	c.Assert(info.InAPod(), check.Equals, false)
	os.Setenv("OSHINKO_REST_POD_NAME", "some-pod-name")
	c.Assert(info.InAPod(), check.Equals, true)
}

func (s *OshinkoUnitTestSuite) TestGetNamespace(c *check.C) {
	expectedNamespace := ""
	expectedErr := error(nil)
	os.Setenv("OSHINKO_REST_POD_NAME", "")
	observedNamespace, observedErr := info.GetNamespace()
	c.Assert(observedNamespace, check.Equals, expectedNamespace)
	c.Assert(observedErr, check.Equals, expectedErr)
	expectedNamespace = "testnamespace"
	os.Setenv("OSHINKO_CLUSTER_NAMESPACE", expectedNamespace)
	observedNamespace, observedErr = info.GetNamespace()
	c.Assert(observedNamespace, check.Equals, expectedNamespace)
	c.Assert(observedErr, check.Equals, expectedErr)
	os.Setenv("OSHINKO_REST_POD_NAME", "some-pod-name")
	observedNamespace, observedErr = info.GetNamespace()
	c.Assert(observedNamespace, check.Equals, expectedNamespace)
	c.Assert(observedErr, check.Equals, expectedErr)
	// TODO(elmiko) add a test to demonstrate a working GetServiceAccountNS
}

func (s *OshinkoUnitTestSuite) TestGetSparkImage(c *check.C) {
	expectedImage := version.GetSparkImage()
	os.Setenv("OSHINKO_CLUSTER_IMAGE", "")
	observedImage := info.GetSparkImage()
	c.Assert(observedImage, check.Equals, expectedImage)
	expectedImage = "some/test/image"
	os.Setenv("OSHINKO_CLUSTER_IMAGE", expectedImage)
	observedImage = info.GetSparkImage()
	c.Assert(observedImage, check.Equals, expectedImage)
}

func (s *OshinkoUnitTestSuite) TestGetKubeConfigPath(c *check.C) {
	expectedPath := "test/path"
	os.Setenv("OSHINKO_KUBE_CONFIG", expectedPath)
	observedPath := info.GetKubeConfigPath()
	c.Assert(observedPath, check.Equals, expectedPath)
}

func (s *OshinkoUnitTestSuite) TestGetKubeProxyAddress(c *check.C) {
	expectedProxy := ""
	os.Setenv("KUBERNETES_SERVICE_HOST", "")
	observedProxy, observedErr := info.GetKubeProxyAddress()
	c.Assert(observedProxy, check.Equals, expectedProxy)
	c.Assert(observedErr, check.Not(check.Equals), nil)
	expectedProxy = "test-host-proxy"
	os.Setenv("KUBERNETES_SERVICE_HOST", expectedProxy)
	observedProxy, observedErr = info.GetKubeProxyAddress()
	c.Assert(observedProxy, check.Equals, expectedProxy)
	c.Assert(observedErr, check.Equals, nil)
}

func (s *OshinkoUnitTestSuite) TestGetKubeProxyPort(c *check.C) {
	expectedPort := ""
	os.Setenv("KUBERNETES_SERVICE_PORT", "")
	observedPort, observedErr := info.GetKubeProxyPort()
	c.Assert(observedPort, check.Equals, expectedPort)
	c.Assert(observedErr, check.Not(check.Equals), nil)
	expectedPort = "12345"
	os.Setenv("KUBERNETES_SERVICE_PORT", expectedPort)
	observedPort, observedErr = info.GetKubeProxyPort()
	c.Assert(observedPort, check.Equals, expectedPort)
	c.Assert(observedErr, check.Equals, nil)
}

func (s *OshinkoUnitTestSuite) TestGetServiceAccountCAPath(c *check.C) {
	observedPath := info.GetServiceAccountCAPath()
	c.Assert(observedPath, check.Equals, info.CA_PATH)
}

func (s *OshinkoUnitTestSuite) TestGetServiceAccountTokenPath(c *check.C) {
	observedPath := info.GetServiceAccountTokenPath()
	c.Assert(observedPath, check.Equals, info.TOKEN_PATH)
}

func (s *OshinkoUnitTestSuite) TestGetServiceAccountNSPath(c *check.C) {
	observedPath := info.GetServiceAccountNSPath()
	c.Assert(observedPath, check.Equals, info.NS_PATH)
}

func (s *OshinkoUnitTestSuite) TestGetServiceAccountToken(c *check.C) {
	expectedToken := []byte(nil)
	observedToken, observedErr := info.GetServiceAccountToken()
	c.Assert(observedToken, check.DeepEquals, expectedToken)
	c.Assert(observedErr, check.Not(check.Equals), nil)
	// INFO(elmiko) this cannot be tested further given that a file read is
	// required based on a const path set in the package.
}

func (s *OshinkoUnitTestSuite) TestGetServiceAccountNS(c *check.C) {
	expectedNS := []byte(nil)
	observedNS, observedErr := info.GetServiceAccountNS()
	c.Assert(observedNS, check.DeepEquals, expectedNS)
	c.Assert(observedErr, check.Not(check.Equals), nil)
	// INFO(elmiko) this cannot be tested further given that a file read is
	// required based on a const path set in the package.
}

func (s *OshinkoUnitTestSuite) TestGetWebServiceName(c *check.C) {
	expectedName := "test-service-name"
	os.Setenv("OSHINKO_WEB_NAME", expectedName)
	observedName := info.GetWebServiceName()
	c.Assert(observedName, check.Equals, expectedName)
}
