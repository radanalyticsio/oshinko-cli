package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	middleware "github.com/go-openapi/runtime/middleware"
	oclient "github.com/openshift/origin/pkg/client"
	osa "github.com/redhatanalytics/oshinko-rest/helpers/authentication"
	ocon "github.com/redhatanalytics/oshinko-rest/helpers/containers"
	odc "github.com/redhatanalytics/oshinko-rest/helpers/deploymentconfigs"
	"github.com/redhatanalytics/oshinko-rest/helpers/info"
	opt "github.com/redhatanalytics/oshinko-rest/helpers/podtemplates"
	osv "github.com/redhatanalytics/oshinko-rest/helpers/services"
	"github.com/redhatanalytics/oshinko-rest/models"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations/clusters"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)

const namespacemsg = "Cannot determine target openshift namespace"
const clientmsg = "Unable to create an openshift client"

const type_label = "oshinko-type"
const cluster_label = "oshinko-cluster"

const worker_type = "worker"
const master_type = "master"
const webui_type = "webui"

const master_port_name = "spark-master"
const web_port_name = "spark-webui"

func generalerr(err error, title, msg string, code int32) *models.ErrorResponse {
	if err != nil {
		msg = msg + ", err: " + err.Error()
	}
	return makeSingleErrorResponse(code, title, msg)
}

func responsefailure(err error, msg string, code int32) *models.ErrorResponse {
	return generalerr(err, "Cannot build response", msg, code)
}

func sparkMasterURL(name string, port *kapi.ServicePort) string {
	return "spark://" + name + ":" + strconv.Itoa(port.Port)
}

func makeselector(otype string, clustername string) kapi.ListOptions {
	// Build a selector list based on type and/or cluster name
	ls := labels.NewSelector()
	if otype != "" {
		ot, _ := labels.NewRequirement(type_label, labels.EqualsOperator, sets.NewString(otype))
		ls = ls.Add(*ot)
	}
	if clustername != "" {
		cname, _ := labels.NewRequirement(cluster_label, labels.EqualsOperator, sets.NewString(clustername))
		ls = ls.Add(*cname)
	}
	return kapi.ListOptions{LabelSelector: ls}
}

func countworkers(client kclient.PodInterface, clustername string) (int64, *kapi.PodList, error) {
	// If we are  unable to retrieve a list of worker pods, return -1 for count
	// This is an error case, differnt from a list of length 0. Let the caller
	// decide whether to report the error or the -1 count
	cnt := int64(-1)
	selectorlist := makeselector(worker_type, clustername)
	pods, err := client.List(selectorlist)
	if pods != nil {
		cnt = int64(len(pods.Items))
	}
	return cnt, pods, err
}

func retrievemasterurl(client kclient.ServiceInterface, clustername string) string {
	selectorlist := makeselector(master_type, clustername)
	srvs, err := client.List(selectorlist)
	if err == nil && len(srvs.Items) != 0 {
		srv := srvs.Items[0]
		return sparkMasterURL(srv.Name, &srv.Spec.Ports[0])
	}
	return ""
}

func tostrptr(val string) *string {
	v := val
	return &v
}

func toint64ptr(val int64) *int64 {
	v := val
	return &v
}

func singleclusterresponse(clustername string,
	pc kclient.PodInterface,
	sc kclient.ServiceInterface, masterurl string) (*models.SingleCluster, error) {

	addpod := func(p kapi.Pod) *models.ClusterModelPodsItems0 {
		pod := new(models.ClusterModelPodsItems0)
		pod.IP = tostrptr(p.Status.PodIP)
		pod.Status = tostrptr(string(p.Status.Phase))
		pod.Type = tostrptr(p.Labels[type_label])
		return pod
	}

	// Note, we never expect "nil, nil" returned from the routine

	// Build the response
	cluster := &models.SingleCluster{&models.ClusterModel{}}
	cluster.Cluster.Name = tostrptr(clustername)

	// If we passed in a master url, just use it
	if masterurl == "" {
		// If the developer passed in a nil sc, error!
		if sc == nil {
			return nil, errors.New(
				"Programming error," +
					"building cluster response but masterurl and service client are empty ")
		}
		masterurl = retrievemasterurl(sc, clustername)
	}
	cluster.Cluster.MasterURL = tostrptr(masterurl)

	//TODO make something real for status
	cluster.Cluster.Status = tostrptr("Running")

	cluster.Cluster.Pods = []*models.ClusterModelPodsItems0{}

	// Report the master pod
	selectorlist := makeselector(master_type, clustername)
	pods, err := pc.List(selectorlist)
	if err != nil {
		return nil, err
	}
	cluster.Cluster.MasterCount = toint64ptr(int64(len(pods.Items)))
	for i := range pods.Items {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(pods.Items[i]))
	}

	// Report the worker pods
	cnt, workers, err := countworkers(pc, clustername)
	if err != nil {
		return nil, err
	}
	cluster.Cluster.WorkerCount = toint64ptr(cnt)
	for i := range workers.Items {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(workers.Items[i]))
	}

	return cluster, nil
}

