package unittest

import (
	"errors"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"gopkg.in/check.v1"
	"strconv"
	//"github.com/radanalyticsio/oshinko-cli/rest/models"
	"fmt"
	api "k8s.io/api/core/v1"
	//"k8s.io/kubernetes/pkg/api"
	"k8s.io/apimachinery/pkg/watch"
	//"k8s.io/kubernetes/pkg/watch"
)

var tiny clusters.ClusterConfig = clusters.ClusterConfig{
	MastersCount: 1,
	WorkersCount: clusters.SentinelCountValue,
	ConfigName:   "tiny"}
var small clusters.ClusterConfig = clusters.ClusterConfig{
	MastersCount:      1,
	WorkersCount:      3,
	SparkMasterConfig: "master-config",
	SparkWorkerConfig: "worker-config",
	ConfigName:        "small"}
var large clusters.ClusterConfig = clusters.ClusterConfig{
	MastersCount: clusters.SentinelCountValue,
	WorkersCount: 10,
	ConfigName:   "large"}
var brokenMaster clusters.ClusterConfig = clusters.ClusterConfig{
	MastersCount: 2,
	WorkersCount: clusters.SentinelCountValue,
	ConfigName:   "brokenmaster"}

var nonIntMaster clusters.ClusterConfig = clusters.ClusterConfig{ConfigName: "cow"}
var nonIntWorker clusters.ClusterConfig = clusters.ClusterConfig{ConfigName: "pig"}
var userDefault = clusters.ClusterConfig{MastersCount: 3, WorkersCount: 3,
	SparkMasterConfig: "master-default", SparkWorkerConfig: "worker-default", ConfigName: "default-oshinko-cluster-config",
	ExposeWebUI: "false", Metrics: "true"}

func makeConfigMap(cfg clusters.ClusterConfig) *api.ConfigMap {
	var res api.ConfigMap = api.ConfigMap{Data: map[string]string{}}
	if cfg.SparkMasterConfig != "" {
		res.Data["sparkmasterconfig"] = cfg.SparkMasterConfig
	}
	if cfg.SparkWorkerConfig != "" {
		res.Data["sparkworkerconfig"] = cfg.SparkWorkerConfig
	}
	if cfg.MastersCount != 0 {
		res.Data["mastercount"] = strconv.Itoa(cfg.MastersCount)
	}
	if cfg.WorkersCount != 0 {
		res.Data["workercount"] = strconv.Itoa(cfg.WorkersCount)
	}
	if cfg.Metrics != "" {
		res.Data["metrics"] = cfg.Metrics
	}
	if cfg.ExposeWebUI != "" {
		res.Data["exposeui"] = cfg.ExposeWebUI
	}
	res.Name = cfg.ConfigName
	return &res
}

func addLineFeeds(cmap *api.ConfigMap) {
	for _, v := range cmap.Data {
		v = "\n" + v + "\n"
	}
}

// We need something that implements the kube ConfigMapsInterface since we
// are not conntected to a real client
type FakeConfigMapsClient struct {
	Configs api.ConfigMapList
}

func (f *FakeConfigMapsClient) Get(name string) (*api.ConfigMap, error) {
	for c := range f.Configs.Items {
		if f.Configs.Items[c].Name == name {
			return &f.Configs.Items[c], nil
		}
	}
	return nil, errors.New(fmt.Sprintf("configmaps \"%s\" not found", name))
}

func (f *FakeConfigMapsClient) List(opts api.ListOptions) (*api.ConfigMapList, error) {
	return nil, nil
}

func (c *FakeConfigMapsClient) Create(cfg *api.ConfigMap) (*api.ConfigMap, error) {
	c.Configs.Items = append(c.Configs.Items, *cfg)
	return cfg, nil
}

func (f *FakeConfigMapsClient) Delete(string) error {
	return nil
}

func (f *FakeConfigMapsClient) Update(*api.ConfigMap) (*api.ConfigMap, error) {
	return nil, nil
}

