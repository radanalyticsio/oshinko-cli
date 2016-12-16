package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

func NewCmdScale(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdScale(f, in, out)
	return cmd
}

func CmdScale(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	authOptions := &AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
	}

	cmd := &cobra.Command{
		Use:   ScaleCmdUsage,
		Short: ScaleCmdShort,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunScale(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().Int("masters", -1, "Numbers of workers in spark cluster")
	cmd.Flags().Int("workers", -1, "Numbers of workers in spark cluster")
	cmd.MarkFlagRequired("workers")
	return cmd
}

func (o *CmdOptions) RunScale() error {
	allErrs := []error{}

	ok, err := checkForDeploymentConfigs(o.Client.DeploymentConfigs(o.Project), o.Name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("No such cluster " + o.Name)
	}

	rcc := o.KClient.ReplicationControllers(o.Project)
	wrepl, err := getReplController(rcc, o.Name, workerType)
	if err != nil || wrepl == nil {
		return err
	}

	//mrepl, err := getReplController(rcc, o.Name, masterType)
	//if err != nil || mrepl == nil {
	//	return err
	//}

	// If the current replica count does not match the request, update the replication controller
	if o.WorkerCount >= 0 && o.WorkerCount <= maxWorkers &&
		wrepl.Spec.Replicas != o.WorkerCount {
		wrepl.Spec.Replicas = o.WorkerCount
		_, err = rcc.Update(wrepl)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Cannot Scale Cluster \n")
	}

	//if o.MasterCount != "" && mrepl.Spec.Replicas != o.MasterCount {
	//	mrepl.Spec.Replicas = o.MasterCount
	//	_, err = rcc.Update(mrepl)
	//	if err != nil {
	//		return err
	//	}
	//}

	if _, err := fmt.Fprintf(o.Out, "cluster \"%s\" scaled \n",
		o.Name,
	); err != nil {
		allErrs = append(allErrs, err)
	}
	return utilerrors.NewAggregate(allErrs)
}
