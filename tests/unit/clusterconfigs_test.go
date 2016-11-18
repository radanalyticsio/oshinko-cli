package unittest

import (
	"path"
	"gopkg.in/check.v1"
	"github.com/radanalyticsio/oshinko-rest/helpers/clusterconfigs"
	"github.com/radanalyticsio/oshinko-rest/models"
	"fmt"
)

func (s *OshinkoUnitTestSuite) TestGetConfigPath(c *check.C) {
	// Test the ability to override the config path
	configpath := clusterconfigs.GetConfigPath()
	c.Assert(configpath, check.Equals, clusterconfigs.DefaultConfigPath)
	clusterconfigs.SetConfigPath(s.Configpath)
	newconfigpath := clusterconfigs.GetConfigPath()
	c.Assert(newconfigpath, check.Equals, s.Configpath)
}

func (s *OshinkoUnitTestSuite) TestNoLocalDefault(c *check.C) {
	// Test that if we ask for a named config "default" we do not
	// get an error if there is no local override of default.
	// For all other named configs, an error is returned if the local
	// definition is not found.
	DeleteDefaultConfig(s)
	defconfig := clusterconfigs.GetDefaultConfig()
	configarg := models.NewClusterConfig{Name: clusterconfigs.Defaultname}
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(err, check.IsNil)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
}

func (s *OshinkoUnitTestSuite) TestDefaultConfig(c *check.C) {
	// Test that with no config object passed in, we get the default config
	defconfig := clusterconfigs.GetDefaultConfig()
	myconfig, err := clusterconfigs.GetClusterConfig(nil)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(myconfig.Name, check.Equals, "")
	c.Assert(err, check.IsNil)

	// Test that with a config object containing zeroes, we get the default config
	configarg := models.NewClusterConfig{WorkerCount: 0, MasterCount: 0}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(myconfig.Name, check.Equals, "")
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamed(c *check.C) {
	// Test that named configs can inherit and override parts of the default config
	defconfig := clusterconfigs.GetDefaultConfig()
	clusterconfigs.SetConfigPath(s.Configpath)

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := models.NewClusterConfig{WorkerCount: 0, MasterCount: 0}

	// tiny should inherit the default worker count
	configarg.Name = s.Tiny.Name
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, s.Tiny.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, defconfig.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, defconfig.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

	// small supplies values for everything
	configarg.Name = s.Small.Name
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, s.Small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, s.Small.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, s.Small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, s.Small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)

	// large should inherit everything but the workercount
	configarg.Name = s.Large.Name
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, s.Large.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, s.Large.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, s.Large.SparkWorkerConfig)
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigArgs(c *check.C) {
	// Test that a config object with no name but with args will
	// inherit and override defaults
	defconfig := clusterconfigs.GetDefaultConfig()
	clusterconfigs.SetConfigPath(s.Configpath)

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := models.NewClusterConfig{WorkerCount: 7, MasterCount: 0,
		SparkMasterConfig: "test-master-config", SparkWorkerConfig: "test-worker-config"}

	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, int64(7))
	c.Assert(myconfig.SparkMasterConfig, check.Equals, "test-master-config")
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, "test-worker-config")
	c.Assert(err, check.IsNil)

	configarg = models.NewClusterConfig{WorkerCount: 0, MasterCount: 7}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, int64(7))
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(err, check.NotNil) // master count is illegal ...

	configarg = models.NewClusterConfig{WorkerCount: 7, MasterCount: 7}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, int64(7))
	c.Assert(myconfig.WorkerCount, check.Equals, int64(7))
	c.Assert(err, check.NotNil) // master count is illegal ...
}

func (s *OshinkoUnitTestSuite) TestGetClusterConfigNamedArgs(c *check.C) {
	// Test that a named config with args will override and inherit
	// defaults, and that the args will take precedence
	defconfig := clusterconfigs.GetDefaultConfig()
	clusterconfigs.SetConfigPath(s.Configpath)

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := models.NewClusterConfig{Name: s.BrokenMaster.Name, WorkerCount: 7, MasterCount: 1}
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, int64(1))
	c.Assert(myconfig.WorkerCount, check.Equals, int64(7))
	c.Assert(s.BrokenMaster.MasterCount, check.Not(check.Equals), int64(1))
	c.Assert(err, check.IsNil)

	configarg = models.NewClusterConfig{Name: s.BrokenMaster.Name, WorkerCount: 0, MasterCount: 5}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, int64(5))
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(s.BrokenMaster.MasterCount, check.Not(check.Equals), defconfig.WorkerCount)
	c.Assert(err, check.NotNil) // master count is wrong

	configarg = models.NewClusterConfig{Name: s.Small.Name, SparkMasterConfig: "test-master-config", SparkWorkerConfig: "test-worker-config"}
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, "test-master-config")
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, "test-worker-config")
	c.Assert(s.Small.SparkMasterConfig, check.Not(check.Equals), "test-master-config")
	c.Assert(s.Small.SparkWorkerConfig, check.Not(check.Equals), "test-worker-config")
	c.Assert(myconfig.MasterCount, check.Equals, s.Small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, s.Small.WorkerCount)
	c.Assert(err, check.IsNil)
}

