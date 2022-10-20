#!/usr/bin/make -f

BUILDDIR ?= $(CURDIR)/build
build=s
cache=false
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.7.0

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

all: lint check-dependencies build-local

###############################################################################
###                            Build & Clean                                ###
###############################################################################

build:
	mkdir -p $(BUILDDIR)/
	go build -mod=readonly -ldflags '$(ldflags)' -trimpath -o $(BUILDDIR) ./...;

install: go.sum
		go install $(BUILD_FLAGS) ./cmd/strided

clean: 
	rm -rf $(BUILDDIR)/* 

clean-state:
	rm -rf scripts-local/state

###############################################################################
###                                CI                                       ###
###############################################################################

ci: lint check-dependencies test-unit gosec build-local

gosec:
	gosec -exclude-dir=deps -severity=high ./...

lint:
	golangci-lint run

###############################################################################
###                                Tests                                    ###
###############################################################################

test-unit:
	@go test -mod=readonly ./x/$(module)/...

test-cover:
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./x/$(module)/...

test-integration-local:
	bash scripts-local/tests/run_all_tests.sh

test-integration-docker:
	bash scripts/tests/run_all_tests.sh

###############################################################################
###                                DockerNet                                ###
###############################################################################

build-docker: 
	@bash scripts/build.sh -${build} ${BUILDDIR}
	
start-docker: build-docker
	@bash scripts/start_network.sh 

clean-docker: 
	@docker-compose stop
	@docker-compose down
	rm -rf scripts/state
	docker image prune -a
	
stop-docker:
	@-pkill -f "docker-compose logs" 
	@-pkill -f "/bin/bash.*create_logs.sh" 
	docker-compose down

###############################################################################
###                                Protobuf                                 ###
###############################################################################

containerProtoVer=v0.7
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(containerProtoVer)
containerProtoGenSwagger=cosmos-sdk-proto-gen-swagger-$(containerProtoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./proto -name "*.proto" -exec clang-format -i {} \; ; fi

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

.PHONY: proto-all proto-gen proto-format proto-lint

###############################################################################
###                              LocalStride                                ###
###############################################################################


###############################################################################
###                             LocalStride                                 ###
###############################################################################

localnet-keys:
	. testutil/localstride/scripts/add_keys.sh

localnet-init: localnet-clean localnet-build

localnet-clean:
	@rm -rfI $(HOME)/.stride/

localnet-build:
	@docker-compose -f testutil/localstride/docker-compose.yml build

localnet-start:
	@docker-compose -f testutil/localstride/docker-compose.yml up

localnet-startd:
	@docker-compose -f testutil/localstride/docker-compose.yml up -d

localnet-stop:
	@docker-compose -f testutil/localstride/docker-compose.yml down

localnet-state-export-init: localnet-state-export-clean localnet-state-export-build

localnet-state-export-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f testutil/localstride/state_export/docker-compose.yml build

localnet-state-export-start:
	@docker-compose -f testutil/localstride/state_export/docker-compose.yml up

localnet-state-export-startd:
	@docker-compose -f testutil/localstride/state_export/docker-compose.yml up -d

localnet-state-export-stop:
	@docker-compose -f testutil/localstride/docker-compose.yml down

localnet-state-export-clean: localnet-clean
