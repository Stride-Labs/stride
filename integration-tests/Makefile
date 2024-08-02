K8S_NAMESPACE=integration
VENV_NAME=integration

CONDA_BASE := $(shell conda info --base)/envs
KUBECTL := $(shell which kubectl)
DOCKER := $(shell which docker)
VENV_BIN := $(CONDA_BASE)/$(VENV_NAME)/bin
PYTHON := $(VENV_BIN)/python

python-install:
	conda create --name $(VENV_NAME) python=3.11 -y
	$(PYTHON) -m pip install -r api/requirements.txt

start-api:
	@$(PYTHON) -m uvicorn api.main:app --proxy-headers 

build-api:
	@echo "Building docker image: integration-test-api"
	@$(DOCKER) buildx build --platform linux/amd64 --tag integration-test-api -f dockerfiles/Dockerfile.api api
	@$(DOCKER) tag integration-test-api gcr.io/stride-nodes/integration-test-api
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-test-api

build-stride:
	@echo "Building docker image: stride-test"
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride-test ..
	@$(DOCKER) tag stride-test gcr.io/stride-nodes/stride-test
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/stride-test

build-cosmos:
	@echo "Building docker image"
	@$(DOCKER) buildx build --platform linux/amd64 --tag cosmos-test -f dockerfiles/Dockerfile.cosmos .
	@$(DOCKER) tag cosmos-test gcr.io/stride-nodes/cosmos-test
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/cosmos-test

update-scripts:
	@echo "Updating scripts configmap"
	@$(KUBECTL) delete configmap scripts -n $(K8S_NAMESPACE) --ignore-not-found=true
	@$(KUBECTL) create configmap scripts --from-file=scripts -n $(K8S_NAMESPACE)

update-config:
	@echo "Updating config configmap"
	@$(KUBECTL) delete configmap config -n $(K8S_NAMESPACE) --ignore-not-found=true
	@$(KUBECTL) create configmap config --from-file=config -n $(K8S_NAMESPACE)

start:
	@$(KUBECTL) apply -f api.yaml
	@$(KUBECTL) apply -f validator.yaml

stop:
	@$(KUBECTL) delete -f api.yaml
	@$(KUBECTL) delete -f validator.yaml