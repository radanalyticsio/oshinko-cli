package admin

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/api/latest"
	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	"github.com/openshift/origin/pkg/template/api"
)

const (
	DefaultPolicyFile                    = "openshift.local.config/master/policy.json"
	CreateBootstrapPolicyFileCommand     = "create-bootstrap-policy-file"
	CreateBootstrapPolicyFileFullCommand = "openshift admin " + CreateBootstrapPolicyFileCommand
)

type CreateBootstrapPolicyFileOptions struct {
	File string

	OpenShiftSharedResourcesNamespace string
}

func NewCommandCreateBootstrapPolicyFile(commandName string, fullName string, out io.Writer) *cobra.Command {
	options := &CreateBootstrapPolicyFileOptions{}

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create the default bootstrap policy",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.CreateBootstrapPolicyFile(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&options.File, "filename", DefaultPolicyFile, "The policy template file that will be written with roles and bindings.")
	flags.StringVar(&options.OpenShiftSharedResourcesNamespace, "openshift-namespace", "openshift", "Namespace for shared resources.")

	// autocompletion hints
	cmd.MarkFlagFilename("filename")

	return cmd
}

func (o CreateBootstrapPolicyFileOptions) Validate(args []string) error {
	if len(args) != 0 {
		return errors.New("no arguments are supported")
	}
	if len(o.File) == 0 {
		return errors.New("filename must be provided")
	}
	if len(o.OpenShiftSharedResourcesNamespace) == 0 {
		return errors.New("openshift-namespace must be provided")
	}

	return nil
}

func (o CreateBootstrapPolicyFileOptions) CreateBootstrapPolicyFile() error {
	if err := os.MkdirAll(path.Dir(o.File), os.FileMode(0755)); err != nil {
		return err
	}

	policyTemplate := &api.Template{}

	clusterRoles := bootstrappolicy.GetBootstrapClusterRoles()
	for i := range clusterRoles {
		versionedObject, err := kapi.Scheme.ConvertToVersion(&clusterRoles[i], latest.Version)
		if err != nil {
			return err
		}
		policyTemplate.Objects = append(policyTemplate.Objects, versionedObject)
	}

	clusterRoleBindings := bootstrappolicy.GetBootstrapClusterRoleBindings()
	for i := range clusterRoleBindings {
		versionedObject, err := kapi.Scheme.ConvertToVersion(&clusterRoleBindings[i], latest.Version)
		if err != nil {
			return err
		}
		policyTemplate.Objects = append(policyTemplate.Objects, versionedObject)
	}

	openshiftRoles := bootstrappolicy.GetBootstrapOpenshiftRoles(o.OpenShiftSharedResourcesNamespace)
	for i := range openshiftRoles {
		versionedObject, err := kapi.Scheme.ConvertToVersion(&openshiftRoles[i], latest.Version)
		if err != nil {
			return err
		}
		policyTemplate.Objects = append(policyTemplate.Objects, versionedObject)
	}

	openshiftRoleBindings := bootstrappolicy.GetBootstrapOpenshiftRoleBindings(o.OpenShiftSharedResourcesNamespace)
	for i := range openshiftRoleBindings {
		versionedObject, err := kapi.Scheme.ConvertToVersion(&openshiftRoleBindings[i], latest.Version)
		if err != nil {
			return err
		}
		policyTemplate.Objects = append(policyTemplate.Objects, versionedObject)
	}

	versionedPolicyTemplate, err := kapi.Scheme.ConvertToVersion(policyTemplate, latest.Version)
	if err != nil {
		return err
	}

	buffer := &bytes.Buffer{}
	(&kubectl.JSONPrinter{}).PrintObj(versionedPolicyTemplate, buffer)

	if err := ioutil.WriteFile(o.File, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
