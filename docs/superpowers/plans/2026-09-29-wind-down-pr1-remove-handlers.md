# Wind-Down PR 1: Remove Tx Handlers — Implementation Plan

> Note: stakeibc `ResumeHostZone` is listed as removed throughout this plan; it was restored after implementation (kept per spec §5).

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Branching:** PR 1 branches off `wind-down-design-consolidation`. Each later PR branches
> off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on) and the PRs are
> implemented and merged strictly in order: 1, 2, 3, 4, 5, 6. Branch names:
> `wind-down-pr1-remove-handlers`, `wind-down-pr2-freeze-by-code`,
> `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`, `wind-down-pr5-sweep-tx`,
> `wind-down-pr6-release-gate`.
> The Go module path stays `github.com/Stride-Labs/stride/v34` in every PR; the bump to
> `/v35` is a manual step after all six land and is out of scope for every plan.

**Goal:** Remove the transaction handlers the protocol no longer needs (spec §5) so that a submitted `MsgLiquidStake`, `MsgRedeemStake`, etc. fails with "can't route message", while every historical transaction that contains one of those messages still decodes.

**Architecture:** For each removed message, delete the `rpc` line from the module's `Msg` service, regenerate `tx.pb.go`, delete the msg-server method, its CLI command and the tests that went through the handler. Message types, their `RegisterImplementations` entries and their amino names stay untouched (four stakeibc types are added to `RegisterImplementations` because today they are registered only through the service descriptor that is about to lose them). Keeper logic that something still calls survives: stakeibc's `LiquidStake` moves from the msg server onto the `Keeper` because autopilot, the community pool and the reward collector call it; `RedeemStake` and `RegisterHostZone` are already keeper methods. Two new tests guard the outcome: a router/reflection test proving no handler exists for the removed messages, and a decode test over three real mainnet transactions.

**Tech Stack:** Go 1.25, cosmos-sdk v0.54.3, ibc-go v11.2.0, gogoproto via `make proto-gen` (docker + buf), testify suites (`apptesting.AppTestHelper`).

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` §5 ("Message handlers removed"), §11, §12 item 1, §13 "Removals (PR 1)".
- Removed handlers, exactly (§5): stakeibc `LiquidStake`, `LSMLiquidStake`, `RedeemStake`, `RegisterHostZone`, `CreateTradeRoute`, `UpdateTradeRoute`, `DeleteTradeRoute`, `SetCommunityPoolRebate`, `ToggleTradeController`, `RebalanceValidators`, `ClearBalance`, `ResumeHostZone`; staketia and stakedym `LiquidStake` (stub in staketia, live in stakedym), `RedeemStake`, `ResumeHostZone`; icaoracle `AddOracle`, `InstantiateOracle`; icqoracle `RegisterTokenPriceQuery`, `RemoveTokenPriceQuery`; auction `PlaceBid`, `CreateAuction`, `UpdateAuction`; airdrop all seven; claim all four.
- Kept, exactly: stakeibc `ClaimUndelegatedTokens` (keeper/claim.go), `RestoreInterchainAccount`, `CloseDelegationChannel`, `UpdateValidatorSharesExchRate`, `CalibrateDelegation`, `AddValidators`, `ChangeValidatorWeight`, `DeleteValidator`, `UpdateInnerRedemptionRateBounds`, `UpdateHostZoneParams`, `DeprecateHostZone`; icaoracle `RestoreOracleICA`, `ToggleOracle`, `RemoveOracle`; icqoracle `UpdateParams`; every staketia/stakedym operator message.
- Every removed message keeps its Go type, its `registry.RegisterImplementations((*sdk.Msg)(nil), ...)` entry and its `legacy.RegisterAminoMsg` name. Never delete a `message_*.go` type file that a non-deleted file references (`message_liquid_stake.go`, `message_redeem_stake.go`, `message_register_host_zone.go`, `message_lsm_liquid_stake.go` all stay).
- Keeper functions stay wherever something still calls them: stakeibc `Keeper.LiquidStake` (new), `Keeper.RedeemStake`, `Keeper.RegisterHostZone`, stakedym `Keeper.LiquidStake` (fee distribution). Per §5, staketia's and stakedym's `Keeper.RedeemStake` go with their handlers; the helpers they leave unreferenced (`BurnRedeemedStTokens` etc.) are the §12 follow-up cleanup, not this PR.
- Module path stays `github.com/Stride-Labs/stride/v34`. `make proto-gen` may rewrite descriptor bytes in other `*.pb.go`; commit only the `tx.pb.go` of the module whose `tx.proto` changed and `git checkout --` the rest.
- The tree must compile after every commit. Pre-existing failure to ignore: `utils` `TestCreateModuleAccount` fails on the base branch too.
- Commit messages end with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`.
- Scripts under `dockernet/` and `scripts/local-to-mainnet/` that call removed CLI commands (`liquid-stake`, `redeem-stake`, `register-host-zone`, `lsm-liquid-stake`, `set-rebate`, `toggle-trade-controller`, airdrop/claim scripts) are left alone; they are dev tooling and out of scope.

---

## File map

| Path | Responsibility in this PR |
|---|---|
| `x/stakeibc/types/codec.go` | Add `MsgLSMLiquidStake`, `MsgCreateTradeRoute`, `MsgDeleteTradeRoute`, `MsgUpdateTradeRoute` to `RegisterImplementations` |
| `app/testdata/historical_txs.json` (new) | Three real mainnet tx bytes (LiquidStake, RedeemStake, LSMLiquidStake) with each type URL and amino name |
| `app/historical_tx_decode_test.go` (new) | Decode test over the fixture: protobuf decode, type URL, amino JSON carrying the registered name |
| `app/removed_handlers_test.go` (new) | Router + reflection guard for every removed message |
| `x/stakeibc/keeper/liquid_stake.go` (new) | `Keeper.LiquidStake` (moved verbatim from the msg server) |
| `x/stakeibc/keeper/msg_server.go` | Delete 12 handlers |
| `x/stakeibc/keeper/community_pool.go`, `reward_allocation.go`, `x/autopilot/keeper/liquidstake.go`, `redeem_stake.go` | Call the keeper instead of `NewMsgServerImpl` |
| `x/stakeibc/handler.go` | Delete `NewMessageHandler`; keep `NewStakeibcProposalHandler` |
| `proto/stride/<module>/tx.proto` + `x/<module>/types/tx.pb.go` | rpc deletions, eight modules |
| `x/<module>/keeper/msg_server.go`, `x/<module>/client/cli/tx*.go`, tests | Handler, CLI and handler-test deletions per module |

---

### Task 1: Register the four service-descriptor-only stakeibc types and add the historical decode test

**Files:**
- Modify: `x/stakeibc/types/codec.go:41-61`
- Create: `app/testdata/historical_txs.json`
- Create: `app/historical_tx_decode_test.go`

**Interfaces:**
- Consumes: `apptesting.AppTestHelper` (`s.App.TxDecode(bz)`, `s.App.LegacyAmino()`), `sdk.MsgTypeURL`.
- Produces: the fixture file and the decode test that every later task must keep green.
- Review: yes (this is the test that protects chain history).

- [ ] **Step 1: Write the fixture with three real mainnet transactions**

The three transactions below were fetched from `https://stride-rpc.polkachu.com` on 2026-09-29 (the `tx` field of `/tx?hash=0x<HASH>` is the raw signed tx, base64). Create `app/testdata/historical_txs.json`:

