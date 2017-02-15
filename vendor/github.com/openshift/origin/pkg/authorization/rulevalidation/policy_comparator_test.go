package rulevalidation

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/util/sets"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

type escalationTest struct {
	ownerRules   []authorizationapi.PolicyRule
	servantRules []authorizationapi.PolicyRule

	expectedCovered        bool
	expectedUncoveredRules []authorizationapi.PolicyRule
}

func TestExactMatch(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("builds")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("builds")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestMultipleRulesCoveringSingleRule(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("deployments")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("builds")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("builds", "deployments")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "deployments")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)

}

func TestMultipleAPIGroupsCoveringSingleRule(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"group1"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("deployments")},
			{APIGroups: []string{"group1"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("builds")},
			{APIGroups: []string{"group1"}, Verbs: sets.NewString("update"), Resources: sets.NewString("builds", "deployments")},
			{APIGroups: []string{"group2"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("deployments")},
			{APIGroups: []string{"group2"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("builds")},
			{APIGroups: []string{"group2"}, Verbs: sets.NewString("update"), Resources: sets.NewString("builds", "deployments")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"group1", "group2"}, Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "deployments")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)

}

func TestSingleAPIGroupsCoveringMultiple(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"group1", "group2"}, Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "deployments")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"group1"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("deployments")},
			{APIGroups: []string{"group1"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("builds")},
			{APIGroups: []string{"group1"}, Verbs: sets.NewString("update"), Resources: sets.NewString("builds", "deployments")},
			{APIGroups: []string{"group2"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("deployments")},
			{APIGroups: []string{"group2"}, Verbs: sets.NewString("delete"), Resources: sets.NewString("builds")},
			{APIGroups: []string{"group2"}, Verbs: sets.NewString("update"), Resources: sets.NewString("builds", "deployments")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)

}

func TestMultipleRulesMissingSingleVerbResourceCombination(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "deployments")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("pods")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "deployments", "pods")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("update"), Resources: sets.NewString("pods")},
		},
	}.test(t)
}

func TestResourceGroupCoveringEnumerated(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("create", "delete", "update"), Resources: sets.NewString("resourcegroup:builds")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "buildconfigs")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestEnumeratedCoveringResourceGroup(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "buildconfigs", "buildlogs", "buildconfigs/instantiate", "buildconfigs/instantiatebinary", "builds/log", "builds/clone", "buildconfigs/webhooks")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("resourcegroup:builds")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestEnumeratedMissingPartOfResourceGroup(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("builds", "buildconfigs")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete", "update"), Resources: sets.NewString("resourcegroup:builds")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("buildlogs")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("buildlogs")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("buildconfigs/instantiate")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("buildconfigs/instantiate")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("buildconfigs/instantiatebinary")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("buildconfigs/instantiatebinary")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("builds/log")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("builds/log")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("builds/clone")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("builds/clone")},
			{Verbs: sets.NewString("delete"), Resources: sets.NewString("buildconfigs/webhooks")},
			{Verbs: sets.NewString("update"), Resources: sets.NewString("buildconfigs/webhooks")},
		},
	}.test(t)
}

func TestAPIGroupStarCoveringMultiple(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"group1", "group2"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestEnumerationNotCoveringAPIGroupStar(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"dummy-group"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},
	}.test(t)
}

func TestAPIGroupStarCoveringStar(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{APIGroups: []string{"*"}, Verbs: sets.NewString("get"), Resources: sets.NewString("roles")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestVerbStarCoveringMultiple(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("*"), Resources: sets.NewString("roles")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("watch", "list"), Resources: sets.NewString("roles")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestEnumerationNotCoveringVerbStar(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get", "list", "watch", "create", "update", "delete", "exec"), Resources: sets.NewString("roles")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("*"), Resources: sets.NewString("roles")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("*"), Resources: sets.NewString("roles")},
		},
	}.test(t)
}

func TestVerbStarCoveringStar(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("*"), Resources: sets.NewString("roles")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("*"), Resources: sets.NewString("roles")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestResourceStarCoveringMultiple(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("*")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("resourcegroup:deployments")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestEnumerationNotCoveringResourceStar(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("roles", "resourcegroup:deployments")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("*")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("*")},
		},
	}.test(t)
}

