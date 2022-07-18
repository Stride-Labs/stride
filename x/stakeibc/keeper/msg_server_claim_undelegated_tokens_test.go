package keeper_test

func (suite *KeeperTestSuite) SetupClaimUndelegatedTokens() {
	// Create user balance
	// Create user redemption records
	// Create host zone with redemption account
	// Create ICA state
	// Create valid message
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensSuccessful() {
	// Confirm isClaimable is false in redemption record
	// Confirm pending claims added
	// Anything else to check?
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensNoUserRedemptionRecord() {
	// Remove redemption record
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensRecordNotClaimable() {
	// Mark redemption record as not claimable
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensHostZoneNotFound() {
	// Change host zone in message
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensNoRedemptionAccount() {
	// Remove redemption account from host zone
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensNoEpochTracker() {
	// Remove epoch tracker
}

func (suite *KeeperTestSuite) TestClaimUndelegatedTokensSubmitTxFailure() {
	// Alter ICA State
}
