package v36

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v36"

	// UpgradeAuthority is the team multisig, the chain authority since v35. Its Osmosis address (same bytes) is the
	// wind-down vault, so the sweep runner can move what lands here straight into the vault.
	UpgradeAuthority = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"

	// The governance-earmarked treasury accounts. They are 32-byte module accounts with no key, so no transaction can
	// spend them and every earlier release from them went through an upgrade handler (v7, v10, v17, v18, v25, v28).
	CommunityPoolGrowthAddress   = "stride1lj0m72d70qerts9ksrsphy9nmsd4h0s88ll9gfphmhemh8ewet5qj44jc9"
	IncentiveProgramAddress      = "stride1tlxk4as9sgpqkh42cfaxqja0mdj6qculqshy0gg3glazmrnx3y8s8gsvqk"
	CommunityPoolSecurityAddress = "stride14p6dyylhlmr445w9shcp36r5qcypc5d7rftpr2v0dt3pcjvy66zq60qfgg"
)

// TreasuryAddresses are swept to the authority in this order
var TreasuryAddresses = []string{
	CommunityPoolGrowthAddress,
	IncentiveProgramAddress,
	CommunityPoolSecurityAddress,
}
