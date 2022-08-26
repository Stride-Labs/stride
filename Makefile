#!/usr/bin/make -f

BUILDDIR ?= $(CURDIR)/build
build=s
cache=false

.PHONY: build

all: lint check-dependencies build-local

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

test-integration:
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
	@killall gaiad strided junod osmosisd rly hermes interchain-queries || true
	@pkill -f "/bin/bash.*create_logs.sh" || true
	@pkill -f "sh.*start_network.sh" || true
