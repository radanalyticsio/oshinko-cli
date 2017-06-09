// +build extended

package cli

import (
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"

	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
)

func GetCommandGroups(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) (
	templates.CommandGroups,
	*cobra.Command) {
	first := oshinkocmd.NewCmdGetExtended(fullName, f, in, out)
	return templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				first,
				oshinkocmd.NewCmdDeleteExtended(fullName, f, in, out),
				oshinkocmd.NewCmdCreateExtended(fullName, f, in, out),
				oshinkocmd.NewCmdScale(fullName, f, in, out),
			},
		},
		{
			Message: "Extended Commands:",
			Commands: []*cobra.Command{
				oshinkocmd.NewCmdConfigMap(fullName, f, in, out),
			},
		},
	}, first
}