```json
[
  {
    "name": "stakeibc MsgLiquidStake",
    "hash": "DBDD161AF1D0B9FC158ADA2CCCDC50F965F289F181D47AA5C3B5DEDDC9410B7A",
    "height": 40815423,
    "type_url": "/stride.stakeibc.MsgLiquidStake",
    "amino_name": "stakeibc/MsgLiquidStake",
    "tx_base64": "CmUKYwofL3N0cmlkZS5zdGFrZWliYy5Nc2dMaXF1aWRTdGFrZRJACi1zdHJpZGUxZmU0bXZoNWY3dDA4MzBmd2N1Y3V1bGo1dzIydnh0ZWNobHh6dTQSCDE5OTI2MzAwGgV1YmFuZBJzClEKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiEDgP8WjsoBIbpROHVmlEY2kGJFnggb/2GUENv9j64jFqISBAoCCAEYqgQSHgoYCgVzdGluahIPNjE4ODQwNjAwMDAwMDAwENr6NRpAB/z4hQnq4bYOn7QmWvYFx2F+3gGXgHJyixIFGjoGP1ASTcFF3oYHXNL/qFxT/89ZT+kstK4BZvYX+CqyvC5a/w=="
  },
  {
    "name": "stakeibc MsgRedeemStake",
    "hash": "7D90D8E247D833F0F4486B74AB41B9F5BA96B6C4CC46C703513E553F816C6229",
    "height": 40824022,
    "type_url": "/stride.stakeibc.MsgRedeemStake",
    "amino_name": "stakeibc/MsgRedeemStake",
    "tx_base64": "CpYBCpMBCh8vc3RyaWRlLnN0YWtlaWJjLk1zZ1JlZGVlbVN0YWtlEnAKLXN0cmlkZTE3cWNrZnJ4dWxlM3BlZnhtMHpwMnNrdjZrazdoc3hjM3hkdHhhcBIHMjA3MzY3MBoJb3Ntb3Npcy0xIitvc21vMTdxY2tmcnh1bGUzcGVmeG0wenAyc2t2NmtrN2hzeGMzZGFjMmxsEmcKUApGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQPK4L+sCuEUCEcbeS7NVu9F/BihArMFHksISmXx3rgzMhIECgIIfxhVEhMKDQoFdXN0cmQSBDE1MTQQ4rwSGkDKTehUP2Gmltpig0j2I9d6PwiZfmBB9Q6Hyc8Dx7d3JhvTji6QeebawQT+olPHlJDXGvUI7Ihhd0AvXPE1aHXy"
  },
  {
    "name": "stakeibc MsgLSMLiquidStake",
    "hash": "4AEA3294520649BA0DD55E6F7262D2E6608A7F592460D75BE34D3B2CDB9ACC97",
    "height": 40824588,
    "type_url": "/stride.stakeibc.MsgLSMLiquidStake",
    "amino_name": "stakeibc/MsgLSMLiquidStake",
    "tx_base64": "CqgBCqUBCiIvc3RyaWRlLnN0YWtlaWJjLk1zZ0xTTUxpcXVpZFN0YWtlEn8KLXN0cmlkZTF1YXV2Y2M4N3d2ZXk1ZGNqN2tnd3NweDV6d2pwY2ZlNTNqc3NrdRIIMjUwMDAwMDAaRGliYy84NDRFOUZBOEEwMDY0MzcxQkY4NjgwRjc3QzI1OTc1N0ZDRDM5NDVDMzczNTI2QkEyMjdDNjg5NTJDQjk4Nzk4EmgKUQpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQN8FlofzxaPR7INHxiHvyvGdo0LUfFItHGsQCxtzSxL4xIECgIIARi8AhITCg0KBXVzdHJkEgQxOTgxEPzteBpAM2kRXPXIb6+HPW9X+WCsJUSTvg3GcbpmXHdVTjQ0MYkU7LARE/HRNWPpaItblBvUuV2X952Q2i6kakXLlwGUOg=="
  }
]
```

To refresh or extend the fixture (the RPC rate-limits aggressively; wait ~45 s between calls):

```bash
# find the newest tx for a message family
curl -s -A 'Mozilla/5.0' "https://stride-rpc.polkachu.com/tx_search?query=%22message.action%3D%27/stride.stakeibc.MsgLiquidStake%27%22&per_page=1&order_by=%22desc%22" | python3 -c 'import sys,json; t=json.load(sys.stdin)["result"]["txs"][0]; print(t["hash"], t["height"]); print(t["tx"])'
# or fetch a known hash
curl -s -A 'Mozilla/5.0' "https://stride-rpc.polkachu.com/tx?hash=0x<HASH>" | python3 -c 'import sys,json; r=json.load(sys.stdin)["result"]; print(r["height"]); print(r["tx"])'
```

- [ ] **Step 2: Write the failing decode test**

Create `app/historical_tx_decode_test.go`:

```go
package app_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
)

// historicalTx is one entry of app/testdata/historical_txs.json: a real signed mainnet
// transaction whose message type no longer has a handler after the wind-down removals.
type historicalTx struct {
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	Height   int64  `json:"height"`
	TypeURL   string `json:"type_url"`
	AminoName string `json:"amino_name"` // the name registered in the module's RegisterCodec
	TxBase64  string `json:"tx_base64"`
}

type HistoricalTxDecodeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *HistoricalTxDecodeTestSuite) SetupTest() {
	s.Setup()
}

func TestHistoricalTxDecodeTestSuite(t *testing.T) {
	suite.Run(t, new(HistoricalTxDecodeTestSuite))
}

// Removing a message's rpc removes its handler; it must not remove the ability to decode
// the chain's history. The message types stay registered in the interface registry and
// with amino, and this test is what fails if a registration is dropped by mistake (the
// first dry run of these removals broke `strided q tx` on old hashes exactly that way).
func (s *HistoricalTxDecodeTestSuite) TestHistoricalTxsStillDecode() {
	raw, err := os.ReadFile("testdata/historical_txs.json")
	s.Require().NoError(err, "fixture must exist")

	var fixtures []historicalTx
	s.Require().NoError(json.Unmarshal(raw, &fixtures))
	s.Require().NotEmpty(fixtures)

	for _, fixture := range fixtures {
		s.Run(fixture.Name, func() {
			txBytes, err := base64.StdEncoding.DecodeString(fixture.TxBase64)
			s.Require().NoError(err)

			// Protobuf decode through the app's tx decoder (what `strided q tx` uses)
			tx, err := s.App.TxDecode(txBytes)
			s.Require().NoError(err, "tx %s must still decode", fixture.Hash)

			msgs := tx.GetMsgs()
			s.Require().Len(msgs, 1)
			s.Require().Equal(fixture.TypeURL, sdk.MsgTypeURL(msgs[0]))

			// Legacy amino JSON rendering must still carry the registered name. NoError alone
			// proves nothing: go-amino's MarshalJSON on a concrete type that lost its
			// RegisterAminoMsg line just omits the {"type": ...} wrapper and returns no error,
			// so the name check is the assertion that fails when a registration is dropped
			aminoJson, err := s.App.LegacyAmino().MarshalJSON(msgs[0])
			s.Require().NoError(err, "amino JSON for %s", fixture.TypeURL)
			s.Require().Contains(string(aminoJson), fmt.Sprintf(`"type":"%s"`, fixture.AminoName),
				"amino JSON for %s must carry its registered name", fixture.TypeURL)
		})
	}
}
```

- [ ] **Step 3: Run the test to see it pass on the current registrations**

Run: `go test ./app/ -run 'TestHistoricalTxDecodeTestSuite' -v`
Expected: PASS (the types are registered today; the test is the tripwire for the rest of the PR). If it fails, the fixture bytes are wrong: refetch them with the commands in Step 1.

- [ ] **Step 4: Add the four missing `RegisterImplementations` entries**

In `x/stakeibc/types/codec.go`, change the `RegisterImplementations((*sdk.Msg)(nil), ...)` block to:

```go
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgLSMLiquidStake{},
		&MsgClearBalance{},
		&MsgRegisterHostZone{},
		&MsgRedeemStake{},
		&MsgClaimUndelegatedTokens{},
		&MsgRebalanceValidators{},
		&MsgAddValidators{},
		&MsgChangeValidatorWeights{},
		&MsgDeleteValidator{},
		&MsgRestoreInterchainAccount{},
		&MsgCloseDelegationChannel{},
		&MsgUpdateValidatorSharesExchRate{},
		&MsgCalibrateDelegation{},
		&MsgUpdateInnerRedemptionRateBounds{},
		&MsgResumeHostZone{},
		&MsgCreateTradeRoute{},
		&MsgDeleteTradeRoute{},
		&MsgUpdateTradeRoute{},
		&MsgSetCommunityPoolRebate{},
		&MsgToggleTradeController{},
		&MsgUpdateHostZoneParams{},
		&MsgDeprecateHostZone{},
	)
```

Add a comment above the block:

```go
	// Every message type stays registered here even after its rpc is removed (v35 wind-down):
	// the interface registry is what decodes historical transactions, and once the rpc is gone
	// msgservice.RegisterMsgServiceDesc no longer registers the type for us
```

- [ ] **Step 5: Verify build and test**

Run: `go build ./... && go test ./app/ -run 'TestHistoricalTxDecodeTestSuite' -v`
Expected: build ok, PASS.

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/types/codec.go app/testdata/historical_txs.json app/historical_tx_decode_test.go
git commit -m "test(app): historical tx decode fixture; register the four service-desc-only stakeibc msg types

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 2: Move stakeibc `LiquidStake` onto the keeper and rewrite its callers and tests

