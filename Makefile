#!/usr/bin/make -f
VERSION := $(shell echo $(shell git describe --tags))
BUILDDIR ?= $(CURDIR)/build
build=s
cache=false
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.7.0
STRIDE_HOME=./
DOCKERNET_HOME=./dockernet
DOCKERNET_COMPOSE_FILE=$(DOCKERNET_HOME)/docker-compose.yml
LOCALSTRIDE_HOME=./testutil/localstride
LOCALNET_COMPOSE_FILE=$(LOCALSTRIDE_HOME)/localnet/docker-compose.yml
STATE_EXPORT_COMPOSE_FILE=$(LOCALSTRIDE_HOME)/state-export/docker-compose.yml
LOCAL_TO_MAIN_COMPOSE_FILE=./scripts/local-to-mainnet/docker-compose.yml

# process build tags
LEDGER_ENABLED ?= true
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=stride \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=strided \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" 

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

.PHONY: build

###############################################################################
###                            Build & Clean                                ###
###############################################################################

build:
	which go
	mkdir -p $(BUILDDIR)/
	go build -mod=readonly $(BUILD_FLAGS) -trimpath -o $(BUILDDIR) ./...;

build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

install: go.sum
	go install $(BUILD_FLAGS) ./cmd/strided

clean:
	rm -rf $(BUILDDIR)/*

###############################################################################
###                                CI                                       ###
###############################################################################

gosec:
	gosec -exclude-dir=deps -severity=high ./...

lint:
	golangci-lint run

###############################################################################
###                                Tests                                    ###
###############################################################################

test-unit:
	@go test -mod=readonly ./x/... ./app/...

test-unit-path:
	@go test -mod=readonly ./x/$(path)/...

test-cover:
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./x/$(path)/...

test-integration-docker:
	bash $(DOCKERNET_HOME)/tests/run_all_tests.sh

test-integration-docker-all:
	@ALL_HOST_CHAINS=true bash $(DOCKERNET_HOME)/tests/run_all_tests.sh

###############################################################################
###                                DockerNet                                ###
###############################################################################

sync:
	@git submodule sync --recursive
	@git submodule update --init --recursive

build-docker:
	@bash $(DOCKERNET_HOME)/build.sh -${build} ${BUILDDIR}

start-docker: stop-docker build-docker
	@bash $(DOCKERNET_HOME)/start_network.sh

start-docker-all: stop-docker build-docker
	@ALL_HOST_CHAINS=true bash $(DOCKERNET_HOME)/start_network.sh

clean-docker:
	@docker-compose -f $(DOCKERNET_COMPOSE_FILE) stop
	@docker-compose -f $(DOCKERNET_COMPOSE_FILE) down
	rm -rf $(DOCKERNET_HOME)/state
	docker image prune -a

stop-docker:
	@bash $(DOCKERNET_HOME)/pkill.sh
	docker-compose -f $(DOCKERNET_COMPOSE_FILE) down

upgrade-build-old-binary:
	@DOCKERNET_HOME=$(DOCKERNET_HOME) BUILDDIR=$(BUILDDIR) bash $(DOCKERNET_HOME)/upgrades/build_old_binary.sh

submit-upgrade-immediately:
	UPGRADE_HEIGHT=150 bash $(DOCKERNET_HOME)/upgrades/submit_upgrade.sh

submit-upgrade-after-tests:
	UPGRADE_HEIGHT=500 bash $(DOCKERNET_HOME)/upgrades/submit_upgrade.sh

start-upgrade-integration-tests:
	PART=1 bash $(DOCKERNET_HOME)/tests/run_tests_upgrade.sh

finish-upgrade-integration-tests:
	PART=2 bash $(DOCKERNET_HOME)/tests/run_tests_upgrade.sh

upgrade-integration-tests-part-1: start-docker-all start-upgrade-integration-tests submit-upgrade-after-tests

setup-ics:
	UPGRADE_HEIGHT=150 bash $(DOCKERNET_HOME)/upgrades/setup_ics.sh

###############################################################################
###                              LocalNet                                   ###
###############################################################################
start-local-node:
	@bash scripts/start_local_node.sh

###############################################################################
###                           Local to Mainnet                              ###
###############################################################################
start-local-to-main:
	bash scripts/local-to-mainnet/start.sh

stop-local-to-main:
	docker-compose -f $(LOCAL_TO_MAIN_COMPOSE_FILE) down

###############################################################################
###                                Protobuf                                 ###
###############################################################################

containerProtoVer=0.14.0
containerProtoImage=ghcr.io/cosmos/proto-builder:$(containerProtoVer)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(DOCKER) run --user $(id -u):$(id -g) --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protocgen.sh; 

proto-format:
	@echo "Formatting Protobuf files"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./proto -name "*.proto" -exec clang-format -i {} \;  

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protoc-swagger-gen.sh; 

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main

###############################################################################
###                             LocalStride                                 ###
###############################################################################

localnet-keys:
	. $(LOCALSTRIDE_HOME)/localnet/add_keys.sh

localnet-init: localnet-clean localnet-build

localnet-clean:
	@rm -rfI $(HOME)/.stride/

localnet-build:
	@docker-compose -f $(LOCALNET_COMPOSE_FILE) build

localnet-start:
	@docker-compose -f $(LOCALNET_COMPOSE_FILE) up

localnet-startd:
	@docker-compose -f $(LOCALNET_COMPOSE_FILE) up -d

localnet-stop:
	@docker-compose -f $(LOCALNET_COMPOSE_FILE) down

localnet-state-export-init: localnet-state-export-clean localnet-state-export-build

localnet-state-export-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f $(STATE_EXPORT_COMPOSE_FILE) build

localnet-state-export-start:
	@docker-compose -f $(STATE_EXPORT_COMPOSE_FILE) up

localnet-state-export-startd:
	@docker-compose -f $(STATE_EXPORT_COMPOSE_FILE) up -d

localnet-state-export-upgrade:
	bash $(LOCALSTRIDE_HOME)/state-export/scripts/submit_upgrade.sh

localnet-state-export-stop:
	@docker-compose -f $(STATE_EXPORT_COMPOSE_FILE) down

localnet-state-export-clean: localnet-clean
