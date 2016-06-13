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

func sparkWorker(namespace string,
	image string,
	replicas int, masterurl, clustername string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name.
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(clustername+"-spark-worker", namespace).
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

func sparkMaster(namespace string, image string, clustername string) *odc.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := odc.DeploymentConfig(clustername+"-spark-master", namespace).
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
		image).Command("/start-master",
		dc.Name).Ports(masterp, webp)

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func Service(name string, port int,
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
	masterdc := sparkMaster(namespace, image, *params.Cluster.Name)

	// Make master services
	mastersv, masterp := Service(masterdc.Name,
		masterdc.FindPort("spark-master"),
		*params.Cluster.Name, "master",
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := Service(masterdc.Name+"-webui",
		masterdc.FindPort("spark-webui"),
		*params.Cluster.Name, "webui",
		masterdc.GetPodTemplateSpecLabels())

	// Make worker deployment config
	masterurl := sparkMasterURL(mastersv.Name, &masterp.ServicePort)
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
	cluster := &models.SingleCluster{&models.ClusterModel{}}
	cluster.Cluster.Name = params.Cluster.Name
	cluster.Cluster.WorkerCount = params.Cluster.WorkerCount
	cluster.Cluster.MasterCount = params.Cluster.MasterCount

	// TODO can we fill in some pods here, maybe the deployment pods?
	return clusters.NewCreateClusterCreated().WithLocation(masterurl).WithPayload(cluster)
}

func WaitForZero(client kclient.ReplicationControllerInterface, name string) {
	// TODO probably should not spin forever here
	for {
		r, _ := client.Get(name)
		if r.Status.Replicas == 0 {
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
	requirement, err := labels.NewRequirement("oshinko-cluster", labels.EqualsOperator, sets.NewString(params.Name))
	selectorlist := kapi.ListOptions{LabelSelector: labels.NewSelector().Add(*requirement)}

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
		WaitForZero(rcc, repls.Items[i].Name)
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

func countworkers(client kclient.PodInterface, clustername string) int {

	// Build a selector list for the "oshinko-type" label
	otype, _ := labels.NewRequirement("oshinko-type", labels.EqualsOperator, sets.NewString("worker"))
	cluster, _ := labels.NewRequirement("oshinko-cluster", labels.EqualsOperator, sets.NewString(clustername))
	selectorlist := kapi.ListOptions{LabelSelector: labels.NewSelector().Add(*otype).Add(*cluster)}
	pods, _ := client.List(selectorlist)
	return len(pods.Items)
}

func retrievemasterurl(client kclient.ServiceInterface, clustername string) string {
	// Build a selector list for the "oshinko-type" label
	otype, _ := labels.NewRequirement("oshinko-type", labels.EqualsOperator, sets.NewString("master"))
	cluster, _ := labels.NewRequirement("oshinko-cluster", labels.EqualsOperator, sets.NewString(clustername))
	selectorlist := kapi.ListOptions{LabelSelector: labels.NewSelector().Add(*otype).Add(*cluster)}
	srvs, _ := client.List(selectorlist)
	srv := srvs.Items[0]
	return sparkMasterURL(srv.Name, &srv.Spec.Ports[0])
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
	requirement, err := labels.NewRequirement("oshinko-type", labels.EqualsOperator, sets.NewString("master"))
	selectorlist := kapi.ListOptions{LabelSelector: labels.NewSelector().Add(*requirement)}
	pods, err := pc.List(selectorlist)

	// From the list of master pods, figure out which clusters we have
	for i := range pods.Items {

		clustername := pods.Items[i].Labels["oshinko-cluster"]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(clusters.ClustersItems0)
			citem = clist[clustername]
			citem.Name = new(string)
			*citem.Name = clustername
			citem.Href = new(string)
			*citem.Href = "/clusters/" + clustername
			citem.WorkerCount = new(int64)
			// TODO we only want to count running pods
			*citem.WorkerCount = int64(countworkers(pc, clustername))
			citem.Status = new(string)
			*citem.Status = "Running"
			citem.MasterURL = new(string)
			*citem.MasterURL = retrievemasterurl(sc, clustername)
		}
	}

	for _, value := range clist {
		payload.Clusters = append(payload.Clusters, value)
	}

	return clusters.NewFindClustersOK().WithPayload(payload)
}

// FindSingleClusterResponse find a cluster and return its representation
func FindSingleClusterResponse(clusters.FindSingleClusterParams) middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.FindSingleCluster has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}

// UpdateSingleClusterResponse update a cluster and return the new representation
func UpdateSingleClusterResponse(params clusters.UpdateSingleClusterParams) middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.UpdateSingleCluster has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}
