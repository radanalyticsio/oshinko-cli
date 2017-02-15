package test

import (
	"errors"
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/watch"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	policybindingregistry "github.com/openshift/origin/pkg/authorization/registry/policybinding"
	"github.com/openshift/origin/pkg/client"
)

type PolicyBindingRegistry struct {
	// policyBindings is a of namespace->name->PolicyBinding
	policyBindings map[string]map[string]authorizationapi.PolicyBinding
	Err            error
}

func NewPolicyBindingRegistry(bindings []authorizationapi.PolicyBinding, err error) *PolicyBindingRegistry {
	bindingMap := make(map[string]map[string]authorizationapi.PolicyBinding)

	for _, binding := range bindings {
		addPolicyBinding(bindingMap, binding)
	}

	return &PolicyBindingRegistry{bindingMap, err}
}

func (r *PolicyBindingRegistry) PolicyBindings(namespace string) client.PolicyBindingLister {
	return policyBindingLister{registry: r, namespace: namespace}
}

type policyBindingLister struct {
	registry  policybindingregistry.Registry
	namespace string
}

func (s policyBindingLister) List(options kapi.ListOptions) (*authorizationapi.PolicyBindingList, error) {
	return s.registry.ListPolicyBindings(kapi.WithNamespace(kapi.NewContext(), s.namespace), &options)
}

func (s policyBindingLister) Get(name string) (*authorizationapi.PolicyBinding, error) {
	return s.registry.GetPolicyBinding(kapi.WithNamespace(kapi.NewContext(), s.namespace), name)
}

// ListPolicyBindings obtains a list of policyBinding that match a selector.
func (r *PolicyBindingRegistry) ListPolicyBindings(ctx kapi.Context, options *kapi.ListOptions) (*authorizationapi.PolicyBindingList, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	list := make([]authorizationapi.PolicyBinding, 0)

	if namespace == kapi.NamespaceAll {
		for _, curr := range r.policyBindings {
			for _, binding := range curr {
				list = append(list, binding)
			}
		}

	} else {
		if namespacedBindings, ok := r.policyBindings[namespace]; ok {
			for _, curr := range namespacedBindings {
				list = append(list, curr)
			}
		}
	}

	return &authorizationapi.PolicyBindingList{
			Items: list,
		},
		nil
}

// GetPolicyBinding retrieves a specific policyBinding.
func (r *PolicyBindingRegistry) GetPolicyBinding(ctx kapi.Context, id string) (*authorizationapi.PolicyBinding, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return nil, errors.New("invalid request.  Namespace parameter required.")
	}

	if namespacedBindings, ok := r.policyBindings[namespace]; ok {
		if binding, ok := namespacedBindings[id]; ok {
			return &binding, nil
		}
	}

	return nil, kapierrors.NewNotFound(authorizationapi.Resource("policybinding"), id)
}

// CreatePolicyBinding creates a new policyBinding.
func (r *PolicyBindingRegistry) CreatePolicyBinding(ctx kapi.Context, policyBinding *authorizationapi.PolicyBinding) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return errors.New("invalid request.  Namespace parameter required.")
	}
	if existing, _ := r.GetPolicyBinding(ctx, policyBinding.Name); existing != nil {
		return fmt.Errorf("PolicyBinding %v::%v already exists", namespace, policyBinding.Name)
	}

	addPolicyBinding(r.policyBindings, *policyBinding)

	return nil
}

// UpdatePolicyBinding updates a policyBinding.
func (r *PolicyBindingRegistry) UpdatePolicyBinding(ctx kapi.Context, policyBinding *authorizationapi.PolicyBinding) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return errors.New("invalid request.  Namespace parameter required.")
	}
	if existing, _ := r.GetPolicyBinding(ctx, policyBinding.Name); existing == nil {
		return kapierrors.NewNotFound(authorizationapi.Resource("policybinding"), policyBinding.Name)
	}

	addPolicyBinding(r.policyBindings, *policyBinding)

	return nil
}

// DeletePolicyBinding deletes a policyBinding.
func (r *PolicyBindingRegistry) DeletePolicyBinding(ctx kapi.Context, id string) error {
	if r.Err != nil {
		return r.Err
	}

	namespace := kapi.NamespaceValue(ctx)
	if len(namespace) == 0 {
		return errors.New("invalid request.  Namespace parameter required.")
	}

	namespacedBindings, ok := r.policyBindings[namespace]
	if ok {
		delete(namespacedBindings, id)
	}

	return nil
}

func (r *PolicyBindingRegistry) WatchPolicyBindings(ctx kapi.Context, options *kapi.ListOptions) (watch.Interface, error) {
	return nil, errors.New("unsupported action for test registry")
}

func addPolicyBinding(bindings map[string]map[string]authorizationapi.PolicyBinding, binding authorizationapi.PolicyBinding) {
	resourceVersion += 1
	binding.ResourceVersion = fmt.Sprintf("%d", resourceVersion)

	namespacedBindings, ok := bindings[binding.Namespace]
	if !ok {
		namespacedBindings = make(map[string]authorizationapi.PolicyBinding)
		bindings[binding.Namespace] = namespacedBindings
	}

	namespacedBindings[binding.Name] = binding
}