**Files:**
- Create: `x/stakeibc/keeper/liquid_stake.go`
- Modify: `x/stakeibc/keeper/msg_server.go:213-309` (handler becomes a one-line delegate; deleted in Task 4)
- Modify: `x/stakeibc/keeper/community_pool.go:176-190,212-226`
- Modify: `x/stakeibc/keeper/reward_allocation.go:47`
- Modify: `x/autopilot/keeper/liquidstake.go:90-96`
- Modify: `x/autopilot/keeper/redeem_stake.go:75-82`
- Modify: `x/stakeibc/keeper/msg_server_test.go` (the 14 `TestLiquidStake_*` tests, lines 639-836, plus `LiquidStakeState`, `LiquidStakeTestCase` and `SetupLiquidStake`, lines 567-637)

**Interfaces:**
- Produces: `func (k Keeper) LiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error)`.
- Consumes: existing `func (k Keeper) RedeemStake(ctx sdk.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error)` in `x/stakeibc/keeper/redeem_stake.go:20`.
- Review: yes (moves the mint path; callers are autopilot and the community pool).

- [ ] **Step 1: Create `x/stakeibc/keeper/liquid_stake.go` with the handler body moved verbatim**

```go
package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
	epochtypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// LiquidStake exchanges native tokens for stTokens at the current redemption rate.
//
// The user-facing MsgLiquidStake handler was removed in the v35 wind-down; this keeper
// method is what autopilot (inbound IBC liquid stakes), the community pool staking flow and
// the reward collector still call. The native tokens must live on Stride with an IBC
// denomination before this function is called.
//
// WARNING: This function is invoked from the begin/end blocker in a way that does not revert
// partial state when an error is thrown (i.e. the execution is non-atomic). As a result, the
// validation steps are positioned at the top of the function, and logic that creates state
// changes (bank sends, mint) appears towards the end.
func (k Keeper) LiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	// Get the host zone from the base denom in the message (e.g. uatom)
	hostZone, err := k.GetHostZoneFromHostDenom(ctx, msg.HostDenom)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "no host zone found for denom (%s)", msg.HostDenom)
	}

	// Error immediately if the host zone is halted
	if hostZone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "halted host zone found for denom (%s)", msg.HostDenom)
	}

	// Get user and module account addresses
	liquidStakerAddress, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "user's address is invalid")
	}
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, *hostZone)
	if !rateIsSafe || (err != nil) {
		return nil, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, "HostZone: %s, err: %s", hostZone.ChainId, err.Error())
	}

	// Grab the deposit record that will be used for record keeping
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	depositRecord, found := k.RecordsKeeper.GetTransferDepositRecordByEpochAndChain(ctx, strideEpochTracker.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no deposit record for epoch (%d)", strideEpochTracker.EpochNumber)
	}

	// The tokens that are sent to the protocol are denominated in the ibc hash of the native token on stride (e.g. ibc/xxx)
	nativeDenom := hostZone.IbcDenom
	nativeCoin := sdk.NewCoin(nativeDenom, msg.Amount)
	if !types.IsIBCToken(nativeDenom) {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "denom is not an IBC token (%s)", nativeDenom)
	}

	// Confirm the user has a sufficient balance to execute the liquid stake
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, nativeDenom)
	if balance.IsLT(nativeCoin) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stAmount := (sdkmath.LegacyNewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Transfer the native tokens from the user to module account
	// Note: checkBlockedAddr=false because hostZoneDepositAddress is a module
	if err := utils.SafeSendCoins(false, k.bankKeeper, ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(nativeCoin)); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Mint the stTokens and transfer them to the user
	stDenom := types.StAssetDenomFromHostZoneDenom(msg.HostDenom)
	stCoin := sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to mint coins")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account", stCoin.String())
	}

	// Update the liquid staked amount on the deposit record
	depositRecord.Amount = depositRecord.Amount.Add(msg.Amount)
	k.RecordsKeeper.SetDepositRecord(ctx, *depositRecord)

	// Emit liquid stake event
	EmitSuccessfulLiquidStakeEvent(ctx, msg, *hostZone, stAmount)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return &types.MsgLiquidStakeResponse{StToken: stCoin}, nil
}
```

- [ ] **Step 2: Turn the msg-server handler into a delegate (keeps the tree compiling until Task 4)**

In `x/stakeibc/keeper/msg_server.go` replace lines 213-309 (the comment block and the whole `func (k msgServer) LiquidStake`) with:

```go
func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.LiquidStake(ctx, msg)
}
```

- [ ] **Step 3: Rewrite the three in-tree callers to the keeper**

`x/stakeibc/keeper/community_pool.go`, around line 180, replace

```go
	msgServer := NewMsgServerImpl(k)
	liquidStakeRequest := types.MsgLiquidStake{
		Creator:   hostZone.CommunityPoolStakeHoldingAddress,
		Amount:    nativeTokens.Amount,
		HostDenom: hostZone.HostDenom,
	}
	resp, err := msgServer.LiquidStake(ctx, &liquidStakeRequest)
```

with

```go
	liquidStakeRequest := types.MsgLiquidStake{
		Creator:   hostZone.CommunityPoolStakeHoldingAddress,
		Amount:    nativeTokens.Amount,
		HostDenom: hostZone.HostDenom,
	}
	resp, err := k.LiquidStake(ctx, &liquidStakeRequest)
```

and delete the `// TODO: Move LS function to keeper method instead of message server` line above it. Around line 216, replace

```go
	msgServer := NewMsgServerImpl(k)
	redeemStakeRequest := types.MsgRedeemStake{
		Creator:  hostZone.CommunityPoolRedeemHoldingAddress,
		Amount:   stTokens.Amount,
		HostZone: hostZone.ChainId,
		Receiver: hostZone.CommunityPoolReturnIcaAddress,
	}
	if _, err := msgServer.RedeemStake(ctx, &redeemStakeRequest); err != nil {
```

with

```go
	redeemStakeRequest := types.MsgRedeemStake{
		Creator:  hostZone.CommunityPoolRedeemHoldingAddress,
		Amount:   stTokens.Amount,
		HostZone: hostZone.ChainId,
		Receiver: hostZone.CommunityPoolReturnIcaAddress,
	}
	if _, err := k.RedeemStake(ctx, &redeemStakeRequest); err != nil {
```

and delete the `// TODO: Move Redeem function to keeper method instead of message server` line.

`x/stakeibc/keeper/reward_allocation.go:47`: replace `liquidStakeResp, err := NewMsgServerImpl(k).LiquidStake(ctx, msg)` with `liquidStakeResp, err := k.LiquidStake(ctx, msg)`.

`x/autopilot/keeper/liquidstake.go` around line 92: replace

```go
	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	msgResponse, err := msgServer.LiquidStake(
		ctx,
		msg,
	)
```

with

```go
	msgResponse, err := k.stakeibcKeeper.LiquidStake(ctx, msg)
```

`x/autopilot/keeper/redeem_stake.go` around line 77: replace

```go
	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	if _, err = msgServer.RedeemStake(ctx, msg); err != nil {
```

with

```go
	if _, err = k.stakeibcKeeper.RedeemStake(ctx, msg); err != nil {
```

Then check the autopilot keeper's `stakeibcKeeper` field: run `grep -n "stakeibcKeeper" x/autopilot/keeper/keeper.go`. If it is typed as `stakeibckeeper.Keeper` nothing else changes. If it is an interface in `x/autopilot/types/expected_keepers.go`, add the two methods to that interface:

```go
	LiquidStake(ctx sdk.Context, msg *stakeibctypes.MsgLiquidStake) (*stakeibctypes.MsgLiquidStakeResponse, error)
	RedeemStake(ctx sdk.Context, msg *stakeibctypes.MsgRedeemStake) (*stakeibctypes.MsgRedeemStakeResponse, error)
```

Remove the now-unused `stakeibckeeper` import from both autopilot files if the compiler reports it.

- [ ] **Step 4: Build**

Run: `go build ./... && go vet ./x/stakeibc/... ./x/autopilot/...`
Expected: no output (clean).

- [ ] **Step 5: Rewrite the fourteen LiquidStake handler tests to the keeper**

In `x/stakeibc/keeper/msg_server_test.go` the tests `TestLiquidStake_Successful` through `TestLiquidStake_HaltedZone` (14 tests, lines 639-836) call `s.GetMsgServer().LiquidStake(`. Rewrite every occurrence:

```bash
sed -i '' 's/s\.GetMsgServer()\.LiquidStake(/s.App.StakeibcKeeper.LiquidStake(/g' x/stakeibc/keeper/msg_server_test.go
```

Then move those tests, the `LiquidStakeState` and `LiquidStakeTestCase` types (lines 567-577; `LiquidStakeTestCase` embeds `LiquidStakeState`, so both move) and `SetupLiquidStake` (line 579) out of `msg_server_test.go` into a new `x/stakeibc/keeper/liquid_stake_test.go` (same package `keeper_test`, same imports as needed) so the tests live beside the keeper file. `sed` leaves `s.Ctx` as the first argument, which is already an `sdk.Context`, so no further edits are needed.

