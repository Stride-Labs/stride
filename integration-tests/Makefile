K8S_NAMESPACE=integration
VENV_NAME=integration

CONDA_BASE := $(shell conda info --base)/envs
KUBECTL := $(shell which kubectl)
DOCKER := $(shell which docker)
HELM := $(shell which helm)
VENV_BIN := $(CONDA_BASE)/$(VENV_NAME)/bin
PYTHON := $(VENV_BIN)/python

HELM_CHART=network
PLATFORM=linux/amd64
GCR_REPO = gcr.io/stride-nodes/integration-tests

# Builds and pushes a docker image
# args: [image-file-suffix] [context] [image-name] 
# e.g. $(call build_and_push_docker,stride,.,chains/stride:latest
define build_and_push_docker
	@echo "Building docker image: $(1)"
	@$(DOCKER) buildx build --platform $(PLATFORM) --tag stride-tests:$(1) -f dockerfiles/Dockerfile.$(1) $(2)
	@$(DOCKER) tag stride-tests:$(1) $(GCR_REPO)/$(3)
	@echo "Pushing image to GCR: $(GCR_REPO)/$(3)"
	@$(DOCKER) push $(GCR_REPO)/$(3)
endef

python-install:
	conda create --name $(VENV_NAME) python=3.11 -y
	$(PYTHON) -m pip install -r api/requirements.txt

start-api: docker-bridge
	@$(DOCKER) compose up api

build-api:
	$(call build_and_push_docker,api,api,api:latest)

build-stride:
	@$(DOCKER) buildx build --platform linux/amd64 --tag core:stride ..
	$(call build_and_push_docker,stride,.,chains/stride:latest)

build-cosmos:
	$(call build_and_push_docker,cosmos,.,chains/cosmoshub:v18.1.0)

build-relayer:
	$(call build_and_push_docker,relayer,.,relayer:v2.5.2)

build-hermes:
	$(call build_and_push_docker,hermes,.,hermes:v1.9.0)

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