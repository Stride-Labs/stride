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

lint:
	golangci-lint run


###############################################################################
###                                  Proto                                  ###
###############################################################################

proto-all: proto-format proto-gen

proto:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/protocgen.sh
	@echo
	@echo "=========== Generate Complete ============"
	@echo

protoVer=v0.7
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(protoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(protoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \; ; fi

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
	@killall gaiad strided hermes interchain-queries
	@pkill -f "/bin/bash.*create_logs.sh"
