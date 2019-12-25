VERSION=`git describe --tags`
BUILD=`date +%FT%T%z`
BRANCH=`git branch | sed -n '/\* /s///p'`

LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}"
GOSRC = $(shell find . -type f -name '*.go')

REGISTRY_NAME = zdnscloud
IMAGE_Name = storage-operator
IMAGE_VERSION = v3.5.7

build:
	CGO_ENABLED=0 GOOS=linux go build cmd/operator.go

image:
	docker build -t $(REGISTRY_NAME)/$(IMAGE_NAME):${BRANCH} --build-arg version=${VERSION} --build-arg buildtime=${BUILD} .
	docker image prune -f
	docker push $(REGISTRY_NAME)/$(IMAGE_NAME):${BRANCH}

clean:
	rm -f operator

clean-image:
	docker rmi $(REGISTRY_NAME)/$(IMAGE_NAME):${BRANCH}

.PHONY: build
