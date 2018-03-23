package unittest

import (
	"gopkg.in/check.v1"
	kapi "k8s.io/api/core/v1"

	api  "github.com/openshift/api/apps/v1"
	"github.com/radanalyticsio/oshinko-cli/core/clusters/deploymentconfigs"
	"github.com/radanalyticsio/oshinko-cli/core/clusters/podtemplates"
)

func (s *OshinkoUnitTestSuite) TestDeploymentConfig(c *check.C) {
	expectedName, expectedNamespace := "testname", "testnamespace"
	newDeploymentConfig := deploymentconfigs.DeploymentConfig(expectedName, expectedNamespace)
	c.Assert(newDeploymentConfig.Name, check.Equals, expectedName)
	c.Assert(newDeploymentConfig.Namespace, check.Equals, expectedNamespace)
}

func (s *OshinkoUnitTestSuite) TestReplicas(c *check.C) {
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	expectedReplicas := 12345
	newDeploymentConfig.Replicas(expectedReplicas)
	c.Assert(newDeploymentConfig.Spec.Replicas, check.Equals, int32(expectedReplicas))
}

func (s *OshinkoUnitTestSuite) TestLabel(c *check.C) {
	expectedName, expectedLabel := "testname", "testlabel"
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.Label(expectedName, expectedLabel)
	c.Assert(newDeploymentConfig.Labels[expectedName], check.Equals, expectedLabel)
}

func (s *OshinkoUnitTestSuite) TestDeploymentPodSelector(c *check.C) {
	expectedSelector, expectedValue := "testselector", "testvalue"
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.PodSelector(expectedSelector, expectedValue)
	c.Assert(newDeploymentConfig.Spec.Selector[expectedSelector], check.Equals, expectedValue)
}

func (s *OshinkoUnitTestSuite) TestDeploymentPodSelectors(c *check.C) {
	expectedSelectors := map[string]string{"selector1": "value1", "selector2": "value2"}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.PodSelectors(expectedSelectors)
	c.Assert(newDeploymentConfig.Spec.Selector, check.DeepEquals, expectedSelectors)
}

func (s *OshinkoUnitTestSuite) TestGetPodSelectors(c *check.C) {
	expectedSelectors := map[string]string{"selector1": "value1", "selector2": "value2"}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.PodSelectors(expectedSelectors)
	c.Assert(newDeploymentConfig.GetPodSelectors(), check.DeepEquals, expectedSelectors)
}

func (s *OshinkoUnitTestSuite) TestRollingStrategy(c *check.C) {
	expectedStrategy := api.DeploymentStrategy{Type: api.DeploymentStrategyTypeRolling}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.RollingStrategy()
	c.Assert(newDeploymentConfig.Spec.Strategy, check.DeepEquals, expectedStrategy)
}

func (s *OshinkoUnitTestSuite) TestRollingStrategyParams(c *check.C) {
	rp := api.RollingDeploymentStrategyParams{}
	req := kapi.ResourceRequirements{}
	lbls := map[string]string{"test": "value"}
	anttns := map[string]string{"moretest": "morevalue"}
	expectedStrategy := api.DeploymentStrategy{
		Type:          api.DeploymentStrategyTypeRolling,
		RollingParams: &rp,
		Resources:     req,
		Labels:        lbls,
		Annotations:   anttns}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.RollingStrategyParams(&rp, req, lbls, anttns)
	c.Assert(newDeploymentConfig.Spec.Strategy, check.DeepEquals, expectedStrategy)
}

func (s *OshinkoUnitTestSuite) TestRecreateStrategy(c *check.C) {
	expectedStrategy := api.DeploymentStrategy{Type: api.DeploymentStrategyTypeRecreate}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.RecreateStrategy()
	c.Assert(newDeploymentConfig.Spec.Strategy, check.DeepEquals, expectedStrategy)
}

func (s *OshinkoUnitTestSuite) TestRecreateStrategyParams(c *check.C) {
	rp := api.RecreateDeploymentStrategyParams{}
	req := kapi.ResourceRequirements{}
	lbls := map[string]string{"test": "value"}
	anttns := map[string]string{"moretest": "morevalue"}
	expectedStrategy := api.DeploymentStrategy{
		Type:           api.DeploymentStrategyTypeRecreate,
		RecreateParams: &rp,
		Resources:      req,
		Labels:         lbls,
		Annotations:    anttns}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.RecreateStrategyParams(&rp, req, lbls, anttns)
	c.Assert(newDeploymentConfig.Spec.Strategy, check.DeepEquals, expectedStrategy)
}

