// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stride/stakeibc/host_zone.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// next id: 22
type HostZone struct {
	ChainId               string       `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	ConnectionId          string       `protobuf:"bytes,2,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	Bech32Prefix          string       `protobuf:"bytes,17,opt,name=bech32prefix,proto3" json:"bech32prefix,omitempty"`
	TransferChannelId     string       `protobuf:"bytes,12,opt,name=transfer_channel_id,json=transferChannelId,proto3" json:"transfer_channel_id,omitempty"`
	Validators            []*Validator `protobuf:"bytes,3,rep,name=validators,proto3" json:"validators,omitempty"`
	BlacklistedValidators []*Validator `protobuf:"bytes,4,rep,name=blacklisted_validators,json=blacklistedValidators,proto3" json:"blacklisted_validators,omitempty"`
	WithdrawalAccount     *ICAAccount  `protobuf:"bytes,5,opt,name=withdrawal_account,json=withdrawalAccount,proto3" json:"withdrawal_account,omitempty"`
	FeeAccount            *ICAAccount  `protobuf:"bytes,6,opt,name=fee_account,json=feeAccount,proto3" json:"fee_account,omitempty"`
	DelegationAccount     *ICAAccount  `protobuf:"bytes,7,opt,name=delegation_account,json=delegationAccount,proto3" json:"delegation_account,omitempty"`
	RedemptionAccount     *ICAAccount  `protobuf:"bytes,16,opt,name=redemption_account,json=redemptionAccount,proto3" json:"redemption_account,omitempty"`
	// ibc denom on stride
	IbcDenom string `protobuf:"bytes,8,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom,omitempty"`
	// native denom on host zone
	HostDenom string `protobuf:"bytes,9,opt,name=host_denom,json=hostDenom,proto3" json:"host_denom,omitempty"`
	// TODO(TEST-68): Should we make this an array and store the last n redemption
	// rates then calculate a TWARR?
	LastRedemptionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,10,opt,name=last_redemption_rate,json=lastRedemptionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"last_redemption_rate"`
	RedemptionRate     github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,11,opt,name=redemption_rate,json=redemptionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"redemption_rate"`
	// stores how many days we should wait before issuing unbondings
	UnbondingFrequency uint64 `protobuf:"varint,14,opt,name=unbonding_frequency,json=unbondingFrequency,proto3" json:"unbonding_frequency,omitempty"`
	// TODO(TEST-101) int to dec
	StakedBal         github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,13,opt,name=staked_bal,json=stakedBal,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"staked_bal"`
	Address           string                                 `protobuf:"bytes,18,opt,name=address,proto3" json:"address,omitempty" yaml:"address"`
	Halted            bool                                   `protobuf:"varint,19,opt,name=halted,proto3" json:"halted,omitempty"`
	MinRedemptionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,20,opt,name=min_redemption_rate,json=minRedemptionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"min_redemption_rate"`
	MaxRedemptionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,21,opt,name=max_redemption_rate,json=maxRedemptionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"max_redemption_rate"`
}

func (m *HostZone) Reset()         { *m = HostZone{} }
func (m *HostZone) String() string { return proto.CompactTextString(m) }
func (*HostZone) ProtoMessage()    {}
func (*HostZone) Descriptor() ([]byte, []int) {
	return fileDescriptor_f81bf5b42c61245a, []int{0}
}
func (m *HostZone) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HostZone) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_HostZone.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *HostZone) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HostZone.Merge(m, src)
}
func (m *HostZone) XXX_Size() int {
	return m.Size()
}
func (m *HostZone) XXX_DiscardUnknown() {
	xxx_messageInfo_HostZone.DiscardUnknown(m)
}

var xxx_messageInfo_HostZone proto.InternalMessageInfo

func (m *HostZone) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

func (m *HostZone) GetConnectionId() string {
	if m != nil {
		return m.ConnectionId
	}
	return ""
}

func (m *HostZone) GetBech32Prefix() string {
	if m != nil {
		return m.Bech32Prefix
	}
	return ""
}

func (m *HostZone) GetTransferChannelId() string {
	if m != nil {
		return m.TransferChannelId
	}
	return ""
}

func (m *HostZone) GetValidators() []*Validator {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *HostZone) GetBlacklistedValidators() []*Validator {
	if m != nil {
		return m.BlacklistedValidators
	}
	return nil
}

func (m *HostZone) GetWithdrawalAccount() *ICAAccount {
	if m != nil {
		return m.WithdrawalAccount
	}
	return nil
}

func (m *HostZone) GetFeeAccount() *ICAAccount {
	if m != nil {
		return m.FeeAccount
	}
	return nil
}

func (m *HostZone) GetDelegationAccount() *ICAAccount {
	if m != nil {
		return m.DelegationAccount
	}
	return nil
}

func (m *HostZone) GetRedemptionAccount() *ICAAccount {
	if m != nil {
		return m.RedemptionAccount
	}
	return nil
}

func (m *HostZone) GetIbcDenom() string {
	if m != nil {
		return m.IbcDenom
	}
	return ""
}

func (m *HostZone) GetHostDenom() string {
	if m != nil {
		return m.HostDenom
	}
	return ""
}

func (m *HostZone) GetUnbondingFrequency() uint64 {
	if m != nil {
		return m.UnbondingFrequency
	}
	return 0
}

func (m *HostZone) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *HostZone) GetHalted() bool {
	if m != nil {
		return m.Halted
	}
	return false
}

func init() {
	proto.RegisterType((*HostZone)(nil), "stride.stakeibc.V3HostZone")
}

func init() { proto.RegisterFile("stride/stakeibc/host_zone.proto", fileDescriptor_f81bf5b42c61245a) }

var fileDescriptor_f81bf5b42c61245a = []byte{
	// 671 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0x4f, 0x4f, 0xdb, 0x30,
	0x18, 0xc6, 0x9b, 0xc1, 0xa0, 0x75, 0xf9, 0x57, 0x17, 0x50, 0x00, 0xad, 0xed, 0x3a, 0x69, 0xea,
	0x61, 0xa4, 0x1a, 0x5c, 0x26, 0xc4, 0x85, 0x3f, 0x9a, 0x56, 0xb6, 0x1d, 0x96, 0x49, 0x1c, 0xb8,
	0x44, 0x8e, 0xfd, 0xb6, 0xb5, 0x48, 0xed, 0x2e, 0x76, 0xa1, 0xec, 0x23, 0xec, 0xb4, 0x0f, 0xb3,
	0x0f, 0xc1, 0x11, 0xed, 0x34, 0xed, 0x80, 0x26, 0xf8, 0x06, 0xfb, 0x04, 0x53, 0x9d, 0xa4, 0x0d,
	0xed, 0x01, 0x26, 0x71, 0x4a, 0xfc, 0x3e, 0xcf, 0xfb, 0x7b, 0xac, 0xd7, 0x89, 0x51, 0x59, 0xe9,
	0x90, 0x33, 0xa8, 0x2b, 0x4d, 0x4e, 0x81, 0xfb, 0xb4, 0xde, 0x96, 0x4a, 0x7b, 0x5f, 0xa5, 0x00,
	0xa7, 0x1b, 0x4a, 0x2d, 0xf1, 0x62, 0x64, 0x70, 0x12, 0xc3, 0xfa, 0x44, 0xc7, 0x19, 0x09, 0x38,
	0x23, 0x5a, 0x86, 0x51, 0xc7, 0xfa, 0xf3, 0x71, 0x03, 0xa7, 0xc4, 0x23, 0x94, 0xca, 0x9e, 0xd0,
	0xb1, 0x65, 0xb9, 0x25, 0x5b, 0xd2, 0xbc, 0xd6, 0x07, 0x6f, 0x71, 0x75, 0x8d, 0x4a, 0xd5, 0x91,
	0xca, 0x8b, 0x84, 0x68, 0x11, 0x49, 0xd5, 0x6f, 0x08, 0x65, 0xdf, 0x49, 0xa5, 0x4f, 0xa4, 0x00,
	0xbc, 0x86, 0xb2, 0xb4, 0x4d, 0xb8, 0xf0, 0x38, 0xb3, 0xad, 0x8a, 0x55, 0xcb, 0xb9, 0xb3, 0x66,
	0xdd, 0x60, 0xf8, 0x05, 0x9a, 0xa7, 0x52, 0x08, 0xa0, 0x9a, 0x4b, 0xa3, 0x3f, 0x31, 0xfa, 0xdc,
	0xa8, 0xd8, 0x60, 0xb8, 0x8a, 0xe6, 0x7c, 0xa0, 0xed, 0xed, 0xad, 0x6e, 0x08, 0x4d, 0xde, 0xb7,
	0x0b, 0x91, 0x27, 0x5d, 0xc3, 0x0e, 0x2a, 0xea, 0x90, 0x08, 0xd5, 0x84, 0xd0, 0xa3, 0x6d, 0x22,
	0x04, 0x04, 0x03, 0xdc, 0x9c, 0xb1, 0x16, 0x12, 0xe9, 0x20, 0x52, 0x1a, 0x0c, 0xef, 0x20, 0x34,
	0x9c, 0x83, 0xb2, 0xa7, 0x2a, 0x53, 0xb5, 0xfc, 0xd6, 0xba, 0x33, 0x36, 0x3b, 0xe7, 0x38, 0xb1,
	0xb8, 0x29, 0x37, 0xfe, 0x84, 0x56, 0xfd, 0x80, 0xd0, 0xd3, 0x80, 0x2b, 0x0d, 0xcc, 0x4b, 0x71,
	0xa6, 0xef, 0xe5, 0xac, 0xa4, 0x3a, 0x8f, 0x47, 0xc8, 0x23, 0x84, 0xcf, 0xb9, 0x6e, 0xb3, 0x90,
	0x9c, 0x93, 0x20, 0x19, 0xbe, 0xfd, 0xb4, 0x62, 0xd5, 0xf2, 0x5b, 0x1b, 0x13, 0xb8, 0xc6, 0xc1,
	0xde, 0x5e, 0x64, 0x71, 0x0b, 0xa3, 0xb6, 0xb8, 0x84, 0x77, 0x51, 0xbe, 0x09, 0x30, 0x84, 0xcc,
	0xdc, 0x0f, 0x41, 0x4d, 0x80, 0xa4, 0xfb, 0x08, 0x61, 0x06, 0x01, 0xb4, 0x88, 0x39, 0x91, 0x04,
	0x32, 0xfb, 0x80, 0x9d, 0x8c, 0xda, 0x52, 0xac, 0x10, 0x18, 0x74, 0xba, 0x77, 0x58, 0x4b, 0x0f,
	0x60, 0x8d, 0xda, 0x12, 0xd6, 0x06, 0xca, 0x71, 0x9f, 0x7a, 0x0c, 0x84, 0xec, 0xd8, 0x59, 0x73,
	0xac, 0x59, 0xee, 0xd3, 0xc3, 0xc1, 0x1a, 0x3f, 0x43, 0xc8, 0xfc, 0x07, 0x91, 0x9a, 0x33, 0x6a,
	0x6e, 0x50, 0x89, 0x64, 0x81, 0x96, 0x03, 0xa2, 0xb4, 0x97, 0xda, 0x4c, 0x48, 0x34, 0xd8, 0x68,
	0x60, 0xdc, 0xdf, 0xbd, 0xbc, 0x2e, 0x67, 0x7e, 0x5f, 0x97, 0x5f, 0xb6, 0xb8, 0x6e, 0xf7, 0x7c,
	0x87, 0xca, 0x4e, 0xfc, 0x31, 0xc7, 0x8f, 0x4d, 0xc5, 0x4e, 0xeb, 0xfa, 0xa2, 0x0b, 0xca, 0x39,
	0x04, 0xfa, 0xf3, 0xc7, 0x26, 0x8a, 0xbf, 0xf5, 0x43, 0xa0, 0x2e, 0x1e, 0x90, 0xdd, 0x21, 0xd8,
	0x25, 0x1a, 0x30, 0xa0, 0xc5, 0xf1, 0xa8, 0xfc, 0x23, 0x44, 0x2d, 0x84, 0x77, 0x63, 0xea, 0xa8,
	0xd8, 0x13, 0xbe, 0x14, 0x8c, 0x8b, 0x96, 0xd7, 0x0c, 0xe1, 0x4b, 0x0f, 0x04, 0xbd, 0xb0, 0x17,
	0x2a, 0x56, 0x6d, 0xda, 0xc5, 0x43, 0xe9, 0x6d, 0xa2, 0xe0, 0x8f, 0x08, 0x99, 0x69, 0x33, 0xcf,
	0x27, 0x81, 0x3d, 0x6f, 0xb6, 0xe4, 0xfc, 0xc7, 0x96, 0x1a, 0x42, 0xbb, 0xb9, 0x88, 0xb0, 0x4f,
	0x02, 0xfc, 0x0a, 0xcd, 0x12, 0xc6, 0x42, 0x50, 0xca, 0xc6, 0x86, 0x85, 0xff, 0x5e, 0x97, 0x17,
	0x2e, 0x48, 0x27, 0xd8, 0xa9, 0xc6, 0x42, 0xd5, 0x4d, 0x2c, 0x78, 0x15, 0xcd, 0xb4, 0x49, 0xa0,
	0x81, 0xd9, 0xc5, 0x8a, 0x55, 0xcb, 0xba, 0xf1, 0x0a, 0x07, 0xa8, 0xd8, 0xe1, 0x62, 0xe2, 0x6c,
	0x96, 0x1f, 0x61, 0x60, 0x85, 0x0e, 0x17, 0x63, 0x47, 0x33, 0x48, 0x23, 0xfd, 0x89, 0xb4, 0x95,
	0x47, 0x49, 0x23, 0xfd, 0xbb, 0x69, 0x47, 0xd3, 0xd9, 0xc5, 0xa5, 0xa5, 0xfd, 0xf7, 0x97, 0x37,
	0x25, 0xeb, 0xea, 0xa6, 0x64, 0xfd, 0xb9, 0x29, 0x59, 0xdf, 0x6f, 0x4b, 0x99, 0xab, 0xdb, 0x52,
	0xe6, 0xd7, 0x6d, 0x29, 0x73, 0xf2, 0x3a, 0x15, 0xf4, 0xd9, 0xfc, 0x0e, 0x9b, 0x1f, 0x88, 0xaf,
	0xea, 0xf1, 0x8d, 0x7c, 0xf6, 0xa6, 0xde, 0x1f, 0x5d, 0xcb, 0x26, 0xd7, 0x9f, 0x31, 0x17, 0xec,
	0xf6, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xeb, 0x33, 0x2e, 0xd1, 0x09, 0x06, 0x00, 0x00,
}

func (m *HostZone) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HostZone) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *HostZone) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.MaxRedemptionRate.Size()
		i -= size
		if _, err := m.MaxRedemptionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintHostZone(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1
	i--
	dAtA[i] = 0xaa
	{
		size := m.MinRedemptionRate.Size()
		i -= size
		if _, err := m.MinRedemptionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintHostZone(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1
	i--
	dAtA[i] = 0xa2
	if m.Halted {
		i--
		if m.Halted {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x98
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x92
	}
	if len(m.Bech32Prefix) > 0 {
		i -= len(m.Bech32Prefix)
		copy(dAtA[i:], m.Bech32Prefix)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.Bech32Prefix)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x8a
	}
	if m.RedemptionAccount != nil {
		{
			size, err := m.RedemptionAccount.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHostZone(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x82
	}
	if m.UnbondingFrequency != 0 {
		i = encodeVarintHostZone(dAtA, i, uint64(m.UnbondingFrequency))
		i--
		dAtA[i] = 0x70
	}
	{
		size := m.StakedBal.Size()
		i -= size
		if _, err := m.StakedBal.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintHostZone(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x6a
	if len(m.TransferChannelId) > 0 {
		i -= len(m.TransferChannelId)
		copy(dAtA[i:], m.TransferChannelId)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.TransferChannelId)))
		i--
		dAtA[i] = 0x62
	}
	{
		size := m.RedemptionRate.Size()
		i -= size
		if _, err := m.RedemptionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintHostZone(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x5a
	{
		size := m.LastRedemptionRate.Size()
		i -= size
		if _, err := m.LastRedemptionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintHostZone(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x52
	if len(m.HostDenom) > 0 {
		i -= len(m.HostDenom)
		copy(dAtA[i:], m.HostDenom)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.HostDenom)))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.IbcDenom) > 0 {
		i -= len(m.IbcDenom)
		copy(dAtA[i:], m.IbcDenom)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.IbcDenom)))
		i--
		dAtA[i] = 0x42
	}
	if m.DelegationAccount != nil {
		{
			size, err := m.DelegationAccount.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHostZone(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x3a
	}
	if m.FeeAccount != nil {
		{
			size, err := m.FeeAccount.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHostZone(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x32
	}
	if m.WithdrawalAccount != nil {
		{
			size, err := m.WithdrawalAccount.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintHostZone(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if len(m.BlacklistedValidators) > 0 {
		for iNdEx := len(m.BlacklistedValidators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BlacklistedValidators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintHostZone(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Validators) > 0 {
		for iNdEx := len(m.Validators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Validators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintHostZone(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.ConnectionId) > 0 {
		i -= len(m.ConnectionId)
		copy(dAtA[i:], m.ConnectionId)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.ConnectionId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ChainId) > 0 {
		i -= len(m.ChainId)
		copy(dAtA[i:], m.ChainId)
		i = encodeVarintHostZone(dAtA, i, uint64(len(m.ChainId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintHostZone(dAtA []byte, offset int, v uint64) int {
	offset -= sovHostZone(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *HostZone) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ChainId)
	if l > 0 {
		n += 1 + l + sovHostZone(uint64(l))
	}
	l = len(m.ConnectionId)
	if l > 0 {
		n += 1 + l + sovHostZone(uint64(l))
	}
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovHostZone(uint64(l))
		}
	}
	if len(m.BlacklistedValidators) > 0 {
		for _, e := range m.BlacklistedValidators {
			l = e.Size()
			n += 1 + l + sovHostZone(uint64(l))
		}
	}
	if m.WithdrawalAccount != nil {
		l = m.WithdrawalAccount.Size()
		n += 1 + l + sovHostZone(uint64(l))
	}
	if m.FeeAccount != nil {
		l = m.FeeAccount.Size()
		n += 1 + l + sovHostZone(uint64(l))
	}
	if m.DelegationAccount != nil {
		l = m.DelegationAccount.Size()
		n += 1 + l + sovHostZone(uint64(l))
	}
	l = len(m.IbcDenom)
	if l > 0 {
		n += 1 + l + sovHostZone(uint64(l))
	}
	l = len(m.HostDenom)
	if l > 0 {
		n += 1 + l + sovHostZone(uint64(l))
	}
	l = m.LastRedemptionRate.Size()
	n += 1 + l + sovHostZone(uint64(l))
	l = m.RedemptionRate.Size()
	n += 1 + l + sovHostZone(uint64(l))
	l = len(m.TransferChannelId)
	if l > 0 {
		n += 1 + l + sovHostZone(uint64(l))
	}
	l = m.StakedBal.Size()
	n += 1 + l + sovHostZone(uint64(l))
	if m.UnbondingFrequency != 0 {
		n += 1 + sovHostZone(uint64(m.UnbondingFrequency))
	}
	if m.RedemptionAccount != nil {
		l = m.RedemptionAccount.Size()
		n += 2 + l + sovHostZone(uint64(l))
	}
	l = len(m.Bech32Prefix)
	if l > 0 {
		n += 2 + l + sovHostZone(uint64(l))
	}
	l = len(m.Address)
	if l > 0 {
		n += 2 + l + sovHostZone(uint64(l))
	}
	if m.Halted {
		n += 3
	}
	l = m.MinRedemptionRate.Size()
	n += 2 + l + sovHostZone(uint64(l))
	l = m.MaxRedemptionRate.Size()
	n += 2 + l + sovHostZone(uint64(l))
	return n
}

func sovHostZone(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozHostZone(x uint64) (n int) {
	return sovHostZone(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *HostZone) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHostZone
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: HostZone: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HostZone: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConnectionId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConnectionId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validators = append(m.Validators, &Validator{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlacklistedValidators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BlacklistedValidators = append(m.BlacklistedValidators, &Validator{})
			if err := m.BlacklistedValidators[len(m.BlacklistedValidators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WithdrawalAccount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.WithdrawalAccount == nil {
				m.WithdrawalAccount = &ICAAccount{}
			}
			if err := m.WithdrawalAccount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeeAccount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.FeeAccount == nil {
				m.FeeAccount = &ICAAccount{}
			}
			if err := m.FeeAccount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegationAccount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.DelegationAccount == nil {
				m.DelegationAccount = &ICAAccount{}
			}
			if err := m.DelegationAccount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IbcDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HostDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.HostDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastRedemptionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.LastRedemptionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RedemptionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RedemptionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TransferChannelId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TransferChannelId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 13:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakedBal", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.StakedBal.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 14:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UnbondingFrequency", wireType)
			}
			m.UnbondingFrequency = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UnbondingFrequency |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 16:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RedemptionAccount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.RedemptionAccount == nil {
				m.RedemptionAccount = &ICAAccount{}
			}
			if err := m.RedemptionAccount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 17:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bech32Prefix", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bech32Prefix = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 18:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 19:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Halted", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Halted = bool(v != 0)
		case 20:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinRedemptionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinRedemptionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 21:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxRedemptionRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHostZone
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthHostZone
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MaxRedemptionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipHostZone(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthHostZone
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipHostZone(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowHostZone
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHostZone
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthHostZone
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupHostZone
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthHostZone
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthHostZone        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowHostZone          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupHostZone = fmt.Errorf("proto: unexpected end of group")
)