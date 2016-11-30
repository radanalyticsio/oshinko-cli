package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

func NewCmdScale(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdScale(f, in, out)
	return cmd
}

func CmdScale(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	options := &AuthOptions{
		Reader: reader,
		Out:    out,
	}

	cmd := &cobra.Command{
		Use:   ScaleCmdUsage,
		Short: ScaleCmdShort,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunScale(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().String("masters", "", "Numbers of workers in spark cluster")
	cmd.Flags().String("workers", "", "Numbers of workers in spark cluster")
	cmd.MarkFlagRequired("workers")
	return cmd
}

func (o *AuthOptions) RunScale(out io.Writer, cmd *cobra.Command, args []string) error {
	allErrs := []error{}
	if err := o.GatherInfo(); err != nil {
		return err
	}
	kubeclient := o.KClient

	//fmt.Println("Scale : ", args, o.Project)
	currentCluster, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	//fmt.Println("Scale : ", currentCluster, o.Project)

	rcc := kubeclient.ReplicationControllers(o.Project)
	repl, err := getReplController(rcc, currentCluster, workerType)
	if err != nil || repl == nil {
		return err
	}

	//pc := kubeclient.Pods(o.Project)
	// get existing workers count
	//workercount, _, _ := countWorkers(pc, currentCluster)

	workers := "1"
	//masters := "1"

	if kcmdutil.GetFlagString(cmd, "masters") != "" &&
		kcmdutil.GetFlagString(cmd, "workers") != "" {
		if _, err := fmt.Fprintf(out, "cluster \"%s\" scaled \n",
			args[0],
		); err != nil {
			allErrs = append(allErrs, err)
		}
		return utilerrors.NewAggregate(allErrs)
	}

	if kcmdutil.GetFlagString(cmd, "workers") != "" {
		workers = kcmdutil.GetFlagString(cmd, "workers")
	}

	//if (kcmdutil.GetFlagString(cmd, "masters")!="") {
	//	masters = kcmdutil.GetFlagString(cmd, "masters")
	//}
	workersInt, _ := resolveWorkers(workers)
	//, _ := resolveWorkers(masters)

	// If the current replica count does not match the request, update the replication controller
	if repl.Spec.Replicas != workersInt {
		repl.Spec.Replicas = workersInt
		_, err = rcc.Update(repl)
		if err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(out, "cluster \"%s\" scaled \n",
		currentCluster,
	); err != nil {
		allErrs = append(allErrs, err)
	}
	return utilerrors.NewAggregate(allErrs)
}

//TODO move to struct
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
	return &newestRepl, err
}
