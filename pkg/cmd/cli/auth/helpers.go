package auth

import (
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"

	userapi "github.com/openshift/api/user/v1"
	userclient "github.com/openshift/client-go/user/clientset/versioned"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	projectclient "github.com/openshift/origin/pkg/project/generated/internalclientset/typed/project/internalversion"
)


func whoAmI(clientConfig *rest.Config) (*userapi.User, error) {
	client, err := userclient.NewForConfig(clientConfig)

	me, err := client.UserV1().Users().Get("~", metav1.GetOptions{})

	// if we're talking to kube (or likely talking to kube),
	if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
		switch {
		case len(clientConfig.BearerToken) > 0:
			// the user has already been willing to provide the token on the CLI, so they probably
			// don't mind using it again if they switch to and from this user
			return &userapi.User{ObjectMeta: metav1.ObjectMeta{Name: clientConfig.BearerToken}}, nil

		case len(clientConfig.Username) > 0:
			return &userapi.User{ObjectMeta: metav1.ObjectMeta{Name: clientConfig.Username}}, nil

		}
	}

	if err != nil {
		return nil, err
	}

	return me, nil
}

func confirmProjectAccess(currentProject string, projectClient projectclient.ProjectInterface, kClient kclientset.Interface) error {
	_, projectErr := projectClient.Projects().Get(currentProject, metav1.GetOptions{})
	if !kapierrors.IsNotFound(projectErr) && !kapierrors.IsForbidden(projectErr) {
		return projectErr
	}

	// at this point we know the error is a not found or forbidden, but we'll test namespaces just in case we're running on kube
	if _, err := kClient.Core().Namespaces().Get(currentProject, metav1.GetOptions{}); err == nil {
		return nil
	}

	// otherwise return the openshift error default
	return projectErr
}
