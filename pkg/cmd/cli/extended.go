// +build extended

package cli

import (
	"github.com/spf13/cobra"
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"io"

	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
)

func GetCommandGroups(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) templates.CommandGroups {
	return templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				oshinkocmd.NewCmdGet(fullName, f, in, out),
				oshinkocmd.NewCmdDelete(fullName, f, in, out),
				oshinkocmd.NewCmdCreate(fullName, f, in, out),
				oshinkocmd.NewCmdScale(fullName, f, in, out),
			},
		},
		{
			Message: "Extended Commands:",
			Commands: []*cobra.Command{
				oshinkocmd.NewCmdConfigMap(fullName, f, in, out),
			},
		},
	}
}
