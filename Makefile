DRIVER_VERSION ?= 0.1.0
DEVTAG=$(DRIVER_VERSION)-dev

NAME=oras-csi-plugin
DOCKER_REGISTRY=ghcr.io/converged-computing

.PHONY: help
help: ## Generates help for all targets
	@grep -E '^[^#[:space:]].*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

ready: clean compile ## Clean pods and compiles the driver
publish: uninstall compile build push install ## Publish DRIVER_VERSION
dev: uninstall compile build-dev push-dev install ## Publish DEVTAG

uninstall: ## Uninstalls the plugin from the cluster
	@echo "==> Uninstalling plugin"
	kubectl delete -f deploy/kubernetes/csi-oras.yaml || true 
	kubectl delete -f deploy/kubernetes/csi-oras-config.yaml || true

install: ## Install plugin into the cluster
	@echo "==> Installing plugin"
	kubectl apply -f deploy/kubernetes/csi-oras.yaml || true
	kubectl apply -f deploy/kubernetes/csi-oras-config.yaml || true

compile:
	@echo "==> Building the project"
	@env CGO_ENABLED=0 GOCACHE=/tmp/go-cache GOOS=linux GOARCH=amd64 go build -a -o cmd/oras-csi-plugin/${NAME} cmd/oras-csi-plugin/main.go

push-dev: build-dev ## Build and push images to DEVTAG
	@echo "==> Publishing DEV $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)"
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:latest
	@echo "==> Your DEV image is now available at $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)"

push: build ## Build and push images for DRIVER_VERSION
	@echo "==> Publishing $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION)"
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION)
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:latest
	@echo "==> Your published image is now available at $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION)"

build: compile ## Build images for DRIVER_VERSION
	@echo "==> Building docker images"
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:$(DRIVER_VERSION) cmd/oras-csi-plugin
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:latest cmd/oras-csi-plugin

build-dev: compile ## Build images for DEVTAG
	@echo "==> Building DEV docker images"
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG) cmd/oras-csi-plugin
	@docker build -t $(DOCKER_REGISTRY)/oras-csi-plugin:latest cmd/oras-csi-plugin

clean: ## Deletes driver
	@echo "==> Cleaning releases"
	@GOOS=linux go clean -i -x ./...
	kubectl delete -f deploy/kubernetes/csi-oras.yaml || true 
	kubectl delete -f deploy/kubernetes/csi-oras-config.yaml || true

.PHONY: clean