func sparkWorker(namespace string,
	image string,
	replicas int, masterurl, clustername string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name.
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(clustername+"-w", namespace).
		TriggerOnConfigChange().RollingStrategy().Label(cluster_label, clustername).
		Label(type_label, worker_type).
		PodSelector(cluster_label, clustername).Replicas(replicas)

	// Create a pod template spec with the matching label
	pt := opt.PodTemplateSpec().Label(cluster_label, clustername).Label(type_label, worker_type)

	// Create a container with the correct start command
	cont := ocon.Container(
		dc.Name,
		image).Command("/start-worker", masterurl)

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
		TriggerOnConfigChange().RollingStrategy().Label(cluster_label, clustername).
		Label(type_label, master_type).
		PodSelector(cluster_label, clustername)

	// Create a pod template spec with the matching label
	pt := opt.PodTemplateSpec().Label(cluster_label, clustername).
		Label(type_label, master_type)

	// Create a container with the correct ports and start command
	masterp := ocon.ContainerPort(master_port_name, 7077)
	webp := ocon.ContainerPort(web_port_name, 8080)
	cont := ocon.Container(
		dc.Name,
		image).Command("/start-master", masterhost).Ports(masterp, webp)

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func Service(name string,
	port int,
	clustername, otype string,
	podselectors map[string]string) (*osv.OService, *osv.OServicePort) {

	p := osv.ServicePort(port).TargetPort(port)
	return osv.Service(name).Label(cluster_label, clustername).
		Label(type_label, otype).PodSelectors(podselectors).Ports(p), p
}

// CreateClusterResponse create a cluster and return the representation
func CreateClusterResponse(params clusters.CreateClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.CreateClusterDefault {
		return clusters.NewCreateClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for create failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalerr(err, "Cannot create cluster", msg, code)
	}

	const mdepconfigmsg = "Unable to create naster deployment configuration"
	const wdepconfigmsg = "Unable to create worker deployment configuration"
	const mastersrvmsg = "Unable to create spark master service endpoint"
	const imagemsg = "Cannot determine name of spark image"
	const respmsg = "Created cluster but failed to construct a response object"

	clustername := *params.Cluster.Name
	// pre spark 2, the name the master calls itself must match
	// the name the workers use and the service name created
	masterhost := *params.Cluster.Name
	workercount := int(*params.Cluster.WorkerCount)

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, namespacemsg, 409))
	}

	image, err := info.GetSparkImage()
	if image == "" || err != nil {
		return reterr(fail(err, imagemsg, 409))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
	}

	osclient, err := osa.GetOpenShiftClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
	}

	// Create the master deployment config
	dcc := osclient.DeploymentConfigs(namespace)
	masterdc := sparkMaster(namespace, image, clustername, masterhost)

	// Create the services that will be associated with the master pod
	// They will be created with selectors based on the pod labels
	mastersv, masterp := Service(masterhost,
		masterdc.FindPort(master_port_name),
		clustername, master_type,
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := Service(masterhost+"-ui",
		masterdc.FindPort(web_port_name),
		clustername, webui_type,
		masterdc.GetPodTemplateSpecLabels())

	// Create the worker deployment config
	masterurl := sparkMasterURL(masterhost, &masterp.ServicePort)
	workerdc := sparkWorker(namespace, image, workercount, masterurl, clustername)

	// Launch all of the objects
	_, err = dcc.Create(&masterdc.DeploymentConfig)
	if err != nil {
		return reterr(fail(err, mdepconfigmsg, 409))
	}
	_, err = dcc.Create(&workerdc.DeploymentConfig)
	if err != nil {
		// Since we created the master deployment config, try to clean up
		deletecluster(clustername, namespace, osclient, client)
		return reterr(fail(err, wdepconfigmsg, 409))
	}

	// If we've gotten this far, then likely the cluster naming is not in conflict so
	// assume at this point that we should use a 500 error code
	sc := client.Services(namespace)
	_, err = sc.Create(&mastersv.Service)
	if err != nil {
		// Since we create the master and workers, try to clean up
		deletecluster(clustername, namespace, osclient, client)
		return reterr(fail(err, mastersrvmsg, 500))
	}

	// Note, if spark webui service fails for some reason we can live without it
	// TODO ties into cluster status, make a note if the service is missing
	sc.Create(&websv.Service)

	// Since we already know what the masterurl is, pass it in explicitly and do not pass a service client
	cluster, err := singleclusterresponse(clustername, client.Pods(namespace), nil, masterurl)
	if err != nil {
		return reterr(responsefailure(err, respmsg, 500))
	}
	return clusters.NewCreateClusterCreated().WithLocation(masterurl).WithPayload(cluster)
}

