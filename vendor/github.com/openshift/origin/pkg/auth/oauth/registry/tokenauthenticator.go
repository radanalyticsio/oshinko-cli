package registry

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kauthenticator "k8s.io/apiserver/pkg/authentication/authenticator"
	kuser "k8s.io/apiserver/pkg/authentication/user"

	"github.com/openshift/origin/pkg/auth/authenticator"
	"github.com/openshift/origin/pkg/auth/userregistry/identitymapper"
	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	oauthclient "github.com/openshift/origin/pkg/oauth/generated/internalclientset/typed/oauth/internalversion"
	userclient "github.com/openshift/origin/pkg/user/generated/internalclientset/typed/user/internalversion"
)

type tokenAuthenticator struct {
	tokens      oauthclient.OAuthAccessTokenInterface
	users       userclient.UserResourceInterface
	groupMapper identitymapper.UserToGroupMapper
	validators  authenticator.OAuthTokenValidator
}

func NewTokenAuthenticator(tokens oauthclient.OAuthAccessTokenInterface, users userclient.UserResourceInterface, groupMapper identitymapper.UserToGroupMapper, validators ...authenticator.OAuthTokenValidator) kauthenticator.Token {
	return &tokenAuthenticator{
		tokens:      tokens,
		users:       users,
		groupMapper: groupMapper,
		validators:  authenticator.OAuthTokenValidators(validators),
	}
}

func (a *tokenAuthenticator) AuthenticateToken(name string) (kuser.Info, bool, error) {
	token, err := a.tokens.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	user, err := a.users.Get(token.UserName, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	if err := a.validators.Validate(token, user); err != nil {
		return nil, false, err
	}

	groups, err := a.groupMapper.GroupsFor(user.Name)
	if err != nil {
		return nil, false, err
	}
	groupNames := make([]string, 0, len(groups)+len(user.Groups))
	for _, group := range groups {
		groupNames = append(groupNames, group.Name)
	}
	groupNames = append(groupNames, user.Groups...)

	return &kuser.DefaultInfo{
		Name:   user.Name,
		UID:    string(user.UID),
		Groups: groupNames,
		Extra: map[string][]string{
			authorizationapi.ScopesKey: token.Scopes,
		},
	}, true, nil
}
