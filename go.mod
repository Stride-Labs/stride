module github.com/Stride-Labs/stride

go 1.16

require (
	github.com/DataDog/zstd v1.4.8 // indirect
	github.com/armon/go-metrics v0.3.10
	github.com/cosmos/cosmos-proto v1.0.0-alpha7
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/cosmos/ibc-go/v3 v3.1.0
	github.com/dgraph-io/badger/v2 v2.2007.3 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/ignite-hq/cli v0.21.0
	github.com/improbable-eng/grpc-web v0.15.0 // indirect
	github.com/jhump/protoreflect v1.12.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.11.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.7
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220610221304-9f5ed59c137d // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/genproto v0.0.0-20220719170305-83ca9fad585f
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
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
	// replace distribution module
	github.com/cosmos/cosmos-sdk/x/distribution => ./x/distribution
)
