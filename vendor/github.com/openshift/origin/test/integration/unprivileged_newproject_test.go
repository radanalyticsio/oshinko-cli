package integration

import (
	"io/ioutil"
	"testing"
	"time"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	"github.com/openshift/origin/pkg/client"
	oc "github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
	projectapi "github.com/openshift/origin/pkg/project/api"
	testutil "github.com/openshift/origin/test/util"
	testserver "github.com/openshift/origin/test/util/server"
)

func TestUnprivilegedNewProject(t *testing.T) {
	testutil.RequireEtcd(t)
	defer testutil.DumpEtcdOnFailure(t)
	_, clusterAdminKubeConfig, err := testserver.StartTestMasterAPI()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clusterAdminClientConfig, err := testutil.GetClusterAdminClientConfig(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valerieClientConfig := *clusterAdminClientConfig
	valerieClientConfig.Username = ""
	valerieClientConfig.Password = ""
	valerieClientConfig.BearerToken = ""
	valerieClientConfig.CertFile = ""
	valerieClientConfig.KeyFile = ""
	valerieClientConfig.CertData = nil
	valerieClientConfig.KeyData = nil

	accessToken, err := tokencmd.RequestToken(&valerieClientConfig, nil, "valerie", "security!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valerieClientConfig.BearerToken = accessToken
	valerieOpenshiftClient, err := client.New(&valerieClientConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// confirm that we have access to request the project
	allowed, err := valerieOpenshiftClient.ProjectRequests().List(kapi.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed.Status != unversioned.StatusSuccess {
		t.Fatalf("expected %v, got %v", unversioned.StatusSuccess, allowed.Status)
	}

	requestProject := oc.NewProjectOptions{
		ProjectName: "new-project",
		DisplayName: "display name here",
		Description: "the special description",

		Client: valerieOpenshiftClient,
		Out:    ioutil.Discard,
	}

	if err := requestProject.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	waitForProject(t, valerieOpenshiftClient, "new-project", 5*time.Second, 10)

	actualProject, err := valerieOpenshiftClient.Projects().Get("new-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e, a := "valerie", actualProject.Annotations[projectapi.ProjectRequester]; e != a {
		t.Errorf("incorrect project requester: expected %v, got %v", e, a)
	}

	if err := requestProject.Run(); !kapierrors.IsAlreadyExists(err) {
		t.Fatalf("expected an already exists error, but got %v", err)
	}

}
func TestUnprivilegedNewProjectFromTemplate(t *testing.T) {
	testutil.RequireEtcd(t)
	defer testutil.DumpEtcdOnFailure(t)
	namespace := "foo"
	templateName := "bar"

	masterOptions, err := testserver.DefaultMasterOptions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	masterOptions.ProjectConfig.ProjectRequestTemplate = namespace + "/" + templateName

	clusterAdminKubeConfig, err := testserver.StartConfiguredMaster(masterOptions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clusterAdminClientConfig, err := testutil.GetClusterAdminClientConfig(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	clusterAdminClient, err := testutil.GetClusterAdminClient(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valerieClientConfig := *clusterAdminClientConfig
	valerieClientConfig.Username = ""
	valerieClientConfig.Password = ""
	valerieClientConfig.BearerToken = ""
	valerieClientConfig.CertFile = ""
	valerieClientConfig.KeyFile = ""
	valerieClientConfig.CertData = nil
	valerieClientConfig.KeyData = nil

	accessToken, err := tokencmd.RequestToken(&valerieClientConfig, nil, "valerie", "security!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valerieClientConfig.BearerToken = accessToken
	valerieOpenshiftClient, err := client.New(&valerieClientConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := clusterAdminClient.Projects().Create(&projectapi.Project{ObjectMeta: kapi.ObjectMeta{Name: namespace}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	template, err := testutil.GetTemplateFixture("testdata/project-request-template-with-quota.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	template.Name = templateName
	template.Namespace = namespace

	_, err = clusterAdminClient.Templates(namespace).Create(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	requestProject := oc.NewProjectOptions{
		ProjectName: "new-project",
		DisplayName: "display name here",
		Description: "the special description",

		Client: valerieOpenshiftClient,
		Out:    ioutil.Discard,
	}

	if err := requestProject.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	waitForProject(t, valerieOpenshiftClient, "new-project", 5*time.Second, 10)
	project, err := valerieOpenshiftClient.Projects().Get("new-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Annotations["extra"] != "here" {
		t.Errorf("unexpected project %#v", project)
	}

	if err := clusterAdminClient.Templates(namespace).Delete(templateName); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	requestProject.ProjectName = "different"
	// This should fail during the template retrieve
	if err := requestProject.Run(); !kapierrors.IsNotFound(err) {
		t.Fatalf("expected a not found error, but got %v", err)
	}

}

func TestUnprivilegedNewProjectDenied(t *testing.T) {
	testutil.RequireEtcd(t)
	defer testutil.DumpEtcdOnFailure(t)
	_, clusterAdminKubeConfig, err := testserver.StartTestMasterAPI()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clusterAdminClient, err := testutil.GetClusterAdminClient(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role, err := clusterAdminClient.ClusterRoles().Get(bootstrappolicy.SelfProvisionerRoleName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role.Rules = []authorizationapi.PolicyRule{}
	if _, err := clusterAdminClient.ClusterRoles().Update(role); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clusterAdminClientConfig, err := testutil.GetClusterAdminClientConfig(clusterAdminKubeConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valerieClientConfig := *clusterAdminClientConfig
	valerieClientConfig.Username = ""
	valerieClientConfig.Password = ""
	valerieClientConfig.BearerToken = ""
	valerieClientConfig.CertFile = ""
	valerieClientConfig.KeyFile = ""
	valerieClientConfig.CertData = nil
	valerieClientConfig.KeyData = nil

	accessToken, err := tokencmd.RequestToken(&valerieClientConfig, nil, "valerie", "security!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valerieClientConfig.BearerToken = accessToken
	valerieOpenshiftClient, err := client.New(&valerieClientConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := testutil.WaitForClusterPolicyUpdate(valerieOpenshiftClient, "create", projectapi.Resource("projectrequests"), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// confirm that we have access to request the project
	_, err = valerieOpenshiftClient.ProjectRequests().List(kapi.ListOptions{})
	if err == nil {
		t.Fatalf("expected error: %v", err)
	}
	expectedError := `You may not request a new project via this API.`
	if (err != nil) && (err.Error() != expectedError) {
		t.Fatalf("expected\n\t%v\ngot\n\t%v", expectedError, err.Error())
	}
}
