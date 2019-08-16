REGISTRY_NAME = zdnscloud
IMAGE_Name = storage-operator
IMAGE_VERSION = v1.7

.PHONY: all container

all: container

container: 
	docker build -t $(REGISTRY_NAME)/$(IMAGE_Name):$(IMAGE_VERSION) ./ --no-cache
