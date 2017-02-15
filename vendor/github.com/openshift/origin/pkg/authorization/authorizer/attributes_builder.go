package authorizer

import (
	"net/http"
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"
)

type openshiftAuthorizationAttributeBuilder struct {
	contextMapper kapi.RequestContextMapper
	infoResolver  RequestInfoResolver
}

func NewAuthorizationAttributeBuilder(contextMapper kapi.RequestContextMapper, infoResolver RequestInfoResolver) AuthorizationAttributeBuilder {
	return &openshiftAuthorizationAttributeBuilder{contextMapper, infoResolver}
}

func (a *openshiftAuthorizationAttributeBuilder) GetAttributes(req *http.Request) (Action, error) {
	requestInfo, err := a.infoResolver.GetRequestInfo(req)
	if err != nil {
		return nil, err
	}

	if !requestInfo.IsResourceRequest {
		return DefaultAuthorizationAttributes{
			Verb:           strings.ToLower(req.Method),
			NonResourceURL: true,
			URL:            requestInfo.Path,
		}, nil
	}

	resource := requestInfo.Resource
	if len(requestInfo.Subresource) > 0 {
		resource = requestInfo.Resource + "/" + requestInfo.Subresource
	}

	return DefaultAuthorizationAttributes{
		Verb:              requestInfo.Verb,
		APIGroup:          requestInfo.APIGroup,
		APIVersion:        requestInfo.APIVersion,
		Resource:          resource,
		ResourceName:      requestInfo.Name,
		RequestAttributes: req,
		NonResourceURL:    false,
		URL:               requestInfo.Path,
	}, nil
}
