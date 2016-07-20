TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $TAG ]; then
    TAG='0.0.0'
fi

APP=oshinko-rest-server

TAG_APPNAME_FLAGS="-X github.com/redhatanalytics/oshinko-rest/version.gitTag=$TAG -X github.com/redhatanalytics/oshinko-rest/version.appName=$APP"
