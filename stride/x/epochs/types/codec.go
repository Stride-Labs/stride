package types

import (
	"github.com/Stride-Labs/cosmos-sdk/codec"
	codectypes "github.com/Stride-Labs/cosmos-sdk/codec/types"
)

var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
