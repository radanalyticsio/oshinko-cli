package clusterrole

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// Registry is an interface for things that know how to store ClusterRoles.
type Registry interface {
	// ListClusterRoles obtains list of policyClusterRoles that match a selector.
	ListClusterRoles(ctx kapi.Context, options *kapi.ListOptions) (*authorizationapi.ClusterRoleList, error)
	// GetClusterRole retrieves a specific policyClusterRole.
	GetClusterRole(ctx kapi.Context, id string) (*authorizationapi.ClusterRole, error)
	// CreateClusterRole creates a new policyClusterRole.
	CreateClusterRole(ctx kapi.Context, policyClusterRole *authorizationapi.ClusterRole) (*authorizationapi.ClusterRole, error)
	// UpdateClusterRole updates a policyClusterRole.
	UpdateClusterRole(ctx kapi.Context, policyClusterRole *authorizationapi.ClusterRole) (*authorizationapi.ClusterRole, bool, error)
	// DeleteClusterRole deletes a policyClusterRole.
	DeleteClusterRole(ctx kapi.Context, id string) error
}

// Storage is an interface for a standard REST Storage backend
type Storage interface {
	rest.Getter
	rest.Lister
	rest.CreaterUpdater
	rest.GracefulDeleter

	// CreateRoleWithEscalation creates a new policyRole.  Skipping the escalation check should only be done during bootstrapping procedures where no users are currently bound.
	CreateRoleWithEscalation(ctx kapi.Context, policyRole *authorizationapi.Role) (*authorizationapi.Role, error)
	// UpdateRoleWithEscalation updates a policyRole.  Skipping the escalation check should only be done during bootstrapping procedures where no users are currently bound.
	UpdateRoleWithEscalation(ctx kapi.Context, policyRole *authorizationapi.Role) (*authorizationapi.Role, bool, error)
}

// storage puts strong typing around storage calls
type storage struct {
	Storage
}

// NewRegistry returns a new Registry interface for the given Storage. Any mismatched
// types will panic.
func NewRegistry(s Storage) Registry {
	return &storage{s}
}

func (s *storage) ListClusterRoles(ctx kapi.Context, options *kapi.ListOptions) (*authorizationapi.ClusterRoleList, error) {
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}

	return obj.(*authorizationapi.ClusterRoleList), nil
}

func (s *storage) CreateClusterRole(ctx kapi.Context, node *authorizationapi.ClusterRole) (*authorizationapi.ClusterRole, error) {
	obj, err := s.Create(ctx, node)
	if err != nil {
		return nil, err
	}

	return obj.(*authorizationapi.ClusterRole), err
}

func (s *storage) UpdateClusterRole(ctx kapi.Context, node *authorizationapi.ClusterRole) (*authorizationapi.ClusterRole, bool, error) {
	obj, created, err := s.Update(ctx, node.Name, rest.DefaultUpdatedObjectInfo(node, kapi.Scheme))
	if err != nil {
		return nil, created, err
	}
	return obj.(*authorizationapi.ClusterRole), created, err
}

func (s *storage) GetClusterRole(ctx kapi.Context, name string) (*authorizationapi.ClusterRole, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return obj.(*authorizationapi.ClusterRole), nil
}

func (s *storage) DeleteClusterRole(ctx kapi.Context, name string) error {
	_, err := s.Delete(ctx, name, nil)
	return err
}
