package browsersafe

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

const (
	proxyAction = "proxy"
	unsafeProxy = "unsafeproxy"
)

type browserSafeAuthorizer struct {
	delegate authorizer.Authorizer

	// list of groups, any of which indicate the request is authenticated
	authenticatedGroups sets.String
}

func NewBrowserSafeAuthorizer(delegate authorizer.Authorizer, authenticatedGroups ...string) authorizer.Authorizer {
	return &browserSafeAuthorizer{
		delegate:            delegate,
		authenticatedGroups: sets.NewString(authenticatedGroups...),
	}
}

func (a *browserSafeAuthorizer) Authorize(attributes authorizer.Attributes) (authorizer.Decision, string, error) {
	browserSafeAttributes := a.getBrowserSafeAttributes(attributes)
	return a.delegate.Authorize(browserSafeAttributes)
}

func (a *browserSafeAuthorizer) getBrowserSafeAttributes(attributes authorizer.Attributes) authorizer.Attributes {
	if !attributes.IsResourceRequest() {
		return attributes
	}

	isProxyVerb := attributes.GetVerb() == proxyAction
	isProxySubresource := attributes.GetSubresource() == proxyAction

	if !isProxyVerb && !isProxySubresource {
		// Requests to non-proxy resources don't expose HTML or HTTP-handling user content to browsers
		return attributes
	}

	if user := attributes.GetUser(); user != nil {
		if a.authenticatedGroups.HasAny(user.GetGroups()...) {
			// An authenticated request indicates this isn't a browser page load.
			// Browsers cannot make direct authenticated requests.
			// This depends on the API not enabling basic or cookie-based auth.
			return attributes
		}
	}

	return &browserSafeAttributes{
		Attributes:         attributes,
		isProxyVerb:        isProxyVerb,
		isProxySubresource: isProxySubresource,
	}
}

type browserSafeAttributes struct {
	authorizer.Attributes

	isProxyVerb, isProxySubresource bool
}

func (b *browserSafeAttributes) GetVerb() string {
	if b.isProxyVerb {
		return unsafeProxy
	}
	return b.Attributes.GetVerb()
}

func (b *browserSafeAttributes) GetSubresource() string {
	if b.isProxySubresource {
		return unsafeProxy
	}
	return b.Attributes.GetSubresource()
}
