package handlers

import (
	"strconv"
	"time"

	middleware "github.com/go-openapi/runtime/middleware"

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

func missingnamespace(err error) *models.ErrorResponse {
	msg := "Current namespace must be known for this operation"
	if err != nil {
		msg = msg + ", err: " + err.Error()
	}
	return makeSingleErrorResponse(400, "Cannot determine namespace", msg)
}

func missingimage(err error) *models.ErrorResponse {
	msg := "Spark image must be known to create cluster"
	if err != nil {
		msg = msg + ", err: " + err.Error()
	}
	return makeSingleErrorResponse(400, "Cannot determine image", msg)
}

func missingclient(err error) *models.ErrorResponse {
	msg := "Unable to create an openshift client"
	if err != nil {
		msg = msg + ", err: " + err.Error()
	}
	return makeSingleErrorResponse(400, "Cannot create client", msg)
}

func sparkMasterURL(name string, port *kapi.ServicePort) string {
	return "spark://" + name + ":" + strconv.Itoa(port.Port)
}

func makeselector(otype string, clustername string) kapi.ListOptions {
	// Build a selector list based on type and/or cluster name
	ls := labels.NewSelector()
	if otype != "" {
		ot, _ := labels.NewRequirement("oshinko-type", labels.EqualsOperator, sets.NewString(otype))
		ls = ls.Add(*ot)
	}
	if clustername != "" {
		cname, _ := labels.NewRequirement("oshinko-cluster", labels.EqualsOperator, sets.NewString(clustername))
		ls = ls.Add(*cname)
	}
	return kapi.ListOptions{LabelSelector: ls}
}

func getworkers(client kclient.PodInterface, clustername string) *kapi.PodList {
	selectorlist := makeselector("worker", clustername)
	pods, _ := client.List(selectorlist)
	return pods
}

func countworkers(client kclient.PodInterface, clustername string) (int64, *kapi.PodList) {
	pods := getworkers(client, clustername)
	if pods != nil {
		return int64(len(pods.Items)), pods
	}
	return 0, nil
}

func retrievemasterurl(client kclient.ServiceInterface, clustername string) string {
	selectorlist := makeselector("master", clustername)
	srvs, _ := client.List(selectorlist)
	srv := srvs.Items[0]
	return sparkMasterURL(srv.Name, &srv.Spec.Ports[0])
}

func tostrptr(val string) *string {
	v := val
	return &v
}

func toint64ptr(val int64) *int64 {
	v := val
	return &v
}

func sparkWorker(namespace string,
	image string,
	replicas int, masterurl, clustername string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name.
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(clustername+"-w", namespace).
		TriggerOnConfigChange().RollingStrategy().Label("oshinko-cluster", clustername).
		Label("oshinko-type", "worker").
		PodSelector("oshinko-cluster", clustername).Replicas(replicas)

	// Create a pod template spec with the matching label
	pt := opt.PodTemplateSpec().Label("oshinko-cluster", clustername).Label("oshinko-type", "worker")

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
		TriggerOnConfigChange().RollingStrategy().Label("oshinko-cluster", clustername).
		Label("oshinko-type", "master").
		PodSelector("oshinko-cluster", clustername)

	// Create a pod template spec with the matching label
	// Additionally, we need another label specifically for the master
	// so that the services can find the correct pod
	pt := opt.PodTemplateSpec().Label("oshinko-cluster", clustername).
		Label("oshinko-type", "master")

	// Create a container with the correct ports and start command
	masterp := ocon.ContainerPort("spark-master", 7077)
	webp := ocon.ContainerPort("spark-webui", 8080)
	cont := ocon.Container(
		dc.Name,
		image).Command("/start-master", masterhost).Ports(masterp, webp)

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func Service(name string,
	port int,
	clustername string,
	otype string, podselectors map[string]string) (*osv.OService, *osv.OServicePort) {
	p := osv.ServicePort(port).TargetPort(port)
	return osv.Service(name).Label("oshinko-cluster", clustername).
		Label("oshinko-type", otype).PodSelectors(podselectors).Ports(p), p
}

// CreateClusterResponse create a cluster and return the representation
func CreateClusterResponse(params clusters.CreateClusterParams) middleware.Responder {
	// create a cluster here
	// INFO(elmiko) my thinking on creating clusters is that we should use a
	// label on the items we create with kubernetes so that we can get them
	// all with a request.
	// in addition to labels for general identification, we should then use
	// annotations on objects to help further refine what we are dealing with.

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingnamespace(err))
	}

	image, err := info.GetSparkImage()
	if image == "" || err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingimage(err))
	}

	// kube rest client
	client, err := osa.GetKubeClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}

	// openshift rest client
	osclient, err := osa.GetOpenShiftClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}

	// deployment config client
	dcc := osclient.DeploymentConfigs(namespace)

	// Make master deployment config
	// Ignoring master-count for now, leave it defaulted at 1
	clustername := *params.Cluster.Name

	// pre spark 2, the name the master calls itself must match
	// the name the workers use and the service name created
	masterhost := *params.Cluster.Name
	masterdc := sparkMaster(namespace, image, clustername, masterhost)

	// Make master services
	mastersv, masterp := Service(masterhost,
		masterdc.FindPort("spark-master"),
		clustername, "master",
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := Service(masterhost+"-ui",
		masterdc.FindPort("spark-webui"),
		clustername, "webui",
		masterdc.GetPodTemplateSpecLabels())

	// Make worker deployment config
	masterurl := sparkMasterURL(masterhost, &masterp.ServicePort)
	workerdc := sparkWorker(namespace, image, int(*params.Cluster.WorkerCount), masterurl, *params.Cluster.Name)

	// Launch all of the objects
	_, err = dcc.Create(&masterdc.DeploymentConfig)
	if err != nil {
		payload := makeSingleErrorResponse(409, "Creation failed", err.Error())
		return clusters.NewCreateClusterDefault(409).WithPayload(payload)
	}
	_, err = dcc.Create(&workerdc.DeploymentConfig)
	if err != nil {
		payload := makeSingleErrorResponse(409, "Creation failed", err.Error())
		return clusters.NewCreateClusterDefault(409).WithPayload(payload)
	}

	_, err = client.Services(namespace).Create(&mastersv.Service)
	if err != nil {
		payload := makeSingleErrorResponse(500, "Service creation failed", err.Error())
		return clusters.NewCreateClusterDefault(500).WithPayload(payload)
	}
	_, err = client.Services(namespace).Create(&websv.Service)
	if err != nil {
		payload := makeSingleErrorResponse(500, "Service creation failed", err.Error())
		return clusters.NewCreateClusterDefault(500).WithPayload(payload)
	}

	// Build the response
	// Note, the pod list here will always be empty on create
	cluster := &models.SingleCluster{&models.ClusterModel{}}
	cluster.Cluster.Name = params.Cluster.Name
	cluster.Cluster.WorkerCount = params.Cluster.WorkerCount
	cluster.Cluster.MasterCount = params.Cluster.MasterCount

	return clusters.NewCreateClusterCreated().WithLocation(masterurl).WithPayload(cluster)
}

