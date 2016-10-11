package clienttest

import (
	"fmt"
	"time"

	check "gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-rest/client/clusters"
	"github.com/radanalyticsio/oshinko-rest/helpers/errors"
	"github.com/radanalyticsio/oshinko-rest/models"
)

type lessThanChecker struct {
	*check.CheckerInfo
}

// The LessThan checker attempts to determine if the observed value is
// less than the expected. It will only attempt this on values that can be
// cast as int64.
var LessThan check.Checker = &lessThanChecker{
	&check.CheckerInfo{Name: "LessThan", Params: []string{"obtained", "compare"}},
}

func (checker *lessThanChecker) Check(params []interface{}, names []string) (result bool, error string) {
	defer func() {
		if v := recover(); v != nil {
			result = false
			error = fmt.Sprint(v)
		}
	}()
	return params[0].(int) < params[1].(int), ""
}

func (s *OshinkoRestTestSuite) TestCreateAndDeleteCluster(c *check.C) {
	cname := "e2ecluster"
	mcount := int64(1)
	wcount := int64(3)

	cconfig := models.NewClusterConfig{MasterCount: mcount, WorkerCount: wcount}
	cdetails := models.NewCluster{Name: &cname, Config: &cconfig}
	clparams := clusters.NewCreateClusterParams().WithCluster(&cdetails)

	// create a cluster
	_, err := s.cli.Clusters.CreateCluster(clparams)
	if err != nil {
		msg := err.(*clusters.CreateClusterDefault).Error() + "\n"
		for _, e := range err.(*clusters.CreateClusterDefault).Payload.Errors {
			msg += errors.SingleErrorToString(e)
		}
		c.Fatal(msg)
	}

	// read the cluster details
	// because it may take some time for the pods to become available, we
	// must loop and try multiple times to read them. if we fail to read
	// them after a set number of retries, we consider the test to have
	// failed.
	const retries = 30
	var tries int
	obsmcount, obswcount := int64(0), int64(0)
	cldparams := clusters.NewFindSingleClusterParams().WithName(cname)
	for tries = 0; tries < retries; tries++ {
		cldresult, err := s.cli.Clusters.FindSingleCluster(cldparams)
		if err != nil {
			msg := fmt.Sprintf("%s \n %+v",
				err.(*clusters.FindSingleClusterDefault).Error(),
				err.(*clusters.FindSingleClusterDefault).Payload.Errors)
			c.Fatal(msg)
		}

		// if enough pods are available, we loop through them and exit
		// the retry loop, otherwise sleep for 1 second.
		if len(cldresult.Payload.Cluster.Pods) >= int(mcount+wcount) {
			// loop through the pods to count workers and masters
			obsmcount = 0
			obswcount = 0
			for _, pod := range cldresult.Payload.Cluster.Pods {
				switch *pod.Type {
				case "master":
					if *pod.Status == "Running" {
						obsmcount += 1
					}
				case "worker":
					if *pod.Status == "Running" {
						obswcount += 1
					}
				}
			}
			if obsmcount+obswcount >= mcount+wcount {
				break
			}
		}

		time.Sleep(1 * time.Second)
	}

	c.Assert(tries, LessThan, retries)
	c.Assert(obsmcount, check.Equals, mcount)
	c.Assert(obswcount, check.Equals, wcount)

	// delete the cluster
	delparams := clusters.NewDeleteSingleClusterParams().WithName(cname)
	_, err = s.cli.Clusters.DeleteSingleCluster(delparams)
	if err != nil {
		switch err.(type) {
		case *clusters.DeleteSingleClusterDefault:
			c.Fatal(err.(*clusters.DeleteSingleClusterDefault).Error())
		default:
			c.Fatal(err)
		}
	}

	// confirm delete
	obsclcount := 0
	for tries = 0; tries < retries; tries++ {
		fcresult, err := s.cli.Clusters.FindClusters(nil)
		if err != nil {
			c.Fatal(err.(*clusters.FindClustersDefault).Error())
		}
		if obsclcount = len(fcresult.Payload.Clusters); obsclcount == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	c.Assert(tries, LessThan, retries)
	expclcount := 0
	c.Assert(obsclcount, check.Equals, expclcount)
}
