#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
os::util::environment::setup_all_server_vars "test-extended/clusterup"

os::util::ensure::built_binary_exists 'oc'
os::util::environment::use_sudo
os::util::environment::setup_time_vars

function os::test::extended::clusterup::run_test () {
    local test="${1}"
    local funcname="os::test::extended::clusterup::${test}"
    local global_home="${HOME}"
    os::log::info "starting test -- ${test}"

    os::test::junit::declare_suite_start "extended/clusterup/${test}"

    local test_home="${ARTIFACT_DIR}/${test}/home"
    mkdir -p "${test_home}"
    export HOME="${test_home}"
    pushd "${HOME}" &> /dev/null
    os::log::info "Using ${HOME} as home directory"
    ${funcname} ${2}
    popd &> /dev/null
    export HOME="${global_home}"

    os::log::info "cluster up -- ${test}: ok"
    os::test::junit::declare_suite_end
}

function os::test::extended::clusterup::verify_router_and_registry () {
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success_and_text "oc get svc -n default" "docker-registry"
    os::cmd::expect_success_and_text "oc get svc -n default" "router"
	os::cmd::try_until_text "oc get endpoints docker-registry -o jsonpath='{ .subsets[*].ports[?(@.name==\"5000-tcp\")].port }' -n default" "5000" $(( 10*minute )) 1
	os::cmd::try_until_text "oc get endpoints router -o jsonpath='{ .subsets[*].ports[?(@.name==\"80-tcp\")].port }' -n default" "80" $(( 10*minute )) 1
    os::cmd::expect_success "oc login -u developer"
}

function os::test::extended::clusterup::verify_image_streams () {
    os::cmd::try_until_text "oc get is -n openshift ruby -o jsonpath='{ .status.tags[*].tag }'" "latest" $(( 10*minute )) 1
}

function os::test::extended::clusterup::verify_ruby_build () {
    os::cmd::expect_success "oc new-app https://github.com/openshift/ruby-hello-world.git"
    os::cmd::try_until_text "oc get builds -l app=ruby-hello-world -o jsonpath='{ .items[*].status.phase }'" "Complete" $(( 10*minute )) 1
}

function os::test::extended::clusterup::verify_persistent_volumes () {
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success_and_text "oc get jobs -n default" "persistent-volume-setup"
    os::cmd::try_until_text "oc get job persistent-volume-setup -n default -o jsonpath='{ .status.succeeded }'" "1" $(( 10*minute )) 1
    os::cmd::expect_success "oc get pv/pv0100"
    os::cmd::expect_success "oc login -u developer"
}

function os::test::extended::clusterup::skip_persistent_volumes () {
    mkdir -p /tmp/pv
    touch /tmp/pv/.skip_pv
    arg=$@
    os::cmd::expect_success_and_text "oc cluster up --host-pv-dir=/tmp/pv/ $arg" "Skip persistent volume creation"
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success_and_text "oc get pv | wc -l" "0"
}

function os::test::extended::clusterup::verify_console () {
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success_and_text "oc get svc -n openshift-web-console" "webconsole"
    os::cmd::try_until_text "oc get endpoints webconsole -o jsonpath='{ .subsets[*].ports[?(@.name==\"https\")].port }' -n openshift-web-console" "8443" $(( 10*minute )) 1
    os::cmd::expect_success "oc login -u developer"
}

function os::test::extended::clusterup::verify_metrics () {
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success_and_text "oc get pods -n openshift-infra" "metrics-deployer"
    os::cmd::try_until_text "oc get pods -n openshift-infra -l job-name=metrics-deployer-pod" "Completed" $(( 20*minute )) 2
	os::cmd::try_until_text "oc get endpoints hawkular-metrics -o jsonpath='{ .subsets[*].ports[?(@.name==\"443-tcp\")].port }' -n openshift-infra" "443" $(( 10*minute )) 1
    os::cmd::expect_success "oc login -u developer"
}

