package unittest

/*
INFO(elmiko) This file exists as a helper to ensure that the tests in this
directory are hooked into the go testing infrastructure. The Test function
declaration needs to be included once in a package for tests, hence the
existence of this file.
*/

import (
	"testing"

	"gopkg.in/check.v1"
	"os"
	"path"
	"strconv"
	"github.com/radanalyticsio/oshinko-rest/models"
	"github.com/radanalyticsio/oshinko-rest/helpers/clusterconfigs"
)

// Test connects gocheck to the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

// OshinkoUnitTestSuite can be used for data that may be passed to
// individual tests, or state information that is needed by tests.
type OshinkoUnitTestSuite struct{
	Configpath string
	UserConfigpath string
	Tiny models.NewClusterConfig
	Small models.NewClusterConfig
	Large models.NewClusterConfig
	BrokenMaster models.NewClusterConfig
	BrokenWorker models.NewClusterConfig
	NonIntMaster models.NewClusterConfig
	NonIntWorker models.NewClusterConfig
	UserDefault models.NewClusterConfig
}

var _ = check.Suite(&OshinkoUnitTestSuite{})

const defaultmastercount = clusterconfigs.Defaultname + ".mastercount"
const defaultworkercount = clusterconfigs.Defaultname + ".workercount"
const defaultmasterconfig = clusterconfigs.Defaultname + ".sparkmasterconfig"
const defaultworkerconfig = clusterconfigs.Defaultname + ".sparkworkerconfig"

func makeConfig(dir string, name string, val string) {
	f, err := os.Create(path.Join(dir, name))
	if err == nil {
		f.WriteString(val)
		f.Close()
	}
}

func deleteConfig(dir string, name string) {
	_ = os.Remove(path.Join(dir, name))
}

func DeleteDefaultConfig(s *OshinkoUnitTestSuite) {
	deleteConfig(s.UserConfigpath, defaultmastercount)
	deleteConfig(s.UserConfigpath, defaultworkercount)
}

func MakeDefaultConfig(s *OshinkoUnitTestSuite) {
	makeConfig(s.UserConfigpath, defaultmastercount, strconv.Itoa(int(s.UserDefault.WorkerCount)))
	makeConfig(s.UserConfigpath, defaultworkercount, strconv.Itoa(int(s.UserDefault.WorkerCount)))
	makeConfig(s.UserConfigpath, defaultmasterconfig, s.UserDefault.SparkMasterConfig)
	makeConfig(s.UserConfigpath, defaultworkerconfig, s.UserDefault.SparkWorkerConfig)
}

// SetUpSuite is run once before the entire test suite
func (s *OshinkoUnitTestSuite) SetUpSuite(c *check.C) {
	s.Configpath = path.Join(os.TempDir(), "oshinko-cluster-configs/")
	os.RemoveAll(s.Configpath)
	os.MkdirAll(s.Configpath, os.ModePerm)

	// we'll put a user defined default in oshinko-cluster-configs-user along
	// some extra elements that don't map to properties
	s.UserConfigpath = path.Join(os.TempDir(), "oshinko-cluster-configs-user")
	os.RemoveAll(s.UserConfigpath)
	os.MkdirAll(s.UserConfigpath, os.ModePerm)

	// Save configuration values for reference, they will be written to files too
	s.Tiny = models.NewClusterConfig{MasterCount: 1, WorkerCount: 0, Name: "tiny"}
	s.Small = models.NewClusterConfig{MasterCount: 1, WorkerCount: 3,
		SparkMasterConfig: "master-config", SparkWorkerConfig: "worker-config", Name: "small"}
	s.Large = models.NewClusterConfig{MasterCount: 0, WorkerCount: 10, Name: "large"}
	s.BrokenMaster = models.NewClusterConfig{MasterCount: 2, WorkerCount: 0, Name: "brokenmaster"}
	s.BrokenWorker = models.NewClusterConfig{MasterCount: 1, WorkerCount: 0, Name: "brokenworker"}
	s.NonIntMaster = models.NewClusterConfig{Name: "cow"}
	s.NonIntWorker = models.NewClusterConfig{Name: "pig"}

	// Inherit worker count from default but overwrite master
	makeConfig(s.Configpath, "tiny.mastercount", strconv.Itoa(int(s.Tiny.MasterCount)))

	// Don't inherit either count
	makeConfig(s.Configpath, "small.mastercount", strconv.Itoa(int(s.Small.MasterCount)))
	makeConfig(s.Configpath, "small.workercount", strconv.Itoa(int(s.Small.WorkerCount)))
	makeConfig(s.Configpath, "small.sparkmasterconfig", s.Small.SparkMasterConfig)
	makeConfig(s.Configpath, "small.sparkworkerconfig", s.Small.SparkWorkerConfig)

	// Inherit master count from default but overwrite worker
	makeConfig(s.Configpath, "large.workercount", strconv.Itoa(int(s.Large.WorkerCount)))

	// Set mastercount to something illegal
	makeConfig(s.Configpath, "brokenmaster.mastercount", strconv.Itoa(int(s.BrokenMaster.MasterCount)))

	// Set workercount to something illegal
	makeConfig(s.Configpath, "brokenworker.workercount", strconv.Itoa(int(s.BrokenWorker.WorkerCount)))

	// Create configs with non-int values
	makeConfig(s.Configpath, "cow.mastercount", "cow")
	makeConfig(s.Configpath, "pig.workercount", "pig")

	// Set up a user defined default
	// This will be written at the beginning of each test because some tests may
	// destroy it to change default behavior
	s.UserDefault = models.NewClusterConfig{Name: "",
						MasterCount: 3,
						WorkerCount: 3,
						SparkMasterConfig: "master-default",
						SparkWorkerConfig: "worker-default"}

	// Also create some troublesome name elements in an alternate configpath
	// to make sure it doesn't break anything
	makeConfig(s.UserConfigpath, "small", "fish")
	makeConfig(s.UserConfigpath, "small.somethingelse", "chicken")
	makeConfig(s.UserConfigpath, "small.mastercount", strconv.Itoa(int(s.Small.MasterCount)))
	makeConfig(s.UserConfigpath, "small.workercount", strconv.Itoa(int(s.Small.WorkerCount)))
	makeConfig(s.UserConfigpath, "small.sparkmasterconfig", s.Small.SparkMasterConfig)
	makeConfig(s.UserConfigpath, "small.sparkworkerconfig", s.Small.SparkWorkerConfig)
}

// SetUpTest is run once before each test
func (s *OshinkoUnitTestSuite) SetUpTest(c *check.C) {
}

// TearDownSuite is run once after all tests have finished
func (s *OshinkoUnitTestSuite) TearDownSuite(c *check.C) {}

// TearDownTest is run once after each test has finished
func (s *OshinkoUnitTestSuite) TearDownTest(c *check.C) {}
