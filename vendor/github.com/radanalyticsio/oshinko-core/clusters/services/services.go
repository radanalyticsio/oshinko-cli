package services

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/intstr"
)

/*
        {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {
                "labels": {
                    "name": "spark-master-jpda"
                },
                "name": "spark-master-jpda"
            },
            "spec": {
                "ports": [
                    {
                        "port": 7077,
                        "protocol": "TCP",
                        "targetPort": 7077
                    }
                ],
                "selector": {
                    "name": "spark-master-jpda"
                }
            }
        },

 */


type OService struct {
	kapi.Service
}

func Service(name string) *OService {
	s := OService{}
	s.Kind = "Service"
	s.APIVersion = "v1"
	s.Name = name
	s.Spec.Type = kapi.ServiceTypeClusterIP
	s.Spec.SessionAffinity = kapi.ServiceAffinityNone
	s.Spec.Selector = map[string]string{}
	return &s
}

func (s *OService) SetLabels(selectors map[string]string) *OService {
	s.Service.SetLabels(selectors)
	return s
}

func (s *OService) Label(name, value string) *OService {
	if s.Labels == nil {
		s.Labels = map[string]string{}
	}
	s.Labels[name] = value
	return s
}

func (s *OService) PodSelector(selector, value string) *OService {
	s.Spec.Selector[selector] = value
	return s
}

func (s *OService) PodSelectors(selectors map[string]string) *OService {
	s.Spec.Selector = selectors
	return s
}

func (s *OService) Ports(ports ...*OServicePort) *OService {
	kports := make([]kapi.ServicePort, len(ports))
	for idx, p := range ports {
		kports[idx] = p.ServicePort
	}
	s.Spec.Ports = kports
	return s
}

type OServicePort struct {
	kapi.ServicePort
}

func ServicePort(port int) *OServicePort {
	s := OServicePort{}
	s.ServicePort.Protocol = kapi.ProtocolTCP
	s.ServicePort.Port = port
	return &s
}

func (s *OServicePort) Name(name string) *OServicePort {
	s.ServicePort.Name = name
	return s
}

func (s *OServicePort) Protocol(proto kapi.Protocol) *OServicePort {
	s.ServicePort.Protocol = proto
	return s
}

func (s *OServicePort) TargetPort(port int) *OServicePort {
	s.ServicePort.TargetPort = intstr.FromInt(port)
	return s
}

/*
type ServicePort struct {
	// Optional if only one ServicePort is defined on this service: The
	// name of this port within the service.  This must be a DNS_LABEL.
	// All ports within a ServiceSpec must have unique names.  This maps to
	// the 'Name' field in EndpointPort objects.
	Name string `json:"name"`

	// The IP protocol for this port.  Supports "TCP" and "UDP".
	Protocol Protocol `json:"protocol"`

	// The port that will be exposed on the service.
	Port int `json:"port"`

	// Optional: The target port on pods selected by this service.  If this
	// is a string, it will be looked up as a named port in the target
	// Pod's container ports.  If this is not specified, the value
	// of the 'port' field is used (an identity map).
	// This field is ignored for services with clusterIP=None, and should be
	// omitted or set equal to the 'port' field.
	TargetPort intstr.IntOrString `json:"targetPort"`

	// The port on each node on which this service is exposed.
	// Default is to auto-allocate a port if the ServiceType of this Service requires one.
	NodePort int `json:"nodePort"`
}
 */
