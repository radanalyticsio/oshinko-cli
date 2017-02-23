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
)

// Test connects gocheck to the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

// OshinkoUnitTestSuite can be used for data that may be passed to
// individual tests, or state information that is needed by tests.
type OshinkoUnitTestSuite struct{
}

var _ = check.Suite(&OshinkoUnitTestSuite{})

// SetUpSuite is run once before the entire test suite
func (s *OshinkoUnitTestSuite) SetUpSuite(c *check.C) {}

// SetUpTest is run once before each test
func (s *OshinkoUnitTestSuite) SetUpTest(c *check.C) {}

// TearDownSuite is run once after all tests have finished
func (s *OshinkoUnitTestSuite) TearDownSuite(c *check.C) {}

// TearDownTest is run once after each test has finished
func (s *OshinkoUnitTestSuite) TearDownTest(c *check.C) {}