func TestResourceStarCoveringStar(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("*")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("*")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestResourceNameEmptyCoveringMultiple(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), ResourceNames: sets.NewString()},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), ResourceNames: sets.NewString("foo", "bar")},
		},

		expectedCovered:        true,
		expectedUncoveredRules: []authorizationapi.PolicyRule{},
	}.test(t)
}

func TestEnumerationNotCoveringResourceNameEmpty(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), ResourceNames: sets.NewString("foo", "bar")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), ResourceNames: sets.NewString()},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods")},
		},
	}.test(t)
}

func TestNonResourceCoversExactMatch(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/foo")},
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/bar")},
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/baz")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/foo", "/bar")},
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/baz")},
		},

		expectedCovered: true,
	}.test(t)
}

func TestNonResourceCoversWildcard(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get", "post"), NonResourceURLs: sets.NewString("/foo/*", "/bar/*")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get", "post"), NonResourceURLs: sets.NewString("/foo/", "/bar/")},
			{Verbs: sets.NewString("get", "post"), NonResourceURLs: sets.NewString("/foo/1", "/bar/1")},
			{Verbs: sets.NewString("get", "post"), NonResourceURLs: sets.NewString("/foo/*", "/bar/*")},
		},

		expectedCovered: true,
	}.test(t)
}

func TestNonResourceUncovered(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/foo")},
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/bar/*")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get", "post"), NonResourceURLs: sets.NewString("/foo", "/foo/1", "/foo/*", "/bar/baz", "/bar/*")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/foo")},
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/foo/1")},
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/foo/1")},
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/foo/*")},
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/foo/*")},
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/bar/baz")},
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/bar/*")},
		},
	}.test(t)
}

func TestMixedResourceNonResourceCovered(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), NonResourceURLs: sets.NewString("/api", "/api/*")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), NonResourceURLs: sets.NewString("/api", "/api/v1")},
		},

		expectedCovered: true,
	}.test(t)
}

func TestMixedResourceNonResourceUncovered(t *testing.T) {
	escalationTest{
		ownerRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get"), Resources: sets.NewString("pods"), NonResourceURLs: sets.NewString("/api", "/api/*")},
		},
		servantRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("get", "post"), Resources: sets.NewString("pods", "builds"), NonResourceURLs: sets.NewString("/api", "/apis")},
		},

		expectedCovered: false,
		expectedUncoveredRules: []authorizationapi.PolicyRule{
			{Verbs: sets.NewString("post"), Resources: sets.NewString("pods")},
			{Verbs: sets.NewString("get"), Resources: sets.NewString("builds")},
			{Verbs: sets.NewString("post"), Resources: sets.NewString("builds")},

			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/api")},
			{Verbs: sets.NewString("get"), NonResourceURLs: sets.NewString("/apis")},
			{Verbs: sets.NewString("post"), NonResourceURLs: sets.NewString("/apis")},
		},
	}.test(t)
}

func (test escalationTest) test(t *testing.T) {
	actualCovered, actualUncoveredRules := Covers(test.ownerRules, test.servantRules)

	if actualCovered != test.expectedCovered {
		t.Errorf("expected %v, but got %v", test.expectedCovered, actualCovered)
	}

	if !rulesMatch(test.expectedUncoveredRules, actualUncoveredRules) {
		t.Errorf("expected %v, but got %v", test.expectedUncoveredRules, actualUncoveredRules)
	}
}

func rulesMatch(expectedRules, actualRules []authorizationapi.PolicyRule) bool {
	if len(expectedRules) != len(actualRules) {
		return false
	}

	for _, expectedRule := range expectedRules {
		found := false
		for _, actualRule := range actualRules {
			if reflect.DeepEqual(expectedRule, actualRule) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
