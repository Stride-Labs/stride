K8S_NAMESPACE=integration
VENV_NAME=integration

CONDA_BASE := $(shell conda info --base)/envs
KUBECTL := $(shell which kubectl)
DOCKER := $(shell which docker)
HELM := $(shell which helm)
VENV_BIN := $(CONDA_BASE)/$(VENV_NAME)/bin
PYTHON := $(VENV_BIN)/python

HELM_CHART=network

python-install:
	conda create --name $(VENV_NAME) python=3.11 -y
	$(PYTHON) -m pip install -r api/requirements.txt

start-api:
	@$(PYTHON) -m uvicorn api.main:app --proxy-headers 

build-api:
	@echo "Building docker image: api"
	@$(DOCKER) buildx build --platform linux/amd64 --tag api -f dockerfiles/Dockerfile.api api
	@$(DOCKER) tag api gcr.io/stride-nodes/integration-tests/api:latest
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/api:latest

build-stride:
	@echo "Building docker image: stride-validator"
	@$(DOCKER) buildx build --platform linux/amd64 --tag core:stride ..
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride-validator -f dockerfiles/Dockerfile.stride .
	@$(DOCKER) tag stride-validator gcr.io/stride-nodes/integration-tests/chains/stride:latest
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/chains/stride:latest

build-cosmos:
	@echo "Building docker image"
	@$(DOCKER) buildx build --platform linux/amd64 --tag cosmos-validator -f dockerfiles/Dockerfile.cosmos .
	@$(DOCKER) tag cosmos-validator gcr.io/stride-nodes/integration-tests/chains/cosmoshub:v18.1.0
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/chains/cosmoshub:v18.1.0

update-scripts:
	@echo "Updating scripts configmap"
	@$(KUBECTL) delete configmap scripts -n $(K8S_NAMESPACE) --ignore-not-found=true
	@$(KUBECTL) create configmap scripts --from-file=scripts -n $(K8S_NAMESPACE)

update-config:
	@echo "Updating config configmap"
	@$(KUBECTL) delete configmap config -n $(K8S_NAMESPACE) --ignore-not-found=true
	@$(KUBECTL) create configmap config --from-file=config -n $(K8S_NAMESPACE)

start:
	@$(HELM) install $(HELM_CHART) $(HELM_CHART) --values $(HELM_CHART)/values.yaml

stop:
	@$(HELM) uninstall $(HELM_CHART)

lint: 
	@$(HELM) lint $(HELM_CHART)