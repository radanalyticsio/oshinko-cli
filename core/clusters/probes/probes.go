package probes

import (
	kapi "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewHTTPGetProbe returns a probe object configured for HTTPGet actions,
// it currently only accepts the single required parameter.
func NewHTTPGetProbe(port int) kapi.Probe {
	act := kapi.HTTPGetAction{Port: intstr.FromInt(port)}
	hnd := kapi.Handler{HTTPGet: &act}
	prb := kapi.Probe{Handler: hnd}
	return prb
}

func NewExecProbe(cmd []string) kapi.Probe {
	act := kapi.ExecAction{Command: cmd}
	hnd := kapi.Handler{Exec: &act}
	prb := kapi.Probe{Handler: hnd}
	return prb
}
