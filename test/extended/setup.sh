#!/bin/bash
#
# This abstracts starting up an extended server.

# Launches an extended server for OpenShift
# TODO: this should be doing less, because clusters should be stood up outside
#		and then tests are executed.	Tests that depend on fine grained setup should
#		be done in other contexts.
function os::test::extended::setup () {
	# build binaries
	#os::util::ensure::built_binary_exists 'oc'

	# ensure proper relative directories are set
	export KUBE_REPO_ROOT="${OS_ROOT}/vendor/k8s.io/kubernetes"

	os::util::environment::setup_time_vars

	# TODO: we shouldn't have to do this much work just to get tests to run against a real
	#   cluster, until then

	if [[ -n "${TEST_ONLY-}" ]]; then
		function cleanup() {
			out=$?
			#os::log::info "Exiting"
			return $out
		}
		trap "exit" INT TERM
		trap "cleanup" EXIT

		#os::log::info "Not starting server"
		return 0
	fi

	return 1

}
