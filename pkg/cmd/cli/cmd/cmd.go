package cmd

import (
	"fmt"
	//osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"strconv"

	osclientcmd "github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	"github.com/radanalyticsio/oshinko-cli/version"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type CmdOptions struct {
	Name           string `json:"name,omitempty"`
	Image          string `json:"image"`
	WorkerCount    int    `json:"workerCount"`
	MasterCount    int    `json:"masterCount,omitempty"`
	MasterConfig   string `json:"sparkMasterConfig,omitempty"`
	WorkerConfig   string `json:"workerConfig,omitempty"`
	StoredConfig   string `json:"storedConfig,omitempty"`
	ExposeWebUI    string `json:"exposeui,omitempty"`
	AppStatus      string `json:"appStatus,omitempty"`
	App            string `json:"app,omitempty"`
	Verbose        bool
	Output         string
	Directory      string
	Ephemeral      bool
	NoNameRequired bool
	Metrics        string
	NoPods         bool
	auth.AuthOptions
}

func (o *CmdOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string) error {
	o.Image = version.GetSparkImage()

	// Pod counts will be assigned by the default config in oshinko-core,
	// here values should be defaulted to the sentinel value
	o.WorkerCount = clusters.SentinelCountValue
	o.MasterCount = clusters.SentinelCountValue

	currentCluster, err := NameFromCommandArgs(cmd, args, o.NoNameRequired)
	if err != nil {
		return err
	}
	o.Name = currentCluster

	/*
		#
		#	Fill AuthOptions
		#
	*/
	if err := o.AuthOptions.Complete(f, cmd, args, "oshinko"); err != nil {
		return err
	}
	/*
		#
	*/
	if err := o.GatherInfo(); err != nil {
		return err
	}
	if cmd.Flags().Lookup("verbose") != nil {
		o.Verbose = kcmdutil.GetFlagBool(cmd, "verbose")
	}
	if cmd.Flags().Lookup("output") != nil {
		o.Output = kcmdutil.GetFlagString(cmd, "output")
		if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
			return cmdutil.UsageErrorf(cmd, "INVALID output format, only yaml|json allowed")
		}
	}
	if cmd.Flags().Lookup("workers") != nil {
		o.WorkerCount = kcmdutil.GetFlagInt(cmd, "workers")
	}
	if cmd.Flags().Lookup("masters") != nil {
		o.MasterCount = kcmdutil.GetFlagInt(cmd, "masters")
	}
	if cmd.Flags().Lookup("image") != nil {
		o.Image = kcmdutil.GetFlagString(cmd, "image")
	}
	if cmd.Flags().Lookup("masterconfig") != nil {
		o.MasterConfig = kcmdutil.GetFlagString(cmd, "masterconfig")
	}
	if cmd.Flags().Lookup("workerconfig") != nil {
		o.WorkerConfig = kcmdutil.GetFlagString(cmd, "workerconfig")
	}
	if cmd.Flags().Lookup("storedconfig") != nil {
		o.StoredConfig = kcmdutil.GetFlagString(cmd, "storedconfig")
	}
	if cmd.Flags().Lookup("exposeui") != nil {
		o.ExposeWebUI = kcmdutil.GetFlagString(cmd, "exposeui")
		if o.ExposeWebUI != "" {
			_, err := strconv.ParseBool(o.ExposeWebUI)
			if err != nil {
				return cmdutil.UsageErrorf(cmd, "Value for 'exposeui' must be a boolean")
			}
		}
	}
	if cmd.Flags().Lookup("metrics") != nil {
		o.Metrics = kcmdutil.GetFlagString(cmd, "metrics")
		if o.Metrics != "" {
			_, err := strconv.ParseBool(o.Metrics)
			if err != nil && o.Metrics != "jolokia" && o.Metrics != "prometheus" {
				return cmdutil.UsageErrorf(cmd, "Value for 'metrics' must be 'true', 'false', 'jolokia', or 'prometheus'")
			}
		}
	}
	if cmd.Flags().Lookup("app-status") != nil {
		o.AppStatus = kcmdutil.GetFlagString(cmd, "app-status")
		if o.AppStatus != "" && o.AppStatus != "completed" && o.AppStatus != "terminated" {
			return cmdutil.UsageErrorf(cmd, "INVALID app-status value, only completed|terminated allowed")
		}
	}
	if cmd.Flags().Lookup("app") != nil {
		o.App = kcmdutil.GetFlagString(cmd, "app")
	}

	if cmd.Flags().Lookup("ephemeral") != nil {
		o.Ephemeral = kcmdutil.GetFlagBool(cmd, "ephemeral")
		if o.Ephemeral && o.App == "" {
			return cmdutil.UsageErrorf(cmd, "An app value must be supplied if ephemeral is used")
		}
	}
	if cmd.Flags().Lookup("directory") != nil {
		o.Directory = kcmdutil.GetFlagString(cmd, "directory")
	}
	if cmd.Flags().Lookup("nopods") != nil {
		o.NoPods = kcmdutil.GetFlagBool(cmd, "nopods")
	}
	return nil
}

func (o *CmdOptions) GatherInfo() error {
	var msg string
	var err error
	if msg, err = o.AuthOptions.GatherAuthInfo(); err != nil {
		return err
	}
	if o.Verbose {
		fmt.Fprintf(o.Out, msg)
	}
	if msg, err = o.AuthOptions.GatherProjectInfo(); err != nil {
		return err
	}
	if o.Verbose {
		fmt.Fprintf(o.Out, msg)
	}
	return nil
}
