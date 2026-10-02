package app_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v35/app"
	"github.com/Stride-Labs/stride/v35/app/apptesting"
	"github.com/Stride-Labs/stride/v35/app/upgrades/v35"
)

type BlockedMsgsTestSuite struct {
	apptesting.AppTestHelper
}

func (s *BlockedMsgsTestSuite) SetupTest() {
	s.Setup()
}

func TestBlockedMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(BlockedMsgsTestSuite))
}

// blockedMsgs builds one of each blocked staking message
func (s *BlockedMsgsTestSuite) blockedMsgs() []sdk.Msg {
	delegator := s.TestAccs[0].String()
	validator := sdk.ValAddress(s.TestAccs[1]).String()
	coin := sdk.NewCoin("ustrd", sdkmath.NewInt(1))

	return []sdk.Msg{
		&stakingtypes.MsgDelegate{DelegatorAddress: delegator, ValidatorAddress: validator, Amount: coin},
		&stakingtypes.MsgBeginRedelegate{
			DelegatorAddress: delegator, ValidatorSrcAddress: validator, ValidatorDstAddress: validator, Amount: coin,
		},
		&stakingtypes.MsgCreateValidator{DelegatorAddress: delegator, ValidatorAddress: validator},
		&stakingtypes.MsgCancelUnbondingDelegation{
			DelegatorAddress: delegator, ValidatorAddress: validator, Amount: coin, CreationHeight: 1,
		},
	}
}

func (s *BlockedMsgsTestSuite) buildTx(msgs ...sdk.Msg) sdk.Tx {
	txBuilder := s.App.GetTxConfig().NewTxBuilder()
	s.Require().NoError(txBuilder.SetMsgs(msgs...))
	return txBuilder.GetTx()
}

func (s *BlockedMsgsTestSuite) wrapInExec(msgs ...sdk.Msg) *authz.MsgExec {
	execMsg := authz.NewMsgExec(s.TestAccs[2], msgs)
	return &execMsg
}

// runDecorator calls the decorator in check, simulate and deliver modes and returns the error
// and whether `next` was reached for each mode
func (s *BlockedMsgsTestSuite) runDecorator(tx sdk.Tx) (errs []error, reached []bool) {
	decorator := app.NewBlockedMsgsDecorator(v35.BlockedStakingMsgTypeUrls)

	for _, mode := range []struct{ checkTx, simulate bool }{{true, false}, {false, true}, {false, false}} {
		nextReached := false
		next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
			nextReached = true
			return ctx, nil
		}

		ctx := s.Ctx.WithIsCheckTx(mode.checkTx)
		_, err := decorator.AnteHandle(ctx, tx, mode.simulate, next)
		errs = append(errs, err)
		reached = append(reached, nextReached)
	}
	return errs, reached
}

func (s *BlockedMsgsTestSuite) requireRejected(tx sdk.Tx, typeUrl string) {
	errs, reached := s.runDecorator(tx)
	for i := range errs {
		s.Require().Error(errs[i], "mode %d", i)
		s.Require().True(errors.Is(errs[i], sdkerrors.ErrUnauthorized), "mode %d: %v", i, errs[i])
		s.Require().Contains(errs[i].Error(), typeUrl+" is disabled: the chain is winding down")
		s.Require().False(reached[i], "next reached in mode %d", i)
	}
}

func (s *BlockedMsgsTestSuite) requireAllowed(tx sdk.Tx) {
	errs, reached := s.runDecorator(tx)
	for i := range errs {
		s.Require().NoError(errs[i], "mode %d", i)
		s.Require().True(reached[i], "next not reached in mode %d", i)
	}
}

func (s *BlockedMsgsTestSuite) TestBlockedMsgsRejectedTopLevel() {
	for _, msg := range s.blockedMsgs() {
		typeUrl := sdk.MsgTypeURL(msg)
		s.Run(typeUrl, func() {
			s.requireRejected(s.buildTx(msg), typeUrl)
		})
	}
}

func (s *BlockedMsgsTestSuite) TestBlockedMsgsRejectedInsideExec() {
	for _, msg := range s.blockedMsgs() {
		typeUrl := sdk.MsgTypeURL(msg)
		s.Run(typeUrl, func() {
			s.requireRejected(s.buildTx(s.wrapInExec(msg)), typeUrl)
		})
	}
}

func (s *BlockedMsgsTestSuite) TestBlockedMsgsRejectedInsideNestedExec() {
	for _, msg := range s.blockedMsgs() {
		typeUrl := sdk.MsgTypeURL(msg)
		s.Run(typeUrl, func() {
			s.requireRejected(s.buildTx(s.wrapInExec(s.wrapInExec(msg))), typeUrl)
		})
	}
}

func (s *BlockedMsgsTestSuite) TestBlockedMsgRejectsWholeTx() {
	allowed := &banktypes.MsgSend{
		FromAddress: s.TestAccs[0].String(),
		ToAddress:   s.TestAccs[1].String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("ustrd", 1)),
	}
	blocked := s.blockedMsgs()[0]

	s.requireRejected(s.buildTx(allowed, blocked), sdk.MsgTypeURL(blocked))
	s.requireRejected(s.buildTx(s.wrapInExec(allowed, blocked)), sdk.MsgTypeURL(blocked))
}

func (s *BlockedMsgsTestSuite) TestAllowedMsgsPassThrough() {
	delegator := s.TestAccs[0].String()
	validator := sdk.ValAddress(s.TestAccs[1]).String()

	allowedMsgs := map[string]sdk.Msg{
		"undelegate": &stakingtypes.MsgUndelegate{
			DelegatorAddress: delegator, ValidatorAddress: validator, Amount: sdk.NewInt64Coin("ustrd", 1),
		},
		"withdraw rewards": &distrtypes.MsgWithdrawDelegatorReward{
			DelegatorAddress: delegator, ValidatorAddress: validator,
		},
		"send": &banktypes.MsgSend{
			FromAddress: delegator,
			ToAddress:   s.TestAccs[1].String(),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("ustrd", 1)),
		},
	}

	for name, msg := range allowedMsgs {
		s.Run(name, func() {
			s.requireAllowed(s.buildTx(msg))
			s.requireAllowed(s.buildTx(s.wrapInExec(msg)))
		})
	}
}