function os::test::extended::clusterup::verify_logging () {
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success_and_text "oc get pods -n logging" "logging-deployer"
    os::cmd::try_until_text "oc get pods -n logging -l logging-infra=deployer" "Completed" $(( 20*minute )) 2
	os::cmd::try_until_text "oc get endpoints logging-kibana -o jsonpath='{ .subsets[*].ports[?(@.name==\"3000-tcp\")].port }' -n logging" "3000" $(( 10*minute )) 1
    os::cmd::expect_success "oc login -u developer"
}

function os::test::extended::clusterup::junit_cleanup() {
    # TODO(skuznets): un-hack this nonsense once traps are in a better state
    if [[ -n "${JUNIT_REPORT_OUTPUT:-}" ]]; then
      # get the jUnit output file into a workable state in case we crashed in the middle of testing something
      os::test::junit::reconcile_output

      # check that we didn't mangle jUnit output
      os::test::junit::check_test_counters

      # use the junitreport tool to generate us a report
      os::util::ensure::built_binary_exists 'junitreport'

      cat "${JUNIT_REPORT_OUTPUT}" "${junit_gssapi_output}" \
        | junitreport --type oscmd                          \
                      --suites nested                       \
                      --roots github.com/openshift/origin   \
                      --output "${ARTIFACT_DIR}/report.xml"
      cat "${ARTIFACT_DIR}/report.xml" | junitreport summarize
    fi
}

function os::test::extended::clusterup::cleanup () {
    os::log::info "Cleaning up cluster"
    oc cluster down
}

function os::test::extended::clusterup::cleanup_func () {
    local test="${1:-}"
    test_cleanup="os::test::extended::clusterup::${test}_cleanup"
    if [[ "$(type -t ${test_cleanup})" == "function" ]]; then
        echo "${test_cleanup}"
    else
        echo "os::test::extended::clusterup::cleanup"
    fi
}

function os::test::extended::clusterup::standard_test () {
    arg=$@
    os::cmd::expect_success "oc cluster up $arg"
    os::test::extended::clusterup::verify_router_and_registry
    os::test::extended::clusterup::verify_image_streams
    os::test::extended::clusterup::verify_ruby_build
    os::test::extended::clusterup::verify_persistent_volumes
}


# Test with host dirs specified
function os::test::extended::clusterup::internal::hostdirs () {
    local data_dir="${BASE_DIR}/data"
    local config_dir="${BASE_DIR}/config"
    local pv_dir="${BASE_DIR}/pv"
    local volumes_dir="${BASE_DIR}/volumes"
    os::cmd::expect_success "mkdir -p ${data_dir} ${config_dir} ${pv_dir} ${volumes_dir}"
    local volumes_arg=""
    if [[ "$(uname)" == "Linux" ]]; then
        volumes_arg="--host-volumes-dir=${volumes_dir}"
    fi
    os::test::extended::clusterup::standard_test \
        --host-data-dir="${data_dir}" \
        --host-config-dir="${config_dir}" \
        --host-pv-dir="${pv_dir}" \
        ${volumes_arg} \
        --version="$ORIGIN_COMMIT" \
        ${@}

	local sudo="${USE_SUDO:+sudo}"
    os::cmd::expect_success "${sudo} ls ${config_dir}/master/master-config.yaml"
    os::cmd::expect_success "${sudo} ls ${pv_dir}/pv0100"
    os::cmd::expect_success "${sudo} ls ${data_dir}/member"
}

# Tests the simplest case, no arguments specified
function os::test::extended::clusterup::noargs () {
    os::test::extended::clusterup::standard_test ${@}
}

# Tests creating a cluster with specific host directories
function os::test::extended::clusterup::hostdirs () {
    local base_dir
    if [[ "$(uname)" == "Darwin" ]]; then
        base_dir="$(mktemp -d /tmp/clusterup.XXXXXX)"
    else
        base_dir="$(mktemp -d)"
    fi
    BASE_DIR="${base_dir}" os::test::extended::clusterup::internal::hostdirs ${@}
}

