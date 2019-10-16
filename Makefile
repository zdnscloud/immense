REGISTRY_NAME = zdnscloud
IMAGE_Name = storage-operator
IMAGE_VERSION = v3.1.1

.PHONY: all container

all: container

container: 
	docker build -t $(REGISTRY_NAME)/$(IMAGE_Name):$(IMAGE_VERSION) ./ --no-cache
build:
	CGO_ENABLED=0 GOOS=linux go build cmd/operator.go
