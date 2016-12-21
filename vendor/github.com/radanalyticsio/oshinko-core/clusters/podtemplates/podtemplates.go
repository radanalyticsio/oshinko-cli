package podtemplates

import (
	"github.com/radanalyticsio/oshinko-core/clusters/containers"
	kapi "k8s.io/kubernetes/pkg/api"
)

type OPodTemplateSpec struct {
	kapi.PodTemplateSpec
}

// we might care about volumes
// we might care about terminationGracePeriodSeconds
// we might care about serviceaccountname
// we might care about security context
// we might care about image pull secrets

func PodTemplateSpec() *OPodTemplateSpec {
	// Note, name and namespace can be set on a PodTemplateSpec but
	// I assume that openshift takes care of that based on the DeploymentConfig
	p := OPodTemplateSpec{}
	p.Spec.DNSPolicy = kapi.DNSClusterFirst
	p.Spec.RestartPolicy = kapi.RestartPolicyAlways
	return &p
}

func (pt *OPodTemplateSpec) SetLabels(selectors map[string]string) *OPodTemplateSpec {
	pt.PodTemplateSpec.SetLabels(selectors)
	return pt
}

func (pt *OPodTemplateSpec) Label(name, value string) *OPodTemplateSpec {
	if pt.Labels == nil {
		pt.Labels = map[string]string{}
	}
	pt.Labels[name] = value
	return pt
}

func (pt *OPodTemplateSpec) Containers(cntnrs ...*containers.OContainer) *OPodTemplateSpec {
	kcntnrs := make([]kapi.Container, len(cntnrs))
	for idx, c := range cntnrs {
		kcntnrs[idx] = c.Container
	}
	pt.Spec.Containers = kcntnrs
	return pt
}

func (pt *OPodTemplateSpec) SetConfigMapVolume(configmap string) *OPodTemplateSpec {
	if pt.Spec.Volumes == nil {
		pt.Spec.Volumes = []kapi.Volume{}
	}

	cm := kapi.ConfigMapVolumeSource{}
	cm.LocalObjectReference.Name = configmap

	// This is the source for the volume, there are lots of different kinds.
	// We just need a pointer to the configmap source
	vsource := kapi.VolumeSource{ConfigMap: &cm}

	v := kapi.Volume{Name: configmap, VolumeSource: vsource}
	pt.Spec.Volumes = append(pt.Spec.Volumes, v)
	return pt
}