func WaitForCount(client kclient.ReplicationControllerInterface, name string, count int) {
	// TODO probably should not spin forever here
	for {
		r, _ := client.Get(name)
		if r.Status.Replicas == count {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// DeleteClusterResponse delete a cluster
func DeleteClusterResponse(params clusters.DeleteSingleClusterParams) middleware.Responder {

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return clusters.NewDeleteSingleClusterDefault(400).WithPayload(missingnamespace(err))
	}

	// openshift rest client
	osclient, err := osa.GetOpenShiftClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}

	// kube rest client
	client, err := osa.GetKubeClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}

	// Build a selector list for the "oshinko-cluster" label
	selectorlist := makeselector("", params.Name)

	// Delete all of the deployment configs
	dcc := osclient.DeploymentConfigs(namespace)
	deployments, err := dcc.List(selectorlist)
	// TODO add something about not being able to find the deployment configs
	if err != nil {
	}
	for i := range deployments.Items {
		err = dcc.Delete(deployments.Items[i].Name)
		// TODO add something about not being able to delete deployment config
		if err != nil {
		}
	}

	// Get a list of all the replication controllers for the cluster
	// and set all of the replica values to 0
	rcc := client.ReplicationControllers(namespace)
	repls, err := rcc.List(selectorlist)
	// TODO add something about not being able to find the replication controllers
	if err != nil {
	}
	for i := range repls.Items {
		repls.Items[i].Spec.Replicas = 0
		_, err = rcc.Update(&repls.Items[i])
		// TODO add something about not being able to scale down
		if err != nil {
		}
	}

	// Wait for the replica count to drop to 0 for each
	// TODO if we failed to update the count above, we shouldn't wait for the repl here
	for i := range repls.Items {
		WaitForCount(rcc, repls.Items[i].Name, 0)
	}

	// Delete each replication controller
	for i := range repls.Items {
		err = rcc.Delete(repls.Items[i].Name)
		// TODO add something about not being able to delete the repl here
		if err != nil {
		}
	}

	// Delete the services
	sc := client.Services(namespace)
	srvs, err := sc.List(selectorlist)
	// TODO add something about not being able to get services
	if err != nil {
	}
	for i := range srvs.Items {
		err = sc.Delete(srvs.Items[i].Name)
		// TODO add something about not being able to delete services
		if err != nil {
		}
	}
	return clusters.NewDeleteSingleClusterNoContent()
}