func (f *FakeConfigMapsClient) Watch(api.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (s *OshinkoUnitTestSuite) TestNoLocalDefault(c *check.C) {
	// Test that if we ask for a named config "default" we do not
	// get an error if there is no local override of default.
	// For all other named configs, an error is returned if the local
	// definition is not found.
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}
	var cm1 *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusters.GetDefaultConfig()
	configarg := clusters.ClusterConfig{
		ConfigName:   clusters.DefaultName,
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: clusters.SentinelCountValue}
	myconfig, err := clusters.GetClusterConfig(&configarg, cm1, cm)

	c.Assert(err, check.IsNil)
	c.Assert(myconfig.MastersCount, check.Equals, defconfig.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
}

func (s *OshinkoUnitTestSuite) TestDefaultConfig(c *check.C) {
	// Test that with no config object passed in, we get the default config
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}
	defconfig := clusters.GetDefaultConfig()
	myconfig, err := clusters.GetClusterConfig(nil, cm)
	c.Assert(myconfig.MastersCount, check.Equals, defconfig.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(myconfig.ExposeWebUI, check.Equals, defconfig.ExposeWebUI)
	c.Assert(myconfig.ConfigName, check.Equals, "")
	c.Assert(err, check.IsNil)

	// Test that with a config object containing sentinel values, we get the default config
	configarg := clusters.ClusterConfig{
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: clusters.SentinelCountValue}
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, defconfig.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(myconfig.ConfigName, check.Equals, "")
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamed(c *check.C) {
	// Test that named configs can inherit and override parts of the default config
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}
	defconfig := clusters.GetDefaultConfig()

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := clusters.ClusterConfig{
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: clusters.SentinelCountValue}

	// tiny should inherit the default worker count
	cm.Create(makeConfigMap(tiny))
	configarg.ConfigName = tiny.ConfigName
	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, tiny.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

	// small supplies values for everything
	cm.Create(makeConfigMap(small))
	configarg.ConfigName = small.ConfigName
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, small.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, small.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

	// large should inherit everything but the workercount
	cm.Create(makeConfigMap(large))
	configarg.ConfigName = large.ConfigName
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, defconfig.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, large.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, large.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, large.SparkWorkerConfig)
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamedLinefeed(c *check.C) {
	// Depending on how a configmap is created in openshift, values may
	// have trailing linefeeds. Add linefeeds to the config values and
	// verify that ints are read correctly and strings do not contain
	// linefeeds. For our configs, string values should never legitimately
	// contain a linefeed.
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	sm := makeConfigMap(small)
	addLineFeeds(sm)
	cm.Create(sm)
	configarg := clusters.ClusterConfig{
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: clusters.SentinelCountValue}
	configarg.ConfigName = small.ConfigName

	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, small.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, small.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigArgs(c *check.C) {
	// Test that a config object with no name but with args will
	// inherit and override defaults
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusters.GetDefaultConfig()

	configarg := clusters.ClusterConfig{
		WorkersCount:      7,
		MastersCount:      clusters.SentinelCountValue,
		SparkMasterConfig: "test-master-config",
		SparkWorkerConfig: "test-worker-config"}

	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, defconfig.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, 7)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, "test-master-config")
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, "test-worker-config")
	c.Assert(err, check.IsNil)

	configarg = clusters.ClusterConfig{WorkersCount: clusters.SentinelCountValue, MastersCount: 7}
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, 7)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(err, check.NotNil) // master count is illegal ...

	configarg = clusters.ClusterConfig{WorkersCount: 7, MastersCount: 7}
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, 7)
	c.Assert(myconfig.WorkersCount, check.Equals, 7)
	c.Assert(err, check.NotNil) // master count is illegal ...
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamedArgs(c *check.C) {
	// Test that a named config with args will override and inherit
	// defaults, and that the args will take precedence
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusters.GetDefaultConfig()

	cm.Create(makeConfigMap(brokenMaster))
	configarg := clusters.ClusterConfig{ConfigName: brokenMaster.ConfigName, WorkersCount: 7, MastersCount: 1}
	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, 1)
	c.Assert(myconfig.WorkersCount, check.Equals, 7)
	c.Assert(brokenMaster.MastersCount, check.Not(check.Equals), int64(1))
	c.Assert(err, check.IsNil)

	configarg = clusters.ClusterConfig{
		ConfigName:   brokenMaster.ConfigName,
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: 5}
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, 5)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(brokenMaster.MastersCount, check.Not(check.Equals), defconfig.MastersCount)
	c.Assert(err, check.NotNil) // master count is wrong

	cm.Create(makeConfigMap(small))
	configarg = clusters.ClusterConfig{
		ConfigName:        small.ConfigName,
		SparkMasterConfig: "test-master-config",
		SparkWorkerConfig: "test-worker-config",
		WorkersCount:      clusters.SentinelCountValue,
		MastersCount:      clusters.SentinelCountValue}
	myconfig, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, "test-master-config")
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, "test-worker-config")
	c.Assert(small.SparkMasterConfig, check.Not(check.Equals), "test-master-config")
	c.Assert(small.SparkWorkerConfig, check.Not(check.Equals), "test-worker-config")
	c.Assert(myconfig.MastersCount, check.Equals, small.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, small.WorkersCount)
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterBadConfig(c *check.C) {
	// Test that master count != 1 and worker count < 1 raises an error
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusters.GetDefaultConfig()

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := clusters.ClusterConfig{
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: clusters.SentinelCountValue}

	// brokenmaster should result in an error because the mastercount is != 1
	cm.Create(makeConfigMap(brokenMaster))
	configarg.ConfigName = brokenMaster.ConfigName
	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, brokenMaster.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(brokenMaster.MastersCount, check.Not(check.Equals), 1)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, clusters.MasterCountMustBeZeroOrOne)
}

