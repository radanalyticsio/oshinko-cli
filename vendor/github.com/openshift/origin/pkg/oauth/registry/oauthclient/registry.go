package oauthclient

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"

	"github.com/openshift/origin/pkg/oauth/api"
)

// Registry is an interface for things that know how to store OAuthClient objects.
type Registry interface {
	// ListClients obtains a list of clients that match a selector.
	ListClients(ctx kapi.Context, options *kapi.ListOptions) (*api.OAuthClientList, error)
	// GetClient retrieves a specific client.
	GetClient(ctx kapi.Context, name string) (*api.OAuthClient, error)
	// CreateClient creates a new client.
	CreateClient(ctx kapi.Context, client *api.OAuthClient) (*api.OAuthClient, error)
	// UpdateClient updates a client.
	UpdateClient(ctx kapi.Context, client *api.OAuthClient) (*api.OAuthClient, error)
	// DeleteClient deletes a client.
	DeleteClient(ctx kapi.Context, name string) error
}

// Getter exposes a way to get a specific client.  This is useful for other registries to get scope limitations
// on particular clients.   This interface will make its easier to write a future cache on it
type Getter interface {
	GetClient(ctx kapi.Context, name string) (*api.OAuthClient, error)
}

// storage puts strong typing around storage calls
type storage struct {
	rest.StandardStorage
}

// NewRegistry returns a new Registry interface for the given Storage. Any mismatched
// types will panic.
func NewRegistry(s rest.StandardStorage) Registry {
	return &storage{s}
}

func (s *storage) ListClients(ctx kapi.Context, options *kapi.ListOptions) (*api.OAuthClientList, error) {
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return obj.(*api.OAuthClientList), nil
}

func (s *storage) GetClient(ctx kapi.Context, name string) (*api.OAuthClient, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return obj.(*api.OAuthClient), nil
}

func (s *storage) CreateClient(ctx kapi.Context, client *api.OAuthClient) (*api.OAuthClient, error) {
	obj, err := s.Create(ctx, client)
	if err != nil {
		return nil, err
	}
	return obj.(*api.OAuthClient), nil
}

func (s *storage) UpdateClient(ctx kapi.Context, client *api.OAuthClient) (*api.OAuthClient, error) {
	obj, _, err := s.Update(ctx, client.Name, rest.DefaultUpdatedObjectInfo(client, kapi.Scheme))
	if err != nil {
		return nil, err
	}
	return obj.(*api.OAuthClient), nil
}

func (s *storage) DeleteClient(ctx kapi.Context, name string) error {
	_, err := s.Delete(ctx, name, nil)
	if err != nil {
		return err
	}
	return nil
}
