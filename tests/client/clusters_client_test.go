package clienttest

import (
	check "gopkg.in/check.v1"

	"github.com/redhatanalytics/oshinko-rest/client/clusters"
	"github.com/redhatanalytics/oshinko-rest/models"
)

func (s *OshinkoRestTestSuite) TestCreateCluster(c *check.C) {
	cname := "TestCluster"
	mcount := int64(1)
	wcount := int64(3)
	cdetails := models.NewCluster{Name: &cname, MasterCount: &mcount, WorkerCount: &wcount}
	params := clusters.NewCreateClusterParams().WithCluster(&cdetails)

	_, err := s.cli.Clusters.CreateCluster(params)

	expectedStatusCode := 500
	observedStatusCode := err.(*clusters.CreateClusterDefault).Code()

	c.Assert(observedStatusCode, check.Equals, expectedStatusCode)
}

func (s *OshinkoRestTestSuite) TestDeleteSingleCluster(c *check.C) {
	params := clusters.NewDeleteSingleClusterParams().WithName("TestCluster")

	_, err := s.cli.Clusters.DeleteSingleCluster(params)

	expectedStatusCode := 500
	observedStatusCode := err.(*clusters.DeleteSingleClusterDefault).Code()

	c.Assert(observedStatusCode, check.Equals, expectedStatusCode)
}

func (s *OshinkoRestTestSuite) TestFindClusters(c *check.C) {
	_, err := s.cli.Clusters.FindClusters(nil)

	expectedStatusCode := 500
	observedStatusCode := err.(*clusters.FindClustersDefault).Code()

	c.Assert(observedStatusCode, check.Equals, expectedStatusCode)
}

func (s *OshinkoRestTestSuite) TestFindSingleCluster(c *check.C) {
	params := clusters.NewFindSingleClusterParams().WithName("TestCluster")

	_, err := s.cli.Clusters.FindSingleCluster(params)

	expectedStatusCode := 500
	observedStatusCode := err.(*clusters.FindSingleClusterDefault).Code()

	c.Assert(observedStatusCode, check.Equals, expectedStatusCode)
}

func (s *OshinkoRestTestSuite) TestUpdateSingleCluster(c *check.C) {
	cname := "TestCluster"
	mcount := int64(1)
	wcount := int64(3)
	cdetails := models.NewCluster{Name: &cname, MasterCount: &mcount, WorkerCount: &wcount}
	params := clusters.NewUpdateSingleClusterParams().WithName("TestCluster").WithCluster(&cdetails)

	_, err := s.cli.Clusters.UpdateSingleCluster(params)

	expectedStatusCode := 500
	observedStatusCode := err.(*clusters.UpdateSingleClusterDefault).Code()

	c.Assert(observedStatusCode, check.Equals, expectedStatusCode)
}
