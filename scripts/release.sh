#!/bin/bash
if [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
        result=:$(brew ls coreutils)
        if [ -z "$result" ]; then
          'Error: coreutils is not installed.'
          exit 1
        fi
        TOP_DIR=$(greadlink -f `dirname "$0"` | grep -o '.*/oshinko-cli')
else
        TOP_DIR=$(readlink -f `dirname "$0"` | grep -o '.*/oshinko-cli')
fi
. $TOP_DIR/sparkimage.sh
PROJECT='github.com/radanalyticsio/oshinko-cli'

function usage {
    echo "usage: release.sh VERSION"
    echo ""
    echo "VERSION -- should be set to the git tag being built, for example v0.6.1"
    echo ""
    echo "Builds release zip/tarball containing a versioned oshinko binary for each supported platform."
    echo "Meant to be run from the oshinko-cli root directory."
    echo "Output is in the _release directory."
    echo "Installs gox cross-compiler on the current GOPATH"
}

while getopts h option; do
    case $option in
        h)
            usage
            exit 0
            ;;
    esac
done
shift $((OPTIND-1))

if [ "$#" -ne 1 ]; then
    usage
    exit 1
fi

go get github.com/mitchellh/gox

$GOPATH/bin/gox "-output=_release/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" "-osarch=darwin/amd64 linux/386 linux/amd64" -tags standard -ldflags "-X $PROJECT/version.gitTag=$1 -X $PROJECT/version.appName=oshinko -X $PROJECT/version.sparkImage=$SPARK_IMAGE" ./cmd/oshinko

cd _release
zip oshinko_$1_macosx.zip oshinko_darwin_amd64/oshinko
tar -cvzf oshinko_$1_linux_386.tar.gz oshinko_linux_386
tar -cvzf oshinko_$1_linux_amd64.tar.gz oshinko_linux_amd64

