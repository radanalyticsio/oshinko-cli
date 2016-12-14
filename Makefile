LOCAL_IMAGE=project/oshinko-cli

.PHONY: all build test clean

build:
	scripts/build.sh build

image:
	sudo docker build -t $(LOCAL_IMAGE) .

clean:
	rm -rf _output
	sudo docker rmi $(LOCAL_IMAGE)

install:
	scripts/build.sh install

test:
	scripts/build.sh test

debug:
	scripts/build.sh debug