- [ ] **Step 6: Run the moved tests**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestLiquidStake' -v`
Expected: all 14 PASS.

Run: `go test ./x/autopilot/... ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/(TestOnRecvPacket_LiquidStake|TestOnRecvPacket_RedeemStake|TestLiquidStakeCommunityPoolTokens|TestRedeemCommunityPoolTokens|TestLiquidStakeRewardCollectorBalance)' -v`
Expected: PASS (these are the callers; if a test name above does not exist, run the whole autopilot package and `TestKeeperTestSuite` in stakeibc instead).

- [ ] **Step 7: Commit**

```bash
git add x/stakeibc/keeper/liquid_stake.go x/stakeibc/keeper/liquid_stake_test.go x/stakeibc/keeper/msg_server.go x/stakeibc/keeper/msg_server_test.go x/stakeibc/keeper/community_pool.go x/stakeibc/keeper/reward_allocation.go x/autopilot/keeper/liquidstake.go x/autopilot/keeper/redeem_stake.go x/autopilot/types/expected_keepers.go
git commit -m "refactor(stakeibc): move LiquidStake onto the keeper; autopilot, community pool and reward collector call the keeper

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 3: Point the RegisterHostZone and RedeemStake tests at the keeper

**Files:**
- Modify: `x/stakeibc/keeper/registration_test.go` (24 calls)
- Modify: `x/stakeibc/keeper/redeem_stake_test.go` (18 calls)

**Interfaces:**
- Consumes: `Keeper.RegisterHostZone(ctx, msg)` (`registration.go:26`), `Keeper.RedeemStake(ctx, msg)` (`redeem_stake.go:20`).
- Review: no.

- [ ] **Step 1: Rewrite the calls**

```bash
sed -i '' 's/s\.GetMsgServer()\.RegisterHostZone(/s.App.StakeibcKeeper.RegisterHostZone(/g' x/stakeibc/keeper/registration_test.go
sed -i '' 's/s\.GetMsgServer()\.RedeemStake(/s.App.StakeibcKeeper.RedeemStake(/g' x/stakeibc/keeper/redeem_stake_test.go
grep -c "GetMsgServer" x/stakeibc/keeper/registration_test.go x/stakeibc/keeper/redeem_stake_test.go
```

Expected grep output: `0` for both files.

- [ ] **Step 2: Run them**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/(TestRegisterHostZone|TestRedeemStake)' -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add x/stakeibc/keeper/registration_test.go x/stakeibc/keeper/redeem_stake_test.go
git commit -m "test(stakeibc): register-host-zone and redeem-stake tests call the keeper

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 4: Remove the twelve stakeibc handlers, their rpcs, CLI and handler tests

**Files:**
- Modify: `proto/stride/stakeibc/tx.proto:14-56`
- Regenerate: `x/stakeibc/types/tx.pb.go`
- Modify: `x/stakeibc/keeper/msg_server.go`
- Modify: `x/stakeibc/handler.go`
- Modify: `x/stakeibc/client/cli/tx.go`, `x/stakeibc/client/cli/tx_test.go`
- Modify: `x/stakeibc/keeper/msg_server_test.go`

**Interfaces:**
- Produces: `types.MsgServer` with eleven methods; `stakeibc.NewStakeibcProposalHandler` unchanged.
- Depends on: Tasks 1-3.
- Review: yes (large deletion; the reviewer checks that nothing listed as "kept" went with it).

- [ ] **Step 1: Edit the proto service**

Replace the `service Msg { ... }` block of `proto/stride/stakeibc/tx.proto` with:

```proto
// Msg defines the Msg service.
//
// v35 wind-down: the user and admin transactions that moved funds on behalf of users
// (liquid stake, redeem, register, trade routes, rebate, rebalance, clear balance, resume)
// have no rpc any more. Their message types remain below so historical transactions decode.
service Msg {
  option (cosmos.msg.v1.service) = true;

  rpc ClaimUndelegatedTokens(MsgClaimUndelegatedTokens)
      returns (MsgClaimUndelegatedTokensResponse);
  rpc AddValidators(MsgAddValidators) returns (MsgAddValidatorsResponse);
  rpc ChangeValidatorWeight(MsgChangeValidatorWeights)
      returns (MsgChangeValidatorWeightsResponse);
  rpc DeleteValidator(MsgDeleteValidator) returns (MsgDeleteValidatorResponse);
  rpc RestoreInterchainAccount(MsgRestoreInterchainAccount)
      returns (MsgRestoreInterchainAccountResponse);
  rpc CloseDelegationChannel(MsgCloseDelegationChannel)
      returns (MsgCloseDelegationChannelResponse);
  rpc UpdateValidatorSharesExchRate(MsgUpdateValidatorSharesExchRate)
      returns (MsgUpdateValidatorSharesExchRateResponse);
  rpc CalibrateDelegation(MsgCalibrateDelegation)
      returns (MsgCalibrateDelegationResponse);
  rpc UpdateInnerRedemptionRateBounds(MsgUpdateInnerRedemptionRateBounds)
      returns (MsgUpdateInnerRedemptionRateBoundsResponse);
  rpc UpdateHostZoneParams(MsgUpdateHostZoneParams)
      returns (MsgUpdateHostZoneParamsResponse);
  rpc DeprecateHostZone(MsgDeprecateHostZone)
      returns (MsgDeprecateHostZoneResponse);
}
```

Leave every `message Msg...` definition in the file exactly as it is.

- [ ] **Step 2: Regenerate and keep only stakeibc's generated file**

```bash
make proto-gen
git status --short | grep pb.go
git add x/stakeibc/types/tx.pb.go
files=$(git diff --name-only | grep 'pb.go$' | grep -v '^x/stakeibc/types/tx.pb.go$' || true)
[ -n "$files" ] && git checkout -- $files  # no-op when proto-gen churned nothing else
go build ./... 2>&1 | head -40
```

Expected: the build fails only with "msgServer does not implement types.MsgServer (unexported/extra method ...)"-style errors or `undefined: ...` errors in `msg_server.go`, `handler.go`, `client/cli/tx.go` and tests. Those are what the next steps delete.

- [ ] **Step 3: Delete the handlers from `x/stakeibc/keeper/msg_server.go`**

Line numbers below are from the base branch; those after `LiquidStake` shift up by about 90 lines once Task 2 shrinks that handler, so locate each function by name. Delete these functions (and the comment blocks directly above them): `RegisterHostZone` (lines 42-45), `RebalanceValidators` (156-164), `ClearBalance` (166-211), `LiquidStake` (the delegate from Task 2), `RedeemStake` (311-314), `LSMLiquidStake` (with its 19-line comment, 316-393), `CreateTradeRoute` (with its example-proposal comment, 360-482), `DeleteTradeRoute` (with comment, 463-518), `UpdateTradeRoute` (with comment, 500-536), `ResumeHostZone` (744-767), `SetCommunityPoolRebate` (770-803), `ToggleTradeController` (with comment, 805-847, the end of the file). Keep `NewMsgServerImpl`, `UpdateHostZoneParams`, `DeprecateHostZone`, `AddValidators`, `DeleteValidator`, `ChangeValidatorWeight`, `RestoreInterchainAccount`, `CloseDelegationChannel`, `UpdateValidatorSharesExchRate`, `CalibrateDelegation`, `UpdateInnerRedemptionRateBounds`.

Then run `go build ./x/stakeibc/...` and delete every import the compiler reports as unused (expected: `time`, `proto`, `ibctransfertypes`, `cast`, `banktypes` may survive via RestoreInterchainAccount, `govtypes`, `sdkmath`, `epochtypes`, `recordstypes`/`recordtypes` — keep whichever are still used).

- [ ] **Step 4: Delete the legacy message handler**

In `x/stakeibc/handler.go` delete `NewMessageHandler` (the whole function and its `// TODO: Remove` comment) and the imports it alone used (`baseapp`, `fmt`, `sdk`, `sdkerrors`, `errorsmod` if unused). Keep `NewStakeibcProposalHandler` (used by `app/app.go:855`).

- [ ] **Step 5: Delete the CLI commands and flag constants**

In `x/stakeibc/client/cli/tx.go`:
- delete the `const ( FlagMinRedemptionRate ... FlagLegacy )` block (lines 24-30) entirely: `FlagMinRedemptionRate`, `FlagMaxRedemptionRate`, `FlagCommunityPoolTreasuryAddress`, `FlagMaxMessagesPerIcaTx` are used only by `CmdRegisterHostZone`, and `FlagLegacy` only by `CmdToggleTradeController`;
- delete the `cmd.AddCommand(...)` lines for `CmdLiquidStake`, `CmdLSMLiquidStake`, `CmdRegisterHostZone`, `CmdRedeemStake`, `CmdRebalanceValidators`, `CmdClearBalance`, `CmdResumeHostZone`, `CmdSetCommunityPoolRebate`, `CmdToggleTradeController`;
- delete those nine `func Cmd...` functions.

