package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"strings"
	"time"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
)

func NewCmdDelete(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdDelete(f, in, out)
	return cmd
}

func CmdDelete(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	options := &AuthOptions{
		Reader: reader,
		Out:    out,
	}

	cmd := &cobra.Command{
		Use:   "delete <NAME>",
		Short: "Delete spark cluster by name.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunDelete(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	return cmd
}

func (o *AuthOptions) RunDelete(out io.Writer, cmd *cobra.Command, args []string) error {
	allErrs := []error{}
	if err := o.GatherInfo(); err != nil {
		return err
	}

	kubeclient := o.KClient
	oClient := o.Client

	//fmt.Println("Deletion : ", args, o.Project)
	currentCluster, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	info, _ := deleteCluster(currentCluster, o.Project, oClient, kubeclient)
	if info != "" {
		fmt.Println("Deletion may be incomplete:")
	}

	if _, err := fmt.Fprintf(out, "cluster \"%s\" deleted \n",
		currentCluster,
	); err != nil {
		allErrs = append(allErrs, err)
	}
	return utilerrors.NewAggregate(allErrs)
}

func waitForCount(client kclient.ReplicationControllerInterface, name string, count int32) {

	for i := 0; i < 5; i++ {
		r, _ := client.Get(name)
		if int32(r.Status.Replicas) == count {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func deleteCluster(clustername, namespace string, osclient *client.Client, client *kclient.Client) (string, bool) {
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
	return strings.Join(info, ", "), foundSomething
}
