package containers

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

type OContainer struct {
	kapi.Container
}

func Container(name, image string) *OContainer {
	c := OContainer{}
	c.Name = name
	c.Image = image
	c.TerminationMessagePath = "/dev/termination-log"
	return &c
}

func (c *OContainer) Command(args ...string) *OContainer {
	c.Container.Command = args
	return c
}

func (c *OContainer) EnvVar(name, value string) *OContainer {
	c.Container.Env = append(c.Container.Env, kapi.EnvVar{Name: name, Value: value})
	return c
}

func (c *OContainer) EnvVars(envs []kapi.EnvVar) *OContainer {
	c.Container.Env = envs
	return c
}

// TODO we might want to add some handling around building Quantities too
func (c *OContainer) ResourceLimit(name kapi.ResourceName, q resource.Quantity) *OContainer {
	if c.Resources.Limits == nil {
		c.Resources.Limits = make(kapi.ResourceList, 1)
	}
	c.Resources.Limits[name] = q
	return c
}

func (c *OContainer) ResourceRequest(name kapi.ResourceName, q resource.Quantity) *OContainer {
	if c.Resources.Requests == nil {
		c.Resources.Requests = make(kapi.ResourceList, 1)
	}
	c.Resources.Requests[name] = q
	return c
}

func (c *OContainer) Ports(ports ...*OContainerPort) *OContainer {
	kports := make([]kapi.ContainerPort, len(ports))
	for idx, p := range ports {
		kports[idx] = p.ContainerPort
	}
	c.Container.Ports = kports
	return c
}

func (c *OContainer) SetLivenessProbe(probe kapi.Probe) *OContainer {
	c.LivenessProbe = &probe
	return c
}

func (c *OContainer) SetReadinessProbe(probe kapi.Probe) *OContainer {
	c.ReadinessProbe = &probe
	return c
}

func (c *OContainer) SetVolumeMount(name, mountpath string, ro bool) *OContainer {
	if c.VolumeMounts == nil {
		c.VolumeMounts = []kapi.VolumeMount{}
	}
	vm := kapi.VolumeMount{Name: name, MountPath: mountpath, ReadOnly: ro}
	c.VolumeMounts = append(c.VolumeMounts, vm)
	return c
}

type OContainerPort struct {
	kapi.ContainerPort
}

func ContainerPort(name string, port int) *OContainerPort {
	cp := OContainerPort{}
	cp.Name = name
	cp.ContainerPort.ContainerPort = int32(port)
	cp.ContainerPort.Protocol = kapi.ProtocolTCP
	return &cp
}

func (cp *OContainerPort) Protocol(proto kapi.Protocol) *OContainerPort {
	cp.ContainerPort.Protocol = proto
	return cp
}

func (cp *OContainerPort) SetName(name string) *OContainerPort {
	cp.Name = name
	return cp
}

func (cp *OContainerPort) HostPort(port int) *OContainerPort {
	cp.ContainerPort.HostPort = int32(port)
	return cp
}

func (cp *OContainerPort) HostIP(hostip string) *OContainerPort {
	cp.ContainerPort.HostIP = hostip
	return cp
}
