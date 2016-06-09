package handlers

import (
	"os"
	"strconv"
	"time"

	middleware "github.com/go-openapi/runtime/middleware"

	_ "github.com/openshift/origin/pkg/api/install"
	serverapi "github.com/openshift/origin/pkg/cmd/server/api"
	ocon "github.com/redhatanalytics/oshinko-rest/helpers/containers"
	odc "github.com/redhatanalytics/oshinko-rest/helpers/deploymentconfigs"
	opt "github.com/redhatanalytics/oshinko-rest/helpers/podtemplates"
	osv "github.com/redhatanalytics/oshinko-rest/helpers/services"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations/clusters"

	"github.com/redhatanalytics/oshinko-rest/models"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)

func sparkMasterURL(name string, port *osv.OServicePort) string {
	return "spark://" + name + ":" + strconv.Itoa(port.ServicePort.Port)
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
		PodSelector("oshinko-cluster", clustername).
		Replicas(replicas)

	// Create a pod template spec with the matching label
	pt := opt.PodTemplateSpec().Label("oshinko-cluster", clustername)

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
		PodSelector("oshinko-cluster", clustername)

	// Create a pod template spec with the matching label
	// Additionally, we need another label specifically for the master
	// so that the services can find the correct pod
	pt := opt.PodTemplateSpec().Label("oshinko-cluster", clustername).Label("spark-master", clustername)

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

func Service(name string, port int, clustername string, podselectors map[string]string) (*osv.OService, *osv.OServicePort) {
	p := osv.ServicePort(port).TargetPort(port)
	return osv.Service(name).Label("oshinko-cluster", clustername).PodSelectors(podselectors).Ports(p), p
}

// CreateClusterResponse create a cluster and return the representation
func CreateClusterResponse(params clusters.CreateClusterParams) middleware.Responder {
	// create a cluster here
	// INFO(elmiko) my thinking on creating clusters is that we should use a
	// label on the items we create with kubernetes so that we can get them
	// all with a request.
	// in addition to labels for general identification, we should then use
	// annotations on objects to help further refine what we are dealing with.

	namespace := os.Getenv("OSHINKO_CLUSTER_NAMESPACE")
	configfile := os.Getenv("OSHINKO_KUBE_CONFIG")
	image := os.Getenv("OSHINKO_CLUSTER_IMAGE")
	if namespace == "" || configfile == "" || image == "" {
		payload := makeSingleErrorResponse(400, "Missing Env",
			"OSHIKO_CLUSTER_NAMESPACE, OSHINKO_KUBE_CONFIG, and OSHINKO_CLUSTER_IMAGE env vars must be set")
		return clusters.NewCreateClusterDefault(400).WithPayload(payload)
	}

	// kube rest client
	// TODO add an error on failure to get client (wait for merge of auth stuff)
	client, _, err := serverapi.GetKubeClient(configfile)
	if err != nil {
		// handle error
	}

	// openshift rest client
	osclient, _, err := serverapi.GetOpenShiftClient(configfile)
	if err != nil {
		//handle error
	}

	// deployment config client
	dcc := osclient.DeploymentConfigs(namespace)

	// Make master deployment config
	// Ignoring master-count for now, leave it defaulted at 1
	masterdc := sparkMaster(namespace, image, *params.Cluster.Name)

	// Make master services
	mastersv, masterp := Service(masterdc.Name,
		masterdc.FindPort("spark-master"),
		*params.Cluster.Name,
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := Service(masterdc.Name+"-webui",
		masterdc.FindPort("spark-webui"),
		*params.Cluster.Name,
		masterdc.GetPodTemplateSpecLabels())

	// Make worker deployment config
	masterurl := sparkMasterURL(mastersv.Name, masterp)
	workerdc := sparkWorker(namespace, image, int(*params.Cluster.WorkerCount), masterurl, *params.Cluster.Name)

	// Launch all of the objects
	// TODO if error says that the deploymentconfig already exists return a cluster <name> already exists
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

	namespace := os.Getenv("OSHINKO_CLUSTER_NAMESPACE")
	configfile := os.Getenv("OSHINKO_KUBE_CONFIG")
	if namespace == "" || configfile == "" {
		payload := makeSingleErrorResponse(400, "Missing Env",
			"OSHIKO_CLUSTER_NAMESPACE and OSHINKO_KUBE_CONFIG env vars must be set")
		return clusters.NewDeleteSingleClusterDefault(400).WithPayload(payload)
	}

	// openshift rest client
	// TODO add an error on failure to get client (wait for merge of auth stuff)
	osclient, _, err := serverapi.GetOpenShiftClient(configfile)
	if err != nil {
	}

	// kube rest client
	client, _, err := serverapi.GetKubeClient(configfile)
	if err != nil {
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

// FindClustersResponse find a cluster and return its representation
func FindClustersResponse() middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.FindClusters has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
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
