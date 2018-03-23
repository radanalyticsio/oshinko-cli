package unittest

import (
	"gopkg.in/check.v1"
	kapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/radanalyticsio/oshinko-cli/core/clusters/services"
)

func (s *OshinkoUnitTestSuite) TestService(c *check.C) {
	expectedName := "testservice"
	newService := services.Service(expectedName)
	c.Assert(newService.Name, check.Equals, expectedName)
}

func (s *OshinkoUnitTestSuite) TestServiceSetLabels(c *check.C) {
	expectedLabels := map[string]string{"test": "value"}
	newService := services.Service("testservice")
	newService.SetLabels(expectedLabels)
	c.Assert(newService.Service.GetLabels(), check.DeepEquals, expectedLabels)
}

func (s *OshinkoUnitTestSuite) TestServiceLabel(c *check.C) {
	expectedLabels := map[string]string{"test": "value"}
	newService := services.Service("testservice")
	newService.SetLabels(map[string]string{})
	newService.Label("test", "value")
	c.Assert(newService.Service.GetLabels(), check.DeepEquals, expectedLabels)
}

func (s *OshinkoUnitTestSuite) TestServicePodSelector(c *check.C) {
	expectedSelector, expectedValue := "testselector", "testvalue"
	newService := services.Service("testservice")
	newService.PodSelector(expectedSelector, expectedValue)
	c.Assert(newService.Spec.Selector[expectedSelector], check.Equals, expectedValue)
}

func (s *OshinkoUnitTestSuite) TestServicePodSelectors(c *check.C) {
	expectedSelectors := map[string]string{"selector": "value"}
	newService := services.Service("testservice")
	newService.PodSelectors(expectedSelectors)
	c.Assert(newService.Spec.Selector, check.DeepEquals, expectedSelectors)
}

func (s *OshinkoUnitTestSuite) TestServicePorts(c *check.C) {
	servicePorts := []*services.OServicePort{
		services.ServicePort(1),
		services.ServicePort(2)}
	newService := services.Service("testservice")
	newService.Ports(servicePorts[0], servicePorts[1])
	expectedPorts := make([]kapi.ServicePort, len(servicePorts))
	for idx, p := range servicePorts {
		expectedPorts[idx] = p.ServicePort
	}
	c.Assert(newService.Spec.Ports, check.DeepEquals, expectedPorts)
}

func (s *OshinkoUnitTestSuite) TestServicePort(c *check.C) {
	expectedPort := 12345
	newServicePort := services.ServicePort(expectedPort)
	c.Assert(newServicePort.ServicePort.Port, check.Equals, int32(expectedPort))
}

func (s *OshinkoUnitTestSuite) TestServicePortName(c *check.C) {
	newServicePort := services.ServicePort(12345)
	expectedName := "testname"
	newServicePort.Name(expectedName)
	c.Assert(newServicePort.ServicePort.Name, check.Equals, expectedName)
}

func (s *OshinkoUnitTestSuite) TestServicePortProtocol(c *check.C) {
	newServicePort := services.ServicePort(12345)
	expectedProtocol := kapi.Protocol("testprotocol")
	newServicePort.Protocol(expectedProtocol)
	c.Assert(newServicePort.ServicePort.Protocol, check.Equals, expectedProtocol)
}

func (s *OshinkoUnitTestSuite) TestServicePortTargetPort(c *check.C) {
	newServicePort := services.ServicePort(12345)
	expectedPort := 54321
	newServicePort.TargetPort(expectedPort)
	c.Assert(newServicePort.ServicePort.TargetPort, check.Equals, intstr.FromInt(expectedPort))
}
