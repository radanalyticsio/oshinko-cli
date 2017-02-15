package test

import (
	"errors"
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	policyregistry "github.com/openshift/origin/pkg/authorization/registry/policy"
	"github.com/openshift/origin/pkg/client"
)

type PolicyRegistry struct {
	// policies is a of namespace->name->Policy
	policies map[string]map[string]authorizationapi.Policy
	Err      error
}

func NewPolicyRegistry(policies []authorizationapi.Policy, err error) *PolicyRegistry {
	policyMap := make(map[string]map[string]authorizationapi.Policy)

	for _, policy := range policies {
		addPolicy(policyMap, policy)
	}

	return &PolicyRegistry{policyMap, err}
}

func (r *PolicyRegistry) Policies(namespace string) client.PolicyLister {
	return policyLister{registry: r, namespace: namespace}
}

type policyLister struct {
	registry  policyregistry.Registry
	namespace string
}

func (s policyLister) List(options kapi.ListOptions) (*authorizationapi.PolicyList, error) {
	return s.registry.ListPolicies(kapi.WithNamespace(kapi.NewContext(), s.namespace), &options)
}

func (s policyLister) Get(name string) (*authorizationapi.Policy, error) {
	return s.registry.GetPolicy(kapi.WithNamespace(kapi.NewContext(), s.namespace), name)
}

// ListPolicies obtains a list of policies that match a selector.
func (r *PolicyRegistry) ListPolicies(ctx kapi.Context, options *kapi.ListOptions) (*authorizationapi.PolicyList, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	list := make([]authorizationapi.Policy, 0)

	if namespace == kapi.NamespaceAll {
		for _, curr := range r.policies {
			for _, policy := range curr {
				list = append(list, policy)
			}
		}

	} else {
		if namespacedPolicies, ok := r.policies[namespace]; ok {
			for _, curr := range namespacedPolicies {
				list = append(list, curr)
			}
		}
	}

	return &authorizationapi.PolicyList{
			Items: list,
		},
		nil
}

// GetPolicy retrieves a specific policy.
func (r *PolicyRegistry) GetPolicy(ctx kapi.Context, id string) (*authorizationapi.Policy, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return nil, errors.New("invalid request.  Namespace parameter required.")
	}

	if namespacedPolicies, ok := r.policies[namespace]; ok {
		if policy, ok := namespacedPolicies[id]; ok {
			return &policy, nil
		}
	}

	return nil, fmt.Errorf("Policy %v::%v not found", namespace, id)
}

// CreatePolicy creates a new policy.
func (r *PolicyRegistry) CreatePolicy(ctx kapi.Context, policy *authorizationapi.Policy) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return errors.New("invalid request.  Namespace parameter required.")
	}
	if existing, _ := r.GetPolicy(ctx, policy.Name); existing != nil {
		return fmt.Errorf("Policy %v::%v already exists", namespace, policy.Name)
	}

	addPolicy(r.policies, *policy)

	return nil
}

// UpdatePolicy updates a policy.
func (r *PolicyRegistry) UpdatePolicy(ctx kapi.Context, policy *authorizationapi.Policy) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return errors.New("invalid request.  Namespace parameter required.")
	}
	if existing, _ := r.GetPolicy(ctx, policy.Name); existing == nil {
		return fmt.Errorf("Policy %v::%v not found", namespace, policy.Name)
	}

	addPolicy(r.policies, *policy)

	return nil
}

// DeletePolicy deletes a policy.
func (r *PolicyRegistry) DeletePolicy(ctx kapi.Context, id string) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return errors.New("invalid request.  Namespace parameter required.")
	}

	namespacedPolicies, ok := r.policies[namespace]
	if ok {
		delete(namespacedPolicies, id)
	}

	return nil
}

func (r *PolicyRegistry) WatchPolicies(ctx kapi.Context, options *kapi.ListOptions) (watch.Interface, error) {
	return nil, errors.New("unsupported action for test registry")
}

func addPolicy(policies map[string]map[string]authorizationapi.Policy, policy authorizationapi.Policy) {
	resourceVersion += 1
	policy.ResourceVersion = fmt.Sprintf("%d", resourceVersion)

	namespacedPolicies, ok := policies[policy.Namespace]
	if !ok {
		namespacedPolicies = make(map[string]authorizationapi.Policy)
		policies[policy.Namespace] = namespacedPolicies
	}

	namespacedPolicies[policy.Name] = policy
}
