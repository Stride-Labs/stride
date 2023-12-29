package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

type OnAckPacketTestCase struct {
	ChannelId       string
	Sequence        uint64
	Token           sdk.Coin
	Packet          channeltypes.Packet
	SenderAccount   sdk.AccAddress
	FallbackAccount sdk.AccAddress
}

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
	amount := sdk.NewInt(10000)

	// Test with valid packet data
	expectedToken := sdk.NewCoin(denom, amount)
	transferMetadata := transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: amount.String(),
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
	amount := sdk.NewInt(10000)

	// Fund the sender
	zeroCoin := sdk.NewCoin(denom, sdkmath.ZeroInt())
	balanceCoin := sdk.NewCoin(denom, amount)
	s.FundAccount(senderAccount, balanceCoin)

	// Send to the fallback address with a valid input
	packetDataBz := transfertypes.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:  denom,
		Amount: amount.String(),
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
		Amount: amount.String(),
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

func (s *KeeperTestSuite) SetupTestOnAcknowledgementPacket() OnAckPacketTestCase {
	senderAccount := s.TestAccs[0]
	fallbackAccount := s.TestAccs[1]

	sequence := uint64(1)
	channelId := "channel-0"
	denom := "denom"
	amount := sdk.NewInt(10000)
	token := sdk.NewCoin(denom, amount)

	// Set a fallback addresses
	s.App.AutopilotKeeper.SetTransferFallbackAddress(s.Ctx, channelId, sequence, fallbackAccount.String())

	// Fund the sender account
	s.FundAccount(senderAccount, token)

	// Build the IBC packet
	transferMetadata := transfertypes.FungibleTokenPacketData{
		Denom:  "denom",
		Amount: amount.String(),
		Sender: senderAccount.String(),
	}
	packet := channeltypes.Packet{
		Sequence:      sequence,
		SourceChannel: channelId,
		Data:          transfertypes.ModuleCdc.MustMarshalJSON(&transferMetadata),
	}

	return OnAckPacketTestCase{
		ChannelId:       channelId,
		Sequence:        sequence,
		Token:           token,
		Packet:          packet,
		SenderAccount:   senderAccount,
		FallbackAccount: fallbackAccount,
	}
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_AckSuccess() {
	tc := s.SetupTestOnAcknowledgementPacket()

	// Build a successful ack
	ackSuccess := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{
			Result: []byte{1}, // just has to be non-empty
		},
	})

	// Call OnAckPacket with the successful ack
	err := s.App.AutopilotKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, ackSuccess)
	s.Require().NoError(err, "no error expected during OnAckPacket")

	// Confirm the fallback address was removed
	_, found := s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, tc.ChannelId, tc.Sequence)
	s.Require().False(found, "fallback address should have been removed")

	// Confirm the fallback address has not received any coins
	zeroCoin := sdk.NewCoin(tc.Token.Denom, sdk.ZeroInt())
	fallbackBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.FallbackAccount, tc.Token.Denom)
	s.CompareCoins(zeroCoin, fallbackBalance, "fallback account should not have received funds")
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_AckFailure() {
	tc := s.SetupTestOnAcknowledgementPacket()

	// Build an error ack
	ackFailure := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{},
	})

	// Call OnAckPacket with the successful ack
	err := s.App.AutopilotKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, ackFailure)
	s.Require().NoError(err, "no error expected during OnAckPacket")

	// Confirm tokens were sent to the fallback address
	zeroCoin := sdk.NewCoin(tc.Token.Denom, sdk.ZeroInt())
	senderBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.SenderAccount, tc.Token.Denom)
	fallbackBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.FallbackAccount, tc.Token.Denom)
	s.CompareCoins(zeroCoin, senderBalance, "sender account should have lost funds")
	s.CompareCoins(tc.Token, fallbackBalance, "fallback account should have received funds")

	// Confirm the fallback address was removed
	_, found := s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, tc.ChannelId, tc.Sequence)
	s.Require().False(found, "fallback address should have been removed")
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_InvalidAck() {
	tc := s.SetupTestOnAcknowledgementPacket()

	// Build an invalid ack to force an error
	invalidAck := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{
			Result: []byte{}, // empty result causes an error
		},
	})

	// Call OnAckPacket with the invalid ack
	err := s.App.AutopilotKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, invalidAck)
	s.Require().ErrorContains(err, "invalid acknowledgement")
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_NoOp() {
	tc := s.SetupTestOnAcknowledgementPacket()

	// Remove the fallback address so that there is no action necessary in the callback
	s.App.AutopilotKeeper.RemoveTransferFallbackAddress(s.Ctx, tc.ChannelId, tc.Sequence)

	// Call OnAckPacket and confirm there was no error
	// The ack argument here doesn't matter cause the no-op check is upstream
	err := s.App.AutopilotKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, []byte{})
	s.Require().NoError(err, "no error expected during on ack packet")

	// Check that no funds were moved
	zeroCoin := sdk.NewCoin(tc.Token.Denom, sdk.ZeroInt())
	senderBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.SenderAccount, tc.Token.Denom)
	fallbackBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.FallbackAccount, tc.Token.Denom)
	s.CompareCoins(tc.Token, senderBalance, "sender account should have lost funds")
	s.CompareCoins(zeroCoin, fallbackBalance, "fallback account should have received funds")
}

func (s *KeeperTestSuite) TestOnTimeoutPacket() {

}
