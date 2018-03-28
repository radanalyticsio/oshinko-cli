#!/usr/bin/env groovy

// Used Jenkins plugins:
// * Pipeline GitHub Notify Step Plugin
// * Disable GitHub Multibranch Status Plugin - https://github.com/bluesliverx/disable-github-multibranch-status-plugin
//
// $OCP_HOSTNAME -- hostname of running Openshift cluster
// $OCP_USER     -- Openshift user
// $OCP_PASSWORD -- Openshift user's password

node('radanalytics-test') {
	withEnv(["GOPATH=$WORKSPACE", "KUBECONFIG=$WORKSPACE/client/kubeconfig", "PATH+OC_PATH=$WORKSPACE/client", "API_HOST=$OCP_HOSTNAME"]) {

		// generate build url
		def buildUrl = sh(script: 'curl https://url.corp.redhat.com/new?$BUILD_URL', returnStdout: true)

		stage('Build') {

			try {
				githubNotify(context: 'jenkins-ci/oshinko-cli', description: 'This change is being built', status: 'PENDING', targetUrl: buildUrl)
			} catch (err) {
				echo("Wasn't able to notify Github: ${err}")
			}

			try {
				// wipeout workspace
				deleteDir()

				dir('src/github.com/radanalyticsio/oshinko-cli') {
					checkout scm
				}

				// check golang version
				sh('go version')

				// download oc client
				dir('client') {
					sh('curl -LO https://github.com/openshift/origin/releases/download/v3.7.2/openshift-origin-client-tools-v3.7.2-282e43f-linux-64bit.tar.gz')
					sh('curl -LO https://github.com/openshift/origin/releases/download/v3.7.2/openshift-origin-server-v3.7.2-282e43f-linux-64bit.tar.gz')
					sh('tar -xzf openshift-origin-client-tools-v3.7.2-282e43f-linux-64bit.tar.gz')
					sh('tar -xzf openshift-origin-server-v3.7.2-282e43f-linux-64bit.tar.gz')
					sh('cp openshift-origin-client-tools-v3.7.2-282e43f-linux-64bit/oc .')
					sh('cp openshift-origin-server-v3.7.2-282e43f-linux-64bit/* .')
				}

				// build
				dir('src/github.com/radanalyticsio/oshinko-cli') {
					sh('make build | tee -a build.log && exit ${PIPESTATUS[0]}')
				}
			} catch (err) {
				try {
					githubNotify(context: 'jenkins-ci/oshinko-cli', description: 'This change cannot be built', status: 'ERROR', targetUrl: buildUrl)
				} catch (errNotify) {
					echo("Wasn't able to notify Github: ${errNotify}")
				}
				throw err
			} finally {
				dir('src/github.com/radanalyticsio/oshinko-cli') {
					archiveArtifacts(allowEmptyArchive: true, artifacts: 'build.log')
				}
			}
		}
		stage('Test') {
			try {
				try {
					githubNotify(context: 'jenkins-ci/oshinko-cli', description: 'This change is being tested', status: 'PENDING', targetUrl: buildUrl)
				} catch (err) {
					echo("Wasn't able to notify Github: ${err}")
				}

				// login to openshift instance
				sh('oc login https://$OCP_HOSTNAME:8443 -u $OCP_USER -p $OCP_PASSWORD --insecure-skip-tls-verify=true')
				// let's start on a specific project, to prevent start on a random project which could be deleted in the meantime
				sh('oc project testsuite')

				// run tests
				dir('src/github.com/radanalyticsio/oshinko-cli') {
					sh('./test/run.sh | tee -a test.log && exit ${PIPESTATUS[0]}')
				}
			} catch (err) {
				try {
					githubNotify(context: 'jenkins-ci/oshinko-cli', description: 'There are test failures', status: 'FAILURE', targetUrl: buildUrl)
				} catch (errNotify) {
					echo("Wasn't able to notify Github: ${errNotify}")
				}
				throw err
			} finally {
				dir('src/github.com/radanalyticsio/oshinko-cli') {
					archiveArtifacts(allowEmptyArchive: true, artifacts: 'test.log')
				}
			}

			try {
				githubNotify(context: 'jenkins-ci/oshinko-cli', description: 'This change looks good', status: 'SUCCESS', targetUrl: buildUrl)
			} catch (err) {
				echo("Wasn't able to notify Github: ${err}")
			}
		}
	}
}
