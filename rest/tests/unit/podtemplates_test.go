package unittest

import (
	"gopkg.in/check.v1"
	kapi "k8s.io/api/core/v1"

	"github.com/radanalyticsio/oshinko-cli/core/clusters/containers"
	"github.com/radanalyticsio/oshinko-cli/core/clusters/podtemplates"
)

// This function is named TestCreatePodTemplateSpec because there is another
// function named TestPodTemplateSpec.
func (s *OshinkoUnitTestSuite) TestCreatePodTemplateSpec(c *check.C) {
	newPodTemplateSpec := podtemplates.PodTemplateSpec()
	c.Assert(*newPodTemplateSpec, check.FitsTypeOf, podtemplates.OPodTemplateSpec{})
}

func (s *OshinkoUnitTestSuite) TestPodTemplateSetLabels(c *check.C) {
	expectedLabels := map[string]string{"test": "value"}
	newPodTemplateSpec := podtemplates.PodTemplateSpec()
	newPodTemplateSpec.SetLabels(expectedLabels)
	c.Assert(newPodTemplateSpec.PodTemplateSpec.GetLabels(), check.DeepEquals, expectedLabels)
}

func (s *OshinkoUnitTestSuite) TestPodTemplateLabel(c *check.C) {
	expectedLabels := map[string]string{"test": "value"}
	newPodTemplateSpec := podtemplates.PodTemplateSpec()
	newPodTemplateSpec.SetLabels(map[string]string{})
	newPodTemplateSpec.Label("test", "value")
	c.Assert(newPodTemplateSpec.PodTemplateSpec.GetLabels(), check.DeepEquals, expectedLabels)
}

func (s *OshinkoUnitTestSuite) TestContainers(c *check.C) {
	expectedContainers := []kapi.Container{
		kapi.Container{Name: "container1"},
		kapi.Container{Name: "container2"}}
	newPodTemplateSpec := podtemplates.PodTemplateSpec()
	newPodTemplateSpec.Containers(
		&containers.OContainer{Container: expectedContainers[0]},
		&containers.OContainer{Container: expectedContainers[1]})
	c.Assert(newPodTemplateSpec.Spec.Containers, check.DeepEquals, expectedContainers)
}

func makeCmapVS(name string) kapi.VolumeSource {
	cmapvs := kapi.ConfigMapVolumeSource{}
	cmapvs.LocalObjectReference.Name = name
	return kapi.VolumeSource{ConfigMap: &cmapvs}
}

func (s *OshinkoUnitTestSuite) TestSetConfigMapVolume(c *check.C) {
	ptspec := podtemplates.PodTemplateSpec()
	var names []string = []string{"configmap1", "configmap2"}
	var volumes []kapi.Volume = []kapi.Volume{}

	for _, name :=  range names {
		vs := makeCmapVS(name)
		volumes = append(volumes, kapi.Volume{Name: name, VolumeSource: vs})
		ptspec.SetConfigMapVolume(name)
	}
	c.Assert(ptspec.Spec.Volumes, check.DeepEquals, volumes)

}
