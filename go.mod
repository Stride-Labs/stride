module github.com/Stride-Labs/stride/v16

go 1.21

require (
	cosmossdk.io/math v1.1.2
	github.com/cometbft/cometbft v0.37.2
	github.com/cometbft/cometbft-db v0.8.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.2
	github.com/cosmos/cosmos-sdk v0.47.5
	github.com/cosmos/gogoproto v1.4.10
	github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7 v7.1.2
	github.com/cosmos/ibc-go/v7 v7.3.1
	github.com/cosmos/ics23/go v0.10.0
	github.com/cosmos/interchain-security/v3 v3.2.0
	github.com/evmos/vesting v0.0.0-20230818101748-9ea561e4529c
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.3
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/spf13/cast v1.5.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	google.golang.org/genproto/googleapis/api v0.0.0-20230711160842-782d3b101e98
	google.golang.org/grpc v1.58.2
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	// Use the keyring specified by the cosmos-sdk
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.2 //indirect
	// fork SDK to fix SDKv0.47 Distribution Bug
	// TODO - Remove this patch and update Tokens in a subsequent upgrade handler
	github.com/cosmos/cosmos-sdk => github.com/Stride-Labs/cosmos-sdk v0.47.5-stride-distribution-fix-0

	// Add additional verification check to ensure an account is a BaseAccount type before converting
	// it to a vesting account: https://github.com/Stride-Labs/vesting/pull/1
	github.com/evmos/vesting => github.com/Stride-Labs/vesting v1.0.0-check-base-account

	//github.com/evmos/vesting => ../vesting
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	// fork cast to add additional error checking
	github.com/spf13/cast => github.com/Stride-Labs/cast v0.0.3

	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)
