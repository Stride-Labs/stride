package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	config "github.com/Stride-Labs/stride/v27/cmd/strided/config"
	icacallbacktypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
)

func FilterDepositRecords(arr []recordstypes.DepositRecord, condition func(recordstypes.DepositRecord) bool) (ret []recordstypes.DepositRecord) {
	for _, elem := range arr {
		if condition(elem) {
			ret = append(ret, elem)
		}
	}
	return ret
}

func Int64ToCoinString(amount int64, denom string) string {
	return strconv.FormatInt(amount, 10) + denom
}

func ValidateAdminAddress(address string) error {
	if !Admins[address] {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "address (%s) is not an admin", address)
	}
	return nil
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func StringMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func Int32MapKeys[V any](m map[int32]V) []int32 {
	keys := make([]int32, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func Uint64MapKeys[V any](m map[uint64]V) []uint64 {
	keys := make([]uint64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// Converts from uint64 -> int64 with a panic check for overflow
// This should only be used on values where it is known that overflow
// is not possible (e.g. params, block times, etc.), as in those scenarios
// we want to make sure we don't silently fail
func UintToInt(u uint64) int64 {
	if u > math.MaxInt64 {
		panic(fmt.Sprintf("uint64 value %d too large for int64", u))
	}
	return int64(u)
}

// Converts from int64 -> uint64 with a panic check for underflow
// This should only be used on values where it is known that underflow
// is not possible (e.g. params, block times, etc.), as in those scenarios
// we want to make sure we don't silently fail
func IntToUint(i int64) uint64 {
	if i < 0 {
		panic(fmt.Sprintf("int64 value %d is negative and can't be converted to uint64", i))
	}
	return uint64(i)
}

//==============================  ADDRESS VERIFICATION UTILS  ================================
// ref: https://github.com/cosmos/cosmos-sdk/blob/b75c2ebcfab1a6b535723f1ac2889a2fc2509520/types/address.go#L177

var errBech32EmptyAddress = errors.New("decoding Bech32 address failed: must provide a non empty address")

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errBech32EmptyAddress
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}

// VerifyAddressFormat verifies that the provided bytes form a valid address
// according to the default address rules or a custom address verifier set by
// GetConfig().SetAddressVerifier().
// TODO make an issue to get rid of global Config
// ref: https://github.com/cosmos/cosmos-sdk/issues/9690
func VerifyAddressFormat(bz []byte) error {
	verifier := func(bz []byte) error {
		n := len(bz)
		// Base accounts are length 20, module/ICA accounts are length 32
		if n == 20 || n == 32 {
			return nil
		}
		return fmt.Errorf("incorrect address length %d", n)
	}
	if verifier(bz) != nil {
		return verifier(bz)
	}

	if len(bz) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
	}

	if len(bz) > address.MaxAddrLen {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bz))
	}

	return nil
}

// AccAddress a wrapper around bytes meant to represent an account address.
// When marshaled to a string or JSON, it uses Bech32.
type AccAddress []byte

// AccAddressFromBech32 creates an AccAddress from a Bech32 string.
func AccAddressFromBech32(address string, bech32prefix string) (addr AccAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return AccAddress{}, errors.New("empty address string is not allowed")
	}

	bz, err := GetFromBech32(address, bech32prefix)
	if err != nil {
		return nil, err
	}

	err = VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return AccAddress(bz), nil
}

// ==============================  AIRDROP UTILS  ================================
// max64 returns the maximum of its inputs.
func Max64(i, j int64) int64 {
	if i > j {
		return i
	}
	return j
}

// Min64 returns the minimum of its inputs.
func Min64(i, j int64) int64 {
	if i < j {
		return i
	}
	return j
}

// Compute coin amount for specific period using linear vesting calculation algorithm.
func GetVestedCoinsAt(vAt int64, vStart int64, vLength int64, vCoins sdk.Coins) sdk.Coins {
	var vestedCoins sdk.Coins

	if vAt < 0 || vStart < 0 || vLength < 0 {
		return sdk.Coins{}
	}

	vEnd := vStart + vLength
	if vAt <= vStart {
		return sdk.Coins{}
	} else if vAt >= vEnd {
		return vCoins
	}

	// calculate the vesting scalar
	portion := sdk.NewDec(vAt - vStart).Quo(sdk.NewDec(vLength))

	for _, ovc := range vCoins {
		vestedAmt := sdk.NewDec(ovc.Amount.Int64()).Mul(portion).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins
}

// check string array inclusion
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Convert any bech32 to stride address
func ConvertAddressToStrideAddress(address string) string {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return ""
	}

	bech32Addr, err := bech32.ConvertAndEncode(config.Bech32PrefixAccAddr, bz)
	if err != nil {
		return ""
	}

	return bech32Addr
}