// FindClustersResponse find a cluster and return its representation
func FindClustersResponse() middleware.Responder {

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return clusters.NewDeleteSingleClusterDefault(400).WithPayload(missingnamespace(err))
	}

	// kube rest client
	client, err := osa.GetKubeClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}
	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	// Create the payload that we're going to write into for the response
	payload := clusters.FindClustersOKBodyBody{}
	payload.Clusters = []*clusters.ClustersItems0{}

	// Create a map so that we can track clusters by name while we
	// find out information about them
	clist := map[string]*clusters.ClustersItems0{}

	// Build a selector list for the "oshinko-type" label
	pods, err := pc.List(makeselector("master", ""))

	// From the list of master pods, figure out which clusters we have
	for i := range pods.Items {

		clustername := pods.Items[i].Labels["oshinko-cluster"]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(clusters.ClustersItems0)
			citem = clist[clustername]
			citem.Name = tostrptr(clustername)
			citem.Href = tostrptr("/clusters/" + clustername)
			cnt, _ := countworkers(pc, clustername)
			citem.WorkerCount = toint64ptr(cnt)
			// TODO we only want to count running pods
			citem.Status = tostrptr("Running")
			// TODO make something real for status
			citem.MasterURL = tostrptr(retrievemasterurl(sc, clustername))
		}
	}

	for _, value := range clist {
		payload.Clusters = append(payload.Clusters, value)
	}

	return clusters.NewFindClustersOK().WithPayload(payload)
}

