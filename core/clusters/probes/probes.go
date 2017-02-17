package probes

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// NewHTTPGetProbe returns a probe object configured for HTTPGet actions,
// it currently only accepts the single required parameter.
func NewHTTPGetProbe(port int) kapi.Probe {
	act := kapi.HTTPGetAction{Port: intstr.FromInt(port)}
	hnd := kapi.Handler{HTTPGet: &act}
	prb := kapi.Probe{Handler: hnd}
	return prb
}
