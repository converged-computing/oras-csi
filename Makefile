DRIVER_VERSION ?= 0.1.0
DEVTAG=$(DRIVER_VERSION)-dev

NAME=oras-csi-plugin
DOCKER_REGISTRY=ghcr.io/converged-computing

ready: clean compile
publish: uninstall compile build push install
dev: uninstall compile build-dev push-dev install

uninstall:
	@echo "==> Uninstalling plugin"
	kubectl delete -f deploy/kubernetes/csi-oras.yaml || true 
	kubectl delete -f deploy/kubernetes/csi-oras-config.yaml || true

install:
	@echo "==> Installing plugin"
	kubectl apply -f deploy/kubernetes/csi-oras.yaml || true
	kubectl apply -f deploy/kubernetes/csi-oras-config.yaml || true

compile:
	@echo "==> Building the project"
	@env CGO_ENABLED=0 GOCACHE=/tmp/go-cache GOOS=linux GOARCH=amd64 go build -a -o cmd/oras-csi-plugin/${NAME} cmd/oras-csi-plugin/main.go

push-dev: build-dev
	@echo "==> Publishing DEV $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)"
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:latest
	@echo "==> Your DEV image is now available at $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)"

push: build
	@echo "==> Publishing DEV $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION)"
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION)
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:latest
	@echo "==> Your published image is now available at $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION)"

build: compile
	@echo "==> Building docker images"
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION) cmd/oras-csi-plugin
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:latest cmd/oras-csi-plugin

build-dev: compile
	@echo "==> Building DEV docker images"
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG) cmd/oras-csi-plugin
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:latest cmd/oras-csi-plugin

clean:
	@echo "==> Cleaning releases"
	@GOOS=linux go clean -i -x ./...
	kubectl delete -f deploy/kubernetes/csi-oras.yaml || true 
	kubectl delete -f deploy/kubernetes/csi-oras-config.yaml || true

.PHONY: clean