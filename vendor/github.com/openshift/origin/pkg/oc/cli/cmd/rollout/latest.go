package rollout

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	kprinters "k8s.io/kubernetes/pkg/printers"

	appsapi "github.com/openshift/origin/pkg/apps/apis/apps"
	appsclientinternal "github.com/openshift/origin/pkg/apps/generated/internalclientset/typed/apps/internalversion"
	appsutil "github.com/openshift/origin/pkg/apps/util"
	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
)

var (
	rolloutLatestLong = templates.LongDesc(`
		Start a new rollout for a deployment config with the latest state from its triggers

		This command is appropriate for running manual rollouts. If you want full control over
		running new rollouts, use "oc set triggers --manual" to disable all triggers in your
		deployment config and then whenever you want to run a new deployment process, use this
		command in order to pick up the latest images found in the cluster that are pointed by
		your image change triggers.`)

	rolloutLatestExample = templates.Examples(`
	# Start a new rollout based on the latest images defined in the image change triggers.
	%[1]s rollout latest dc/nginx

	# Print the rolled out deployment config
	%[1]s rollout latest dc/nginx -o json`)
)

// RolloutLatestOptions holds all the options for the `rollout latest` command.
type RolloutLatestOptions struct {
	mapper meta.RESTMapper
	typer  runtime.ObjectTyper
	infos  []*resource.Info

	DryRun bool
	out    io.Writer
	output string
	again  bool

	appsClient      appsclientinternal.DeploymentConfigsGetter
	kc              kclientset.Interface
	baseCommandName string

	printer kprinters.ResourcePrinter
}

// NewCmdRolloutLatest implements the oc rollout latest subcommand.
func NewCmdRolloutLatest(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	opts := &RolloutLatestOptions{
		baseCommandName: fullName,
	}

	cmd := &cobra.Command{
		Use:     "latest DEPLOYMENTCONFIG",
		Short:   "Start a new rollout for a deployment config with the latest state from its triggers",
		Long:    rolloutLatestLong,
		Example: fmt.Sprintf(rolloutLatestExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			err := opts.Complete(f, cmd, args, out)
			kcmdutil.CheckErr(err)

			if err := opts.Validate(); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageErrorf(cmd, err.Error()))
			}

			err = opts.RunRolloutLatest()
			kcmdutil.CheckErr(err)
		},
		ValidArgs: []string{"deploymentconfig"},
	}

	kcmdutil.AddPrinterFlags(cmd)
	kcmdutil.AddDryRunFlag(cmd)
	cmd.Flags().Bool("again", false, "If true, deploy the current pod template without updating state from triggers")

	return cmd
}

func (o *RolloutLatestOptions) Complete(f *clientcmd.Factory, cmd *cobra.Command, args []string, out io.Writer) error {
	if len(args) != 1 {
		return errors.New("one deployment config name is needed as argument.")
	}

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	o.DryRun = kcmdutil.GetFlagBool(cmd, "dry-run")

	o.kc, err = f.ClientSet()
	if err != nil {
		return err
	}
	appsClient, err := f.OpenshiftInternalAppsClient()
	if err != nil {
		return err
	}
	o.appsClient = appsClient.Apps()

	o.mapper, o.typer = f.Object()
	o.infos, err = f.NewBuilder().
		Internal().
		ContinueOnError().
		NamespaceParam(namespace).
		ResourceNames("deploymentconfigs", args[0]).
		SingleResourceType().
		Do().Infos()
	if err != nil {
		return err
	}

	o.out = out
	o.output = kcmdutil.GetFlagString(cmd, "output")
	o.again = kcmdutil.GetFlagBool(cmd, "again")

	if o.output != "revision" {
		o.printer, err = f.PrinterForOptions(kcmdutil.ExtractCmdPrintOptions(cmd, false))
		if err != nil {
			return err
		}
	}

	return nil
}

func (o RolloutLatestOptions) Validate() error {
	if len(o.infos) != 1 {
		return errors.New("a deployment config name is required.")
	}
	return nil
}

func (o RolloutLatestOptions) RunRolloutLatest() error {
	info := o.infos[0]
	config, ok := info.Object.(*appsapi.DeploymentConfig)
	if !ok {
		return fmt.Errorf("%s is not a deployment config", info.Name)
	}

	// TODO: Consider allowing one-off deployments for paused configs
	// See https://github.com/openshift/origin/issues/9903
	if config.Spec.Paused {
		return fmt.Errorf("cannot deploy a paused deployment config")
	}

	deploymentName := appsutil.LatestDeploymentNameForConfig(config)
	deployment, err := o.kc.Core().ReplicationControllers(config.Namespace).Get(deploymentName, metav1.GetOptions{})
	switch {
	case err == nil:
		// Reject attempts to start a concurrent deployment.
		if !appsutil.IsTerminatedDeployment(deployment) {
			status := appsutil.DeploymentStatusFor(deployment)
			return fmt.Errorf("#%d is already in progress (%s).", config.Status.LatestVersion, status)
		}
	case !kerrors.IsNotFound(err):
		return err
	}

	dc := config
	if !o.DryRun {
		request := &appsapi.DeploymentRequest{
			Name:   config.Name,
			Latest: !o.again,
			Force:  true,
		}

		dc, err = o.appsClient.DeploymentConfigs(config.Namespace).Instantiate(config.Name, request)

		// Pre 1.4 servers don't support the instantiate endpoint. Fallback to incrementing
		// latestVersion on them.
		if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
			config.Status.LatestVersion++
			dc, err = o.appsClient.DeploymentConfigs(config.Namespace).Update(config)
		}

		if err != nil {
			return err
		}

		info.Refresh(dc, true)
	}

	if o.output == "revision" {
		fmt.Fprintf(o.out, fmt.Sprintf("%d", dc.Status.LatestVersion))
		return nil
	} else if len(o.output) > 0 {
		return o.printer.PrintObj(dc, o.out)
	}

	kcmdutil.PrintSuccess(o.mapper, o.output == "name", o.out, info.Mapping.Resource, info.Name, o.DryRun, "rolled out")
	return nil
}
