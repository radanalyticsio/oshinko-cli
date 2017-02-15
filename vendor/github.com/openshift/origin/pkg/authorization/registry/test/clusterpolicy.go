package test

import (
	"errors"
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/watch"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

var resourceVersion = 1

type ClusterPolicyRegistry struct {
	// ClusterPolicies is a of namespace->name->ClusterPolicy
	clusterPolicies map[string]map[string]authorizationapi.ClusterPolicy
	Err             error
}

func NewClusterPolicyRegistry(policies []authorizationapi.ClusterPolicy, err error) *ClusterPolicyRegistry {
	policyMap := make(map[string]map[string]authorizationapi.ClusterPolicy)

	for _, policy := range policies {
		addClusterPolicy(policyMap, policy)
	}

	return &ClusterPolicyRegistry{policyMap, err}
}

func (r *ClusterPolicyRegistry) List(options kapi.ListOptions) (*authorizationapi.ClusterPolicyList, error) {
	return r.ListClusterPolicies(kapi.NewContext(), &options)
}
func (r *ClusterPolicyRegistry) Get(name string) (*authorizationapi.ClusterPolicy, error) {
	return r.GetClusterPolicy(kapi.NewContext(), name)
}

// ListClusterPolicies obtains list of ListClusterPolicy that match a selector.
func (r *ClusterPolicyRegistry) ListClusterPolicies(ctx kapi.Context, options *kapi.ListOptions) (*authorizationapi.ClusterPolicyList, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	list := make([]authorizationapi.ClusterPolicy, 0)

	if namespace == kapi.NamespaceAll {
		for _, curr := range r.clusterPolicies {
			for _, policy := range curr {
				list = append(list, policy)
			}
		}

	} else {
		if namespacedClusterPolicies, ok := r.clusterPolicies[namespace]; ok {
			for _, curr := range namespacedClusterPolicies {
				list = append(list, curr)
			}
		}
	}

	return &authorizationapi.ClusterPolicyList{
			Items: list,
		},
		nil
}

// GetClusterPolicy retrieves a specific policy.
func (r *ClusterPolicyRegistry) GetClusterPolicy(ctx kapi.Context, id string) (*authorizationapi.ClusterPolicy, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) != 0 {
		return nil, errors.New("invalid request.  Namespace parameter disallowed.")
	}

	if namespacedClusterPolicies, ok := r.clusterPolicies[namespace]; ok {
		if policy, ok := namespacedClusterPolicies[id]; ok {
			return &policy, nil
		}
	}

	return nil, kapierrors.NewNotFound(authorizationapi.Resource("clusterpolicy"), id)
}

// CreateClusterPolicy creates a new policy.
func (r *ClusterPolicyRegistry) CreateClusterPolicy(ctx kapi.Context, policy *authorizationapi.ClusterPolicy) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) != 0 {
		return errors.New("invalid request.  Namespace parameter disallowed.")
	}
	if existing, _ := r.GetClusterPolicy(ctx, policy.Name); existing != nil {
		return kapierrors.NewAlreadyExists(authorizationapi.Resource("ClusterPolicy"), policy.Name)
	}

	addClusterPolicy(r.clusterPolicies, *policy)

	return nil
}

// UpdateClusterPolicy updates a policy.
func (r *ClusterPolicyRegistry) UpdateClusterPolicy(ctx kapi.Context, policy *authorizationapi.ClusterPolicy) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) != 0 {
		return errors.New("invalid request.  Namespace parameter disallowed.")
	}
	if existing, _ := r.GetClusterPolicy(ctx, policy.Name); existing == nil {
		return kapierrors.NewNotFound(authorizationapi.Resource("clusterpolicy"), policy.Name)
	}

	addClusterPolicy(r.clusterPolicies, *policy)

	return nil
}

// DeleteClusterPolicy deletes a policy.
func (r *ClusterPolicyRegistry) DeleteClusterPolicy(ctx kapi.Context, id string) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) != 0 {
		return errors.New("invalid request.  Namespace parameter disallowed.")
	}

	namespacedClusterPolicies, ok := r.clusterPolicies[namespace]
	if ok {
		delete(namespacedClusterPolicies, id)
	}

	return nil
}

func (r *ClusterPolicyRegistry) WatchClusterPolicies(ctx kapi.Context, options *kapi.ListOptions) (watch.Interface, error) {
	return nil, errors.New("unsupported action for test registry")
}

func addClusterPolicy(policies map[string]map[string]authorizationapi.ClusterPolicy, policy authorizationapi.ClusterPolicy) {
	resourceVersion += 1
	policy.ResourceVersion = fmt.Sprintf("%d", resourceVersion)

	namespacedClusterPolicies, ok := policies[policy.Namespace]
	if !ok {
		namespacedClusterPolicies = make(map[string]authorizationapi.ClusterPolicy)
		policies[policy.Namespace] = namespacedClusterPolicies
	}

	namespacedClusterPolicies[policy.Name] = policy
}