func (s *OshinkoUnitTestSuite) TestGetClusterBadConfig(c *check.C) {
	// Test that master count != 1 and worker count < 1 raises an error
	defconfig := clusterconfigs.GetDefaultConfig()
	clusterconfigs.SetConfigPath(s.Configpath)

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := models.NewClusterConfig{WorkerCount: 0, MasterCount: 0}

	// brokenmaster should result in an error because the mastercount is != 1
	configarg.Name = s.BrokenMaster.Name
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, s.BrokenMaster.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(s.BrokenMaster.MasterCount, check.Not(check.Equals), 1)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, clusterconfigs.MasterCountMustBeOne)

	// brokenworker should result in an error because the workercount is 0
	configarg.Name = s.BrokenWorker.Name
	myconfig, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, s.BrokenWorker.WorkerCount)
	w := s.BrokenWorker.WorkerCount < 1
	c.Assert(w, check.Equals, true)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, clusterconfigs.WorkerCountMustBeAtLeastOne)
}

func (s *OshinkoUnitTestSuite) TestGetClusterNoConfig(c *check.C) {
	// Test that referencing a named config that doesn't exist fails
	defconfig := clusterconfigs.GetDefaultConfig()
	clusterconfigs.SetConfigPath(s.Configpath)
	configarg := models.NewClusterConfig{WorkerCount: 0, MasterCount: 0, Name: "notthere"}

	// should return an error because the config doesn't exist
	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, defconfig.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, defconfig.WorkerCount)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, fmt.Sprintf(clusterconfigs.NamedConfigDoesNotExist, "notthere"))
}

func (s *OshinkoUnitTestSuite) TestGetClusterNonInts(c *check.C) {
	// Test that master count and worker count must be ints
	clusterconfigs.SetConfigPath(s.Configpath)
	configarg := models.NewClusterConfig{WorkerCount: 0, MasterCount: 0}

	configarg.Name = s.NonIntMaster.Name
	_, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals,
		fmt.Sprintf(clusterconfigs.ErrorWhileProcessing,
			path.Join(s.Configpath, configarg.Name + ".mastercount"), "expected integer"))

	configarg.Name = s.NonIntWorker.Name
	_, err = clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals,
		fmt.Sprintf(clusterconfigs.ErrorWhileProcessing,
			path.Join(s.Configpath, configarg.Name + ".workercount"), "expected integer"))
}

func (s *OshinkoUnitTestSuite) TestGetClusterUserDefault(c *check.C) {
	// Test that defaults can be overridden optionally with a named
	// "default" config in the configdir
	defaultconfig := clusterconfigs.GetDefaultConfig()
	olddefault, err := clusterconfigs.GetClusterConfig(nil)
	c.Assert(err, check.IsNil)
	c.Assert(defaultconfig, check.Equals, olddefault)

	clusterconfigs.SetConfigPath(s.UserConfigpath)
	newdefault, err := clusterconfigs.GetClusterConfig(nil)
	c.Assert(newdefault, check.Equals, s.UserDefault)
}

func (s *OshinkoUnitTestSuite) TestGetClusterBadElements(c *check.C) {
	// Test that bogus config elements don't break anything
	// UserConfigpath contains a "small" configuration with extra elements
	clusterconfigs.SetConfigPath(s.UserConfigpath)

	// configarg will represent a config object passed in a REST
	// request which specifies a named config but leaves counts unset
	configarg := models.NewClusterConfig{Name: s.Small.Name, WorkerCount: 0, MasterCount: 0}

	myconfig, err := clusterconfigs.GetClusterConfig(&configarg)
	c.Assert(myconfig.MasterCount, check.Equals, s.Small.MasterCount)
	c.Assert(myconfig.WorkerCount, check.Equals, s.Small.WorkerCount)
	c.Assert(myconfig.SparkMasterConfig, check.Equals, s.Small.SparkMasterConfig)
	c.Assert(myconfig.SparkWorkerConfig, check.Equals, s.Small.SparkWorkerConfig)
	c.Assert(err, check.IsNil)
}
