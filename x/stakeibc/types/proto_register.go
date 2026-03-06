package types

import proto "github.com/cosmos/gogoproto/proto"

// Register validator.proto before tx.pb.go's init() runs.
// Within a package, Go processes init() functions alphabetically by filename.
// Since tx.proto imports validator.proto, validator.proto must be registered
// in gogoProtoRegistry first, otherwise the Validator message descriptor is
// left as an unresolvable placeholder (due to AllowUnresolvable: true in
// gogoproto's RegisterFile). This causes amino-json signing to fail with
// "unknown protobuf field" errors.
// This file (proto_register.go, "p") sorts before tx.pb.go ("t").
func init() {
	proto.RegisterFile("stride/stakeibc/validator.proto", fileDescriptor_5d2f32e16bd6ab8f)
}
