.PHONY: push build

build:
	sudo docker build -t tmckay-scorpionstare .

push: build
	sudo docker tag tmckay-scorpionstare docker.io/tmckay/scorpionstare
	sudo docker push docker.io/tmckay/scorpionstare