The resulting `GetTxCmd` body:

```go
	cmd.AddCommand(CmdClaimUndelegatedTokens())
	cmd.AddCommand(CmdAddValidators())
	cmd.AddCommand(CmdChangeValidatorWeight())
	cmd.AddCommand(CmdChangeMultipleValidatorWeight())
	cmd.AddCommand(CmdDeleteValidator())
	cmd.AddCommand(CmdRestoreInterchainAccount())
	cmd.AddCommand(CmdCloseDelegationChannel())
	cmd.AddCommand(CmdUpdateValidatorSharesExchRate())
	cmd.AddCommand(CmdCalibrateDelegation())
	cmd.AddCommand(CmdUpdateInnerRedemptionRateBounds())
```

In `x/stakeibc/client/cli/tx_test.go` delete `TestCmdLiquidStake`, `TestCmdLSMLiquidStake`, `TestCmdRegisterHostZone`, `TestCmdRedeemStake`, `TestCmdRebalanceValidators`, `TestCmdClearBalance`, `TestCmdSetCommunityPoolRebate`, `TestCmdToggleTradeController`. Run `go vet ./x/stakeibc/client/...` and drop unused imports.

- [ ] **Step 6: Delete the handler tests and the helpers only they used**

In `x/stakeibc/keeper/msg_server_test.go` delete, with their `*TestCase` types and `Setup*` helpers:
- `ClearBalanceTestCase`, `SetupClearBalance` (469), `TestClearBalance_Successful`, `_HostChainMissing`, `_FeeAccountMissing`, `_ParseCoinError`;
- `LSMLiquidStakeTestCase`, `SetupTestLSMLiquidStake` (872) and all ten `TestLSMLiquidStake*` tests (939-1190). `getLSMTokenIBCDenom` (865) is NOT deleted: `icqcallbacks_validator_exchange_rate_test.go:106` calls it, so move that helper into `icqcallbacks_validator_exchange_rate_test.go` (its only remaining user) with the imports it needs;
- `SetupTestCreateTradeRoute` (1216), `submitCreateTradeRouteAndValidate` (1314), `submitUpdateTradeRouteAndValidate` (1425), `TestDeleteTradeRoute`, `TestCreateTradeRoute_*`, `TestUpdateTradeRoute`;
- `ResumeHostZoneTestCase`, `SetupResumeHostZone` (1991), `TestResumeHostZone_Success`, `_MissingZones`, `_UnhaltedZones`;
- `TestSetCommunityPoolRebate`, `TestToggleTradeController`.

Keep `SetupAddValidators`, `SetupDeleteValidator`, `SetupRestoreInterchainAccount` and its verify helpers, `SetupUpdateInnerRedemptionRateBounds`, `getSharesToTokensRateQueryData`.

- [ ] **Step 7: Build, vet, run the package**

```bash
go build ./... && go vet ./x/stakeibc/... ./x/autopilot/... ./app/...
go test ./x/stakeibc/... ./x/autopilot/... ./app/ 2>&1 | tail -20
```

Expected: `ok` for every package (the `app/` run includes `TestHistoricalTxDecodeTestSuite`).

- [ ] **Step 8: Commit**

```bash
git add proto/stride/stakeibc/tx.proto x/stakeibc/types/tx.pb.go x/stakeibc/keeper/msg_server.go x/stakeibc/handler.go x/stakeibc/client/cli/tx.go x/stakeibc/client/cli/tx_test.go x/stakeibc/keeper/msg_server_test.go
git commit -m "feat(stakeibc)!: remove liquid stake, redeem, register, trade route, rebate, rebalance, clear balance and resume handlers

Message types and registrations stay so historical txs decode.

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 5: Router and reflection guard for every removed message

**Files:**
- Create: `app/removed_handlers_test.go`

**Interfaces:**
- Consumes: `s.App.MsgServiceRouter().Handler(msg)` (nil when no route), `reflect` on each module's `MsgServer` interface.
- Depends on: Task 4 for stakeibc; the other modules' entries are added by Tasks 6-12 (see Step 3).
- Review: yes.

- [ ] **Step 1: Write the test with the stakeibc entries**

```go
package app_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

type RemovedHandlersTestSuite struct {
	apptesting.AppTestHelper
}

func (s *RemovedHandlersTestSuite) SetupTest() {
	s.Setup()
}

func TestRemovedHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(RemovedHandlersTestSuite))
}

// removedMsgs are the messages whose rpc was deleted in the v35 wind-down (spec §5). Each
// must still be a registered type (it decodes) but have no route in the msg service router,
// so a submitted tx fails with "can't route message" regardless of entry path.
var removedMsgs = []sdk.Msg{
	&stakeibctypes.MsgLiquidStake{},
	&stakeibctypes.MsgLSMLiquidStake{},
	&stakeibctypes.MsgRedeemStake{},
	&stakeibctypes.MsgRegisterHostZone{},
	&stakeibctypes.MsgCreateTradeRoute{},
	&stakeibctypes.MsgUpdateTradeRoute{},
	&stakeibctypes.MsgDeleteTradeRoute{},
	&stakeibctypes.MsgSetCommunityPoolRebate{},
	&stakeibctypes.MsgToggleTradeController{},
	&stakeibctypes.MsgRebalanceValidators{},
	&stakeibctypes.MsgClearBalance{},
	&stakeibctypes.MsgResumeHostZone{},
}

// keptMsgs are a sample of messages that must keep routing after the removals.
var keptMsgs = []sdk.Msg{
	&stakeibctypes.MsgClaimUndelegatedTokens{},
	&stakeibctypes.MsgRestoreInterchainAccount{},
	&stakeibctypes.MsgUpdateValidatorSharesExchRate{},
	&stakeibctypes.MsgCalibrateDelegation{},
}

// removedServerMethods maps each module's MsgServer interface to the method names that
// must no longer exist on it. Go cannot assert a method's absence at compile time, so
// this is a runtime reflection guard over each generated MsgServer interface.
var removedServerMethods = map[reflect.Type][]string{
	reflect.TypeOf((*stakeibctypes.MsgServer)(nil)).Elem(): {
		"LiquidStake", "LSMLiquidStake", "RedeemStake", "RegisterHostZone",
		"CreateTradeRoute", "UpdateTradeRoute", "DeleteTradeRoute",
		"SetCommunityPoolRebate", "ToggleTradeController",
		"RebalanceValidators", "ClearBalance", "ResumeHostZone",
	},
}

func (s *RemovedHandlersTestSuite) TestRemovedMessagesHaveNoRoute() {
	for _, msg := range removedMsgs {
		typeUrl := sdk.MsgTypeURL(msg)
		s.Require().Nil(s.App.MsgServiceRouter().Handler(msg), "%s must have no handler", typeUrl)

		// The type itself is still known to the codec (otherwise history stops decoding)
		resolved, err := s.App.InterfaceRegistry().Resolve(typeUrl)
		s.Require().NoError(err, "%s must still resolve in the interface registry", typeUrl)
		s.Require().NotNil(resolved)
	}
}

func (s *RemovedHandlersTestSuite) TestKeptMessagesStillRoute() {
	for _, msg := range keptMsgs {
		s.Require().NotNil(s.App.MsgServiceRouter().Handler(msg), "%s must keep its handler", sdk.MsgTypeURL(msg))
	}
}

func (s *RemovedHandlersTestSuite) TestMsgServerInterfacesLostTheMethods() {
	for serverType, methods := range removedServerMethods {
		for _, method := range methods {
			_, found := serverType.MethodByName(method)
			s.Require().False(found, "%s must not have method %s", serverType.String(), method)
		}
	}
}
```

- [ ] **Step 2: Run it**

Run: `go test ./app/ -run 'TestRemovedHandlersTestSuite' -v`
Expected: PASS (three subtests).

- [ ] **Step 3: Commit**

```bash
git add app/removed_handlers_test.go
git commit -m "test(app): removed messages have no route and no MsgServer method; kept ones still route

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

Note for Tasks 6-12: each module task appends its removed messages to `removedMsgs` and its `MsgServer` entry to `removedServerMethods` in this file (the exact entries are listed in each task). The merge of the parallel branches will conflict textually in this file; resolve by keeping every module's entries.

---

## Parallel-safe tasks

