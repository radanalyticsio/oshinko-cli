package clusters

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	oclient "github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	ocon "github.com/radanalyticsio/oshinko-core/clusters/containers"
	odc "github.com/radanalyticsio/oshinko-core/clusters/deploymentconfigs"
	opt "github.com/radanalyticsio/oshinko-core/clusters/podtemplates"
	"github.com/radanalyticsio/oshinko-core/clusters/probes"
	osv "github.com/radanalyticsio/oshinko-core/clusters/services"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/selection"
	"k8s.io/kubernetes/pkg/util/sets"
)

const clusterConfigMsg = "invalid cluster configuration"
const missingConfigMsg = "unable to find spark configuration '%s'"
const findDepConfigMsg = "unable to find deployment configs"
const createDepConfigMsg = "unable to create deployment config '%s'"
const replMsgWorker = "unable to find replication controller for spark workers"
const replMsgMaster = "unable to find replication controller for spark master"
const masterSrvMsg = "unable to create spark master service endpoint"
const mastermsg = "unable to find spark masters"
const updateReplMsg = "unable to update replication controller for spark workers"
const noSuchClusterMsg = "no such cluster '%s'"
const podListMsg = "unable to retrive pod list"
const sparkImageMsg = "no spark image specified"

const typeLabel = "oshinko-type"
const clusterLabel = "oshinko-cluster"

const workerType = "worker"
const masterType = "master"
const webuiType = "webui"

const masterPortName = "spark-master"
const masterPort = 7077
const webPortName = "spark-webui"
const webPort = 8080

const sparkconfdir = "/etc/oshinko-spark-configs"

// The suffix to add to the spark master hostname (clustername) for the web service
const webServiceSuffix = "-ui"

type SparkPod struct {
	IP     string
	Status string
	Type   string
}

type SparkCluster struct {
	Namespace    string `json:"namespace,omitempty"`
	Name         string `json:"name,omitempty"`
	Href         string `json:"href"`
	Image        string `json:"image"`
	MasterURL    string `json:"masterUrl"`
	MasterWebURL string `json:"masterWebUrl"`
	Status       string `json:"status"`
	WorkerCount  int    `json:"workerCount"`
	MasterCount  int    `json:"masterCount,omitempty"`
	Config       ClusterConfig
	Pods         []SparkPod
}

func generalErr(err error, msg string, code int) ClusterError {
	if err != nil {
		if msg == "" {
			msg = "error: " + err.Error()
		} else {
			msg = msg + ", error: " + err.Error()
		}
	}
	return NewClusterError(msg, code)
}

func makeSelector(otype string, clustername string) kapi.ListOptions {
	// Build a selector list based on type and/or cluster name
	ls := labels.NewSelector()
	if otype != "" {
		ot, _ := labels.NewRequirement(typeLabel, selection.Equals, sets.NewString(otype))
		ls = ls.Add(*ot)
	}
	if clustername != "" {
		cname, _ := labels.NewRequirement(clusterLabel, selection.Equals, sets.NewString(clustername))
		ls = ls.Add(*cname)
	}
	return kapi.ListOptions{LabelSelector: ls}
}

func retrieveServiceURL(client kclient.ServiceInterface, stype, clustername string) string {
	selectorlist := makeSelector(stype, clustername)
	srvs, err := client.List(selectorlist)
	if err == nil && len(srvs.Items) != 0 {
		srv := srvs.Items[0]
		scheme := "http://"
		if stype == masterType {
			scheme = "spark://"
		}
		return scheme + srv.Name + ":" + strconv.Itoa(int(srv.Spec.Ports[0].Port))
	}
	return ""
}