func singleclusterresponse(
	pc kclient.PodInterface,
	sc kclient.ServiceInterface, clustername string) (*models.SingleCluster, error) {

	// Build a selector list to get the master
	selectorlist := makeselector("master", clustername)
	pods, err := pc.List(selectorlist)
	if err != nil {
		//TODO do something here
	}

	// No master pod, we assume the cluster is not there
	if len(pods.Items) == 0 {
		return nil, nil

	} else if len(pods.Items) > 1 {
		//TODO do something here, duplicate clusters?
	}

	// Build the response
	cluster := &models.SingleCluster{&models.ClusterModel{}}
	cluster.Cluster.Name = tostrptr(clustername)
	cluster.Cluster.MasterCount = toint64ptr(1)
	cluster.Cluster.Pods = []*models.ClusterModelPodsItems0{}
	//TODO make something real for status
	cluster.Cluster.Status = tostrptr("Running")
	cluster.Cluster.MasterURL = tostrptr(retrievemasterurl(sc, clustername))

	// Report the master pod
	master := pods.Items[0]
	pod := new(models.ClusterModelPodsItems0)
	pod.IP = tostrptr(master.Status.PodIP)
	pod.Status = tostrptr(string(master.Status.Phase))
	pod.Type = tostrptr(master.Labels["oshinko-type"])
	cluster.Cluster.Pods = append(cluster.Cluster.Pods, pod)

	// Report the worker pods
	cnt, workers := countworkers(pc, clustername)
	cluster.Cluster.WorkerCount = toint64ptr(cnt)
	if workers != nil {
		for i := range workers.Items {
			w := &workers.Items[i]
			pod := new(models.ClusterModelPodsItems0)
			pod.IP = tostrptr(w.Status.PodIP)
			pod.Status = tostrptr(string(w.Status.Phase))
			pod.Type = tostrptr(w.Labels["oshinko-type"])
			cluster.Cluster.Pods = append(cluster.Cluster.Pods, pod)
		}
	}

	return cluster, nil
}


// FindSingleClusterResponse find a cluster and return its representation
func FindSingleClusterResponse(params clusters.FindSingleClusterParams) middleware.Responder {

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return clusters.NewDeleteSingleClusterDefault(400).WithPayload(missingnamespace(err))
	}

	// kube rest client
	client, err := osa.GetKubeClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}
	pc := client.Pods(namespace)
	sc := client.Services(namespace)

	cluster, err := singleclusterresponse(pc, sc, params.Name)
	if err != nil {
		// TODO do something here

	} else if cluster == nil {
		return clusters.NewFindSingleClusterOK()
	}

	return clusters.NewFindSingleClusterOK().WithPayload(cluster)
}

// UpdateSingleClusterResponse update a cluster and return the new representation
func UpdateSingleClusterResponse(params clusters.UpdateSingleClusterParams) middleware.Responder {

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return clusters.NewDeleteSingleClusterDefault(400).WithPayload(missingnamespace(err))
	}

	// kube rest client
	client, err := osa.GetKubeClient()
	if err != nil {
		return clusters.NewCreateClusterDefault(400).WithPayload(missingclient(err))
	}
	rcc := client.ReplicationControllers(namespace)

	// Get the worker replication count
	selectorlist := makeselector("worker", params.Name)
	repls, err := rcc.List(selectorlist)
	// TODO add something about not being able to find the replication controllers
	if err != nil {
	}
	if len(repls.Items) == 0 {
		// TODO do something here
	}

	// Well, there should only be one, but loop anyway
	for i := range repls.Items {
		if repls.Items[i].Spec.Replicas == int(*params.Cluster.WorkerCount) {
			continue
		}
		repls.Items[i].Spec.Replicas = int(*params.Cluster.WorkerCount)
		_, err = rcc.Update(&repls.Items[i])
		// TODO add something about not being able to scale down
		if err != nil {
		}
	}

	for i := range repls.Items {
		WaitForCount(rcc, repls.Items[i].Name, int(*params.Cluster.WorkerCount))
	}

	// throw an error if the name in the new cluster does not match the name in the old cluster
	// (this could be added at some point, but it means changing the names, labels, and selectors on
	// all the objects in the cluster)  (not implemented yet)

	// get the master repl controller and compare master count
	// error on master count change

	cluster, err := singleclusterresponse(client.Pods(namespace), client.Services(namespace), params.Name)

	return clusters.NewFindSingleClusterOK().WithPayload(cluster)
}
