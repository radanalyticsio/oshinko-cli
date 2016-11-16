package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	ocutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

type concept struct {
	Name         string
	Abbreviation string
	Description  string
}

var concepts = []concept{
	{
		"Clusters",
		"cluster",
		heredoc.Doc(`
      Cluster Resource refers to a spark-cluster.
    `),
	},
}

func writeConcept(w io.Writer, c concept) {
	fmt.Fprintf(w, "* %s", c.Name)
	if len(c.Abbreviation) > 0 {
		fmt.Fprintf(w, " [%s]", c.Abbreviation)
	}
	fmt.Fprintln(w, ":")
	for _, s := range strings.Split(c.Description, "\n") {
		fmt.Fprintf(w, "    %s\n", s)
	}
}

var (
	typesLong = heredoc.Doc(`
    Concepts and Types

    Oshinko developers and operators deploy spark clusters
    in a containerized cloud environment. Clusters may be composed
    of all of the components below, although most developers will be concerned with
    Clusters for delivering changes.

    Concepts:

    %[1]sFor more, see https://oshinko.openshift.com
  `)

	typesExample = `  # View all clusters you have access to
  %[1]s clusters

	`
)

func NewCmdTypes(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	buf := &bytes.Buffer{}
	for _, c := range concepts {
		writeConcept(buf, c)
	}
	cmd := &cobra.Command{
		Use:     "types",
		Short:   "An introduction to concepts and types",
		Long:    fmt.Sprintf(typesLong, buf.String()),
		Example: fmt.Sprintf(typesExample, fullName),
		Run:     ocutil.DefaultSubCommandRun(out),
	}
	return cmd
}
