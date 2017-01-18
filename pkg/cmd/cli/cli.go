package cli

import (
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/common"
)

const productName = `Oshinko`
const cliLong = productName + ` Client

The Oshinko client.
`
const cliShort = "Command line tools for managing spark clusters"
const cliExplain = `
To see the full list of commands supported, run '%[1]s help'.
`

func CLICommands(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) templates.CommandGroups {
	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				oshinkocmd.NewCmdGet(fullName, f, in, out),
				oshinkocmd.NewCmdDelete(fullName, f, in, out),
				oshinkocmd.NewCmdCreate(fullName, f, in, out),
				oshinkocmd.NewCmdScale(fullName, f, in, out),
			},
		},
	}
	return groups
}

func NewCommandCLI(name, fullName string, in io.Reader, out, errout io.Writer) *cobra.Command {
	return common.NewCommandCommon(name, fullName, cliShort, cliLong, cliExplain, in, out, errout, CLICommands)
}