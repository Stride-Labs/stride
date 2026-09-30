package app_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	airdroptypes "github.com/Stride-Labs/stride/v34/x/airdrop/types"
	auctiontypes "github.com/Stride-Labs/stride/v34/x/auction/types"
	claimtypes "github.com/Stride-Labs/stride/v34/x/claim/types"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
	icqoracletypes "github.com/Stride-Labs/stride/v34/x/icqoracle/types"
	stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
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
	&staketiatypes.MsgLiquidStake{}, //nolint:staticcheck // deprecated type, kept registered for historical decoding
	&staketiatypes.MsgRedeemStake{},
	&staketiatypes.MsgResumeHostZone{},
	&stakedymtypes.MsgLiquidStake{},
	&stakedymtypes.MsgRedeemStake{},
	&stakedymtypes.MsgResumeHostZone{},
	&icaoracletypes.MsgAddOracle{},
	&icaoracletypes.MsgInstantiateOracle{},
	&icqoracletypes.MsgRegisterTokenPriceQuery{},
	&icqoracletypes.MsgRemoveTokenPriceQuery{},
	&auctiontypes.MsgPlaceBid{},
	&auctiontypes.MsgCreateAuction{},
	&auctiontypes.MsgUpdateAuction{},
	&airdroptypes.MsgClaimDaily{},
	&airdroptypes.MsgClaimEarly{},
	&airdroptypes.MsgCreateAirdrop{},
	&airdroptypes.MsgUpdateAirdrop{},
	&airdroptypes.MsgAddAllocations{},
	&airdroptypes.MsgUpdateUserAllocation{},
	&airdroptypes.MsgLinkAddresses{},
	&claimtypes.MsgSetAirdropAllocations{},
	&claimtypes.MsgClaimFreeAmount{},
	&claimtypes.MsgCreateAirdrop{},
	&claimtypes.MsgDeleteAirdrop{},
}

// keptMsgs are a sample of messages that must keep routing after the removals.
var keptMsgs = []sdk.Msg{
	&stakeibctypes.MsgClaimUndelegatedTokens{},
	&stakeibctypes.MsgRestoreInterchainAccount{},
	&stakeibctypes.MsgUpdateValidatorSharesExchRate{},
	&stakeibctypes.MsgCalibrateDelegation{},
	&staketiatypes.MsgConfirmUnbondedTokenSweep{},
	&stakedymtypes.MsgConfirmUnbondedTokenSweep{},
	&icaoracletypes.MsgToggleOracle{},
	&icqoracletypes.MsgUpdateParams{},
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
	reflect.TypeOf((*staketiatypes.MsgServer)(nil)).Elem():  {"LiquidStake", "RedeemStake", "ResumeHostZone"},
	reflect.TypeOf((*stakedymtypes.MsgServer)(nil)).Elem():  {"LiquidStake", "RedeemStake", "ResumeHostZone"},
	reflect.TypeOf((*icaoracletypes.MsgServer)(nil)).Elem(): {"AddOracle", "InstantiateOracle"},
	reflect.TypeOf((*icqoracletypes.MsgServer)(nil)).Elem(): {"RegisterTokenPriceQuery", "RemoveTokenPriceQuery"},
	reflect.TypeOf((*auctiontypes.MsgServer)(nil)).Elem():   {"PlaceBid", "CreateAuction", "UpdateAuction"},
	reflect.TypeOf((*airdroptypes.MsgServer)(nil)).Elem():   {"ClaimDaily", "ClaimEarly", "CreateAirdrop", "UpdateAirdrop", "AddAllocations", "UpdateUserAllocation", "LinkAddresses"},
	reflect.TypeOf((*claimtypes.MsgServer)(nil)).Elem():     {"SetAirdropAllocations", "ClaimFreeAmount", "CreateAirdrop", "DeleteAirdrop"},
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
