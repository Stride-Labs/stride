package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

func (s *KeeperTestSuite) TestCheckAcknowledgementStatus() {
	// Test with a successful ack
	ackSuccess := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{
			Result: []byte{1}, // just has to be non-empty
		},
	})
	success, err := s.App.AutopilotKeeper.CheckAcknowledgementStatus(s.Ctx, ackSuccess)
	s.Require().True(success, "ack success should return true")
	s.Require().NoError(err)

	// Test with an ack error
	// Success should be false, but there should be no error returned
	errorString := "some error"
	ackFailure := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: errorString,
		},
	})
	success, err = s.App.AutopilotKeeper.CheckAcknowledgementStatus(s.Ctx, ackFailure)
	s.Require().False(success, "ack failure should return false")
	s.Require().NoError(err)

	// Test with an ack result that is missing the "result" field
	// It should return an error
	ackResultError := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{
			Result: []byte{}, // empty result throws an error
		},
	})
	_, err = s.App.AutopilotKeeper.CheckAcknowledgementStatus(s.Ctx, ackResultError)
	s.Require().ErrorContains(err, "acknowledgement result cannot be empty")

	// Test with invalid ack data that can't be unmarshaled
	randomBytes := []byte{1, 2, 3}
	_, err = s.App.AutopilotKeeper.CheckAcknowledgementStatus(s.Ctx, randomBytes)
	s.Require().ErrorContains(err, "cannot unmarshal ICS-20 transfer packet acknowledgement")
}

func (s *KeeperTestSuite) TestBuildCoinFromTransferMetadata() {
	denom := "denom"
	amountString := "10000"
	amountInt := sdk.NewInt(10000)

	// Test with valid packet data
	expectedToken := sdk.NewCoin(denom, amountInt)
	transferMetadata := transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: amountString,
	}
	actualToken, err := s.App.AutopilotKeeper.BuildCoinFromTransferMetadata(transferMetadata)
	s.Require().NoError(err)
	s.Require().Equal(expectedToken, actualToken, "token")

	// Test with invalid packet data
	invalidMetadata := transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: "",
	}
	_, err = s.App.AutopilotKeeper.BuildCoinFromTransferMetadata(invalidMetadata)
	s.Require().ErrorContains(err, "unable to parse amount from transfer packet")
}

func (s *KeeperTestSuite) TestSendToFallbackAddress() {
	senderAccount := s.TestAccs[0]
	fallbackAccount := s.TestAccs[1]

	denom := "denom"
	amountInt := sdk.NewInt(10000)
	amountString := "10000"

	// Fund the sender
	zeroCoin := sdk.NewCoin(denom, sdkmath.ZeroInt())
	balanceCoin := sdk.NewCoin(denom, amountInt)
	s.FundAccount(senderAccount, balanceCoin)

	// Send to the fallback address with a valid input
	packetDataBz := transfertypes.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: amountString,
		Sender: senderAccount.String(),
	})
	err := s.App.AutopilotKeeper.SendToFallbackAddress(s.Ctx, packetDataBz, fallbackAccount.String())
	s.Require().NoError(err, "no error expected when sending to fallback address")

	// Check that the funds were transferred
	senderBalance := s.App.BankKeeper.GetBalance(s.Ctx, senderAccount, denom)
	s.CompareCoins(zeroCoin, senderBalance, "sender should have lost tokens")

	fallbackBalance := s.App.BankKeeper.GetBalance(s.Ctx, fallbackAccount, denom)
	s.CompareCoins(balanceCoin, fallbackBalance, "fallback should have gained tokens")

	// Test with an invalid sender address - it should error
	invalidPacketDataBz := transfertypes.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: amountString,
		Sender: "invalid_sender",
	})
	err = s.App.AutopilotKeeper.SendToFallbackAddress(s.Ctx, invalidPacketDataBz, fallbackAccount.String())
	s.Require().ErrorContains(err, "invalid sender address")

	// Test with an invalid fallback address - it should error
	err = s.App.AutopilotKeeper.SendToFallbackAddress(s.Ctx, packetDataBz, "invalid_fallback")
	s.Require().ErrorContains(err, "invalid fallback address")

	// Test with an invalid amount - it should error
	invalidPacketDataBz = transfertypes.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: "",
		Sender: senderAccount.String(),
	})
	err = s.App.AutopilotKeeper.SendToFallbackAddress(s.Ctx, invalidPacketDataBz, fallbackAccount.String())
	s.Require().ErrorContains(err, "unable to parse amount")

	// Finally, try to call the send function again with a valid input,
	// it should fail since the sender now has an insufficient balance
	err = s.App.AutopilotKeeper.SendToFallbackAddress(s.Ctx, packetDataBz, fallbackAccount.String())
	s.Require().ErrorContains(err, "insufficient funds")
}
