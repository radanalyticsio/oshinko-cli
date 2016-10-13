GIT_TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $GIT_TAG ]; then
    GIT_TAG='unknown'
fi
GIT_COMMIT=`git log -n1 --pretty=format:%h`
TAG="${GIT_TAG}-${GIT_COMMIT}"

APP=oshinko-rest-server

TAG_APPNAME_FLAGS="-X github.com/radanalyticsio/oshinko-rest/version.gitTag=$TAG -X github.com/radanalyticsio/oshinko-rest/version.appName=$APP"
