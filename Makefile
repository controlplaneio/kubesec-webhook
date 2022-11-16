NAME:=kubesec-webhook
DOCKER_REPOSITORY:=controlplaneio
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GITREPO:=github.com/controlplaneio/kubesec-webhook
GITCOMMIT:=$(shell git describe --dirty --always)
VERSION:=0.1-dev

.PHONY: build
build:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) -f Dockerfile .

.PHONY: push
push:
	docker push $(DOCKER_IMAGE_NAME):$(VERSION)

.PHONY: test
test:
	cd pkg/webhook ; go test -v -race ./...

.PHONY: certs
certs:
	cd deploy && ./gen-certs.sh

.PHONY: deploy
deploy:
	kubectl create ns kubesec -oyaml --dry-run=client | kubectl apply -f -
	kubectl apply -f ./deploy/

.PHONY: delete
delete:
	kubectl delete namespace kubesec
	kubectl delete -f ./deploy/webhook-registration.yaml

travis_push:
	@docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_IMAGE_NAME):$(TRAVIS_BRANCH)-$(GITCOMMIT)
	@docker push $(DOCKER_IMAGE_NAME):$(TRAVIS_BRANCH)-$(GITCOMMIT)

travis_release:
	@docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_IMAGE_NAME):$(TRAVIS_TAG)
	@docker push $(DOCKER_IMAGE_NAME):$(TRAVIS_TAG)
