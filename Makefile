LOCAL_IMAGE=project/oshinko-cli

.PHONY: all build test clean

build:
	scripts/build.sh build

image:
	sudo docker build -t $(LOCAL_IMAGE) .

clean:
	rm -rf _output
	sudo docker rmi $(LOCAL_IMAGE)

# Run command tests. Uses whatever binaries are currently built.
#
# Example:
#   make test-cmd
test-cmd: build
	hack/test-cmd.sh

crd:
	scripts/build-crd.sh build
.PHONY: test-cmd
