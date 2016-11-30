package cmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	oshinkomodels "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/models"
	"github.com/radanalyticsio/oshinko-rest/models"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
	"k8s.io/kubernetes/pkg/util/intstr"
)

const mDepConfigMsg = "Unable to create master deployment configuration"
const wDepConfigMsg = "Unable to create worker deployment configuration"
const masterSrvMsg = "Unable to create spark master service endpoint"
const imageMsg = "Cannot determine name of spark image"
const respMsg = "Created cluster but failed to construct a response object"
const defaultImage = "radanalyticsio/openshift-spark"
const defaultProject = "default"

//const masterPort = 7077
//const masterPortName = "spark-master"
//const webPortName = "spark-webui"
//const webPort = 8080

const sparkconfdir = "/etc/oshinko-spark-configs"

// The suffix to add to the spark master hostname (clustername) for the web service
const webServiceSuffix = "-ui"

const (
	createLong = `Create a resource by filename or stdin

JSON and YAML formats are accepted.`

	createExample = `  # Create a cluster using the data in cluster.json.
  %[1]s create -f cluster.json

  # Create a cluster based on the JSON passed into stdin.
  cat cluster.json | %[1]s create -f -`
)

var (
	sparkClusterLong = dedent.Dedent(`
		Create a spark cluster with the specified name.`)

	sparkClusterExample = dedent.Dedent(`
		  # Create a new spark cluster named my-spark-cluster
		  $ oshinko create cluster my-spark-cluster`)
)

func NewCmdCreate(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdCreate(f, in, out)
	return cmd
}

func CmdCreate(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	options := &AuthOptions{
		Reader: reader,
		Out:    out,
	}

	cmd := &cobra.Command{
		Use:   "create <NAME> --masters <MASTER> --workers <WORKERS> --image <IMAGE> --sparkconfdir <DIR>",
		Short: "Create new spark clusters",
		Long:  "Create spark cluster.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if _, err := options.RunCreate(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().String("masters", "", "Numbers of workers in spark cluster")
	cmd.Flags().String("workers", "", "Numbers of workers in spark cluster")
	cmd.Flags().String("sparkconfdir", "", "Config folder for spark")
	cmd.Flags().String("image", "", "spark image to be used.Default value is radanalyticsio/openshift-spark.")
	cmd.MarkFlagRequired("workers")
	return cmd
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

func singleClusterResponse(clustername string,
	pc kclient.PodInterface,
	sc kclient.ServiceInterface, workersInt int, mastersInt int) (*models.SingleCluster, error) {

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
	cluster.Cluster.Config = &models.NewClusterConfig{}

	// Report the master pod
	selectorlist := makeSelector(masterType, clustername)
	pods, err := pc.List(selectorlist)
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(pods.Items[i]))
	}

	// Report the worker pods
	_, workers, err := countWorkers(pc, clustername)
	if err != nil {
		return nil, err
	}
	for i := range workers.Items {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(workers.Items[i]))
	}

	cluster.Cluster.Config.WorkerCount = int64(workersInt)
	cluster.Cluster.Config.MasterCount = int64(mastersInt)
	//if config.SparkWorkerConfig != "" {
	//	cluster.Cluster.Config.SparkWorkerConfig = config.SparkWorkerConfig
	//}
	//if config.SparkMasterConfig != "" {
	//	cluster.Cluster.Config.SparkMasterConfig = config.SparkMasterConfig
	//}
	return cluster, nil
}

