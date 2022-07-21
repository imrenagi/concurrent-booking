APP_NAME=booking
IMAGE_REGISTRY=docker.io/imrenagi
IMAGE_NAME=$(IMAGE_REGISTRY)/$(APP_NAME)
IMAGE_TAG=$(shell git rev-parse --short HEAD)

.PHONY: build test docker

build:
	go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o bin/$(APP_NAME) cmd/main.go

test:
	go test ./... -cover -vet -all

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest

docker-push:
	docker push $(IMAGE_NAME):latest