# Tests bringing up the service catalog and provisioning a template
function os::test::extended::clusterup::service_catalog() {
    arg=$@
    os::cmd::expect_success "oc cluster up --service-catalog $arg"
    os::test::extended::clusterup::verify_router_and_registry
    os::test::extended::clusterup::verify_image_streams
    os::cmd::expect_success "oc login -u system:admin"
    os::cmd::expect_success "oc adm policy add-cluster-role-to-group system:openshift:templateservicebroker-client system:unauthenticated system:authenticated"
    # this is only to allow for the retrieval of the TSB service IP, not actual use of the TSB endpoints
    os::cmd::expect_success "oc policy add-role-to-user view developer -n openshift-template-service-broker"
    os::cmd::expect_success "oc login -u developer"
    os::cmd::expect_success "pushd ${OS_ROOT}/pkg/templateservicebroker/servicebroker/test-scripts; serviceUUID=`oc get template jenkins-ephemeral -n openshift -o template --template '{{.metadata.uid}}'` ./provision.sh"
    os::cmd::try_until_text "oc get pods" "jenkins-1-deploy" $(( 2*minute )) 1
    os::cmd::expect_success "pushd ${OS_ROOT}/pkg/templateservicebroker/servicebroker/test-scripts; serviceUUID=`oc get template jenkins-ephemeral -n openshift -o template --template '{{.metadata.uid}}'` ./bind.sh"
    os::cmd::expect_success "pushd ${OS_ROOT}/pkg/templateservicebroker/servicebroker/test-scripts; serviceUUID=`oc get template jenkins-ephemeral -n openshift -o template --template '{{.metadata.uid}}'` ./unbind.sh"
    os::cmd::expect_success "pushd ${OS_ROOT}/pkg/templateservicebroker/servicebroker/test-scripts; serviceUUID=`oc get template jenkins-ephemeral -n openshift -o template --template '{{.metadata.uid}}'` ./deprovision.sh"
    os::cmd::try_until_text "oc get pods" "Terminating" $(( 2*minute )) 1
}

# Tests creating a cluster with a public hostname
function os::test::extended::clusterup::publichostname () {
    os::test::extended::clusterup::standard_test \
        --public-hostname="myserver.127.0.0.1.nip.io" \
        --version="$ORIGIN_COMMIT" \
        ${@}
    os::cmd::expect_success_and_text "docker exec origin cat /var/lib/origin/openshift.local.config/master/master-config.yaml" "masterPublicURL.*myserver\.127\.0\.0\.1\.nip\.io"
}

# Tests creating a cluster with a numeric public hostname
function os::test::extended::clusterup::numerichostname () {
    os::test::extended::clusterup::standard_test \
        --public-hostname="127.0.0.1" \
        --version="$ORIGIN_COMMIT" \
        ${@}
    os::cmd::expect_success_and_text "docker exec origin cat /var/lib/origin/openshift.local.config/master/master-config.yaml" "masterPublicURL.*127\.0\.0\.1"
}

# Tests installation of console components
function os::test::extended::clusterup::console () {
    arg=$@
    os::cmd::expect_success "oc cluster up $arg"
    os::test::extended::clusterup::verify_console
}

# Tests installation of metrics components
function os::test::extended::clusterup::metrics () {
    os::test::extended::clusterup::standard_test --metrics ${@}
    os::test::extended::clusterup::verify_metrics
}

# Tests installation of aggregated logging components
function os::test::extended::clusterup::logging () {
    os::test::extended::clusterup::standard_test --logging ${@}
    os::test::extended::clusterup::verify_logging
}

# Verifies that a service can be accessed by a peer pod
# and by the pod running the service
function os::test::extended::clusterup::svcaccess () {
    arg=$@
    os::cmd::expect_success "oc cluster up $arg"
    os::cmd::expect_success "oc create -f ${OS_ROOT}/examples/gitserver/gitserver-persistent.yaml"
	os::cmd::try_until_text "oc get endpoints git -o jsonpath='{ .subsets[*].ports[?(@.port==8080)].port }'" "8080" $(( 10*minute )) 1
    # Test that a service can be accessed from a peer pod
    sleep 10
    os::cmd::expect_success "timeout 20s oc run peer --image=openshift/origin-gitserver:latest --attach --restart=Never --command -- curl http://git:8080/_/healthz"

    # Test that a service can be accessed from the same pod
    # This doesn't work in any cluster i've tried, not sure why but it's not a cluster up issue.
    #os::cmd::expect_success "timeout 2m oc rsh dc/git curl http://git:8080/_/healthz"
}