func checkForDeploymentConfigs(client oclient.DeploymentConfigInterface, clustername string) (bool, error) {
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

func makeEnvVars(clustername, sparkconfdir string) []kapi.EnvVar {
	envs := []kapi.EnvVar{}

	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_SPARK_CLUSTER", Value: clustername})
	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_REST_HOST", Value: os.Getenv("OSHINKO_REST_SERVICE_HOST")})
	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_REST_PORT", Value: os.Getenv("OSHINKO_REST_SERVICE_PORT")})
	if sparkconfdir != "" {
		envs = append(envs, kapi.EnvVar{Name: "UPDATE_SPARK_CONF_DIR", Value: sparkconfdir})
	}

	return envs
}

func makeWorkerEnvVars(clustername, sparkconfdir string) []kapi.EnvVar {
	envs := []kapi.EnvVar{}

	envs = makeEnvVars(clustername, sparkconfdir)
	envs = append(envs, kapi.EnvVar{
		Name:  "SPARK_MASTER_ADDRESS",
		Value: "spark://" + clustername + ":" + strconv.Itoa(masterPort)})
	envs = append(envs, kapi.EnvVar{
		Name:  "SPARK_MASTER_UI_ADDRESS",
		Value: "http://" + clustername + webServiceSuffix + ":" + strconv.Itoa(webPort)})
	return envs
}

func sparkWorker(namespace string,
	image string,
	replicas int, clustername, sparkconfdir, sparkworkerconfig string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name.
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(clustername+"-w", namespace).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, clustername).
		Label(typeLabel, workerType).
		PodSelector(clusterLabel, clustername).Replicas(replicas)

	// Create a pod template spec with the matching label
	pt := opt.PodTemplateSpec().Label(clusterLabel, clustername).Label(typeLabel, workerType)

	// Create a container with the correct ports and start command
	webport := 8081
	webp := ocon.ContainerPort(webPortName, webport)
	cont := ocon.Container(dc.Name, image).
		Ports(webp).
		SetLivenessProbe(probes.NewHTTPGetProbe(webport)).EnvVars(makeWorkerEnvVars(clustername, sparkconfdir))

	if sparkworkerconfig != "" {
		pt = pt.SetConfigMapVolume(sparkworkerconfig)
		cont = cont.SetVolumeMount(sparkworkerconfig, sparkconfdir, true)
	}

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func sparkMaster(namespace, image, clustername, sparkconfdir, sparkmasterconfig string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(clustername+"-m", namespace).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, clustername).
		Label(typeLabel, masterType).
		PodSelector(clusterLabel, clustername)

	// Create a pod template spec with the matching label
	pt := opt.PodTemplateSpec().Label(clusterLabel, clustername).
		Label(typeLabel, masterType)

	// Create a container with the correct ports and start command
	httpProbe := probes.NewHTTPGetProbe(webPort)
	masterp := ocon.ContainerPort(masterPortName, masterPort)
	webp := ocon.ContainerPort(webPortName, webPort)
	cont := ocon.Container(dc.Name, image).
		Ports(masterp, webp).
		SetLivenessProbe(httpProbe).
		SetReadinessProbe(httpProbe).EnvVars(makeEnvVars(clustername, sparkconfdir))

	if sparkmasterconfig != "" {
		pt = pt.SetConfigMapVolume(sparkmasterconfig)
		cont = cont.SetVolumeMount(sparkmasterconfig, sparkconfdir, true)
	}

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func service(name string,
	port int,
	clustername, otype string,
	podselectors map[string]string) (*osv.OService, *osv.OServicePort) {

	p := osv.ServicePort(port).TargetPort(port)
	return osv.Service(name).Label(clusterLabel, clustername).
		Label(typeLabel, otype).PodSelectors(podselectors).Ports(p), p
}

func checkForConfigMap(name string, cm kclient.ConfigMapsInterface) error {
	_, err := cm.Get(name)
	if err != nil {
		if strings.Index(err.Error(), "not found") != -1 {
			return generalErr(err, fmt.Sprintf(missingConfigMsg, name), ClusterConfigCode)
		}
		return generalErr(nil, fmt.Sprintf(missingConfigMsg, name), ClientOperationCode)
	}
	return nil
}

func countWorkers(client kclient.PodInterface, clustername string) (int, *kapi.PodList, error) {
	// If we are  unable to retrieve a list of worker pods, return -1 for count
	// This is an error case, differnt from a list of length 0. Let the caller
	// decide whether to report the error or the -1 count
	cnt := -1
	selectorlist := makeSelector(workerType, clustername)
	pods, err := client.List(selectorlist)
	if pods != nil {
		cnt = len(pods.Items)
	}
	return cnt, pods, err
}

// CreateClusterResponse create a cluster and return the representation
func CreateCluster(clustername, namespace, sparkimage string, config *ClusterConfig, osclient *oclient.Client, client *kclient.Client) (SparkCluster, error) {

	var masterconfdir string
	var workerconfdir string
	var result SparkCluster = SparkCluster{}

	createCode := func(err error) int {
		if err != nil && strings.Index(err.Error(), "already exists") != -1 {
			return ComponentExistsCode
		}
		return ClientOperationCode
	}

	masterhost := clustername

	// Copy any named config referenced and update it with any explicit config values
	finalconfig, err := GetClusterConfig(config, client.ConfigMaps(namespace))
	if err != nil {
		return result, generalErr(err, clusterConfigMsg, ErrorCode(err))
	}
	if finalconfig.SparkImage != "" {
		sparkimage = finalconfig.SparkImage
	} else if sparkimage == "" {
		return result, generalErr(nil, sparkImageMsg, ClusterConfigCode)
	}

	workercount := int(finalconfig.WorkerCount)

	// Check if finalconfig contains the names of ConfigMaps to use for spark
	// configuration. If so they must exist. The ConfigMaps will be mounted
	// as volumes on spark pods and the path stored in the environment
	// variable UPDATE_SPARK_CONF_DIR
	cm := client.ConfigMaps(namespace)
	if finalconfig.SparkMasterConfig != "" {
		err := checkForConfigMap(finalconfig.SparkMasterConfig, cm)
		if err != nil {
			return result, err
		}
		masterconfdir = sparkconfdir
	}

	if finalconfig.SparkWorkerConfig != "" {
		err := checkForConfigMap(finalconfig.SparkWorkerConfig, cm)
		if err != nil {
			return result, err
		}
		workerconfdir = sparkconfdir
	}

	// Create the master deployment config
	masterdc := sparkMaster(namespace, sparkimage, clustername, masterconfdir, finalconfig.SparkMasterConfig)

	// Create the services that will be associated with the master pod
	// They will be created with selectors based on the pod labels
	mastersv, _ := service(masterhost,
		masterdc.FindPort(masterPortName),
		clustername, masterType,
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := service(masterhost+webServiceSuffix,
		masterdc.FindPort(webPortName),
		clustername, webuiType,
		masterdc.GetPodTemplateSpecLabels())

	// Create the worker deployment config
	workerdc := sparkWorker(namespace, sparkimage, workercount, clustername, workerconfdir, finalconfig.SparkWorkerConfig)

	// Launch all of the objects
	dcc := osclient.DeploymentConfigs(namespace)
	_, err = dcc.Create(&masterdc.DeploymentConfig)
	if err != nil {
		return result, generalErr(err, fmt.Sprintf(createDepConfigMsg, masterdc.Name), createCode(err))
	}
	_, err = dcc.Create(&workerdc.DeploymentConfig)
	if err != nil {
		// Since we created the master deployment config, try to clean up
		DeleteCluster(clustername, namespace, osclient, client)
		return result, generalErr(err, fmt.Sprintf(createDepConfigMsg, workerdc.Name), createCode(err))
	}

	sc := client.Services(namespace)
	_, err = sc.Create(&mastersv.Service)
	if err != nil {
		// Since we create the master and workers, try to clean up
		DeleteCluster(clustername, namespace, osclient, client)
		return result, generalErr(err, masterSrvMsg, createCode(err))
	}

	// Note, if spark webui service fails for some reason we can live without it
	// TODO ties into cluster status, make a note if the service is missing
	sc.Create(&websv.Service)

	// Wait for the replication controllers to exist before building the response.
	rcc := client.ReplicationControllers(namespace)
	{
		var mrepl, wrepl *kapi.ReplicationController
		mrepl = nil
		wrepl = nil
		for i := 0; i < 4; i++ {
			if mrepl == nil {
				mrepl, _ = getReplController(rcc, clustername, masterType)
			}
			if wrepl == nil {
				wrepl, _ = getReplController(rcc, clustername, workerType)
			}
			if wrepl != nil && mrepl != nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
	}

	result.Name = clustername
	result.Namespace = namespace
	result.Href = "/clusters/" + clustername
	result.Image = sparkimage
	result.MasterURL = retrieveServiceURL(sc, masterType, clustername)
	result.MasterWebURL = retrieveServiceURL(sc, webuiType, clustername)
	if result.MasterURL == "" {
		result.Status = "MasterServiceMissing"

	} else {
		result.Status = "Running"
	}
	result.Config = finalconfig
	result.MasterCount = 0
	result.WorkerCount = 0
	result.Pods = []SparkPod{}

	return result, nil
}

func waitForCount(client kclient.ReplicationControllerInterface, name string, count int) {

	for i := 0; i < 5; i++ {
		r, _ := client.Get(name)
		if int(r.Status.Replicas) == count {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func DeleteCluster(clustername, namespace string, osclient *oclient.Client, client *kclient.Client) (string, error) {
	var foundSomething bool = false
	info := []string{}
	scalerepls := []string{}

	// Build a selector list for the "oshinko-cluster" label
	selectorlist := makeSelector("", clustername)

	// Delete all of the deployment configs
	dcc := osclient.DeploymentConfigs(namespace)
	deployments, err := dcc.List(selectorlist)
	if err != nil {
		info = append(info, "unable to find deployment configs ("+err.Error()+")")
	} else {
		foundSomething = len(deployments.Items) > 0
	}
	for i := range deployments.Items {
		name := deployments.Items[i].Name
		err = dcc.Delete(name)
		if err != nil {
			info = append(info, "unable to delete deployment config "+name+" ("+err.Error()+")")
		}
	}

	// Get a list of all the replication controllers for the cluster
	// and set all of the replica values to 0
	rcc := client.ReplicationControllers(namespace)
	repls, err := rcc.List(selectorlist)
	if err != nil {
		info = append(info, "unable to find replication controllers ("+err.Error()+")")
	} else {
		foundSomething = foundSomething || len(repls.Items) > 0
	}
	for i := range repls.Items {
		name := repls.Items[i].Name
		repls.Items[i].Spec.Replicas = 0
		_, err = rcc.Update(&repls.Items[i])
		if err != nil {
			info = append(info, "unable to scale replication controller "+name+" ("+err.Error()+")")
		} else {
			scalerepls = append(scalerepls, name)
		}
	}

	// Wait for the replica count to drop to 0 for each one we scaled
	for i := range scalerepls {
		waitForCount(rcc, scalerepls[i], 0)
	}

	// Delete each replication controller
	for i := range repls.Items {
		name := repls.Items[i].Name
		err = rcc.Delete(name, nil)
		if err != nil {
			info = append(info, "unable to delete replication controller "+name+" ("+err.Error()+")")
		}
	}

	// Delete the services
	sc := client.Services(namespace)
	srvs, err := sc.List(selectorlist)
	if err != nil {
		info = append(info, "unable to find services ("+err.Error()+")")
	} else {
		foundSomething = foundSomething || len(srvs.Items) > 0
	}
	for i := range srvs.Items {
		name := srvs.Items[i].Name
		err = sc.Delete(name)
		if err != nil {
			info = append(info, "unable to delete service "+name+" ("+err.Error()+")")
		}
	}

	// If we found some part of a cluster, then there is no error
	// even though the cluster may not have been fully complete.
	// If we didn't find any trace of a cluster, then call it an error
	if !foundSomething {
		return "", generalErr(nil, fmt.Sprintf(noSuchClusterMsg, clustername), NoSuchClusterCode)
	}
	return strings.Join(info, ", "), nil
}

// FindSingleClusterResponse find a cluster and return its representation
func FindSingleCluster(name, namespace string, osclient *oclient.Client, client *kclient.Client) (SparkCluster, error) {

	addpod := func(p kapi.Pod) SparkPod {
		return SparkPod{
			IP:     p.Status.PodIP,
			Status: string(p.Status.Phase),
			Type:   p.Labels[typeLabel],
		}
	}

	clustername := name

	var result SparkCluster = SparkCluster{}

	// Before we do further checks, make sure that we have deploymentconfigs
	// If either the master or the worker deploymentconfig are missing, we
	// assume that the cluster is missing. These are the base objects that
	// we use to create a cluster
	ok, err := checkForDeploymentConfigs(osclient.DeploymentConfigs(namespace), clustername)
	if err != nil {
		return result, generalErr(err, findDepConfigMsg, ClientOperationCode)
	}
	if !ok {
		return result, generalErr(nil, fmt.Sprintf(noSuchClusterMsg, clustername), NoSuchClusterCode)
	}

	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	rcc := client.ReplicationControllers(namespace)
	mrepl, err := getReplController(rcc, clustername, masterType)
	if err != nil {
		return result, generalErr(err, replMsgMaster, ClientOperationCode)
	} else if mrepl == nil {
		return result, generalErr(err, replMsgMaster, ClusterIncompleteCode)
	}
	wrepl, err := getReplController(rcc, clustername, workerType)
	if err != nil {
		return result, generalErr(err, replMsgWorker, ClientOperationCode)
	} else if wrepl == nil {
		return result, generalErr(err, replMsgWorker, ClusterIncompleteCode)
	}
	// TODO (tmckay) we should add the spark master and worker configuration values here.
	// the most likely thing to do is store them in an annotation

	result.Name = name
	result.Namespace = namespace
	result.Href = "/clusters/" + clustername
	result.WorkerCount, _, _ = countWorkers(pc, clustername)
	result.MasterCount = 1
	result.Config.WorkerCount = int(wrepl.Spec.Replicas)
	result.Config.MasterCount = int(mrepl.Spec.Replicas)
	result.MasterURL = retrieveServiceURL(sc, masterType, clustername)
	result.MasterWebURL = retrieveServiceURL(sc, webuiType, clustername)
	if result.MasterURL == "" {
		result.Status = "MasterServiceMissing"
	} else {
		result.Status = "Running"
	}

	// Report pos
	result.Pods = []SparkPod{}
	selectorlist := makeSelector(masterType, clustername)
	pods, err := pc.List(selectorlist)
	if err != nil {
		return result, generalErr(err, podListMsg, ClientOperationCode)
	}
	for i := range pods.Items {
		result.Pods = append(result.Pods, addpod(pods.Items[i]))
	}

	_, workers, err := countWorkers(pc, clustername)
	if err != nil {
		return result, generalErr(err, podListMsg, ClientOperationCode)
	}
	for i := range workers.Items {
		result.Pods = append(result.Pods, addpod(workers.Items[i]))
	}

	return result, nil
}

// FindClusters find a cluster and return its representation
func FindClusters(namespace string, client *kclient.Client) ([]SparkCluster, error) {

	var result []SparkCluster = []SparkCluster{}

	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	// Create a map so that we can track clusters by name while we
	// find out information about them
	clist := map[string]*SparkCluster{}

	// Get all of the master pods
	pods, err := pc.List(makeSelector(masterType, ""))
	if err != nil {
		return result, generalErr(err, mastermsg, ClientOperationCode)
	}

	// TODO should we do something else to find the clusters, like count deployment configs?

	// From the list of master pods, figure out which clusters we have
	for i := range pods.Items {

		// Build the cluster record if we don't already have it
		// (theoretically with HA we might have more than 1 master)
		clustername := pods.Items[i].Labels[clusterLabel]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(SparkCluster)
			citem = clist[clustername]
			citem.Name = clustername
			citem.Href = "/clusters/" + clustername

			// Note, we do not report an error here since we are
			// reporting on multiple clusters. Instead cnt will be -1.
			cnt, _, _ := countWorkers(pc, clustername)

			// TODO we only want to count running pods (not terminating)
			citem.WorkerCount = cnt
			citem.MasterURL = retrieveServiceURL(sc, masterType, clustername)
			citem.MasterWebURL = retrieveServiceURL(sc, webuiType, clustername)

			// TODO make something real for status
			if citem.MasterURL == "" {
				citem.Status = "MasterServiceMissing"
			} else {
				citem.Status = "Running"
			}
			result = append(result, *citem)
		}
	}
	return result, nil
}

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
	return &newestRepl, nil
}

func getDepConfig(client oclient.DeploymentConfigInterface, clustername, otype string) (*deployapi.DeploymentConfig, error) {
	selectorlist := makeSelector(otype, clustername)
	deps, err := client.List(selectorlist)
	if err != nil || len(deps.Items) == 0 {
		return nil, err
	}
	// Use the latest replication controller.  There could be more than one
	// if the user did something like oc env to set a new env var on a deployment
	newestDep := deps.Items[0]
	for i := 0; i < len(deps.Items); i++ {
		if deps.Items[i].CreationTimestamp.Unix() > newestDep.CreationTimestamp.Unix() {
			newestDep = deps.Items[i]
		}
	}
	return &newestDep, nil
}

// UpdateSingleClusterResponse update a cluster and return the new representation
func UpdateCluster(name, namespace string, config *ClusterConfig, osclient *oclient.Client, client *kclient.Client) (SparkCluster, error) {

	var result SparkCluster = SparkCluster{}
	clustername := name

	// Before we do further checks, make sure that we have deploymentconfigs
	// If either the master or the worker deploymentconfig are missing, we
	// assume that the cluster is missing. These are the base objects that
	// we use to create a cluster
	ok, err := checkForDeploymentConfigs(osclient.DeploymentConfigs(namespace), clustername)
	if err != nil {
		return result, generalErr(err, findDepConfigMsg, ClientOperationCode)
	}
	if !ok {
		return result, generalErr(nil, fmt.Sprintf(noSuchClusterMsg, clustername), NoSuchClusterCode)
	}

	// Copy any named config referenced and update it with any explicit config values
	finalconfig, err := GetClusterConfig(config, client.ConfigMaps(namespace))
	if err != nil {
		return result, generalErr(err, clusterConfigMsg, ErrorCode(err))
	}
	workercount := int(finalconfig.WorkerCount)

	// TODO(tmckay) we need some way to track the current spark config for a cluster,
	// maybe in annotations. If someone tries to change the spark config for a cluster,
	// that should be an error at this point (unless we spin all the pods down and
	// redeploy)

	dcc := osclient.DeploymentConfigs(namespace)
	dep, err := getDepConfig(dcc, clustername, workerType)
	if err != nil {
		return result, generalErr(err, replMsgWorker, ClientOperationCode)
	} else if dep == nil {
		return result, generalErr(err, replMsgWorker, ClusterIncompleteCode)
	}

	// If the current replica count does not match the request, update the replication controller
	if int(dep.Spec.Replicas) != workercount {
		dep.Spec.Replicas = int32(workercount)
		_, err = dcc.Update(dep)
		if err != nil {
			return result, generalErr(err, updateReplMsg, ClientOperationCode)
		}
	}

	result.Name = name
	result.Namespace = namespace
	result.Config = finalconfig
	return result, nil
}
