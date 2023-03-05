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

validate:
	cosmwasm-check ./target/wasm32-unknown-unknown/release/ica_oracle.wasm

# Uploads the contract to juno
store-contract: 
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/store_contract.sh

# Instantiates the contract directly with the juno dockernet validator as the admin
instantiate-contract: 
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/instantiate_contract.sh

# Adds a metric directly to the contract from the juno dockernet validator
add-metric:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/add_metric.sh

# Add's an oracle to Stride
add-oracle:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/add_oracle.sh

# Instantiates the contract through a Stride ICA
instantiate-oracle: 
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/instantiate_oracle.sh

# Queries all oracles on Stride
query-oracles:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/query_oracles.sh

# Queries all metrics stored in the contract
query-metrics:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/query_metrics.sh

# Queries all ICA errors
query-errors:
	@STRIDE_HOME=$(STRIDE_HOME) bash scripts/query_errors.sh

