package cmd

import (
	"fmt"
	"strconv"
	"github.com/openshift/origin/pkg/client"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)

type SparkCluster struct {
	Namespace       string `json:"namespace,omitempty"`
	Name            string `json:"name,omitempty"`
	Href            string `json:"href"`
	Image           string `json:"image"`
	MasterURL       string `json:"masterUrl"`
	MasterWebURL    string `json:"masterWebUrl"`
	Status          string `json:"status"`
	WorkerCount     int    `json:"workerCount"`
	MasterCount     int    `json:"masterCount,omitempty"`
	MasterConfig    string `json:"sparkMasterConfig,omitempty"`
	MasterConfigDir string `json:"masterConfigDir,omitempty"`
	WorkerConfig    string `json:"workerConfig,omitempty"`
	WorkerConfigDir string `json:"workerConfigDir,omitempty"`
}

type SortByClusterName []SparkCluster

func (p SortByClusterName) Len() int {
	return len(p)
}
func (p SortByClusterName) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p SortByClusterName) Less(i, j int) bool {
	return p[i].Name < p[j].Name
}

func checkForConfigMap(name string, cm kclient.ConfigMapsInterface) error {
	if name == "" {
		return fmt.Errorf("ConfigMap not provided\n")
	}
	cmap, err := cm.Get(name)
	if err == nil && cmap == nil {
		err = fmt.Errorf("ConfigMap '%s' not found", name)
	}
	return err
}

func makeSelector(otype string, clustername string) kapi.ListOptions {

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

func (s *SparkCluster) countWorkers(kclient *kclient.Client) (int, error) {
	// If we are  unable to retrieve a list of worker pods, return -1 for count
	// This is an error case, differnt from a list of length 0. Let the caller
	// decide whether to report the error or the -1 count
	pc := kclient.Pods(s.Namespace)
	cnt := 0
	selectorlist := makeSelector(workerType, s.Name)
	pods, err := pc.List(selectorlist)
	if pods != nil {
		cnt = len(pods.Items)
		s.WorkerCount = cnt
	}
	return s.WorkerCount, err
}

func (s *SparkCluster) retrieveServiceURL(kclient *kclient.Client, stype string) string {
	selectorlist := makeSelector(stype, s.Name)
	sc := kclient.Services(s.Namespace)
	srvs, err := sc.List(selectorlist)
	if err == nil && len(srvs.Items) != 0 {
		srv := srvs.Items[0]
		scheme := "http://"
		if stype == masterType {
			scheme = "spark://"
			s.MasterURL = scheme + srv.Name + ":" + strconv.Itoa(srv.Spec.Ports[0].Port)
			return s.MasterURL
		} else {
			s.MasterWebURL = scheme + srv.Name + ":" + strconv.Itoa(srv.Spec.Ports[0].Port)
			return s.MasterWebURL
		}
	}
	return ""
}

//TODO move to struct
func getReplController(client kclient.ReplicationControllerInterface, clustername, otype string) (*kapi.ReplicationController, error) {

	selectorlist := makeSelector(otype, clustername)
	repls, err := client.List(selectorlist)
	if err != nil || len(repls.Items) == 0 {
		return nil, err
	}
	// Use the latest replication controller.  There could be more than one
	// if the user did something like oc env to set a new env var on a deployment
	newestRepl := repls.Items[0]
	for i := 0; i < len(repls.Items); i++ {
		if repls.Items[i].CreationTimestamp.Unix() > newestRepl.CreationTimestamp.Unix() {
			newestRepl = repls.Items[i]
		}
	}
	return &newestRepl, err
}

func checkForDeploymentConfigs(client client.DeploymentConfigInterface, clustername string) (bool, error) {
	selectorlist := makeSelector(masterType, clustername)
	dcs, err := client.List(selectorlist)
	if err != nil {
		return false, err
	}
	if len(dcs.Items) == 0 {
		return false, nil
	}
	selectorlist = makeSelector(workerType, clustername)
	dcs, err = client.List(selectorlist)
	if err != nil {
		return false, err
	}
	if len(dcs.Items) == 0 {
		return false, nil
	}
	return true, nil
}