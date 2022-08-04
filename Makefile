#!/usr/bin/make -f

BUILDDIR ?= $(CURDIR)/build
build=s
cache=false

.PHONY: build

all: check-dependencies build-local

###############################################################################
###                            Build & Clean                                ###
###############################################################################

build:
	mkdir -p $(BUILDDIR)/
	go build -mod=readonly -trimpath -o $(BUILDDIR) ./...;

clean: 
	rm -rf $(BUILDDIR)/* 

clean-state:
	rm -rf scripts-local/state

###############################################################################
###                                Tests                                    ###
###############################################################################

test-unit:
	@go test -mod=readonly ./x/$(module)/...

test-cover:
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./x/$(module)/...

test-integration:
	sh scripts-local/tests/run_all_tests.sh

test-integration-docker:
	sh scripts/tests/run_all_tests.sh

###############################################################################
###                                DockerNet                                ###
###############################################################################

init-docker:
	sh scripts/init_main.sh -${build}

clean-docker: 
	rm -rf scripts/state
	@docker-compose stop
	@docker-compose down
	docker image prune -a

###############################################################################
###                                LocalNet                                 ###
###############################################################################

check-dependencies:
	sh scripts-local/check_dependencies.sh

build-local: 
	@sh scripts-local/build.sh -${build} ${BUILDDIR}

init-local: build-local
	@sh scripts-local/start_network.sh ${cache}

stop:
	@killall gaiad strided hermes interchain-queries junod osmosisd
	@pkill -f "/bin/bash.*create_logs.sh"
