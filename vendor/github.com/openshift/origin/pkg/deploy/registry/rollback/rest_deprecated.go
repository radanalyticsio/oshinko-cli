package rollback

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/openshift/origin/pkg/deploy/api/validation"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
)

// REST provides a rollback generation endpoint. Only the Create method is implemented.
type DeprecatedREST struct {
	generator GeneratorClient
	codec     runtime.Codec
}

// GeneratorClient defines a local interface to a rollback generator for testability.
type GeneratorClient interface {
	GenerateRollback(from, to *deployapi.DeploymentConfig, spec *deployapi.DeploymentConfigRollbackSpec) (*deployapi.DeploymentConfig, error)
	GetDeployment(ctx kapi.Context, name string) (*kapi.ReplicationController, error)
	GetDeploymentConfig(ctx kapi.Context, name string) (*deployapi.DeploymentConfig, error)
}

// Client provides an implementation of Generator client
type Client struct {
	GRFn func(from, to *deployapi.DeploymentConfig, spec *deployapi.DeploymentConfigRollbackSpec) (*deployapi.DeploymentConfig, error)
	RCFn func(ctx kapi.Context, name string) (*kapi.ReplicationController, error)
	DCFn func(ctx kapi.Context, name string) (*deployapi.DeploymentConfig, error)
}

// GetDeployment returns the deploymentConfig with the provided context and name
func (c Client) GetDeploymentConfig(ctx kapi.Context, name string) (*deployapi.DeploymentConfig, error) {
	return c.DCFn(ctx, name)
}

// GetDeployment returns the deployment with the provided context and name
func (c Client) GetDeployment(ctx kapi.Context, name string) (*kapi.ReplicationController, error) {
	return c.RCFn(ctx, name)
}

// GenerateRollback generates a new deploymentConfig by merging a pair of deploymentConfigs
func (c Client) GenerateRollback(from, to *deployapi.DeploymentConfig, spec *deployapi.DeploymentConfigRollbackSpec) (*deployapi.DeploymentConfig, error) {
	return c.GRFn(from, to, spec)
}

// NewDeprecatedREST safely creates a new REST.
func NewDeprecatedREST(generator GeneratorClient, codec runtime.Codec) *DeprecatedREST {
	return &DeprecatedREST{
		generator: generator,
		codec:     codec,
	}
}

// New creates an empty DeploymentConfigRollback resource
func (s *DeprecatedREST) New() runtime.Object {
	return &deployapi.DeploymentConfigRollback{}
}

// Create generates a new DeploymentConfig representing a rollback.
func (s *DeprecatedREST) Create(ctx kapi.Context, obj runtime.Object) (runtime.Object, error) {
	rollback, ok := obj.(*deployapi.DeploymentConfigRollback)
	if !ok {
		return nil, kerrors.NewBadRequest(fmt.Sprintf("not a rollback spec: %#v", obj))
	}

	if errs := validation.ValidateDeploymentConfigRollbackDeprecated(rollback); len(errs) > 0 {
		return nil, kerrors.NewInvalid(deployapi.Kind("DeploymentConfigRollback"), "", errs)
	}

	// Roll back "from" the current deployment "to" a target deployment

	// Find the target ("to") deployment and decode the DeploymentConfig
	targetDeployment, err := s.generator.GetDeployment(ctx, rollback.Spec.From.Name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, newInvalidDeploymentError(rollback, "Deployment not found")
		}
		return nil, newInvalidDeploymentError(rollback, fmt.Sprintf("%v", err))
	}

	to, err := deployutil.DecodeDeploymentConfig(targetDeployment, s.codec)
	if err != nil {
		return nil, newInvalidDeploymentError(rollback,
			fmt.Sprintf("couldn't decode DeploymentConfig from Deployment: %v", err))
	}

	// Find the current ("from") version of the target deploymentConfig
	from, err := s.generator.GetDeploymentConfig(ctx, to.Name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, newInvalidDeploymentError(rollback,
				fmt.Sprintf("couldn't find a current DeploymentConfig %s/%s", targetDeployment.Namespace, to.Name))
		}
		return nil, newInvalidDeploymentError(rollback,
			fmt.Sprintf("error finding current DeploymentConfig %s/%s: %v", targetDeployment.Namespace, to.Name, err))
	}

	return s.generator.GenerateRollback(from, to, &rollback.Spec)
}

func newInvalidDeploymentError(rollback *deployapi.DeploymentConfigRollback, reason string) error {
	err := field.Invalid(field.NewPath("spec").Child("from").Child("name"), rollback.Spec.From.Name, reason)
	return kerrors.NewInvalid(deployapi.Kind("DeploymentConfigRollback"), "", field.ErrorList{err})
}
