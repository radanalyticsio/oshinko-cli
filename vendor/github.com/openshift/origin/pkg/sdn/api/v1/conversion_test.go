package v1_test

import (
	"testing"

	"github.com/openshift/origin/pkg/sdn/api"
	testutil "github.com/openshift/origin/test/util/api"

	// install all APIs
	_ "github.com/openshift/origin/pkg/api/install"
)

func TestFieldSelectorConversions(t *testing.T) {
	testutil.CheckFieldLabelConversions(t, "v1", "ClusterNetwork",
		// Ensure all currently returned labels are supported
		api.ClusterNetworkToSelectableFields(&api.ClusterNetwork{}),
	)

	testutil.CheckFieldLabelConversions(t, "v1", "HostSubnet",
		// Ensure all currently returned labels are supported
		api.HostSubnetToSelectableFields(&api.HostSubnet{}),
	)

	testutil.CheckFieldLabelConversions(t, "v1", "NetNamespace",
		// Ensure all currently returned labels are supported
		api.NetNamespaceToSelectableFields(&api.NetNamespace{}),
	)

	testutil.CheckFieldLabelConversions(t, "v1", "EgressNetworkPolicy",
		// Ensure all currently returned labels are supported
		api.EgressNetworkPolicyToSelectableFields(&api.EgressNetworkPolicy{}),
	)
}
