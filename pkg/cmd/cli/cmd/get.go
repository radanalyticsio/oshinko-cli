package cmd

import (
	"fmt"
	"io"

	//"github.com/renstrom/dedent"
	"github.com/spf13/cobra"
	//"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	//"k8s.io/kubernetes/pkg/kubectl/resource"
	//"k8s.io/kubernetes/pkg/runtime"
	//utilerrors "k8s.io/kubernetes/pkg/util/errors"
	//"k8s.io/kubernetes/pkg/watch"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

const (
	getLong = `Display one or many resources

Possible resources include builds, buildConfigs, services, pods, etc.
Some resources may omit advanced details that you can see with '-o wide'.
If you want an even more detailed view, use '%[1]s describe'.`

	getExample = `  # List all pods in ps output format.
  %[1]s get pods

  # List a single replication controller with specified ID in ps output format.
  %[1]s get rc redis

  # List all pods and show more details about them.
  %[1]s get -o wide pods

  # List a single pod in JSON output format.
  %[1]s get -o json pod redis-pod

  # Return only the status value of the specified pod.
  %[1]s get -o template pod redis-pod --template={{.currentState.status}}`

	valid_resources = `Valid resource types include
   * clusters (aka 'c')`
)

// GetOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type GetOptions struct {
	Filenames []string
	Recursive bool
}

// NewCmdGet is a wrapper for the Kubernetes cli get command
func NewCmdGet(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	cmd := CmdGet(f.Factory, out)
	cmd.Long = fmt.Sprintf(getLong, fullName)
	cmd.Example = fmt.Sprintf(getExample, fullName)
	cmd.SuggestFor = []string{"list"}
	return cmd
}

// NewCmdGet creates a command object for the generic "get" action, which
// retrieves one or more resources from a server.
func CmdGet(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	options := &GetOptions{}

	// retrieve a list of handled resources from printer as valid args
	validArgs := []string{}
	p, err := f.Printer(nil, false, false, false, false, false, false, []string{})
	cmdutil.CheckErr(err)
	if p != nil {
		validArgs = p.HandledResources()
	}

	cmd := &cobra.Command{
		Use:     "get [(-o|--output=)json|yaml|wide|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...] (TYPE [NAME | -l label] | TYPE/NAME ...) [flags]",
		Short:   "Display one or many resources",
		Long:    getLong,
		Example: getExample,
		Run: func(cmd *cobra.Command, args []string) {
			//err := RunGet(f, out, cmd, args, options)
			//cmdutil.CheckErr(err)
		},
		SuggestFor: []string{"list", "ps"},
		ValidArgs:  validArgs,
	}
	cmdutil.AddPrinterFlags(cmd)
	cmd.Flags().StringP("selector", "l", "", "Selector (label query) to filter on")
	usage := "Filename, directory, or URL to a file identifying the resource to get from a server."
	kubectl.AddJsonFilenameFlag(cmd, &options.Filenames, usage)
	cmdutil.AddRecursiveFlag(cmd, &options.Recursive)
	cmdutil.AddInclude3rdPartyFlags(cmd)
	return cmd
}

