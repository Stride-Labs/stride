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

start-api: docker-bridge
	@$(DOCKER) compose up api

build-api:
	@echo "Building docker image: api"
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride:api -f dockerfiles/Dockerfile.api api
	@$(DOCKER) tag stride:api gcr.io/stride-nodes/integration-tests/api:latest
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/api:latest

build-stride:
	@echo "Building docker image: stride-validator"
	@$(DOCKER) buildx build --platform linux/amd64 --tag core:stride ..
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride:stride-validator -f dockerfiles/Dockerfile.stride .
	@$(DOCKER) tag stride:stride-validator gcr.io/stride-nodes/integration-tests/chains/stride:latest
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/chains/stride:latest

build-cosmos:
	@echo "Building docker image: cosmos-validator"
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride:cosmos-validator -f dockerfiles/Dockerfile.cosmos .
	@$(DOCKER) tag stride:cosmos-validator gcr.io/stride-nodes/integration-tests/chains/cosmoshub:v18.1.0
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/chains/cosmoshub:v18.1.0

build-relayer:
	@echo "Building docker image: relayer"
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride:relayer -f dockerfiles/Dockerfile.relayer .
	@$(DOCKER) tag stride:relayer gcr.io/stride-nodes/integration-tests/relayer:v2.5.2
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/relayer:v2.5.2

build-hermes:
	@echo "Building docker image: hermes"
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride:hermes -f dockerfiles/Dockerfile.hermes .
	@$(DOCKER) tag stride:hermes gcr.io/stride-nodes/integration-tests/hermes:v1.9.0
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/integration-tests/hermes:v1.9.0

local-init-chain: 
	@$(DOCKER) compose up --abort-on-container-exit; \
	EXIT_CODE=$$?; \
	$(DOCKER) compose down; \
	exit $$EXIT_CODE

start:
	@$(HELM) install $(HELM_CHART) $(HELM_CHART) --values $(HELM_CHART)/values.yaml -n $(K8S_NAMESPACE)

stop:
	@$(HELM) uninstall $(HELM_CHART) -n $(K8S_NAMESPACE)

lint: 
	@$(HELM) lint $(HELM_CHART) 