Tasks 6-12 each touch one module's `tx.proto`, `tx.pb.go`, `msg_server.go`, CLI and tests, plus their own entries in `app/removed_handlers_test.go`. None consumes another's interface. They depend only on Tasks 1-5.

Common recipe for each (referenced below as "the module recipe"):

```bash
# 1. edit proto/stride/<module>/tx.proto (rpc lines only)
make proto-gen
git add x/<module>/types/tx.pb.go
files=$(git diff --name-only | grep 'pb.go$' | grep -v "^x/<module>/types/tx.pb.go$" || true)
[ -n "$files" ] && git checkout -- $files  # no-op when proto-gen churned nothing else
# 2. delete the handler methods, CLI commands and handler tests listed in the task
go build ./... && go vet ./x/<module>/... ./app/...
go test ./x/<module>/... ./app/ 2>&1 | tail -5
```

---

### Task 6: staketia — remove `LiquidStake` (stub), `RedeemStake`, `ResumeHostZone`

**Files:**
- Modify: `proto/stride/staketia/tx.proto` (service block lines 22-78)
- Regenerate: `x/staketia/types/tx.pb.go`
- Modify: `x/staketia/keeper/msg_server.go:24-39,148-153`
- Modify: `x/staketia/keeper/unbonding.go:21-151` (delete `Keeper.RedeemStake`, per §5)
- Modify: `x/staketia/keeper/unbonding_test.go` (`TestRedeemStake`)
- Modify: `x/staketia/client/cli/tx.go` (`CmdRedeemStake` at 56, `CmdResumeHostZone` at 328, both `AddCommand` lines)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: yes (deletes a keeper function).

- [ ] **Step 1: Proto**

Remove these three rpcs (and their comment lines) from the `service Msg` block of `proto/stride/staketia/tx.proto`:

```proto
  // User transaction to liquid stake native tokens into stTokens
  rpc LiquidStake(MsgLiquidStake) returns (MsgLiquidStakeResponse);

  // User transaction to redeem stake stTokens into native tokens
  rpc RedeemStake(MsgRedeemStake) returns (MsgRedeemStakeResponse);
```
and
```proto
  // Unhalts the host zone if redemption rates were exceeded
  rpc ResumeHostZone(MsgResumeHostZone) returns (MsgResumeHostZoneResponse);
```

Add above the service: `// v35 wind-down: LiquidStake, RedeemStake and ResumeHostZone have no rpc; their message types remain for historical decoding.`

- [ ] **Step 2: Regenerate (module recipe step 1)**

- [ ] **Step 3: Delete handlers and the keeper redeem path**

In `x/staketia/keeper/msg_server.go` delete `LiquidStake` (the stub, lines 26-28, with its `//nolint` comment), `RedeemStake` (31-38) and `ResumeHostZone` (149-152) with their comments; drop the `errors` import if unused.

In `x/staketia/keeper/unbonding.go` delete `func (k Keeper) RedeemStake(...)` (lines 21-151) and its doc comment. Run `go build ./x/staketia/...`; delete imports that become unused. Do not delete `BurnRedeemedStTokens` or any other helper (§12 cleanup).

In `x/staketia/keeper/unbonding_test.go` delete `TestRedeemStake` (keep `TestBurnRedeemedStTokens_*`).

- [ ] **Step 4: CLI**

In `x/staketia/client/cli/tx.go` delete `CmdRedeemStake` and `CmdResumeHostZone` and their two lines inside `cmd.AddCommand(`. Drop unused imports.

- [ ] **Step 5: Guard entries**

In `app/removed_handlers_test.go` add the import `staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"`, append to `removedMsgs`:

```go
	&staketiatypes.MsgLiquidStake{},
	&staketiatypes.MsgRedeemStake{},
	&staketiatypes.MsgResumeHostZone{},
```

append to `keptMsgs`: `&staketiatypes.MsgConfirmUnbondedTokenSweep{},` and add to `removedServerMethods`:

```go
	reflect.TypeOf((*staketiatypes.MsgServer)(nil)).Elem(): {"LiquidStake", "RedeemStake", "ResumeHostZone"},
```

- [ ] **Step 6: Build and test (module recipe step 2)**

Expected: `ok` for `x/staketia/...` and `app`.

- [ ] **Step 7: Commit**

```bash
git add proto/stride/staketia/tx.proto x/staketia/types/tx.pb.go x/staketia/keeper/msg_server.go x/staketia/keeper/unbonding.go x/staketia/keeper/unbonding_test.go x/staketia/client/cli/tx.go app/removed_handlers_test.go
git commit -m "feat(staketia)!: remove liquid stake stub, redeem stake and resume host zone handlers

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 7: stakedym — remove `LiquidStake`, `RedeemStake`, `ResumeHostZone`

**Files:**
- Modify: `proto/stride/stakedym/tx.proto` (same three rpcs as staketia)
- Regenerate: `x/stakedym/types/tx.pb.go`
- Modify: `x/stakedym/keeper/msg_server.go:25-46,175-205`
- Modify: `x/stakedym/keeper/unbonding.go:20-101` (delete `Keeper.RedeemStake`, per §5; `Keeper.LiquidStake` in `delegation.go:24` STAYS, `LiquidStakeAndDistributeFees` calls it)
- Modify: `x/stakedym/keeper/unbonding_test.go` (`TestRedeemStake`), `x/stakedym/keeper/msg_server_test.go` (`TestMsgServerLiquidStake` at 18, `TestResumeHostZone` at ~330)
- Modify: `x/stakedym/client/cli/tx.go` (`CmdLiquidStake` 57, `CmdRedeemStake` 99, `CmdResumeHostZone` 369 and their `AddCommand` lines)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: yes.

- [ ] **Step 1: Proto** — identical edit to Task 6 Step 1, in `proto/stride/stakedym/tx.proto`.
- [ ] **Step 2: Regenerate (module recipe step 1)**.
- [ ] **Step 3: Delete handlers, keeper redeem path and their tests**

`x/stakedym/keeper/msg_server.go`: delete `LiquidStake` (27-35), `RedeemStake` (37-45), `ResumeHostZone` (176-204) with comments. `x/stakedym/keeper/unbonding.go`: delete `func (k Keeper) RedeemStake` (20-101) and its comment. Tests: delete `TestRedeemStake` in `unbonding_test.go`; delete `TestMsgServerLiquidStake` and `TestResumeHostZone` in `msg_server_test.go`. `go build ./x/stakedym/...`; fix unused imports (`utils`, `errorsmod` may become unused in `msg_server.go`).

- [ ] **Step 4: CLI** — delete `CmdLiquidStake`, `CmdRedeemStake`, `CmdResumeHostZone` and their `AddCommand` lines.
- [ ] **Step 5: Guard entries** — import `stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"`; append `&stakedymtypes.MsgLiquidStake{}, &stakedymtypes.MsgRedeemStake{}, &stakedymtypes.MsgResumeHostZone{},` to `removedMsgs`; `&stakedymtypes.MsgConfirmUnbondedTokenSweep{},` to `keptMsgs`; `reflect.TypeOf((*stakedymtypes.MsgServer)(nil)).Elem(): {"LiquidStake", "RedeemStake", "ResumeHostZone"},` to `removedServerMethods`.
- [ ] **Step 6: Build and test (module recipe step 2)** — Expected `ok`.
- [ ] **Step 7: Commit**

```bash
git add proto/stride/stakedym/tx.proto x/stakedym/types/tx.pb.go x/stakedym/keeper/msg_server.go x/stakedym/keeper/unbonding.go x/stakedym/keeper/unbonding_test.go x/stakedym/keeper/msg_server_test.go x/stakedym/client/cli/tx.go app/removed_handlers_test.go
git commit -m "feat(stakedym)!: remove liquid stake, redeem stake and resume host zone handlers

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 8: icaoracle — remove `AddOracle`, `InstantiateOracle`

**Files:**
- Modify: `proto/stride/icaoracle/tx.proto` (service block lines 12-27)
- Regenerate: `x/icaoracle/types/tx.pb.go`
- Modify: `x/icaoracle/keeper/msg_server.go:36-174` (both handlers, inline logic; keep `RestoreOracleICA`, `ToggleOracle`, `RemoveOracle`)
- Delete: `x/icaoracle/keeper/msg_server_add_oracle_test.go`, `x/icaoracle/keeper/msg_server_instantiate_oracle_test.go`
- Modify: `x/icaoracle/client/cli/tx.go` (`CmdAddOracle` 38, `CmdInstantiateOracle` 74, `AddCommand` lines 29-30; `tx_test.go` only tests `CmdRestoreOracleICA` and stays)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: no.

- [ ] **Step 1: Proto** — service block becomes:

```proto
service Msg {
  option (cosmos.msg.v1.service) = true;

  // Restores the oracle ICA channel after a closure
  rpc RestoreOracleICA(MsgRestoreOracleICA)
      returns (MsgRestoreOracleICAResponse);
  // Toggle's whether an oracle is active and should receive metric updates
  rpc ToggleOracle(MsgToggleOracle) returns (MsgToggleOracleResponse);
  // Removes an oracle completely
  rpc RemoveOracle(MsgRemoveOracle) returns (MsgRemoveOracleResponse);
}
```

- [ ] **Step 2: Regenerate (module recipe step 1)**.
- [ ] **Step 3: Delete the two handlers and both test files**; `go build ./x/icaoracle/...`, fix unused imports.
- [ ] **Step 4: CLI** — delete the two commands and their `AddCommand` lines.
- [ ] **Step 5: Guard entries** — import `icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"`; append `&icaoracletypes.MsgAddOracle{}, &icaoracletypes.MsgInstantiateOracle{},` to `removedMsgs`; `&icaoracletypes.MsgToggleOracle{},` to `keptMsgs`; `reflect.TypeOf((*icaoracletypes.MsgServer)(nil)).Elem(): {"AddOracle", "InstantiateOracle"},` to `removedServerMethods`.
- [ ] **Step 6: Build and test (module recipe step 2)** — Expected `ok`.
- [ ] **Step 7: Commit**

