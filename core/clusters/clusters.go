package clusters

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	oclient "github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	ocon "github.com/radanalyticsio/oshinko-cli/core/clusters/containers"
	odc "github.com/radanalyticsio/oshinko-cli/core/clusters/deploymentconfigs"
	opt "github.com/radanalyticsio/oshinko-cli/core/clusters/podtemplates"
	"github.com/radanalyticsio/oshinko-cli/core/clusters/probes"
	ort "github.com/radanalyticsio/oshinko-cli/core/clusters/routes"
	osv "github.com/radanalyticsio/oshinko-cli/core/clusters/services"
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
const replMsgWorker = "unable to find replication controller for spark worker"
const replMsgMaster = "unable to find replication controller for spark master"
const depMsg = "unable to find deployment config for spark %s"
const masterSrvMsg = "unable to create spark master service endpoint"
const mastermsg = "unable to find spark masters"
const updateDepMsg = "unable to update deployment config for spark %s"
const noSuchClusterMsg = "no such cluster '%s'"
const noClusterForDriverMsg = "no cluster found for app '%s'"
const ephemeralDelMsg = "cluster not deleted '%s'"
const podListMsg = "unable to retrive pod list"
const sparkImageMsg = "no spark image specified"
const noSuchAppMsg = "did not find app '%s'"

const typeLabel = "oshinko-type"
const clusterLabel = "oshinko-cluster"
const driverLabel = "uses-oshinko-cluster"
const ephemeralLabel = "ephemeral"

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
	MasterWebRoute string `json:"masterWebRoute"`
	Status       string `json:"status"`
	WorkerCount  int    `json:"workerCount"`
	MasterCount  int    `json:"masterCount"`
	Config       ClusterConfig
	Ephemeral    string `json:"ephemeral,omitempty"`
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
	var ot *labels.Requirement
	var cname *labels.Requirement
	ls := labels.NewSelector()
	if otype == "" {
		ot, _ = labels.NewRequirement(typeLabel, selection.Exists, sets.String{})
	} else {
		ot, _ = labels.NewRequirement(typeLabel, selection.Equals, sets.NewString(otype))
	}
	ls = ls.Add(*ot)
	if clustername == "" {
		cname, _ = labels.NewRequirement(clusterLabel, selection.Exists, sets.String{})
	} else {
		cname, _ = labels.NewRequirement(clusterLabel, selection.Equals, sets.NewString(clustername))
	}
	ls = ls.Add(*cname)
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

func retrieveRouteForService(client oclient.RouteInterface, stype, clustername string) string {
	selectorlist := makeSelector(stype, clustername)
	routes, err := client.List(selectorlist)
	if err == nil && len(routes.Items) != 0 {
		route := routes.Items[0]
		return route.Spec.Host
	}
	return ""
}

