DRIVER_VERSION ?= 0.1.0
DEVTAG=$(DRIVER_VERSION)-dev
HELM_PLUGIN_NAME=oras-csi
BATS := bats
NAME=oras-csi-plugin
DOCKER_REGISTRY=ghcr.io/converged-computing

.PHONY: help
help: ## Generates help for all targets
	@grep -E '^[^#[:space:]].*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

ready: clean compile ## Clean pods and compiles the driver
publish: uninstall compile build push install ## Publish DRIVER_VERSION
dev: uninstall-dev compile build-dev push-dev install-dev ## Publish DEVTAG
dev-helm: uninstall-helm compile build-dev push-dev install-helm ## Publish with custom DOCKER_REGISTRY and DEVTAG

.PHONY: test
test: # Run basic end to end tests with bats
	bats -t -T test/bats/e2e.bats

uninstall: ## Uninstalls the plugin from the cluster
	@echo "==> Uninstalling plugin"
	kubectl delete -f deploy/driver-csi-oras.yaml || true 
	kubectl delete -f deploy/csi-oras-config.yaml || true

uninstall-dev: ## Uninstalls the development plugin from the cluster
	@echo "==> Uninstalling plugin"
	kubectl delete -f deploy/dev-driver.yaml || true 
	kubectl delete -f deploy/csi-oras-config.yaml || true

install-dev: ## Install development plugin into the cluster
	@echo "==> Installing plugin"
	kubectl apply -f deploy/dev-driver.yaml || true 
	kubectl apply -f deploy/csi-oras-config.yaml || true

install: ## Install plugin into the cluster
	@echo "==> Installing plugin"
	kubectl apply -f deploy/driver-csi-oras.yaml || true 
	kubectl apply -f deploy/csi-oras-config.yaml || true

compile:
	@echo "==> Building the project"
	@env CGO_ENABLED=0 GOCACHE=/tmp/go-cache GOOS=linux GOARCH=amd64 go build -a -o cmd/oras-csi-plugin/${NAME} cmd/oras-csi-plugin/main.go

push-dev: build-dev ## Build and push images to DEVTAG
	@echo "==> Publishing DEV $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)"
	@docker push $(DOCKER_REGISTRY)/oras-csi-plugin:$(DEVTAG)
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

clean: ## Deletes driver
	@echo "==> Cleaning releases"
	@GOOS=linux go clean -i -x ./...
	kubectl delete -f deploy/dev-driver.yaml || true 
	kubectl delete -f deploy/driver-csi-oras.yaml || true 
	kubectl delete -f deploy/csi-oras-config.yaml || true

.PHONY: clean

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

uninstall-helm: ## Uninstalls the plugin using helm
	@echo "==> Uninstalling helm plugin"
	helm uninstall $(HELM_PLUGIN_NAME) || true

install-helm: compile ## Install plugin using helm
	helm install --set node.csiOrasPlugin.image.repository="$(DOCKER_REGISTRY)/oras-csi-plugin" --set node.csiOrasPlugin.image.tag="$(DEVTAG)" $(HELM_PLUGIN_NAME) ./charts

# We will eventually want to add this
# .PHONY: sanity-test
# sanity-test: # RUN CSI sanity checks, assumes cluster is running with driver installed
#	go test -v ./test/sanity -ginkgo.skip=Controller\|should.work\|NodeStageVolume

# Helmify was only used for original base templates

# HELMIFY ?= $(LOCALBIN)/helmify
# .PHONY: helmify
# helmify: $(HELMIFY) ## Download helmify locally if necessary.
# $(HELMIFY): $(LOCALBIN)
#	test -s $(LOCALBIN)/helmify || GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@latest    
# helm: helmify
#	awk 'FNR==1 && NR!=1  {print "---"}{print}' ./deploy/kubernetes/*.yaml | $(HELMIFY)
