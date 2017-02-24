.PHONY : build clean install test validate-api image generate
DOCKERFILE=../Dockerfile.rest
DOCKERCONTEXT=..

build:
	tools/build.sh build

clean:
	rm -rf _output

install:
	tools/build.sh install

test:
	tools/oshinko-test.sh -t unit

validate-api:
	swagger validate api/swagger.yaml

image:
	docker build -t oshinko-rest -f $(DOCKERFILE) $(DOCKERCONTEXT)

generate:
	swagger generate server -f api/swagger.yaml -A oshinko-rest
	swagger generate client -f api/swagger.yaml -A oshinko-rest