func (s *OshinkoUnitTestSuite) TestGetClusterNoConfig(c *check.C) {
	// Test that referencing a named config that doesn't exist fails
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defconfig := clusters.GetDefaultConfig()
	configarg := clusters.ClusterConfig{WorkersCount: 0, MastersCount: 0, ConfigName: "notthere"}

	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, defconfig.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, defconfig.WorkersCount)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, fmt.Sprintf(clusters.NamedConfigDoesNotExist, "notthere"))
}

func (s *OshinkoUnitTestSuite) TestGetClusterNonInts(c *check.C) {
	// Test that master count and worker count must be ints
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := clusters.ClusterConfig{WorkersCount: 0, MastersCount: 0}

	m := makeConfigMap(nonIntMaster)
	m.Data["mastercount"] = "fish"
	cm.Create(m)
	configarg.ConfigName = nonIntMaster.ConfigName
	_, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals,
		fmt.Sprintf(clusters.ErrorWhileProcessing,
			configarg.ConfigName+".mastercount", "expected integer, got 'fish'"))

	w := makeConfigMap(nonIntWorker)
	w.Data["workercount"] = "dog"
	cm.Create(w)
	configarg.ConfigName = nonIntWorker.ConfigName
	_, err = clusters.GetClusterConfig(&configarg, cm)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals,
		fmt.Sprintf(clusters.ErrorWhileProcessing,
			configarg.ConfigName+".workercount", "expected integer, got 'dog'"))
}

func (s *OshinkoUnitTestSuite) TestGetClusterUserDefault(c *check.C) {
	// Test that defaults can be overridden optionally with a named
	// "default-oshinko-cluster-config" config
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	defaultconfig := clusters.GetDefaultConfig()
	olddefault, err := clusters.GetClusterConfig(nil, cm)
	c.Assert(err, check.IsNil)
	c.Assert(defaultconfig, check.Equals, olddefault)

	cm.Create(makeConfigMap(userDefault))
	newdefault, err := clusters.GetClusterConfig(nil, cm)
	c.Assert(newdefault, check.Equals, userDefault)
}

func (s *OshinkoUnitTestSuite) TestGetClusterBadElements(c *check.C) {
	// Test that bogus config elements don't break anything
	var cm *FakeConfigMapsClient = &FakeConfigMapsClient{}

	configarg := clusters.ClusterConfig{
		ConfigName:   small.ConfigName,
		WorkersCount: clusters.SentinelCountValue,
		MastersCount: clusters.SentinelCountValue}

	sm := makeConfigMap(small)
	sm.Data["somethingelse"] = "chicken"
	cm.Create(sm)

	myconfig, err := clusters.GetClusterConfig(&configarg, cm)
	c.Assert(myconfig.MastersCount, check.Equals, small.MastersCount)
	c.Assert(myconfig.WorkersCount, check.Equals, small.WorkersCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)
}
