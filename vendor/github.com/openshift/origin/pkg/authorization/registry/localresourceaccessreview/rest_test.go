package localresourceaccessreview

import (
	"errors"
	"reflect"
	"testing"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/diff"
	"k8s.io/kubernetes/pkg/util/sets"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	"github.com/openshift/origin/pkg/authorization/authorizer"
	"github.com/openshift/origin/pkg/authorization/registry/resourceaccessreview"
)

type resourceAccessTest struct {
	authorizer    *testAuthorizer
	reviewRequest *authorizationapi.LocalResourceAccessReview
}

type testAuthorizer struct {
	users  sets.String
	groups sets.String
	err    string

	actualAttributes authorizer.DefaultAuthorizationAttributes
}

func (a *testAuthorizer) Authorize(ctx kapi.Context, attributes authorizer.Action) (allowed bool, reason string, err error) {
	// allow the initial check for "can I run this RAR at all"
	if attributes.GetResource() == "localresourceaccessreviews" {
		return true, "", nil
	}

	return false, "", errors.New("Unsupported")
}
func (a *testAuthorizer) GetAllowedSubjects(ctx kapi.Context, passedAttributes authorizer.Action) (sets.String, sets.String, error) {
	attributes, ok := passedAttributes.(authorizer.DefaultAuthorizationAttributes)
	if !ok {
		return nil, nil, errors.New("unexpected type for test")
	}

	a.actualAttributes = attributes
	if len(a.err) == 0 {
		return a.users, a.groups, nil
	}
	return a.users, a.groups, errors.New(a.err)
}

func TestNoNamespace(t *testing.T) {
	test := &resourceAccessTest{
		authorizer: &testAuthorizer{
			err: "namespace is required on this type: ",
		},
		reviewRequest: &authorizationapi.LocalResourceAccessReview{
			Action: authorizationapi.Action{
				Namespace: "",
				Verb:      "get",
				Resource:  "pods",
			},
		},
	}

	test.runTest(t)
}

func TestConflictingNamespace(t *testing.T) {
	authorizer := &testAuthorizer{}
	reviewRequest := &authorizationapi.LocalResourceAccessReview{
		Action: authorizationapi.Action{
			Namespace: "foo",
			Verb:      "get",
			Resource:  "pods",
		},
	}

	storage := NewREST(resourceaccessreview.NewRegistry(resourceaccessreview.NewREST(authorizer)))
	ctx := kapi.WithNamespace(kapi.NewContext(), "bar")
	_, err := storage.Create(ctx, reviewRequest)
	if err == nil {
		t.Fatalf("unexpected non-error: %v", err)
	}
	if e, a := `namespace: Invalid value: "foo": namespace must be: bar`, err.Error(); e != a {
		t.Fatalf("expected %v, got %v", e, a)
	}
}

func TestEmptyReturn(t *testing.T) {
	test := &resourceAccessTest{
		authorizer: &testAuthorizer{
			users:  sets.String{},
			groups: sets.String{},
		},
		reviewRequest: &authorizationapi.LocalResourceAccessReview{
			Action: authorizationapi.Action{
				Namespace: "unittest",
				Verb:      "get",
				Resource:  "pods",
			},
		},
	}

	test.runTest(t)
}

func TestNoErrors(t *testing.T) {
	test := &resourceAccessTest{
		authorizer: &testAuthorizer{
			users:  sets.NewString("one", "two"),
			groups: sets.NewString("three", "four"),
		},
		reviewRequest: &authorizationapi.LocalResourceAccessReview{
			Action: authorizationapi.Action{
				Namespace: "unittest",
				Verb:      "delete",
				Resource:  "deploymentConfig",
			},
		},
	}

	test.runTest(t)
}

func (r *resourceAccessTest) runTest(t *testing.T) {
	storage := NewREST(resourceaccessreview.NewRegistry(resourceaccessreview.NewREST(r.authorizer)))

	expectedResponse := &authorizationapi.ResourceAccessReviewResponse{
		Namespace: r.reviewRequest.Action.Namespace,
		Users:     r.authorizer.users,
		Groups:    r.authorizer.groups,
	}

	expectedAttributes := authorizer.ToDefaultAuthorizationAttributes(r.reviewRequest.Action)

	ctx := kapi.WithNamespace(kapi.NewContext(), r.reviewRequest.Action.Namespace)
	obj, err := storage.Create(ctx, r.reviewRequest)
	if err != nil && len(r.authorizer.err) == 0 {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.authorizer.err) != 0 {
		if err == nil {
			t.Fatalf("unexpected non-error: %v", err)
		}
		if e, a := r.authorizer.err, err.Error(); e != a {
			t.Fatalf("expected %v, got %v", e, a)
		}

		return
	}

	switch obj.(type) {
	case *authorizationapi.ResourceAccessReviewResponse:
		if !reflect.DeepEqual(expectedResponse, obj) {
			t.Errorf("diff %v", diff.ObjectGoPrintDiff(expectedResponse, obj))
		}
	case nil:
		if len(r.authorizer.err) == 0 {
			t.Fatal("unexpected nil object")
		}
	default:
		t.Errorf("Unexpected obj type: %v", obj)
	}

	if !reflect.DeepEqual(expectedAttributes, r.authorizer.actualAttributes) {
		t.Errorf("diff %v", diff.ObjectGoPrintDiff(expectedAttributes, r.authorizer.actualAttributes))
	}
}
