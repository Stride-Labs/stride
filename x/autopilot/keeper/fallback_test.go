package keeper_test

// Tests Get/Set/RemoveTransferFallbackAddress
func (s *KeeperTestSuite) TestTransferFallbackAddress() {
	channelId := "channel-0"
	sequence := uint64(100)
	expectedAddress := "stride1xjp08gxef09fck6yj2lg0vrgpcjhqhp055ffhj"

	// Add a new fallback address
	s.App.AutopilotKeeper.SetTransferFallbackAddress(s.Ctx, channelId, sequence, expectedAddress)

	// Confirm we can retrieve it
	actualAddress, found := s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, channelId, sequence)
	s.Require().True(found, "address should have been found")
	s.Require().Equal(expectedAddress, actualAddress, "fallback addres")

	// Remove it and confirm we can no longer retrieve it
	s.App.AutopilotKeeper.RemoveTransferFallbackAddress(s.Ctx, channelId, sequence)

	_, found = s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, channelId, sequence)
	s.Require().False(found, "address should have been removed")
}