func (o *AuthOptions) RunCreate(out io.Writer, cmd *cobra.Command, args []string) (*models.SingleCluster, error) {
	allErrs := []error{}
	if err := o.GatherInfo(); err != nil {
		return nil, err
	}
	kubeclient, err := kclient.New(o.Config)
	if err != nil {
		return nil, err
	}

	oClient, err := client.New(o.Config)
	if err != nil {
		return nil, err
	}

	namespace := defaultProject
	workers := "1"
	masters := "1"
	image := defaultImage
	masterconfdir := sparkconfdir
	workerconfdir := sparkconfdir
	clustername := ""

	if o.Project != "" {
		namespace = o.Project
	}

	if cmdutil.GetFlagString(cmd, "workers") != "" {
		workers = cmdutil.GetFlagString(cmd, "workers")
	}

	if cmdutil.GetFlagString(cmd, "masters") != "" {
		masters = cmdutil.GetFlagString(cmd, "masters")
	}

	if cmdutil.GetFlagString(cmd, "image") != "" {
		image = cmdutil.GetFlagString(cmd, "image")
	}

	if cmdutil.GetFlagString(cmd, "sparkconfdir") != "" {
		masterconfdir = cmdutil.GetFlagString(cmd, "sparkconfdir")
	}

	if cmdutil.GetFlagString(cmd, "sparkconfdir") != "" {
		workerconfdir = cmdutil.GetFlagString(cmd, "sparkconfdir")
	}

	if args[0] != "" {
		clustername = args[0]
	}
	workersInt, _ := resolveWorkers(workers)
	mastersInt, _ := resolveWorkers(masters)

	// pre spark 2, the name the master calls itself must match
	// the name the workers use and the service name created
	masterhost := clustername

	// Create the master deployment config
	dcc := oClient.DeploymentConfigs(namespace)
	masterdc := sparkMaster(namespace, image, clustername, masterconfdir, "")

	// Create the services that will be associated with the master pod
	// They will be created with selectors based on the pod labels
	mastersv, _ := service(masterhost,
		masterdc.FindPort(masterPortName),
		clustername, masterType,
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := service(masterhost+"-ui",
		masterdc.FindPort(webPortName),
		clustername, webuiType,
		masterdc.GetPodTemplateSpecLabels())

	// Create the worker deployment config
	//masterurl := sparkMasterURL(masterhost, &masterp.ServicePort)
	workerdc := sparkWorker(namespace, image, workersInt, clustername, workerconfdir, "")

	// Launch all of the objects
	_, err = dcc.Create(&masterdc.DeploymentConfig)
	if err != nil {
		return nil, fmt.Errorf(mDepConfigMsg, err)
		//return reterr(fail(err, mDepConfigMsg, code(err)))
	}
	_, err = dcc.Create(&workerdc.DeploymentConfig)
	if err != nil {
		// Since we created the master deployment config, try to clean up
		deleteCluster(clustername, namespace, oClient, kubeclient)
		return nil, fmt.Errorf(wDepConfigMsg, err)
	}

	// If we've gotten this far, then likely the cluster naming is not in conflict so
	// assume at this point that we should use a 500 error code
	sc := kubeclient.Services(namespace)
	_, err = sc.Create(&mastersv.Service)
	if err != nil {
		// Since we create the master and workers, try to clean up
		deleteCluster(clustername, namespace, oClient, kubeclient)
		//return reterr(fail(err, masterSrvMsg, code(err)))
		return nil, fmt.Errorf(masterSrvMsg, err)
	}

	// Note, if spark webui service fails for some reason we can live without it
	// TODO ties into cluster status, make a note if the service is missing
	sc.Create(&websv.Service)

	// Since we already know what the masterurl is, pass it in explicitly and do not pass a service client
	cluster, err := singleClusterResponse(clustername, kubeclient.Pods(namespace), sc, workersInt, mastersInt)
	if err != nil {
		//return reterr(responseFailure(err, respMsg, 500))
		return nil, fmt.Errorf(respMsg, err)
	}

	if _, err := fmt.Fprintf(out, "cluster \"%s\" created \n",
		args[0],
	); err != nil {
		allErrs = append(allErrs, err)
	}
	return cluster, utilerrors.NewAggregate(allErrs)
	//fmt.Fprintf(out, "cluster  \"%s\" created  with  %d  workers", options.Name, workersInt)
	//fmt.Fprintf(out, "\nUsing project %q on server %q.\n", namespace, ClientConfig.Host)
	//return nil
}

func ValidateArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageError(cmd, "Unexpected args: %v", args)
	}
	return nil
}

func makeEnvVars(clustername, sparkconfdir string) []kapi.EnvVar {
	envs := []kapi.EnvVar{}

	envs = append(envs, kapi.EnvVar{Name: "OSHINKO_SPARK_CLUSTER", Value: clustername})
	if sparkconfdir != "" {
		envs = append(envs, kapi.EnvVar{Name: "SPARK_CONF_DIR", Value: sparkconfdir})
	}

	return envs
}

func makeWorkerEnvVars(clustername, sparkconfdir string) []kapi.EnvVar {
	envs := []kapi.EnvVar{}
	masterPort := 7077
	webPort := 8080
	envs = makeEnvVars(clustername, sparkconfdir)
	envs = append(envs, kapi.EnvVar{
		Name:  "SPARK_MASTER_ADDRESS",
		Value: "spark://" + clustername + ":" + strconv.Itoa(masterPort)})
	envs = append(envs, kapi.EnvVar{
		Name:  "SPARK_MASTER_UI_ADDRESS",
		Value: "http://" + clustername + webServiceSuffix + ":" + strconv.Itoa(webPort)})
	return envs
}

