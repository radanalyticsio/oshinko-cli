package deploymentconfigs

import (
	appsapi "github.com/openshift/api/apps/v1"
	kapi "k8s.io/api/core/v1"
	"github.com/radanalyticsio/oshinko-cli/core/clusters/podtemplates"
)

type ODeploymentConfig struct {
	appsapi.DeploymentConfig
}

func DeploymentConfig(name, namespace string) *ODeploymentConfig {
	m := ODeploymentConfig{}
	m.Kind = "DeploymentConfig"
	m.APIVersion = "v1"
	m.SetName(name)
	m.SetNamespace(namespace)

	// Default Spec values.
	m.Spec.Replicas = 1
	m.Spec.Selector = map[string]string{}
	return &m
}

func (dc *ODeploymentConfig) Replicas(r int) *ODeploymentConfig {
	dc.Spec.Replicas = int32(r)
	return dc
}

func (dc *ODeploymentConfig) Label(name, value string) *ODeploymentConfig {
	l := dc.GetLabels()
	if l == nil {
		l = map[string]string{}
		dc.SetLabels(l)
	}
	l[name] = value
	return dc
}

func (dc *ODeploymentConfig) Annotate(name, value string) *ODeploymentConfig {
	a := dc.GetAnnotations()
	if a == nil {
		a = map[string]string{}
		dc.SetAnnotations(a)
	}
	a[name] = value
	return dc
}

func (dc *ODeploymentConfig) PodSelector(selector, value string) *ODeploymentConfig {
	dc.Spec.Selector[selector] = value
	return dc
}

func (dc *ODeploymentConfig) PodSelectors(selectors map[string]string) *ODeploymentConfig {
	dc.Spec.Selector = selectors
	return dc
}

func (dc *ODeploymentConfig) GetPodSelectors() map[string]string {
	return dc.Spec.Selector
}

func (dc *ODeploymentConfig) RollingStrategy() *ODeploymentConfig {
	dc.Spec.Strategy = appsapi.DeploymentStrategy{Type: appsapi.DeploymentStrategyTypeRolling}
	return dc
}

func (dc *ODeploymentConfig) RollingStrategyParams(rp *appsapi.RollingDeploymentStrategyParams,
	req kapi.ResourceRequirements,
	lbls, anttns map[string]string) *ODeploymentConfig {
	dc.Spec.Strategy = appsapi.DeploymentStrategy{
		Type:          appsapi.DeploymentStrategyTypeRolling,
		RollingParams: rp,
		//Resources:     req,
		Labels:        lbls,
		Annotations:   anttns,
	}
	dc.Spec.Strategy.Resources = req
	return dc
}

func (dc *ODeploymentConfig) RecreateStrategy() *ODeploymentConfig {
	dc.Spec.Strategy = appsapi.DeploymentStrategy{Type: appsapi.DeploymentStrategyTypeRecreate}
	return dc
}

func (dc *ODeploymentConfig) RecreateStrategyParams(rp *appsapi.RecreateDeploymentStrategyParams,
	req kapi.ResourceRequirements,
	lbls, anttns map[string]string) *ODeploymentConfig {
	dc.Spec.Strategy = appsapi.DeploymentStrategy{
		Type:           appsapi.DeploymentStrategyTypeRecreate,
		RecreateParams: rp,
		//Resources:      req,
		Labels:         lbls,
		Annotations:    anttns,
	}
	dc.Spec.Strategy.Resources = req
	return dc
}

func (dc *ODeploymentConfig) CustomStrategyParams(cp *appsapi.CustomDeploymentStrategyParams,
	req kapi.ResourceRequirements,
	lbls, anttns map[string]string) *ODeploymentConfig {
	dc.Spec.Strategy = appsapi.DeploymentStrategy{
		Type:         appsapi.DeploymentStrategyTypeCustom,
		CustomParams: cp,
		Resources:    req,
		Labels:       lbls,
		Annotations:  anttns,
	}
	return dc
}

func (dc *ODeploymentConfig) TriggerOnConfigChange() *ODeploymentConfig {
	for _, val := range dc.Spec.Triggers {
		if val.Type == appsapi.DeploymentTriggerOnConfigChange {
			return dc
		}
	}
	dc.Spec.Triggers = append(
		dc.Spec.Triggers,
		appsapi.DeploymentTriggerPolicy{Type: appsapi.DeploymentTriggerOnConfigChange})
	return dc
}

func (dc *ODeploymentConfig) TriggerOnImageChange(ic *appsapi.DeploymentTriggerImageChangeParams) *ODeploymentConfig {
	for idx, val := range dc.Spec.Triggers {
		if val.Type == appsapi.DeploymentTriggerOnImageChange {
			// If we pass the same pointer, ignore
			if val.ImageChangeParams == ic {
				return dc
			}
			// If the Name matches, update
			// TODO Namespace is allowed to be blank, we should probably handle that case at some point
			if val.ImageChangeParams.From.Name == ic.From.Name &&
				val.ImageChangeParams.From.Namespace == ic.From.Namespace {
				dc.Spec.Triggers[idx].ImageChangeParams = ic
				return dc
			}
		}
	}
	dc.Spec.Triggers = append(
		dc.Spec.Triggers,
		appsapi.DeploymentTriggerPolicy{Type: appsapi.DeploymentTriggerOnImageChange,
			ImageChangeParams: ic})
	return dc
}

func (dc *ODeploymentConfig) PodTemplateSpec(pt *podtemplates.OPodTemplateSpec) *ODeploymentConfig {
	dc.Spec.Template = &pt.PodTemplateSpec
	return dc
}

func (dc *ODeploymentConfig) GetPodTemplateSpecLabels() map[string]string {
	if dc.Spec.Template == nil {
		return map[string]string{}
	}
	return dc.Spec.Template.Labels
}

func (dc *ODeploymentConfig) FindPort(name string) int {
	if dc.Spec.Template != nil {
		for _, val := range dc.Spec.Template.Spec.Containers {
			for _, port := range val.Ports {
				if port.Name == name {
					return int(port.ContainerPort)
				}
			}
		}
	}
	return 0
}
