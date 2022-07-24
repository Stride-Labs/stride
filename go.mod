module github.com/Stride-Labs/stride

go 1.16

require (
	github.com/DataDog/zstd v1.4.8 // indirect
	github.com/cosmos/cosmos-proto v1.0.0-alpha7
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/cosmos/ibc-go/v3 v3.1.0
	github.com/dgraph-io/badger/v2 v2.2007.3 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dustin/go-humanize v1.0.1-0.20200219035652-afde56e7acac // indirect
	github.com/gogo/protobuf v1.3.3
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/improbable-eng/grpc-web v0.15.0 // indirect
	github.com/jhump/protoreflect v1.12.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tendermint/tendermint v0.35.9
	github.com/tendermint/tm-db v0.6.7
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	google.golang.org/genproto v0.0.0-20220719170305-83ca9fad585f
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.7 // indirect
)

replace (
	// TODO(TEST-54): Should we delete this replace statement and use the core cosmos-sdk for mainnet?
	// NOTE: If you need to bump the cosmos-sdk version, create a branch at the commit hash
	// of the target version on github.com/Stride-Labs/cosmos-sdk, then remove the error redaction
	// logic and push a new tag and the branch to github (use that tag below)
	github.com/cosmos/cosmos-sdk => github.com/Stride-Labs/cosmos-sdk v0.45.4-debug-2
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/ignite-hq/cli => github.com/ignite-hq/cli v0.21.0
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