```bash
git add proto/stride/icaoracle/tx.proto x/icaoracle/types/tx.pb.go x/icaoracle/keeper/ x/icaoracle/client/cli/tx.go app/removed_handlers_test.go
git commit -m "feat(icaoracle)!: remove add-oracle and instantiate-oracle handlers

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 9: icqoracle — remove `RegisterTokenPriceQuery`, `RemoveTokenPriceQuery`

**Files:**
- Modify: `proto/stride/icqoracle/tx.proto` (keep only `UpdateParams`)
- Regenerate: `x/icqoracle/types/tx.pb.go`
- Modify: `x/icqoracle/keeper/msg_server.go:29-60` (keep `UpdateParams` at 61)
- Delete: `x/icqoracle/keeper/msg_server_test.go` (its only two tests are the removed handlers)
- Modify: `x/icqoracle/client/cli/tx.go` (`CmdAddTokenPrice` 36, `CmdRemoveTokenPrice` 81, `AddCommand` lines 29-30)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: no.

- [ ] **Step 1: Proto** — service block becomes:

```proto
service Msg {
  option (cosmos.msg.v1.service) = true;

  // UpdateParams defines a governance operation for updating the x/icqoracle
  // module parameters. The authority is defined in the keeper.
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);
}
```

- [ ] **Step 2: Regenerate (module recipe step 1)**.
- [ ] **Step 3: Delete the two handlers and `msg_server_test.go`**; fix imports.
- [ ] **Step 4: CLI** — delete both commands and their `AddCommand` lines; `GetTxCmd` keeps the empty parent command.
- [ ] **Step 5: Guard entries** — import `icqoracletypes "github.com/Stride-Labs/stride/v34/x/icqoracle/types"`; append `&icqoracletypes.MsgRegisterTokenPriceQuery{}, &icqoracletypes.MsgRemoveTokenPriceQuery{},` to `removedMsgs`; `&icqoracletypes.MsgUpdateParams{},` to `keptMsgs`; `reflect.TypeOf((*icqoracletypes.MsgServer)(nil)).Elem(): {"RegisterTokenPriceQuery", "RemoveTokenPriceQuery"},` to `removedServerMethods`.
- [ ] **Step 6: Build and test (module recipe step 2)** — Expected `ok`.
- [ ] **Step 7: Commit**

```bash
git add proto/stride/icqoracle/tx.proto x/icqoracle/types/tx.pb.go x/icqoracle/keeper/ x/icqoracle/client/cli/tx.go app/removed_handlers_test.go
git commit -m "feat(icqoracle)!: remove token price query register/remove handlers

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 10: auction — remove `PlaceBid`, `CreateAuction`, `UpdateAuction`

**Files:**
- Modify: `proto/stride/auction/tx.proto` (service block keeps only the `option`)
- Regenerate: `x/auction/types/tx.pb.go`
- Modify: `x/auction/keeper/msg_server.go` (delete all three handlers; keep `msgServer` struct and `NewMsgServerImpl` so `module.go` still registers an empty server; `Keeper.PlaceBid` in `auction.go:55` stays, nothing else calls it, §12 cleanup)
- Delete: `x/auction/keeper/msg_server_test.go` (all twelve tests go through the handlers)
- Modify: `x/auction/client/cli/tx.go` (delete `CmdPlaceBid`, `CmdCreateAuction`, `CmdUpdateAuction` and the `AddCommand` block contents)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: no.

- [ ] **Step 1: Proto** — service block becomes:

```proto
// v35 wind-down: every auction transaction was removed; message types remain for
// historical decoding.
service Msg {
  option (cosmos.msg.v1.service) = true;
}
```

- [ ] **Step 2: Regenerate (module recipe step 1)**. `go build ./x/auction/...` — the generated `MsgServer` interface is now empty; `msgServer` still satisfies it.
- [ ] **Step 3: Delete the three handlers and `msg_server_test.go`**; fix imports (`msg_server.go` keeps only the struct, `NewMsgServerImpl` and `var _ types.MsgServer = msgServer{}` if present).
- [ ] **Step 4: CLI** — delete the three commands; `GetTxCmd` keeps the parent command with an empty `AddCommand()` removed.
- [ ] **Step 5: Guard entries** — import `auctiontypes "github.com/Stride-Labs/stride/v34/x/auction/types"`; append `&auctiontypes.MsgPlaceBid{}, &auctiontypes.MsgCreateAuction{}, &auctiontypes.MsgUpdateAuction{},` to `removedMsgs`; `reflect.TypeOf((*auctiontypes.MsgServer)(nil)).Elem(): {"PlaceBid", "CreateAuction", "UpdateAuction"},` to `removedServerMethods`.
- [ ] **Step 6: Build and test (module recipe step 2)** — Expected `ok` (remaining auction keeper tests stay green).
- [ ] **Step 7: Commit**

```bash
git add proto/stride/auction/tx.proto x/auction/types/tx.pb.go x/auction/keeper/ x/auction/client/cli/tx.go app/removed_handlers_test.go
git commit -m "feat(auction)!: remove place-bid, create-auction and update-auction handlers

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 11: airdrop — remove all seven handlers

**Files:**
- Modify: `proto/stride/airdrop/tx.proto` (service block keeps only the `option`)
- Regenerate: `x/airdrop/types/tx.pb.go`
- Modify: `x/airdrop/keeper/msg_server.go` (delete `ClaimDaily`, `ClaimEarly`, `CreateAirdrop`, `UpdateAirdrop`, `AddAllocations`, `UpdateUserAllocation`, `LinkAddresses`; keep struct + `NewMsgServerImpl`; `Keeper.ClaimDaily/ClaimEarly/LinkAddresses` in `claim.go` stay)
- Delete: `x/airdrop/keeper/msg_server_test.go`
- Modify: `x/airdrop/client/cli/tx.go` (delete the seven `Cmd*` functions and the `AddCommand` block contents; `parser.go`/`parser_test.go` stay)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: no.

- [ ] **Step 1: Proto** — service block becomes:

```proto
// v35 wind-down: every airdrop transaction was removed (both airdrops ended December
// 2024); message types remain for historical decoding.
service Msg {
  option (cosmos.msg.v1.service) = true;
}
```

- [ ] **Step 2: Regenerate (module recipe step 1)**.
- [ ] **Step 3: Delete the seven handlers and `msg_server_test.go`**; fix imports.
- [ ] **Step 4: CLI** — delete the seven commands and the `AddCommand` block contents; run `go vet ./x/airdrop/...` and delete any helper in `tx.go` that only those commands used (leave `parser.go`).
- [ ] **Step 5: Guard entries** — import `airdroptypes "github.com/Stride-Labs/stride/v34/x/airdrop/types"`; append `&airdroptypes.MsgClaimDaily{}, &airdroptypes.MsgClaimEarly{}, &airdroptypes.MsgCreateAirdrop{}, &airdroptypes.MsgUpdateAirdrop{}, &airdroptypes.MsgAddAllocations{}, &airdroptypes.MsgUpdateUserAllocation{}, &airdroptypes.MsgLinkAddresses{},` to `removedMsgs`; `reflect.TypeOf((*airdroptypes.MsgServer)(nil)).Elem(): {"ClaimDaily", "ClaimEarly", "CreateAirdrop", "UpdateAirdrop", "AddAllocations", "UpdateUserAllocation", "LinkAddresses"},` to `removedServerMethods`.
- [ ] **Step 6: Build and test (module recipe step 2)** — Expected `ok`.
- [ ] **Step 7: Commit**

```bash
git add proto/stride/airdrop/tx.proto x/airdrop/types/tx.pb.go x/airdrop/keeper/ x/airdrop/client/cli/tx.go app/removed_handlers_test.go
git commit -m "feat(airdrop)!: remove all seven airdrop handlers

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 12: claim (legacy) — remove all four handlers

