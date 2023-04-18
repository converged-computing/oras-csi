JAR_FILE_NAME = HelloWorldServer.jar
JAR_FILE_PATH = app/$(JAR_FILE_NAME)
APP_BASE_REPO = java-app-base
APP_BASE_TAG = 1.0 
DOCKER_CONTAINER_NAME = java-app
REGISTRY = localhost:5001
APP_REPO = java-app
APP_TAG = 1.0

APP_BASE = $(REGISTRY)/$(APP_BASE_REPO):$(APP_BASE_TAG)
APP = $(REGISTRY)/$(APP_REPO):$(APP_TAG)
.PHONY: help
help: ## Generates help for all targets
	@grep -E '^[^#[:space:]].*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: build-app
build-app: ## Build the java app
	docker run --rm -v $(shell pwd)/app:/app -w /app openjdk:8-jdk-alpine sh -c 'javac HelloWorldServer.java && jar cfe HelloWorldServer.jar HelloWorldServer .'
	@echo "JAR file created successfully."

.PHONY: push-appd
push-app: ## Push the app to the registry 
	@echo "Pushing JAR file to OCI Registry"
	cd app && oras push --artifact-type application/java-archive $(APP) $(JAR_FILE_NAME)

.PHONY: clean
clean: ## Delete the classes and jar files
	rm -f $(JAR_FILE_PATH)
	rm -f app/*.class

.PHONY: run-app 
run: ## Run the java app in the base container
	docker run --rm -p 8080:8080 -v $(shell pwd)/$(JAR_FILE_PATH):/app.jar --name $(DOCKER_CONTAINER_NAME) $(APP_BASE)

build-base-image: ## Build the base container image
	docker build -t $(APP_BASE) .

.PHONY: push-base-image
push-base-image: ## Tag and push the java base app
	docker push $(APP_BASE)

.PHONY: deploy
deploy: ## Deploy the app to the cluster
	kubectl apply -f ./pod.yaml 