////// RunGet implements the generic Get command
////// TODO: convert all direct flag accessors to a struct and pass that instead of cmd
////func RunGet(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command, args []string, options *GetOptions) error {
////	selector := cmdutil.GetFlagString(cmd, "selector")
////	allNamespaces := cmdutil.GetFlagBool(cmd, "all-namespaces")
////	allKinds := cmdutil.GetFlagBool(cmd, "show-kind")
////	mapper, typer := f.Object(cmdutil.GetIncludeThirdPartyAPIs(cmd))
////
////	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
////	if err != nil {
////		return err
////	}
////
////	if allNamespaces {
////		enforceNamespace = false
////	}
////
////	if len(args) == 0 && len(options.Filenames) == 0 {
////		fmt.Fprint(out, "You must specify the type of resource to get. ", valid_resources)
////		return cmdutil.UsageError(cmd, "Required resource not specified.")
////	}
////
////	// always show resources when getting by name or filename
////	argsHasNames, err := resource.HasNames(args)
////	if err != nil {
////		return err
////	}
////	if len(options.Filenames) > 0 || argsHasNames {
////		cmd.Flag("show-all").Value.Set("true")
////	}
////	export := cmdutil.GetFlagBool(cmd, "export")
////
////	// handle watch separately since we cannot watch multiple resource types
////	//isWatch, isWatchOnly := cmdutil.GetFlagBool(cmd, "watch"), cmdutil.GetFlagBool(cmd, "watch-only")
////	//if isWatch || isWatchOnly {
////	//	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
////	//	NamespaceParam(cmdNamespace).DefaultNamespace().AllNamespaces(allNamespaces).
////	//	FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
////	//	SelectorParam(selector).
////	//	ExportParam(export).
////	//	ResourceTypeOrNameArgs(true, args...).
////	//	SingleResourceType().
////	//	Latest().
////	//	Do()
////	//	err := r.Err()
////	//	if err != nil {
////	//		return err
////	//	}
////	//	infos, err := r.Infos()
////	//	if err != nil {
////	//		return err
////	//	}
////	//	if len(infos) != 1 {
////	//		return fmt.Errorf("watch is only supported on individual resources and resource collections - %d resources were found", len(infos))
////	//	}
////	//	info := infos[0]
////	//	mapping := info.ResourceMapping()
////	//	printer, err := f.PrinterForMapping(cmd, mapping, allNamespaces)
////	//	if err != nil {
////	//		return err
////	//	}
////	//
////	//	obj, err := r.Object()
////	//	if err != nil {
////	//		return err
////	//	}
////	//	rv, err := mapping.MetadataAccessor.ResourceVersion(obj)
////	//	if err != nil {
////	//		return err
////	//	}
////	//
////	//	// print the current object
////	//	if !isWatchOnly {
////	//		if err := printer.PrintObj(obj, out); err != nil {
////	//			return fmt.Errorf("unable to output the provided object: %v", err)
////	//		}
////	//	}
////	//
////	//	// print watched changes
////	//	w, err := r.Watch(rv)
////	//	if err != nil {
////	//		return err
////	//	}
////	//
////	//	kubectl.WatchLoop(w, func(e watch.Event) error {
////	//		return printer.PrintObj(e.Object, out)
////	//	})
////	//	return nil
////	//}
////
////	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
////	NamespaceParam(cmdNamespace).DefaultNamespace().AllNamespaces(allNamespaces).
////	FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
////	SelectorParam(selector).
////	ExportParam(export).
////	ResourceTypeOrNameArgs(true, args...).
////	ContinueOnError().
////	Latest().
////	Flatten().
////	Do()
////	err = r.Err()
////	if err != nil {
////		return err
////	}
////
////	printer, generic, err := cmdutil.PrinterForCommand(cmd)
////	if err != nil {
////		return err
////	}
////
////	if generic {
////		clientConfig, err := f.ClientConfig()
////		if err != nil {
////			return err
////		}
////
////		allErrs := []error{}
////		singular := false
////		infos, err := r.IntoSingular(&singular).Infos()
////		if err != nil {
////			if singular {
////				return err
////			}
////			allErrs = append(allErrs, err)
////		}
////
////		// the outermost object will be converted to the output-version, but inner
////		// objects can use their mappings
////		version, err := cmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
////		if err != nil {
////			return err
////		}
////
////		obj, err := resource.AsVersionedObject(infos, !singular, version, f.JSONEncoder())
////		if err != nil {
////			return err
////		}
////
////		if err := printer.PrintObj(obj, out); err != nil {
////			allErrs = append(allErrs, err)
////		}
////		return utilerrors.NewAggregate(allErrs)
////	}
////
////	allErrs := []error{}
////	infos, err := r.Infos()
////	if err != nil {
////		allErrs = append(allErrs, err)
////	}
////
////	objs := make([]runtime.Object, len(infos))
////	for ix := range infos {
////		objs[ix] = infos[ix].Object
////	}
////
////	sorting, err := cmd.Flags().GetString("sort-by")
////	var sorter *kubectl.RuntimeSort
////	if err == nil && len(sorting) > 0 && len(objs) > 1 {
////		clientConfig, err := f.ClientConfig()
////		if err != nil {
////			return err
////		}
////
////		version, err := cmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
////		if err != nil {
////			return err
////		}
////
////		for ix := range infos {
////			objs[ix], err = infos[ix].Mapping.ConvertToVersion(infos[ix].Object, version)
////			if err != nil {
////				allErrs = append(allErrs, err)
////				continue
////			}
////		}
////
////		// TODO: questionable
////		if sorter, err = kubectl.SortObjects(f.Decoder(true), objs, sorting); err != nil {
////			return err
////		}
////	}
////
////	// use the default printer for each object
////	printer = nil
////	var lastMapping *meta.RESTMapping
////	var withKind bool = allKinds
////	w := kubectl.GetNewTabWriter(out)
////	defer w.Flush()
////
////	// determine if printing multiple kinds of
////	// objects and enforce "show-kinds" flag if so
////	for ix := range objs {
////		var mapping *meta.RESTMapping
////		if sorter != nil {
////			mapping = infos[sorter.OriginalPosition(ix)].Mapping
////		} else {
////			mapping = infos[ix].Mapping
////		}
////
////		// display "kind" column only if we have mixed resources
////		if lastMapping != nil && mapping.Resource != lastMapping.Resource {
////			withKind = true
////		}
////		lastMapping = mapping
////	}
////
////	lastMapping = nil
////
////	for ix := range objs {
////		var mapping *meta.RESTMapping
////		var original runtime.Object
////		if sorter != nil {
////			mapping = infos[sorter.OriginalPosition(ix)].Mapping
////			original = infos[sorter.OriginalPosition(ix)].Object
////		} else {
////			mapping = infos[ix].Mapping
////			original = infos[ix].Object
////		}
////		if printer == nil || lastMapping == nil || mapping == nil || mapping.Resource != lastMapping.Resource {
////			printer, err = f.PrinterForMapping(cmd, mapping, allNamespaces)
////			if err != nil {
////				allErrs = append(allErrs, err)
////				continue
////			}
////			lastMapping = mapping
////		}
////		if resourcePrinter, found := printer.(*kubectl.HumanReadablePrinter); found {
////			resourceName := mapping.Resource
////			if alias, ok := kubectl.ResourceShortFormFor(mapping.Resource); ok {
////				resourceName = alias
////			} else if resourceName == "" {
////				resourceName = "none"
////			}
////
////			resourcePrinter.Options.WithKind = withKind
////			resourcePrinter.Options.KindName = resourceName
////			if err := printer.PrintObj(original, w); err != nil {
////				allErrs = append(allErrs, err)
////			}
////			continue
////		}
////		if err := printer.PrintObj(original, w); err != nil {
////			allErrs = append(allErrs, err)
////			continue
////		}
////	}
////	return utilerrors.NewAggregate(allErrs)
//}
