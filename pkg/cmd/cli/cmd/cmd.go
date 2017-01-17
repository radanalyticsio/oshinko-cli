package cli

import (
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/common"
	"github.com/spf13/cobra"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type CmdOptions struct {

	Image           string `json:"image"`
	WorkerCount     int    `json:"workerCount"`
	MasterCount     int    `json:"masterCount,omitempty"`
	MasterConfig    string `json:"sparkMasterConfig,omitempty"`
	WorkerConfig    string `json:"workerConfig,omitempty"`
	StoredConfig    string `json:"storedConfig,omitempty"`

	common.CommonOptions

}

func (o *CmdOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string) error {

	if err := o.CommonOptions.Complete(f, cmd, args); err != nil {
		kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
	}

	o.Image = defaultImage

	// Pod counts will be assigned by the default config in oshinko-core,
	// here values should be defaulted to 0
	o.WorkerCount = 0
	o.MasterCount = 0

	if cmd.Flags().Lookup("workers") != nil {
		o.WorkerCount = kcmdutil.GetFlagInt(cmd, "workers")
	}
	if cmd.Flags().Lookup("masters") != nil {
		o.MasterCount = kcmdutil.GetFlagInt(cmd, "masters")
	}
	if cmd.Flags().Lookup("image") != nil {
		o.Image = kcmdutil.GetFlagString(cmd, "image")
	}
        if cmd.Flags().Lookup("storedconfig") != nil {
		o.StoredConfig = kcmdutil.GetFlagString(cmd, "storedconfig")
	}
	return nil
}