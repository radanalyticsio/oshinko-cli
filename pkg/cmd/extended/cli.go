package extended

import (
	"io"
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/extended/cmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/common"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli"
)

const productName = `Oshinko`
const cliLong = productName + ` Extended Client

The extended Oshinko utility client.
`
const cliShort = "Command line tools for managing spark clusters"
const cliExplain = `
To see the full list of commands supported, run '%[1]s help'.
`
func ExtendedCommands(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) templates.CommandGroups {
	groups := cli.CLICommands(fullName, f, in, out)
	groups = append(groups,
		templates.CommandGroup{
			Message: "Extended Commands:",
			Commands: []*cobra.Command{
				extended.NewCmdConfigMap(fullName, f, in, out),
			},
		},
	)
	return groups
}

func NewCommandExtended(name, fullName string, in io.Reader, out, errout io.Writer) *cobra.Command {
	return common.NewCommandCommon(name, fullName, cliShort, cliLong, cliExplain, in, out, errout, ExtendedCommands)
}