func checkForDeploymentConfigs(client oclient.DeploymentConfigInterface, clustername string) (bool, *deployapi.DeploymentConfig, error) {
	selectorlist := makeSelector(masterType, clustername)
	dcs, err := client.List(selectorlist)
	if err != nil {
		return false, nil, err
	}
	if len(dcs.Items) == 0 {
		return false, nil, nil
	}
	m := dcs.Items[0]
	selectorlist = makeSelector(workerType, clustername)
	dcs, err = client.List(selectorlist)
	if err != nil {
		return false, &m, err
	}
	if len(dcs.Items) == 0 {
		return false, &m, nil
	}
	return true, &m, nil

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

func sparkWorker(namespace, image string, replicas int, clustername, sparkconfdir, sparkworkerconfig string) *odc.ODeploymentConfig {

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

func mastername(clustername string) string {
	return clustername + "-m"
}

func workername(clustername string) string {
	return clustername + "-w"
}

func sparkMaster(namespace, image string, replicas int, clustername, sparkconfdir, sparkmasterconfig, driverdc string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(mastername(clustername), namespace).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, clustername).
		Label(typeLabel, masterType).
		PodSelector(clusterLabel, clustername).Replicas(replicas)

	if driverdc != "" {
		dc = dc.Label(ephemeralLabel, driverdc)
	}

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

func getDriverDeployment(app, namespace string, client *kclient.Client) (string, string) {

	// When we call from a driver pod, the most likely value we have is a pod name so
	// check that first
	pc := client.Pods(namespace)
	pod, err := pc.Get(app)
	if err == nil && pod != nil {
		return pod.Labels["deployment"], pod.Labels["deploymentconfig"]
	} else if err != nil{
		println(err.Error())
	}

	// Okay, it wasn't a pod, maybe it's a deployment (rc)
	rcc := client.ReplicationControllers(namespace)
	rc, err := rcc.Get(app)
	if err == nil && rc != nil {
		return app, rc.Labels["openshift.io/deployment-config.name"]
	} else if err != nil {
		println(err.Error())
	}

	// Alright, it might be a deploymentconfig. See if we can find
	// an rc that references it.
	// Build a selector list based on type and/or cluster name
	ls := labels.NewSelector()
	dc, _ := labels.NewRequirement("openshift.io/deployment-config.name", selection.Equals, sets.NewString(app))
	ls = ls.Add(*dc)
	rcs, err := rcc.List(kapi.ListOptions{LabelSelector: ls})
	if err == nil && len(rcs.Items) != 0 {
		rc = newestRepl(rcs)
		return rc.Name, app
	}
	return "", ""
}

// Create a cluster and return the representation
func CreateCluster(
	clustername, namespace, sparkimage string,
	config *ClusterConfig, osclient *oclient.Client, client kclient.Interface, app string, ephemeral bool) (SparkCluster, error) {

	var driverrc string
	var driverdc string
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

	mastercount := int(finalconfig.MasterCount)
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

	// If an app value was passed find the deployment and deploymentconfig
	// so we can do the cluster association and potentially mark the cluster
	// as ephemeral
	if app != "" {
		driverrc, driverdc = getDriverDeployment(app, namespace, client)
	}

	// Create the master deployment config
	if ephemeral {
		// If we couldn't find an rc (deployment) of a dc it's
		// an error
		if driverrc == "" || driverdc == "" {
			return result, generalErr(err, fmt.Sprintf(noSuchAppMsg, app), ClientOperationCode)
		}
	} else {
		// If the ephemeral flag is not set, wipe out driverrc. This will cause the sparkMaster to
		// be constructed without the ephemeral label
		driverrc = ""
	}
	// Create the master deployment config
	masterdc := sparkMaster(namespace, sparkimage, mastercount, clustername, masterconfdir, finalconfig.SparkMasterConfig, driverrc)

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

	webuiroute := ort.NewRoute(websv.GetName() + "-route", websv.GetName(), clustername, "webui")

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
		DeleteCluster(clustername, namespace, osclient, client, "", "")
		return result, generalErr(err, fmt.Sprintf(createDepConfigMsg, workerdc.Name), createCode(err))
	}

	sc := client.Services(namespace)
	_, err = sc.Create(&mastersv.Service)
	if err != nil {
		// Since we create the master and workers, try to clean up
		DeleteCluster(clustername, namespace, osclient, client, "", "")
		return result, generalErr(err, masterSrvMsg, createCode(err))
	}

	// Note, if spark webui service fails for some reason we can live without it
	// TODO ties into cluster status, make a note if the service is missing
	sc.Create(&websv.Service)

	// We will expose the Spark master webui unless we are told not to do it
	if config.ExposeWebUI {
		rc := osclient.Routes(namespace)
		_, err = rc.Create(webuiroute)
	}

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

	// Now that the creation actually worked, label the dc if the app value was passed
	if driverdc != "" {
		driver, err := dcc.Get(driverdc)
		if err == nil {
			if driver.Labels == nil {
				driver.Labels = map[string]string{}
			}
			driver.Labels[driverLabel] = clustername
			dcc.Update(driver)
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

func DeleteCluster(clustername, namespace string, osclient *oclient.Client, client kclient.Interface, app, appstatus string) (string, error) {
	var foundSomething bool = false
	//var zero int32 = 0
	info := []string{}
	rcnames := []string{}


	dcc := osclient.DeploymentConfigs(namespace)
	rcc := client.ReplicationControllers(namespace)

	// If we have supplied an appstatus flag, then we only delete the cluster if it is marked as ephemeral
	// If it's not marked as ephemeral then we skip the delete
	if appstatus == "completed" || appstatus == "terminated" {
		var delete bool = false

		// See if the master dc has the ephemeral label
		master, err := dcc.Get(mastername(clustername))
		if err != nil {
			// We can't get the dc for the master to look up whether it's ephemeral.
			// But this means the cluster is partially broken anyway. Let the normal delete
			// fall through and cleanup
			delete = true
		} else if ephemeral, ok := master.Labels[ephemeralLabel]; ok {
			// app may be a pod name, get the dc value
			deployment, driverdc := getDriverDeployment(app, namespace, client)
			if deployment != ephemeral {
				info = append(info, "cluster is not linked to app")
			} else {
				// Either the driver dc has been deleted, or it's been scaled to zero,
				// or the application completed and the driver has not been scaled.
				// In all cases we consider the app complete and delete the cluster
				driver, err := dcc.Get(driverdc)
				delete = err != nil ||
					driver.Spec.Replicas == 0 ||
					(appstatus == "completed" && driver.Spec.Replicas == 1)
				if !delete {
					info = append(info, "driver replica count > 0 (or > 1 for completed app)")
				}
			}

		} else {
			info = append(info, "cluster is not ephemeral")
		}
		if !delete {
			return strings.Join(info, ", "), generalErr(nil, fmt.Sprintf(ephemeralDelMsg, clustername), EphemeralCode)
		}
	}

	// Put a label on the master dc as soon as possible that says "Deleting"
	// just in case this takes some time and someone does a get on the cluster.
	// This may never be seen.
	dc, err := dcc.Get(mastername(clustername))
	if err == nil && dc != nil {
		tmp := odc.ODeploymentConfig{*dc}
		tmp.Label("delete_pending", "true")
	}

	// Build a selector list for the "oshinko-cluster" label
	selectorlist := makeSelector("", clustername)

	// Delete the dcs
	deployments, err := dcc.List(selectorlist)
	for i := range deployments.Items {
		err = dcc.Delete(deployments.Items[i].Name)
		if err != nil {
			info = append(info, "unable to delete deployment config "+deployments.Items[i].Name+" ("+err.Error()+")")
		}
	}

	// Delete the rcs
	rcc = client.ReplicationControllers(namespace)
	repls, err := rcc.List(selectorlist)
	for i := range repls.Items {
		rcnames = append(rcnames, repls.Items[i].Name)
		err = rcc.Delete(repls.Items[i].Name, nil)
		if err != nil {
			info = append(info, "unable to delete replication controller " + repls.Items[i].Name + " (" + err.Error() + ")")
		}
	}

	pc := client.Pods(namespace)
	for i := range rcnames {
		ls := labels.NewSelector()
		plist, _ := labels.NewRequirement("openshift.io/deployer-pod-for.name", selection.Equals, sets.NewString(rcnames[i]))
		ls = ls.Add(*plist)
		pods, err := pc.List(kapi.ListOptions{LabelSelector: ls})
		if err == nil && len(pods.Items) != 0 {
			for p := range pods.Items {
				err = pc.Delete(pods.Items[p].Name, nil)
				if err != nil {
					info = append(info, "unable to delete deployer pod " + pods.Items[p].Name + " ("+err.Error()+")")
				} else {
					info = append(info, "deleted deployer pod " + pods.Items[p].Name)
				}
			}
		}
	}

	pods, err := pc.List(selectorlist)
	for i := range pods.Items {
		err = pc.Delete(pods.Items[i].Name, nil)
		if err != nil {
			info = append(info, "unable to delete replication controller " + repls.Items[i].Name + " (" + err.Error() + ")")
		}
	}

	rc := osclient.Routes(namespace)
	webUIRouteName := clustername + "-ui-route"
	err = rc.Delete(webUIRouteName)
	if err != nil {
		info = append(info, "unable to delete route " + webUIRouteName + " (" + err.Error() + ")")
	}

	// Delete the services
	sc := client.Services(namespace)
	srvs, err := sc.List(selectorlist)
	for i := range srvs.Items {
		err = sc.Delete(srvs.Items[i].Name)
		if err != nil {
			info = append(info, "unable to delete service " + srvs.Items[i].Name + " (" + err.Error() + ")")
		}
	}
	// If we found some part of a cluster, then there is no error
	// even though the cluster may not have been fully complete.
	// If we didn't find any trace of a cluster, then call it an error
	if !foundSomething {
		return strings.Join(info, ", "), generalErr(nil, fmt.Sprintf(noSuchClusterMsg, clustername), NoSuchClusterCode)
	}
	return strings.Join(info, ", "), nil
}

// Find a cluster and return its representation
func FindSingleCluster(name, namespace string, osclient *oclient.Client, client kclient.Interface) (SparkCluster, error) {

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
	ok, master, err := checkForDeploymentConfigs(osclient.DeploymentConfigs(namespace), clustername)
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

	// TODO (tmckay) we need to figure out how to do desired/actual count like the oc
	result.Name = name
	result.Namespace = namespace
	result.Href = "/clusters/" + clustername
	result.WorkerCount = int(wrepl.Status.Replicas)
	result.MasterCount = int(mrepl.Status.Replicas)
	result.Config.WorkerCount = int(wrepl.Spec.Replicas)
	result.Config.MasterCount = int(mrepl.Spec.Replicas)
	result.MasterURL = retrieveServiceURL(sc, masterType, clustername)
	result.MasterWebURL = retrieveServiceURL(sc, webuiType, clustername)
	result.MasterWebRoute = retrieveRouteForService(osclient.Routes(namespace), webuiType, clustername)
	if result.MasterURL == "" {
		result.Status = "MasterServiceMissing"
	} else {
		result.Status = "Running"
	}
	if ephemeral, ok := master.Labels[ephemeralLabel]; ok {
		result.Ephemeral = ephemeral
	} else {
		result.Ephemeral = "shared"
	}

	// Report pods
	result.Pods = []SparkPod{}
	selectorlist := makeSelector(masterType, clustername)
	pods, err := pc.List(selectorlist)
	if err != nil {
		return result, generalErr(err, podListMsg, ClientOperationCode)
	}
	for i := range pods.Items {
		result.Pods = append(result.Pods, addpod(pods.Items[i]))
	}

	selectorlist = makeSelector(workerType, clustername)
	pods, err = pc.List(selectorlist)
	if err != nil {
		return result, generalErr(err, podListMsg, ClientOperationCode)
	}
	for i := range pods.Items {
		result.Pods = append(result.Pods, addpod(pods.Items[i]))
	}

	return result, nil
}

// Find all clusters and return their representation
func FindClusters(namespace string, osclient *oclient.Client, client kclient.Interface, app string) ([]SparkCluster, error) {

	var result []SparkCluster = []SparkCluster{}
	var mcount, wcount int
	dcc := osclient.DeploymentConfigs(namespace)
	sc := client.Services(namespace)
	dc := osclient.DeploymentConfigs(namespace)

	// If app is not null, find the matching dc and look for a driver label.
	// If we find it get the name of the cluster and call FindSingleCluster.
	// If not, the list is empty
	if app != "" {
		_, dcname := getDriverDeployment(app, namespace, client)
		if dcname != "" {
			driver, err := dc.Get(dcname)
			if err == nil && driver != nil {
				if clustername, ok := driver.Labels[driverLabel]; ok {
					c, err := FindSingleCluster(clustername, namespace, osclient, client)
					if err == nil {
						result = append(result, c)
					}
					return result, err
				}
			}
		}
		return result, generalErr(nil, fmt.Sprintf(noClusterForDriverMsg, app), NoSuchClusterCode)
	}

	// Create a map so that we can track clusters by name while we
	// find out information about them
	clist := map[string]*SparkCluster{}

	// Get all of the master dcs
	dcs, err := dcc.List(makeSelector(masterType, ""))
	if err != nil {
		return result, generalErr(err, mastermsg, ClientOperationCode)
	}

	// From the list of master pods, figure out which clusters we have
	for i := range dcs.Items {

		// Build the cluster record if we don't already have it
		// (theoretically with HA we might have more than 1 master)
		clustername := dcs.Items[i].Labels[clusterLabel]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(SparkCluster)
			citem = clist[clustername]
			citem.Name = clustername
			citem.Href = "/clusters/" + clustername

			mcount = int(dcs.Items[i].Status.Replicas)
			wdc, err := getDepConfig(dcc, clustername, workerType)
			if err == nil && wdc != nil {
				wcount = int(wdc.Status.Replicas)
			} else {
				wcount = -1
			}

			// TODO we only want to count running pods (not terminating)
			citem.MasterCount = mcount
			citem.WorkerCount = wcount
			citem.MasterURL = retrieveServiceURL(sc, masterType, clustername)
			citem.MasterWebURL = retrieveServiceURL(sc, webuiType, clustername)
			citem.MasterWebRoute = retrieveRouteForService(osclient.Routes(namespace), webuiType, clustername)

			// TODO make something real for status
			if citem.MasterURL == "" {
				citem.Status = "MasterServiceMissing"
			} else {
				citem.Status = "Running"
			}

			master, err := dc.Get(mastername(clustername))
			if err == nil {
				if ephemeral, ok := master.Labels[ephemeralLabel]; ok {
					citem.Ephemeral = ephemeral
				} else {
					citem.Ephemeral = "shared"
				}
			}
			result = append(result, *citem)
		}
	}
	return result, nil
}

func newestRepl(list *kapi.ReplicationControllerList ) *kapi.ReplicationController {
	newestRepl := list.Items[0]
	for i := 0; i < len(list.Items); i++ {
		if list.Items[i].CreationTimestamp.Unix() > newestRepl.CreationTimestamp.Unix() {
			newestRepl = list.Items[i]
		}
	}
	return &newestRepl
}

func getReplController(client kclient.ReplicationControllerInterface, clustername, otype string) (*kapi.ReplicationController, error) {

	selectorlist := makeSelector(otype, clustername)
	repls, err := client.List(selectorlist)
	if err != nil || len(repls.Items) == 0 {
		return nil, err
	}
	// Use the latest replication controller.  There could be more than one
	// if the user did something like oc env to set a new env var on a deployment
	return newestRepl(repls), nil
}

func getDepConfig(client oclient.DeploymentConfigInterface, clustername, otype string) (*deployapi.DeploymentConfig, error) {
	var dep *deployapi.DeploymentConfig
	var err error
	if otype == masterType {
		dep, err = client.Get(clustername+"-m")
	} else {
		dep, err = client.Get(clustername+"-w")
	}
	return dep, err
}

func scaleDep(dcc oclient.DeploymentConfigInterface, clustername string, count int, otype string) error {
	if count <= SentinelCountValue {
		return nil
	}
	dep, err := getDepConfig(dcc, clustername, otype)
	if err != nil {
		return generalErr(err, fmt.Sprintf(depMsg, otype), ClientOperationCode)
	} else if dep == nil {
		return generalErr(err, fmt.Sprintf(depMsg, otype), ClusterIncompleteCode)
	}

	// If the current replica count does not match the request, update the replication controller
	if int(dep.Spec.Replicas) != count {
		dep.Spec.Replicas = int32(count)
		_, err = dcc.Update(dep)
		if err != nil {
			return generalErr(err, fmt.Sprintf(updateDepMsg, otype), ClientOperationCode)
		}
	}
	return nil
}

// Update a cluster and return the new representation
// This routine supports the same stored config semantics as used in cluster creation
// but at this point only allows updating the master and worker counts.
func UpdateCluster(name, namespace string, config *ClusterConfig, osclient *oclient.Client, client kclient.Interface) (SparkCluster, error) {

	var result SparkCluster = SparkCluster{}
	clustername := name

	// Before we do further checks, make sure that we have deploymentconfigs
	// If either the master or the worker deploymentconfig are missing, we
	// assume that the cluster is missing. These are the base objects that
	// we use to create a cluster
	ok, _, err := checkForDeploymentConfigs(osclient.DeploymentConfigs(namespace), clustername)
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
	mastercount := int(finalconfig.MasterCount)

	// TODO(tmckay) we need some way to track the current spark config for a cluster,
	// maybe in annotations. If someone tries to change the spark config for a cluster,
	// that should be an error at this point (unless we spin all the pods down and
	// redeploy)

	dcc := osclient.DeploymentConfigs(namespace)
	err = scaleDep(dcc, clustername, workercount, workerType)
	if err != nil {
		return result, err
	}
	err = scaleDep(dcc, clustername, mastercount, masterType)
	if err != nil {
		return result, err
	}

	result.Name = name
	result.Namespace = namespace
	result.Config = finalconfig
	return result, nil
}


// Scale a cluster
// This routine supports a specific scale operation based on immediate values for
// master and worker counts and does not consider stored configs.
func ScaleCluster(name, namespace string, masters, workers int, osclient *oclient.Client, client kclient.Interface) error {

	clustername := name

	// Before we do further checks, make sure that we have deploymentconfigs
	// If either the master or the worker deploymentconfig are missing, we
	// assume that the cluster is missing. These are the base objects that
	// we use to create a cluster
	ok, _, err := checkForDeploymentConfigs(osclient.DeploymentConfigs(namespace), clustername)
	if err != nil {
		return generalErr(err, findDepConfigMsg, ClientOperationCode)
	}
	if !ok {
		return generalErr(nil, fmt.Sprintf(noSuchClusterMsg, clustername), NoSuchClusterCode)
	}

	// Allow sale to zero, allow sentinel values
	if masters > 1 {
		return NewClusterError(MasterCountMustBeZeroOrOne, ClusterConfigCode)
	}

	dcc := osclient.DeploymentConfigs(namespace)
	err = scaleDep(dcc, clustername, workers, workerType)
	if err != nil {
		return err
	}
	return scaleDep(dcc, clustername, masters, masterType)
}
