NAME:=kubesec
DOCKER_REPOSITORY:=stefanprodan
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GITREPO:=github.com/stefanprodan/kubesec-webhook
GITCOMMIT:=$(shell git describe --dirty --always)
VERSION:=0.1-test0

.PHONY: build
build:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) -f Dockerfile .

.PHONY: push
push:
	docker push $(DOCKER_IMAGE_NAME):$(VERSION)

.PHONY: test
test:
	cd pkg/server ; go test -v -race ./...

.PHONY: certs
certs:
	cd deploy && ./gen-certs.sh
