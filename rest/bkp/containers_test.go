package unittest

import (
	"gopkg.in/check.v1"
	kapi "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/api/resource"

	"github.com/radanalyticsio/oshinko-cli/core/clusters/containers"
	"github.com/radanalyticsio/oshinko-cli/core/clusters/probes"
)

func (s *OshinkoUnitTestSuite) TestContainer(c *check.C) {
	expectedName, expectedImage := "testname", "testimage"
	newContainer := containers.Container(expectedName, expectedImage)
	c.Assert(newContainer.Name, check.Equals, expectedName)
	c.Assert(newContainer.Image, check.Equals, expectedImage)
}

func (s *OshinkoUnitTestSuite) TestCommand(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedArgs := []string{"Arg1", "Arg2"}
	newContainer.Command(expectedArgs[0], expectedArgs[1])
	c.Assert(newContainer.Container.Command, check.DeepEquals, expectedArgs)
}

func (s *OshinkoUnitTestSuite) TestEnvVar(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedName, expectedValue := "testname", "testvalue"
	expectedEnv := []kapi.EnvVar{kapi.EnvVar{Name: expectedName, Value: expectedValue}}
	newContainer.EnvVar(expectedName, expectedValue)
	c.Assert(newContainer.Container.Env, check.DeepEquals, expectedEnv)
}

func (s *OshinkoUnitTestSuite) TestEnvVars(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedEnv := []kapi.EnvVar{
		kapi.EnvVar{Name: "name1", Value: "value1"},
		kapi.EnvVar{Name: "name2", Value: "value2"}}
	newContainer.EnvVars(expectedEnv)
	c.Assert(newContainer.Container.Env, check.DeepEquals, expectedEnv)
}

func (s *OshinkoUnitTestSuite) TestResourceLimit(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedName := kapi.ResourceName("testname")
	expectedQuantity := resource.Quantity{}
	newContainer.ResourceLimit(expectedName, expectedQuantity)
	c.Assert(newContainer.Resources.Limits[expectedName], check.DeepEquals, expectedQuantity)
}

func (s *OshinkoUnitTestSuite) TestResourceRequest(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedName := kapi.ResourceName("testname")
	expectedQuantity := resource.Quantity{}
	newContainer.ResourceRequest(expectedName, expectedQuantity)
	c.Assert(newContainer.Resources.Requests[expectedName], check.DeepEquals, expectedQuantity)
}

func (s *OshinkoUnitTestSuite) TestContainerPorts(c *check.C) {
	containerPorts := []*containers.OContainerPort{
		containers.ContainerPort("port1", 1),
		containers.ContainerPort("port2", 2)}
	newContainer := containers.Container("name", "image")
	newContainer.Ports(containerPorts[0], containerPorts[1])
	expectedPorts := make([]kapi.ContainerPort, len(containerPorts))
	for idx, p := range containerPorts {
		expectedPorts[idx] = p.ContainerPort
	}
	c.Assert(newContainer.Container.Ports, check.DeepEquals, expectedPorts)
}

func (s *OshinkoUnitTestSuite) TestContainerPort(c *check.C) {
	expectedName, expectedPort := "testname", 1234
	newContainerPort := containers.ContainerPort(expectedName, expectedPort)
	c.Assert(newContainerPort.ContainerPort.Name, check.Equals, expectedName)
	c.Assert(newContainerPort.ContainerPort.ContainerPort, check.Equals, int32(expectedPort))
}

func (s *OshinkoUnitTestSuite) TestProtocol(c *check.C) {
	expectedProtocol := kapi.Protocol("testprotocol")
	newContainerPort := containers.ContainerPort("name", 1)
	newContainerPort.Protocol(expectedProtocol)
	c.Assert(newContainerPort.ContainerPort.Protocol, check.Equals, expectedProtocol)
}

func (s *OshinkoUnitTestSuite) TestSetName(c *check.C) {
	newContainerPort := containers.ContainerPort("name", 1)
	expectedName := "newtestname"
	newContainerPort.SetName(expectedName)
	c.Assert(newContainerPort.Name, check.Equals, expectedName)
}

func (s *OshinkoUnitTestSuite) TestHostPort(c *check.C) {
	newContainerPort := containers.ContainerPort("name", 1)
	expectedHostPort := 12345
	newContainerPort.HostPort(expectedHostPort)
	c.Assert(newContainerPort.ContainerPort.HostPort, check.Equals, int32(expectedHostPort))
}

func (s *OshinkoUnitTestSuite) TestHostIP(c *check.C) {
	newContainerPort := containers.ContainerPort("name", 1)
	expectedHostIP := "some.test.ip"
	newContainerPort.HostIP(expectedHostIP)
	c.Assert(newContainerPort.ContainerPort.HostIP, check.Equals, expectedHostIP)
}

func (s *OshinkoUnitTestSuite) TestSetLivenessProbe(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedPort := 8080
	expectedProbe := probes.NewHTTPGetProbe(expectedPort)
	newContainer.SetLivenessProbe(expectedProbe)
	c.Assert(newContainer.LivenessProbe, check.DeepEquals, &expectedProbe)
	c.Assert(newContainer.LivenessProbe.Handler.HTTPGet.Port.IntValue(),
		check.Equals, expectedPort)
}

func (s *OshinkoUnitTestSuite) TestSetReadinessProbe(c *check.C) {
	newContainer := containers.Container("name", "image")
	expectedPort := 8080
	expectedProbe := probes.NewHTTPGetProbe(expectedPort)
	newContainer.SetReadinessProbe(expectedProbe)
	c.Assert(newContainer.ReadinessProbe, check.DeepEquals, &expectedProbe)
	c.Assert(newContainer.ReadinessProbe.Handler.HTTPGet.Port.IntValue(),
		check.Equals, expectedPort)
}

func (s *OshinkoUnitTestSuite) TestSetVolumeMount(c *check.C) {
	newContainer := containers.Container("name", "image")
	vmounts := []kapi.VolumeMount{
		{Name: "secrets", MountPath: "/etc/secrets", ReadOnly: true},
		{Name: "mysteries", MountPath: "/etc/mysteries", ReadOnly: false}}
	for _, v := range vmounts {
		newContainer.SetVolumeMount(v.Name, v.MountPath, v.ReadOnly)
	}
	c.Assert(newContainer.Container.VolumeMounts, check.DeepEquals, vmounts)
}
