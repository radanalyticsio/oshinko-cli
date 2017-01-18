// +build standard

package cli

import (
	"github.com/spf13/cobra"
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"io"

	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
)

func GetCommandGroups(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) (
											templates.CommandGroups,
											*cobra.Command) {
	first := oshinkocmd.NewCmdGet(fullName, f, in, out)
	return templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				first,
				oshinkocmd.NewCmdDelete(fullName, f, in, out),
				oshinkocmd.NewCmdCreate(fullName, f, in, out),
				oshinkocmd.NewCmdScale(fullName, f, in, out),
			},
		},
	}, first
}
