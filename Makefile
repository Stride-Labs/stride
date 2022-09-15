#!/usr/bin/make -f

BUILDDIR ?= $(CURDIR)/build
build=s
cache=false
COMMIT := $(shell git log -1 --format='%H')

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
	sh scripts-local/tests/run_all_tests.sh

test-integration-docker:
	sh scripts/tests/run_all_tests.sh

###############################################################################
###                                DockerNet                                ###
###############################################################################

build-docker: 
	@sh scripts/build.sh -${build} ${BUILDDIR}
	
start-docker: build-docker
	@sh scripts/start_network.sh 

clean-docker: 
	@docker-compose stop
	@docker-compose down
	rm -rf scripts/state
	docker image prune -a
	
stop-docker:
	@pkill -f "docker-compose logs" || true
	@pkill -f "/bin/bash.*create_logs.sh" || true
	docker-compose down

###############################################################################
###                                LocalNet                                 ###
###############################################################################

check-dependencies:
	sh scripts-local/check_dependencies.sh

build-local: 
	@sh scripts-local/build.sh -${build} ${BUILDDIR}

start-local: build-local
	@sh scripts-local/start_network.sh ${cache}

stop-local:
	@killall gaiad strided junod osmosisd rly hermes interchain-queries icq-startup.sh || true
	@pkill -f "/bin/bash.*create_logs.sh" || true
	@pkill -f "sh.*start_network.sh" || true