func WaitForCount(client kclient.ReplicationControllerInterface, name string, count int) {

	for i := 0; i < 5; i++ {
		r, _ := client.Get(name)
		if r.Status.Replicas == count {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func deletecluster(clustername, namespace string, osclient *oclient.Client, client *kclient.Client) string {

	info := []string{}
	scalerepls := []string{}

	// Build a selector list for the "oshinko-cluster" label
	selectorlist := makeselector("", clustername)

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
		WaitForCount(rcc, scalerepls[i], 0)
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
		return generalerr(err, "Cluster deletion failed", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, namespacemsg, 409))
	}

	osclient, err := osa.GetOpenShiftClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
	}

	info := deletecluster(params.Name, namespace, osclient, client)
	if info != "" {
		return reterr(fail(nil, "Deletion may be incomplete: "+info, 500))
	}
	return clusters.NewDeleteSingleClusterNoContent()
}

// FindClustersResponse find a cluster and return its representation
func FindClustersResponse() middleware.Responder {

	const mastermsg = "Unable to find spark masters"

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *clusters.FindClustersDefault {
		return clusters.NewFindClustersDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for list failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalerr(err, "Cannot list clusters", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, namespacemsg, 409))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
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
	pods, err := pc.List(makeselector(master_type, ""))
	if err != nil {
		return reterr(fail(err, mastermsg, 500))
	}

	// TODO should we do something else to find the clusters, like count deployment configs?

	// From the list of master pods, figure out which clusters we have
	for i := range pods.Items {

		// Build the cluster record if we don't already have it
		// (theoretically with HA we might have more than 1 master)
		clustername := pods.Items[i].Labels[cluster_label]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(clusters.ClustersItems0)
			citem = clist[clustername]
			citem.Name = tostrptr(clustername)
			citem.Href = tostrptr("/clusters/" + clustername)

			// Note, we do not report an error here since we are
			// reporting on multiple clusters. Instead cnt will be -1.
			cnt, _, _ := countworkers(pc, clustername)

			// TODO we only want to count running pods (not terminating)
			citem.WorkerCount = toint64ptr(cnt)
			// TODO make something real for status
			citem.Status = tostrptr("Running")
			citem.MasterURL = tostrptr(retrievemasterurl(sc, clustername))
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
		return generalerr(err, "Cannot get cluster", msg, code)
	}

	const respmsg = "Failed to construct a response object"
	const progmsg = "Programming error, nil cluster returned and no error reported"

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, namespacemsg, 409))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
	}
	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	cluster, err := singleclusterresponse(clustername, pc, sc, "")
	if err != nil {
		// In this case, the entire purpose of this call is to create this
		// response object (as opposed to create and update which might fail
		// in the response but have actually done something)
		return reterr(fail(err, respmsg, 500))

	} else if cluster == nil {
		// If we returned a nil cluster object but there was no error returned,
		// that is a programing error. Note it for development.
		return reterr(fail(err, progmsg, 500))
	}

	// If there are no pods and no master url, there may not be a cluster at all.
	// Check for the existence of replication controllers as a final check
	// If we don't find any, just return an empty response
	if len(cluster.Cluster.Pods) == 0 && *cluster.Cluster.MasterURL == "" {
		rcc := client.ReplicationControllers(namespace)
		// make a selector for label "oshinko-cluster"
		selectorlist := makeselector("", clustername)
		repls, err := rcc.List(selectorlist)
		if err != nil || len(repls.Items) == 0 {
			return clusters.NewFindSingleClusterOK()
		}
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
		return generalerr(err, "Cannot update cluster", msg, code)
	}

	const findreplmsg = "Unable to find cluster components (is cluster name correct?)"
	const updatereplmsg = "Unable to update replication controller for spark workers"
	const clusternamemsg = "Changing the cluster name is not supported"
	const mastermsg = "Changing the master count is not supported"
	const respmsg = "Updated cluster but failed to construct a response object"

	clustername := params.Name
	workercount := int(*params.Cluster.WorkerCount)
	mastercount := int(*params.Cluster.MasterCount)

	// Simple things first. At this time we do not support cluster name change and
	// we do not suppport scaling the master count (likely need HA setup for that to make sense)
	if clustername != *params.Cluster.Name {
		return reterr(fail(nil, clusternamemsg, 409))
	}

	if mastercount != 1 {
		return reterr(fail(nil, mastermsg, 409))
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, namespacemsg, 409))
	}

	client, err := osa.GetKubeClient()
	if err != nil {
		return reterr(fail(err, clientmsg, 400))
	}
	rcc := client.ReplicationControllers(namespace)

	// Get the replication controller for the cluster (there should only be 1)
	// (it's unlikely we would get more than 1 since it is created by the deploymentconfig)
	selectorlist := makeselector(worker_type, clustername)
	repls, err := rcc.List(selectorlist)
	if err != nil || len(repls.Items) == 0 {
		return reterr(fail(err, findreplmsg, 400))
	}
	repl := repls.Items[0]

	// If the current replica count does not match the request, update the replication controller
	if repl.Spec.Replicas != workercount {
		repl.Spec.Replicas = workercount
		_, err = rcc.Update(&repl)
		if err != nil {
			return reterr(fail(err, updatereplmsg, 500))
		}
	}

	cluster, err := singleclusterresponse(clustername, client.Pods(namespace), client.Services(namespace), "")
	if err != nil {
		return reterr(responsefailure(err, respmsg, 500))
	}
	return clusters.NewUpdateSingleClusterAccepted().WithPayload(cluster)
}