// Returns a log string with a chainId and tab as the prefix
// Ex:
//
//	| COSMOSHUB-4   |   string
func LogWithHostZone(chainId string, s string, a ...any) string {
	msg := fmt.Sprintf(s, a...)
	return fmt.Sprintf("|   %-13s |  %s", strings.ToUpper(chainId), msg)
}

// Returns a log string with a chain Id and callback as a prefix
// callbackType is either ICACALLBACK or ICQCALLBACK
// Format:
//
//	|   CHAIN-ID    |  {CALLBACK_ID} {CALLBACK_TYPE}  |  string
func logCallbackWithHostZone(chainId string, callbackId string, callbackType string, s string, a ...any) string {
	msg := fmt.Sprintf(s, a...)
	return fmt.Sprintf("|   %-13s |  %s %s  |  %s", strings.ToUpper(chainId), strings.ToUpper(callbackId), callbackType, msg)
}

// Returns a log string with a chain Id and icacallback as a prefix
// Ex:
//
//	| COSMOSHUB-4   |  DELEGATE ICACALLBACK  |  string
func LogICACallbackWithHostZone(chainId string, callbackId string, s string, a ...any) string {
	return logCallbackWithHostZone(chainId, callbackId, "ICACALLBACK", s, a...)
}

// Returns a log string with a chain Id and icacallback as a prefix, and status of the callback
// Ex:
//
//	| COSMOSHUB-4   |  DELEGATE ICACALLBACK  |  ICA SUCCESS, Packet: ...
func LogICACallbackStatusWithHostZone(chainId string, callbackId string, status icacallbacktypes.AckResponseStatus, packet channeltypes.Packet) string {
	var statusMsg string
	switch status {
	case icacallbacktypes.AckResponseStatus_SUCCESS:
		statusMsg = "ICA SUCCESSFUL"
	case icacallbacktypes.AckResponseStatus_TIMEOUT:
		statusMsg = "ICA TIMEOUT"
	default:
		statusMsg = "ICA FAILED (ack error)"
	}
	return logCallbackWithHostZone(chainId, callbackId, "ICACALLBACK", "%s, Packet: %+v", statusMsg, packet)
}

// Returns a log string with a chain Id and icqcallback as a prefix
// Ex:
//
//	| COSMOSHUB-4   |  WITHDRAWALHOSTBALANCE ICQCALLBACK  |  string
func LogICQCallbackWithHostZone(chainId string, callbackId string, s string, a ...any) string {
	return logCallbackWithHostZone(chainId, callbackId, "ICQCALLBACK", s, a...)
}

// Returns a log header string with a dash padding on either side
// Ex:
//
//	------------------------------ string ------------------------------
func LogHeader(s string, a ...any) string {
	lineLength := 120
	header := fmt.Sprintf(s, a...)
	pad := strings.Repeat("-", (lineLength-len(header))/2)
	return fmt.Sprintf("%s %s %s", pad, header, pad)
}

// Logs a module's migration info
func LogModuleMigration(ctx sdk.Context, versionMap module.VersionMap, moduleName string) {
	currentVersion := versionMap[moduleName]
	ctx.Logger().Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, currentVersion-1, currentVersion))
}

// isIBCToken checks if the token came from the IBC module
// Each IBC token starts with an ibc/ denom, the check is rather simple
func IsIBCToken(denom string) bool {
	return strings.HasPrefix(denom, "ibc/")
}

// Returns the stDenom from a native denom by appending a st prefix
func StAssetDenomFromHostZoneDenom(hostZoneDenom string) string {
	return "st" + hostZoneDenom
}

// Returns the native denom from an stDenom by removing the st prefix
func HostZoneDenomFromStAssetDenom(stAssetDenom string) string {
	return stAssetDenom[2:]
}

// Verifies a tx hash is valid
func VerifyTxHash(txHash string) (err error) {
	if txHash == "" {
		return errorsmod.Wrapf(sdkerrors.ErrTxDecode, "tx hash is empty")
	}
	_, err = hex.DecodeString(txHash)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrTxDecode, "tx hash is invalid %s", txHash)
	}
	return nil
}
