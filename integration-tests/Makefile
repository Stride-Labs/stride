K8S_NAMESPACE=integration

KUBECTL := $(shell which kubectl)
DOCKER := $(shell which docker)

upload: docker-build docker-push

docker-build:
	@echo "Building docker image: stride-test"
	@$(DOCKER) buildx build --platform linux/amd64 --tag stride-test ..
	@$(DOCKER) tag stride-test gcr.io/stride-nodes/stride-test

docker-push:
	@echo "Pushing image to GCR"
	@$(DOCKER) push gcr.io/stride-nodes/stride-test

update-scripts:
	@echo "Updating scripts configmap"
	@$(KUBECTL) delete configmap scripts -n $(K8S_NAMESPACE) --ignore-not-found=true
	@$(KUBECTL) create configmap scripts --from-file=scripts -n $(K8S_NAMESPACE)

update-config:
	@echo "Updating config configmap"
	@$(KUBECTL) delete configmap config -n $(K8S_NAMESPACE) --ignore-not-found=true
	@$(KUBECTL) create configmap config --from-file=config -n $(K8S_NAMESPACE)

start:
	@$(KUBECTL) apply -f validator.yaml

stop:
	@$(KUBECTL) delete -f validator.yaml