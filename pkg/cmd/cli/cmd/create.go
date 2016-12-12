package cmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	oshinkomodels "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/models"
	kapi "k8s.io/kubernetes/pkg/api"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
	"k8s.io/kubernetes/pkg/util/intstr"
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
	authOptions := &AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
	}

	cmd := &cobra.Command{
		Use: "create <NAME> ",
		//--masters <MASTER> --workers <WORKERS> --image <IMAGE> --sparkmasterconfig <DIR>
		Short: "Create new spark clusters",
		Long:  "Create spark cluster.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunCreate(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().Int("masters", 1, "Numbers of workers in spark cluster")
	cmd.Flags().Int("workers", 1, "Numbers of workers in spark cluster")
	cmd.Flags().String("masterconfigdir", defaultsparkconfdir, "Config folder for spark master")
	cmd.Flags().String("workerconfigdir", defaultsparkconfdir, "Config folder for spark worker")
	cmd.Flags().String("masterconfig", "", "ConfigMap name for spark master")
	cmd.Flags().String("workerconfig", "", "ConfigMap name for spark worker")
	cmd.Flags().String("image", defaultImage, "spark image to be used.Default value is radanalyticsio/openshift-spark.")
	//cmd.MarkFlagRequired("workers")
	return cmd
}

func (o *CmdOptions) RunCreate(out io.Writer, cmd *cobra.Command, args []string) error {
	allErrs := []error{}
	// pre spark 2, the name the master calls itself must match
	// the name the workers use and the service name created
	masterhost := o.Name

	// Create the master deployment config
	dcc := o.Client.DeploymentConfigs(o.Project)
	masterdc := sparkMaster(o)

	// Create the services that will be associated with the master pod
	// They will be created with selectors based on the pod labels
	mastersv, _ := service(masterhost,
		masterdc.FindPort(masterPortName),
		o.Name, masterType,
		masterdc.GetPodTemplateSpecLabels())

	websv, _ := service(masterhost+"-ui",
		masterdc.FindPort(webPortName),
		o.Name, webuiType,
		masterdc.GetPodTemplateSpecLabels())

	// Create the worker deployment config
	//masterurl := sparkMasterURL(masterhost, &masterp.ServicePort)
	workerdc := sparkWorker(o)

	// Launch all of the objects
	_, err := dcc.Create(&masterdc.DeploymentConfig)
	if err != nil {
		return fmt.Errorf(mDepConfigMsg, err)
	}
	_, err = dcc.Create(&workerdc.DeploymentConfig)
	if err != nil {
		// Since we created the master deployment config, try to clean up
		deleteCluster(o)
		return fmt.Errorf(wDepConfigMsg, err)
	}

	// If we've gotten this far, then likely the cluster naming is not in conflict so
	// assume at this point that we should use a 500 error code
	sc := o.KClient.Services(o.Project)
	_, err = sc.Create(&mastersv.Service)
	if err != nil {
		// Since we create the master and workers, try to clean up
		deleteCluster(o)
		return fmt.Errorf(masterSrvMsg, err)
	}

	// Note, if spark webui service fails for some reason we can live without it
	// TODO ties into cluster status, make a note if the service is missing
	sc.Create(&websv.Service)

	if _, err := fmt.Fprintf(out, "cluster \"%s\" created \n",
		o.Name,
	); err != nil {
		allErrs = append(allErrs, err)
	}
	return utilerrors.NewAggregate(allErrs)
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
	envs = makeEnvVars(clustername, sparkconfdir)
	envs = append(envs, kapi.EnvVar{
		Name:  "SPARK_MASTER_ADDRESS",
		Value: "spark://" + clustername + ":" + strconv.Itoa(masterPort)})
	envs = append(envs, kapi.EnvVar{
		Name:  "SPARK_MASTER_UI_ADDRESS",
		Value: "http://" + clustername + webServiceSuffix + ":" + strconv.Itoa(webPort)})
	return envs
}

func sparkMaster(o *CmdOptions) *oshinkomodels.ODeploymentConfig {
	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := oshinkomodels.DeploymentConfig(o.Name+"-m", o.Project).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, o.Name).
		Label(typeLabel, masterType).
		PodSelector(clusterLabel, o.Name)

	// Create a pod template spec with the matching label
	pt := oshinkomodels.PodTemplateSpec().Label(clusterLabel, o.Name).
		Label(typeLabel, masterType)

	// Create a container with the correct ports and start command
	httpProbe := NewHTTPGetProbe(webPort)
	masterp := oshinkomodels.ContainerPort(masterPortName, masterPort)
	webp := oshinkomodels.ContainerPort(webPortName, webPort)
	cont := oshinkomodels.Container(dc.Name, o.Image).
		Ports(masterp, webp).
		SetLivenessProbe(httpProbe).
		SetReadinessProbe(httpProbe).
		EnvVars(makeEnvVars(o.Name, o.MasterConfigDir))

	if o.MasterConfig != "" {
		pt = pt.SetConfigMapVolume(o.MasterConfig)
		cont = cont.SetVolumeMount(o.MasterConfig, o.MasterConfigDir, true)
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

func sparkWorker(o *CmdOptions) *oshinkomodels.ODeploymentConfig {

	// Create the basic deployment config
	// We will use a label and pod selector based on the cluster name.
	// Openshift will add additional labels and selectors to distinguish pods handled by
	// this deploymentconfig from pods beloning to another.
	dc := oshinkomodels.DeploymentConfig(o.Name+"-w", o.Project).
		TriggerOnConfigChange().RollingStrategy().Label(clusterLabel, o.Name).
		Label(typeLabel, workerType).
		PodSelector(clusterLabel, o.Name).Replicas(o.WorkerCount)

	// Create a pod template spec with the matching label
	pt := oshinkomodels.PodTemplateSpec().Label(clusterLabel, o.Name).Label(typeLabel, workerType)

	// Create a container with the correct ports and start command
	webport := 8081
	webp := oshinkomodels.ContainerPort(webPortName, webport)
	cont := oshinkomodels.Container(dc.Name, o.Image).Ports(webp).
		EnvVars(makeWorkerEnvVars(o.Name, o.WorkerConfigDir)).
		SetLivenessProbe(NewHTTPGetProbe(webport))

	if o.WorkerConfig != "" {
		pt = pt.SetConfigMapVolume(o.WorkerConfig)
		cont = cont.SetVolumeMount(o.WorkerConfig, o.WorkerConfigDir, true)
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