**Files:**
- Modify: `proto/stride/claim/tx.proto` (service block keeps only the `option`)
- Regenerate: `x/claim/types/tx.pb.go`
- Modify: `x/claim/keeper/msg_server.go` (delete `SetAirdropAllocations`, `ClaimFreeAmount`, `CreateAirdrop`, `DeleteAirdrop`; keep struct + `NewMsgServerImpl`)
- Delete: `x/claim/keeper/msg_server_test.go` (its five tests go through the handlers)
- Delete: `x/claim/client/cli/tx_claim_free_amount.go`, `tx_create_airdrop.go`, `tx_delete_airdrop.go`, `tx_set_airdrop_allocations.go`, and `x/claim/client/cli/cli_test.go` (an in-process network suite whose `SetupSuite` creates an airdrop through `CmdCreateAirdrop`; its one remaining query test cannot run without it)
- Modify: `x/claim/client/cli/tx.go` (delete the four `AddCommand` lines)
- Modify: `app/removed_handlers_test.go`

**Interfaces:**
- Depends on: Tasks 1-5.
- Review: no.

- [ ] **Step 1: Proto** — service block becomes:

```proto
// v35 wind-down: every legacy claim transaction was removed; message types remain for
// historical decoding.
service Msg {
  option (cosmos.msg.v1.service) = true;
}
```

- [ ] **Step 2: Regenerate (module recipe step 1)**.
- [ ] **Step 3: Delete the four handlers, `msg_server_test.go`, the four CLI files and `cli_test.go`**; edit `tx.go` to drop the four `AddCommand` lines; `go build ./x/claim/...`, fix imports. Check `go vet ./x/claim/...` passes (the `query.go` CLI and `query_test`-style helpers stay).
- [ ] **Step 4: Guard entries** — import `claimtypes "github.com/Stride-Labs/stride/v34/x/claim/types"`; append `&claimtypes.MsgSetAirdropAllocations{}, &claimtypes.MsgClaimFreeAmount{}, &claimtypes.MsgCreateAirdrop{}, &claimtypes.MsgDeleteAirdrop{},` to `removedMsgs`; `reflect.TypeOf((*claimtypes.MsgServer)(nil)).Elem(): {"SetAirdropAllocations", "ClaimFreeAmount", "CreateAirdrop", "DeleteAirdrop"},` to `removedServerMethods`.
- [ ] **Step 5: Build and test (module recipe step 2)** — Expected `ok`.
- [ ] **Step 6: Commit**

```bash
git add proto/stride/claim/tx.proto x/claim/types/tx.pb.go x/claim/keeper/ x/claim/client/cli/ app/removed_handlers_test.go
git commit -m "feat(claim)!: remove the four legacy claim handlers and their CLI

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 13: Merge gate — full build, vet, lint and unit suite (serial, after Tasks 6-12 merge)

**Files:** none new.

**Interfaces:**
- Depends on: Tasks 1-12.
- Review: yes (the final whole-branch review).

- [ ] **Step 1: Confirm nothing references a removed handler**

```bash
grep -rn "GetMsgServer()\.\(LiquidStake\|LSMLiquidStake\|RedeemStake\|RegisterHostZone\|RebalanceValidators\|ClearBalance\|ResumeHostZone\|CreateTradeRoute\|UpdateTradeRoute\|DeleteTradeRoute\|SetCommunityPoolRebate\|ToggleTradeController\|AddOracle\|InstantiateOracle\|RegisterTokenPriceQuery\|RemoveTokenPriceQuery\|PlaceBid\|CreateAuction\|UpdateAuction\|ClaimDaily\|ClaimEarly\|CreateAirdrop\|UpdateAirdrop\|AddAllocations\|UpdateUserAllocation\|LinkAddresses\|SetAirdropAllocations\|ClaimFreeAmount\|DeleteAirdrop\)(" x/ app/
grep -rn "NewMessageHandler\|NewMsgServerImpl(k)\." x/ app/
```

Expected: no output for both.

- [ ] **Step 2: Confirm every registration survived**

```bash
for t in MsgLSMLiquidStake MsgCreateTradeRoute MsgDeleteTradeRoute MsgUpdateTradeRoute; do grep -c "&$t{}" x/stakeibc/types/codec.go; done   # expect 2 each (amino + interface registry)
for m in staketia stakedym icaoracle icqoracle auction airdrop claim; do echo "$m $(git diff wind-down-design-consolidation -- x/$m/types/codec.go | wc -l)"; done
```

Expected: every non-stakeibc `codec.go` diff is `0` lines; stakeibc's diff adds exactly four `&Msg...{}` lines and a comment.

- [ ] **Step 3: Build, vet, lint, test**

```bash
go build ./... && go vet ./...
make lint
make test-unit 2>&1 | tail -30
```

Expected: build and vet clean; lint clean; every package `ok` except `github.com/Stride-Labs/stride/v34/utils` (`TestCreateModuleAccount`, pre-existing on the base branch — confirm with `git stash; go test ./utils/ -run TestCreateModuleAccount; git stash pop` if in doubt).

- [ ] **Step 4: Smoke the binary**

```bash
make build
./build/strided tx stakeibc --help | grep -c "liquid-stake\|redeem-stake\|register-host-zone"
./build/strided tx stakeibc --help | grep -c "claim-undelegated-tokens"
```

Expected: `0` then `1`.

- [ ] **Step 5: No commit needed** unless Steps 1-4 forced a fix; if so, commit it as `fix(wind-down): <what>` with the trailer.

---

## Self-review

**1. Spec coverage.** §5 removed list: stakeibc twelve → Task 4; staketia/stakedym `LiquidStake`, `RedeemStake`, `ResumeHostZone` → Tasks 6-7 (redeem keeper paths deleted per §5, helpers left for §12); icaoracle two → Task 8; icqoracle two → Task 9; auction three → Task 10; airdrop seven → Task 11; claim four → Task 12. "Types and both registrations stay" → Task 1 (adds the four missing ones) and Task 13 Step 2 (proves no codec diff elsewhere). `ClaimUndelegatedTokens` and `RestoreInterchainAccount` kept → Task 4 proto block and Task 5 `keptMsgs`. §11 "compile-time guarantee that the removed messages no longer exist" → Task 5's runtime reflection guard on each `MsgServer` interface (Go cannot assert a method's absence at compile time; the guard plus the router check is the equivalent); "decode test proving historical txs still parse" → Task 1. §13: `RegisterHostZone` tests rewritten not deleted → Task 3; `community_pool.go`, `handler.go`, `tx_test.go` → Tasks 2, 4; `msgServer.LiquidStake` → `Keeper.LiquidStake` → Task 2; keep `message_liquid_stake.go`/`message_redeem_stake.go` → Global Constraints; CLI flag constants → Task 4 Step 5; proto-gen descriptor churn → module recipe; the first dry run's `RegisterImplementations` mistake → Task 1's test.

**2. Placeholders.** None: every deletion names the function and line range; every rewrite shows the replacement; the fixture holds real bytes; the fetch commands are exact.

**3. Type consistency.** `Keeper.LiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error)` is defined in Task 2 and consumed by the same signature in Tasks 2 (callers, tests) and 4. `removedMsgs`/`keptMsgs`/`removedServerMethods` are defined in Task 5 and appended by Tasks 6-12 with matching types.

**4. Review tags.** Task 1 (history), Task 2 (mint path), Task 4 (largest deletion), Task 5 (the guard), Task 6-7 (keeper deletion), Task 13 (final) are `yes`; Tasks 3, 8-12 are mechanical deletions and `no`.

**Noted, out of scope:** dockernet and `scripts/local-to-mainnet` scripts still invoke removed CLI commands (dev tooling); `x/stakeibc/types/message_*_test.go` ValidateBasic tests for removed messages stay because the types stay; unreferenced keeper code left behind (`StartLSMLiquidStake`, `SubmitValidatorSlashQuery`, `BuildTradeAuthzMsg`, `GetTradeRouteFromTradeAccountChainId`, `RebalanceDelegationsForHostZone` callers, auction `Keeper.PlaceBid`, airdrop claim keeper functions, staketia/stakedym `BurnRedeemedStTokens`) is the §12 cleanup.