# Verifies that cluster up can start when a process is bound
# to the loopback interface with tcp port 53
function os::test::extended::clusterup::portinuse () {
    # Start listening on the host's 127.0.0.1 interface, port 53
    os::cmd::expect_success "docker run -d --name=port53 --entrypoint=/bin/bash --net=host openshift/origin-gitserver:latest -c \"socat TCP-LISTEN:53,bind=127.0.0.1,fork SYSTEM:'echo hello'\""
    arg=$@
    os::cmd::expect_success "oc cluster up $arg"
}

function os::test::extended::clusterup::portinuse_cleanup () {
    os::test::extended::clusterup::cleanup
    docker rm -f port53
}


# Verifies that clusterup handle different scenarios with persistent volumes setup
function os::test::extended::clusterup::persistentvolumes () {
    os::test::extended::clusterup::skip_persistent_volumes
    rm -rf /tmp/pv
}

readonly default_tests=(
    "service_catalog"
    "noargs"
    "hostdirs"
    "publichostname"
    "numerichostname"
    "portinuse"
    "svcaccess"
    "persistentvolumes"
    "console"

# logging+metrics team needs to fix/enable these tests.
#    "metrics"
#    "logging"
)

ORIGIN_COMMIT=${ORIGIN_COMMIT:-latest}

# run each test with each of these set of additional args.  Primarily
# intended to run the tests against different cluster versions.
readonly extra_args=(
    # Test the previous OCP release
    # TODO - enable this once v3.9 ships, v3.7 didn't have a TSB image so it's
    # annoying to test.
    #"--loglevel=5 --image=registry.access.redhat.com/openshift3/ose --version=v3.7"

    # Test the previous origin release
    # TODO - enable this once oc cluster up v3.9 supports modifiying cluster
    # roles on a 3.7 cluster image (https://github.com/openshift/origin/issues/17867)
    # "--loglevel=5 --image=docker.io/openshift/origin --version=v3.7.0"

    # Test the current published release
    "--loglevel=5"  # can't be empty, so pass something benign

    # Test the code being delivered
    "--loglevel=5 --version=${ORIGIN_COMMIT}"

)
tests=("${1:-"${default_tests[@]}"}")

# re-tag the latest service catalog image w/ the origin commit because we didn't
# build it locally, so we need a tag that aligns with the other images we're going to test here.
docker pull openshift/origin-service-catalog:latest
docker tag openshift/origin-service-catalog:latest openshift/origin-service-catalog:${ORIGIN_COMMIT}

echo "Running cluster up tests using tag $ORIGIN_COMMIT"

# Tag the docker registry image with the same tag as the other origin images
docker pull openshift/origin-docker-registry:latest
docker tag openshift/origin-docker-registry:latest openshift/origin-docker-registry:${ORIGIN_COMMIT}

# Tag the web console image with the same tag as the other origin images
docker pull openshift/origin-web-console:latest
docker tag openshift/origin-web-console:latest openshift/origin-web-console:${ORIGIN_COMMIT}

# Ensure that KUBECONFIG is not set
unset KUBECONFIG
for test in "${tests[@]}"; do
    # Special case the console tests, which only run in the current release.
    # N-1 tests for the console should be enabled next release.
    if [ "$test" == "console" ]; then
        cleanup_func=$("os::test::extended::clusterup::cleanup_func" "${test}")
        # trap "${cleanup_func}; os::test::extended::clusterup::junit_cleanup" EXIT
        os::test::extended::clusterup::run_test "${test}" "--loglevel=2 --version=${ORIGIN_COMMIT}"
        # trap - EXIT
        ${cleanup_func}
    else
        for extra_arg in "${extra_args[@]}"; do
            cleanup_func=$("os::test::extended::clusterup::cleanup_func" "${test}")
            # trap "${cleanup_func}; os::test::extended::clusterup::junit_cleanup" EXIT
            os::test::extended::clusterup::run_test "${test}" "${extra_arg}"
            # trap - EXIT
            ${cleanup_func}
        done
    fi
done

# os::test::extended::clusterup::junit_cleanup