func (s *OshinkoUnitTestSuite) TestCustomStrategyParams(c *check.C) {
	rp := api.CustomDeploymentStrategyParams{}
	req := kapi.ResourceRequirements{}
	lbls := map[string]string{"test": "value"}
	anttns := map[string]string{"moretest": "morevalue"}
	expectedStrategy := api.DeploymentStrategy{
		Type:         api.DeploymentStrategyTypeCustom,
		CustomParams: &rp,
		Resources:    req,
		Labels:       lbls,
		Annotations:  anttns}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.CustomStrategyParams(&rp, req, lbls, anttns)
	c.Assert(newDeploymentConfig.Spec.Strategy, check.DeepEquals, expectedStrategy)
}

func (s *OshinkoUnitTestSuite) TestTriggerOnConfigChange(c *check.C) {
	expectedTriggers := []api.DeploymentTriggerPolicy{
		api.DeploymentTriggerPolicy{Type: api.DeploymentTriggerOnConfigChange}}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.TriggerOnConfigChange()
	c.Assert(newDeploymentConfig.Spec.Triggers, check.DeepEquals, expectedTriggers)
}

func (s *OshinkoUnitTestSuite) TestTriggerOnImageChange(c *check.C) {
	imageChangeParams1 := api.DeploymentTriggerImageChangeParams{
		From: kapi.ObjectReference{
			Name:      "imagename1",
			Namespace: "namespace1"}}
	imageChangeParams1Copy := api.DeploymentTriggerImageChangeParams{
		From: kapi.ObjectReference{
			Name:      "imagename1",
			Namespace: "namespace1"}}
	imageChangeParams2 := api.DeploymentTriggerImageChangeParams{
		From: kapi.ObjectReference{
			Name:      "imagename2",
			Namespace: "namespace2"}}
	expectedTriggers := []api.DeploymentTriggerPolicy{
		api.DeploymentTriggerPolicy{
			Type:              api.DeploymentTriggerOnImageChange,
			ImageChangeParams: &imageChangeParams1}}
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	newDeploymentConfig.TriggerOnImageChange(&imageChangeParams1)
	c.Assert(newDeploymentConfig.Spec.Triggers, check.DeepEquals, expectedTriggers)
	newDeploymentConfig.TriggerOnImageChange(&imageChangeParams1Copy)
	c.Assert(newDeploymentConfig.Spec.Triggers, check.DeepEquals, expectedTriggers)
	newDeploymentConfig.TriggerOnImageChange(&imageChangeParams2)
	expectedTriggers = append(expectedTriggers, api.DeploymentTriggerPolicy{
		Type:              api.DeploymentTriggerOnImageChange,
		ImageChangeParams: &imageChangeParams2})
	c.Assert(newDeploymentConfig.Spec.Triggers, check.DeepEquals, expectedTriggers)
}

func (s *OshinkoUnitTestSuite) TestPodTemplateSpec(c *check.C) {
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	expectedPodTemplateSpec := podtemplates.OPodTemplateSpec{}
	newDeploymentConfig.PodTemplateSpec(&expectedPodTemplateSpec)
	c.Assert(newDeploymentConfig.Spec.Template, check.Equals, &expectedPodTemplateSpec.PodTemplateSpec)
	c.Assert(*newDeploymentConfig.Spec.Template, check.DeepEquals, expectedPodTemplateSpec.PodTemplateSpec)
}

func (s *OshinkoUnitTestSuite) TestGetPodTemplateSpecLabels(c *check.C) {
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	expectedLabels := map[string]string{}
	c.Assert(newDeploymentConfig.GetPodTemplateSpecLabels(), check.DeepEquals, expectedLabels)
	podTemplateSpec := podtemplates.OPodTemplateSpec{}
	expectedLabels["test"] = "value"
	podTemplateSpec.SetLabels(expectedLabels)
	newDeploymentConfig.PodTemplateSpec(&podTemplateSpec)
	c.Assert(newDeploymentConfig.GetPodTemplateSpecLabels(), check.DeepEquals, expectedLabels)
}

func (s *OshinkoUnitTestSuite) TestFindPort(c *check.C) {
	newDeploymentConfig := deploymentconfigs.DeploymentConfig("name", "namespace")
	c.Assert(newDeploymentConfig.FindPort("testport"), check.Equals, 0)
	podTemplateSpec := podtemplates.OPodTemplateSpec{
		kapi.PodTemplateSpec{
			Spec: kapi.PodSpec{
				Containers: []kapi.Container{
					kapi.Container{
						Ports: []kapi.ContainerPort{
							kapi.ContainerPort{
								Name:          "testport",
								ContainerPort: 12345}}}}}}}
	newDeploymentConfig.PodTemplateSpec(&podTemplateSpec)
	c.Assert(newDeploymentConfig.FindPort("testport"), check.Equals, 12345)
	c.Assert(newDeploymentConfig.FindPort("badport"), check.Equals, 0)
}
