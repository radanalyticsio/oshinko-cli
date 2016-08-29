package handlers

import (
	"os"
	"strconv"
	"strings"
	"time"

	middleware "github.com/go-openapi/runtime/middleware"
	oclient "github.com/openshift/origin/pkg/client"
	osa "github.com/redhatanalytics/oshinko-rest/helpers/authentication"
	ocon "github.com/redhatanalytics/oshinko-rest/helpers/containers"
	odc "github.com/redhatanalytics/oshinko-rest/helpers/deploymentconfigs"
	oe "github.com/redhatanalytics/oshinko-rest/helpers/errors"
	"github.com/redhatanalytics/oshinko-rest/helpers/info"
	opt "github.com/redhatanalytics/oshinko-rest/helpers/podtemplates"
	"github.com/redhatanalytics/oshinko-rest/helpers/probes"
	osv "github.com/redhatanalytics/oshinko-rest/helpers/services"
	"github.com/redhatanalytics/oshinko-rest/models"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations/clusters"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)

const nameSpaceMsg = "Cannot determine target openshift namespace"
const clientMsg = "Unable to create an openshift client"

const typeLabel = "oshinko-type"
const clusterLabel = "oshinko-cluster"

const workerType = "worker"
const masterType = "master"
const webuiType = "webui"

const masterPortName = "spark-master"
const masterPort = 7077
const webPortName = "spark-webui"
const webPort = 8080

// The suffix to add to the spark master hostname (clustername) for the web service
const webServiceSuffix = "-ui"

func generalErr(err error, title, msg string, code int32) *models.ErrorResponse {
	if err != nil {
		msg = msg + ", err: " + err.Error()
	}
	return oe.NewSingleErrorResponse(code, title, msg)
}

func responseFailure(err error, msg string, code int32) *models.ErrorResponse {
	return generalErr(err, "Cannot build response", msg, code)
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

func retrieveServiceURL(client kclient.ServiceInterface, stype, clustername string) string {
	selectorlist := makeSelector(stype, clustername)
	srvs, err := client.List(selectorlist)
	if err == nil && len(srvs.Items) != 0 {
		srv := srvs.Items[0]
		scheme := "http://"
		if stype == masterType {
			scheme = "spark://"
		}
		return scheme + srv.Name + ":" + strconv.Itoa(srv.Spec.Ports[0].Port)
	}
	return ""
}

func checkForDeploymentConfigs(client oclient.DeploymentConfigInterface, clustername, namespace string) (bool, error) {
	if client == nil {
		osclient, err := osa.GetOpenShiftClient()
		if err != nil {
			return false, err
		}
		client = osclient.DeploymentConfigs(namespace)
	}
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

func tostrptr(val string) *string {
	v := val
	return &v
}

func toint64ptr(val int64) *int64 {
	v := val
	return &v
}

func singleClusterResponse(clustername string,
	pc kclient.PodInterface,
	sc kclient.ServiceInterface) (*models.SingleCluster, error) {

	addpod := func(p kapi.Pod) *models.ClusterModelPodsItems0 {
		pod := new(models.ClusterModelPodsItems0)
		pod.IP = tostrptr(p.Status.PodIP)
		pod.Status = tostrptr(string(p.Status.Phase))
		pod.Type = tostrptr(p.Labels[typeLabel])
		return pod
	}

	// Note, we never expect "nil, nil" returned from the routine
	// We should always return a cluster, or an error

	// Build the response
	cluster := &models.SingleCluster{&models.ClusterModel{}}
	cluster.Cluster.Name = tostrptr(clustername)

	masterurl := retrieveServiceURL(sc, masterType, clustername)
	masterweburl := retrieveServiceURL(sc, webuiType, clustername)
	cluster.Cluster.MasterURL = tostrptr(masterurl)
	cluster.Cluster.MasterWebURL = tostrptr(masterweburl)

	//TODO make something real for status
	if masterurl == "" {
		cluster.Cluster.Status = tostrptr("MasterServiceMissing")

	} else {
		cluster.Cluster.Status = tostrptr("Running")
	}

	cluster.Cluster.Pods = []*models.ClusterModelPodsItems0{}

	// Report the master pod
	selectorlist := makeSelector(masterType, clustername)
	pods, err := pc.List(selectorlist)
	if err != nil {
		return nil, err
	}
	cluster.Cluster.MasterCount = toint64ptr(int64(len(pods.Items)))
	for i := range pods.Items {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(pods.Items[i]))
	}

	// Report the worker pods
	cnt, workers, err := countWorkers(pc, clustername)
	if err != nil {
		return nil, err
	}
	cluster.Cluster.WorkerCount = toint64ptr(cnt)
	for i := range workers.Items {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(workers.Items[i]))
	}

	return cluster, nil
}

func makeEnvVars(clustername string) []kapi.EnvVar {
	envs := []kapi.EnvVar{}

	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_SPARK_CLUSTER", Value: clustername})
	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_REST_HOST", Value: os.Getenv("OSHINKO_REST_SERVICE_HOST")})
	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_REST_PORT", Value: os.Getenv("OSHINKO_REST_SERVICE_PORT")})
	return envs
}

func makeWorkerEnvVars(clustername string) []kapi.EnvVar {
	envs := []kapi.EnvVar{}

	envs = makeEnvVars(clustername)
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
	replicas int, clustername string) *odc.ODeploymentConfig {

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
		SetLivenessProbe(probes.NewHTTPGetProbe(webport)).EnvVars(makeWorkerEnvVars(clustername))

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func sparkMaster(namespace, image, clustername, masterhost string) *odc.ODeploymentConfig {

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
		SetReadinessProbe(httpProbe).EnvVars(makeEnvVars(clustername))

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

// CreateClusterResponse create a cluster and return the representation
func CreateClusterResponse(params clusters.CreateClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.CreateClusterDefault {
		return clusters.NewCreateClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for create failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "Cannot create cluster", msg, code)
	}

	code := func(err error) int32 {
		if strings.Index(err.Error(), "already exists") != -1 {
			return 409
		}
		return 500
	}

	const mDepConfigMsg = "Unable to create master deployment configuration"
	const wDepConfigMsg = "Unable to create worker deployment configuration"
	const masterSrvMsg = "Unable to create spark master service endpoint"
	const imageMsg = "Cannot determine name of spark image"
	const respMsg = "Created cluster but failed to construct a response object"

	clustername := *params.Cluster.Name
	// pre spark 2, the name the master calls itself must match
	// the name the workers use and the service name created
	masterhost := *params.Cluster.Name
	workercount := int(*params.Cluster.WorkerCount)

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	image, err := info.GetSparkImage()
	if image == "" || err != nil {
		return reterr(fail(err, imageMsg, 500))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	osclient, err := osa.GetOpenShiftClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	// Create the master deployment config
	dcc := osclient.DeploymentConfigs(namespace)
	masterdc := sparkMaster(namespace, image, clustername, masterhost)

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
	workerdc := sparkWorker(namespace, image, workercount, clustername)

	// Launch all of the objects
	_, err = dcc.Create(&masterdc.DeploymentConfig)
	if err != nil {
		return reterr(fail(err, mDepConfigMsg, code(err)))
	}
	_, err = dcc.Create(&workerdc.DeploymentConfig)
	if err != nil {
		// Since we created the master deployment config, try to clean up
		deleteCluster(clustername, namespace, osclient, client)
		return reterr(fail(err, wDepConfigMsg, code(err)))
	}

	// If we've gotten this far, then likely the cluster naming is not in conflict so
	// assume at this point that we should use a 500 error code
	sc := client.Services(namespace)
	_, err = sc.Create(&mastersv.Service)
	if err != nil {
		// Since we create the master and workers, try to clean up
		deleteCluster(clustername, namespace, osclient, client)
		return reterr(fail(err, masterSrvMsg, code(err)))
	}

	// Note, if spark webui service fails for some reason we can live without it
	// TODO ties into cluster status, make a note if the service is missing
	sc.Create(&websv.Service)

	cluster, err := singleClusterResponse(clustername, client.Pods(namespace), sc)
	if err != nil {
		return reterr(responseFailure(err, respMsg, 500))
	}
	return clusters.NewCreateClusterCreated().WithLocation("/clusters/"+clustername).WithPayload(cluster)
}

func waitForCount(client kclient.ReplicationControllerInterface, name string, count int) {

	for i := 0; i < 5; i++ {
		r, _ := client.Get(name)
		if r.Status.Replicas == count {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func deleteCluster(clustername, namespace string, osclient *oclient.Client, client *kclient.Client) string {

	info := []string{}
	scalerepls := []string{}

	// Build a selector list for the "oshinko-cluster" label
	selectorlist := makeSelector("", clustername)

	// Delete all of the deployment configs
	dcc := osclient.DeploymentConfigs(namespace)
	deployments, err := dcc.List(selectorlist)
	if err != nil {
		info = append(info, "unable to find deployment configs ("+err.Error()+")")
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
		err = rcc.Delete(name)
		if err != nil {
			info = append(info, "unable to delete replication controller "+name+" ("+err.Error()+")")
		}
	}

	// Delete the services
	sc := client.Services(namespace)
	srvs, err := sc.List(selectorlist)
	if err != nil {
		info = append(info, "unable to find services ("+err.Error()+")")
	}
	for i := range srvs.Items {
		name := srvs.Items[i].Name
		err = sc.Delete(name)
		if err != nil {
			info = append(info, "unable to delete service "+name+" ("+err.Error()+")")
		}
	}
	return strings.Join(info, ", ")
}

// DeleteClusterResponse delete a cluster
func DeleteClusterResponse(params clusters.DeleteSingleClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.DeleteSingleClusterDefault {
		return clusters.NewDeleteSingleClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for delete failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "Cluster deletion failed", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	osclient, err := osa.GetOpenShiftClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	info := deleteCluster(params.Name, namespace, osclient, client)
	if info != "" {
		return reterr(fail(nil, "Deletion may be incomplete: "+info, 500))
	}
	return clusters.NewDeleteSingleClusterNoContent()
}

// FindClustersResponse find a cluster and return its representation
func FindClustersResponse(params clusters.FindClustersParams) middleware.Responder {

	const mastermsg = "Unable to find spark masters"

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.FindClustersDefault {
		return clusters.NewFindClustersDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for list failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "Cannot list clusters", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}
	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	// Create the payload that we're going to write into for the response
	payload := clusters.FindClustersOKBodyBody{}
	payload.Clusters = []*clusters.ClustersItems0{}

	// Create a map so that we can track clusters by name while we
	// find out information about them
	clist := map[string]*clusters.ClustersItems0{}

	// Get all of the master pods
	pods, err := pc.List(makeSelector(masterType, ""))
	if err != nil {
		return reterr(fail(err, mastermsg, 500))
	}

	// TODO should we do something else to find the clusters, like count deployment configs?

	// From the list of master pods, figure out which clusters we have
	for i := range pods.Items {

		// Build the cluster record if we don't already have it
		// (theoretically with HA we might have more than 1 master)
		clustername := pods.Items[i].Labels[clusterLabel]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(clusters.ClustersItems0)
			citem = clist[clustername]
			citem.Name = tostrptr(clustername)
			citem.Href = tostrptr("/clusters/" + clustername)

			// Note, we do not report an error here since we are
			// reporting on multiple clusters. Instead cnt will be -1.
			cnt, _, _ := countWorkers(pc, clustername)

			// TODO we only want to count running pods (not terminating)
			citem.WorkerCount = toint64ptr(cnt)
			citem.MasterURL = tostrptr(retrieveServiceURL(sc, masterType, clustername))
			citem.MasterWebURL = tostrptr(retrieveServiceURL(sc, webuiType, clustername))

			// TODO make something real for status
			if *citem.MasterURL == "" {
				citem.Status = tostrptr("MasterServiceMissing")
			} else {
				citem.Status = tostrptr("Running")
			}
			payload.Clusters = append(payload.Clusters, citem)
		}
	}
	return clusters.NewFindClustersOK().WithPayload(payload)
}

// FindSingleClusterResponse find a cluster and return its representation
func FindSingleClusterResponse(params clusters.FindSingleClusterParams) middleware.Responder {

	clustername := params.Name

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.FindSingleClusterDefault {
		return clusters.NewFindSingleClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for get failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "Cannot get cluster", msg, code)
	}

	const respMsg = "Failed to construct a response object"
	const progMsg = "Programming error, nil cluster returned and no error reported"

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	// Before we do further checks, make sure that we have deploymentconfigs
	// If either the master or the worker deploymentconfig are missing, we
	// assume that the cluster is missing. These are the base objects that
	// we use to create a cluster
	ok, err := checkForDeploymentConfigs(nil, clustername, namespace)
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}
	if !ok {
		return reterr(fail(nil, "No such cluster", 404))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}
	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	cluster, err := singleClusterResponse(clustername, pc, sc)
	if err != nil {
		// In this case, the entire purpose of this call is to create this
		// response object (as opposed to create and update which might fail
		// in the response but have actually done something)
		return reterr(fail(err, respMsg, 500))

	} else if cluster == nil {
		// If we returned a nil cluster object but there was no error returned,
		// that is a programing error. Note it for development.
		return reterr(fail(err, progMsg, 500))
	}

	return clusters.NewFindSingleClusterOK().WithPayload(cluster)
}

// UpdateSingleClusterResponse update a cluster and return the new representation
func UpdateSingleClusterResponse(params clusters.UpdateSingleClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.UpdateSingleClusterDefault {
		return clusters.NewUpdateSingleClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for update failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "Cannot update cluster", msg, code)
	}

	const findReplMsg = "Unable to find cluster components (is cluster name correct?)"
	const updateReplMsg = "Unable to update replication controller for spark workers"
	const clusterNameMsg = "Changing the cluster name is not supported"
	const masterMsg = "Changing the master count is not supported"
	const respMsg = "Updated cluster but failed to construct a response object"

	clustername := params.Name
	workercount := int(*params.Cluster.WorkerCount)
	mastercount := int(*params.Cluster.MasterCount)

	// Simple things first. At this time we do not support cluster name change and
	// we do not suppport scaling the master count (likely need HA setup for that to make sense)
	if clustername != *params.Cluster.Name {
		return reterr(fail(nil, clusterNameMsg, 409))
	}

	if mastercount != 1 {
		return reterr(fail(nil, masterMsg, 409))
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}
	rcc := client.ReplicationControllers(namespace)

	// Get the replication controller for the cluster (there should only be 1)
	// (it's unlikely we would get more than 1 since it is created by the deploymentconfig)
	selectorlist := makeSelector(workerType, clustername)
	repls, err := rcc.List(selectorlist)
	if err != nil || len(repls.Items) == 0 {
		return reterr(fail(err, findReplMsg, 500))
	}
	// Use the latest replication controller.  There could be more than one
	// if the user did something like oc env to set a new env var on a deployment
	newestRepl := repls.Items[0]
	for i := 0; i < len(repls.Items); i++ {
		if repls.Items[i].CreationTimestamp.Unix() > newestRepl.CreationTimestamp.Unix() {
			newestRepl = repls.Items[i]
		}
	}
	repl := newestRepl

	// If the current replica count does not match the request, update the replication controller
	if repl.Spec.Replicas != workercount {
		repl.Spec.Replicas = workercount
		_, err = rcc.Update(&repl)
		if err != nil {
			return reterr(fail(err, updateReplMsg, 500))
		}
	}

	cluster, err := singleClusterResponse(clustername, client.Pods(namespace), client.Services(namespace))
	if err != nil {
		return reterr(responseFailure(err, respMsg, 500))
	}
	return clusters.NewUpdateSingleClusterAccepted().WithPayload(cluster)
}
