.PHONY: all build test clean

build:
	scripts/build.sh build

clean:
	rm -rf _output

install:
	scripts/build.sh install

test:
	scripts/build.sh test

debug:
	scripts/build.sh debug


