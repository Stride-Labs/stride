BUILDDIR ?= $(CURDIR)/build
STRIDE_HOME ?= $(CURDIR)/../../../..

KEY ?= ""
VALUE ?= ""

check-key-value:
	ifndef KEY
	$(error KEY is not set)
	endif

	ifndef VALUE
	$(error VALUE is not set)
	endif

.PHONY: build
build:
	cargo wasm

build-debug:
	cargo wasm-debug

build-optimized:
	docker run --rm -v "$(CURDIR)":/code \
		--mount type=volume,source="$(notdir $(CURDIR))_cache",target=/code/target \
		--mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
		cosmwasm/rust-optimizer:0.12.11

build-wasmd:
	cd deps/wasmd && go build -mod=readonly -o $(BUILDDIR)/wasmd ./cmd/wasmd
	docker build -t wasmd:local .

start-wasmd:
	@bash wasmd/init_state.sh
	docker-compose up 

stop-wasmd:
	docker-compose down

validate:
	cosmwasm-check ./target/wasm32-unknown-unknown/release/ica_oracle.wasm

store-contract: 
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/store_contract.sh

instantiate: 
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/instantiate.sh

deploy-contract:
	@make store-contract 
	@echo ""
	@make instantiate

query-oracles:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/query_oracles.sh

query-metrics:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/query_metrics.sh

add-metric:
	@KEY=$(KEY) VALUE=$(VALUE) STRIDE_HOME=$(STRIDE_HOME) bash scripts/add_metric.sh

add-oracle:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/add_oracle.sh

query-errors:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/query_errors.sh