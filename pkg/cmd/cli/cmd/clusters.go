package cmd

import (

	"github.com/radanalyticsio/oshinko-rest/restapi/operations/clusters"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)


// SortByProjectName is sort
type SortByClusterName []*clusters.ClustersItems0

//func (p SortByClusterName) Len() int {
//	return len(p)
//}
//func (p SortByClusterName) Swap(i, j int) {
//	p[i], p[j] = p[j], p[i]
//}
//func (p SortByClusterName) Less(i, j int) bool {
//	return p[i].Name < p[j].Name
//}
func (p SortByClusterName) Len() int {
	return len(p)
}
func (p SortByClusterName) Swap(i, j int) {
	*p[i], *p[j] = *p[j], *p[i]
}
func (p SortByClusterName) Less(i, j int) bool {
	return *(p[i].Name) < *(p[j].Name)
}

const (
	clustersLong = `
Display information about the spark clusters on the server.`
	clustersExample = `  # Display the spark cluster %[1]s`
)

const nameSpaceMsg = "Cannot determine target openshift namespace"
const clientMsg = "Unable to create an openshift client"

const typeLabel = "oshinko-type"
const clusterLabel = "oshinko-cluster"

const workerType = "worker"
const masterType = "master"
const webuiType = "webui"

const masterPortName = "spark-master"
const webPortName = "spark-webui"

func makeSelector(otype string, clustername string) kapi.ListOptions {
	const typeLabel = "oshinko-type"
	const clusterLabel = "oshinko-cluster"
	// Build a selector list based on type and/or cluster name
	ls := labels.NewSelector()
	if otype != "" {
		ot, _ := labels.NewRequirement(typeLabel, labels.EqualsOperator, sets.NewString(otype))
		ls = ls.Add(*ot)
	}
	if clustername != "" {
		cname, _ := labels.NewRequirement(clusterLabel, labels.EqualsOperator, sets.NewString(clustername))
		ls = ls.Add(*cname)
	}
	return kapi.ListOptions{LabelSelector: ls}
}

func tostrptr(val string) *string {
	v := val
	return &v
}
func toint64ptr(val int64) *int64 {
	v := val
	return &v
}

func sparkMasterURL(name string, port *kapi.ServicePort) string {
	return "spark://" + name + ":"
	//+ strconv.Itoa(port.Port)
}

func countWorkers(client kclient.PodInterface, clustername string) (int64, *kapi.PodList, error) {
	// If we are  unable to retrieve a list of worker pods, return -1 for count
	// This is an error case, differnt from a list of length 0. Let the caller
	// decide whether to report the error or the -1 count
	cnt := int64(-1)
	selectorlist := makeSelector(workerType, clustername)
	pods, err := client.List(selectorlist)
	if pods != nil {
		cnt = int64(len(pods.Items))
	}
	return cnt, pods, err
}

func retrieveMasterURL(client kclient.ServiceInterface, clustername string) string {
	selectorlist := makeSelector(masterType, clustername)
	srvs, err := client.List(selectorlist)
	if err == nil && len(srvs.Items) != 0 {
		srv := srvs.Items[0]
		return sparkMasterURL(srv.Name, &srv.Spec.Ports[0])
	}
	return ""
}