func sparkMaster(namespace, image, clustername, sparkconfdir, sparkmasterconfig string) *oshinkomodels.ODeploymentConfig {
	masterPort := 7077
	webPort := 8080
	masterPortName := "spark-master"
	webPortName := "spark-webui"
	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := oshinkomodels.DeploymentConfig(clustername+"-m", namespace).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, clustername).
		Label(typeLabel, masterType).
		PodSelector(clusterLabel, clustername)

	// Create a pod template spec with the matching label
	pt := oshinkomodels.PodTemplateSpec().Label(clusterLabel, clustername).
		Label(typeLabel, masterType)

	// Create a container with the correct ports and start command
	httpProbe := NewHTTPGetProbe(webPort)
	masterp := oshinkomodels.ContainerPort(masterPortName, masterPort)
	webp := oshinkomodels.ContainerPort(webPortName, webPort)
	cont := oshinkomodels.Container(dc.Name, image).
		Ports(masterp, webp).
		SetLivenessProbe(httpProbe).
		SetReadinessProbe(httpProbe).
		EnvVars(makeEnvVars(clustername, sparkconfdir))

	if sparkmasterconfig != "" {
		pt = pt.SetConfigMapVolume(sparkmasterconfig)
		cont = cont.SetVolumeMount(sparkmasterconfig, sparkconfdir, true)
	}

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func NewHTTPGetProbe(port int) kapi.Probe {
	act := kapi.HTTPGetAction{Port: intstr.FromInt(port)}
	hnd := kapi.Handler{HTTPGet: &act}
	prb := kapi.Probe{Handler: hnd}
	return prb
}

func sparkWorker(namespace string,
	image string,
	replicas int, clustername, sparkconfdir, sparkworkerconfig string) *oshinkomodels.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name.
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := oshinkomodels.DeploymentConfig(clustername+"-w", namespace).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, clustername).
		Label(typeLabel, workerType).
		PodSelector(clusterLabel, clustername).Replicas(replicas)

	// Create a pod template spec with the matching label
	pt := oshinkomodels.PodTemplateSpec().Label(clusterLabel, clustername).Label(typeLabel, workerType)

	// Create a container with the correct ports and start command
	webport := 8081
	webp := oshinkomodels.ContainerPort(webPortName, webport)
	cont := oshinkomodels.Container(dc.Name, image).Ports(webp).
		EnvVars(makeWorkerEnvVars(clustername, sparkconfdir)).
		SetLivenessProbe(NewHTTPGetProbe(webport))

	if sparkworkerconfig != "" {
		pt = pt.SetConfigMapVolume(sparkworkerconfig)
		cont = cont.SetVolumeMount(sparkworkerconfig, sparkconfdir, true)
	}

	// Finally, assign the container to the pod template spec and
	// assign the pod template spec to the deployment config
	return dc.PodTemplateSpec(pt.Containers(cont))
}

func service(name string,
	port int,
	clustername, otype string,
	podselectors map[string]string) (*oshinkomodels.OService, *oshinkomodels.OServicePort) {

	p := oshinkomodels.ServicePort(port).TargetPort(port)
	return oshinkomodels.Service(name).Label(clusterLabel, clustername).
		Label(typeLabel, otype).PodSelectors(podselectors).Ports(p), p
}

func checkForConfigMap(name string, cm kclient.ConfigMapsInterface) error {
	cmap, err := cm.Get(name)
	if err == nil && cmap == nil {
		err = fmt.Errorf("ConfigMap '%s' not found", name)
	}
	return err
}

func resolveWorkers(workers string) (int, error) {
	if len(workers) == 0 {
		return 0, nil
	}
	var worker intstr.IntOrString
	integer, err := strconv.Atoi(workers)
	if err != nil {
		worker = intstr.FromString(workers)
	} else {
		worker = intstr.FromInt(integer)
	}
	return worker.IntValue(), nil
}

func PrintObject(cmd *cobra.Command, out io.Writer) error {
	allErrs := []error{}
	//if _, err := fmt.Fprintf(out, "%s\t%d/%d\t%s\t%d\t%s",
	//	"abc",
	//	1,
	//	2,
	//	"",
	//	1,
	//	"",
	//); err != nil {
	//	allErrs = append(allErrs, err)
	//}
	return utilerrors.NewAggregate(allErrs)
}
