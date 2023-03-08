<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState 
given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v7.0.0](https://github.com/Stride-Labs/stride/releases/tag/v7.0.0) - 2023-03-08
### On-Chain changes
1. 
### Off-Chain changes
These changes do not affect any on-chain functionality, but have been implemented since `v6.0.0`.

1. 


## [v6.0.0](https://github.com/Stride-Labs/stride/releases/tag/v6.0.0) - 2023-02-22
### On-Chain changes
1. Reset Stride Airdrops to kick off Season 2.

## [v5.0.0](https://github.com/Stride-Labs/stride/releases/tag/v5.0.0) - 2023-01-19
### On-Chain changes
1. Added query for next send packet sequence number ([fec8b45](https://github.com/Stride-Labs/stride/commit/fec8b4570d9a4a85efd944a1b859cec72d1bf9ea))
2. Added additional fields to InitGenesis and ExportGenesis in stakeibc ([ea5a9a4](https://github.com/Stride-Labs/stride/commit/ea5a9a4babbe958c6309fefedc30df4c21a03028))
3. Incremented module version to v5 ([50eb2b4](https://github.com/Stride-Labs/stride/commit/50eb2b4ec541e94ccd46c02ea543ac9293094cef))
4. Upgrade store migration ([f635c1a](https://github.com/Stride-Labs/stride/commit/f635c1a872c9947df52e6f3f8e3ad4741d5d91e2))
5. Upgrade handler, remove Authz from store ([cd96f99](https://github.com/Stride-Labs/stride/commit/cd96f99c84ad2870528d55febad26dc05f82fa96))
6. Upgraded to cosmos-sdk v0.46.7 and ibc-go v5.1.0 ([e55f6f2](https://github.com/Stride-Labs/stride/commit/e55f6f2f8644ebd754ce32d3ced85b1e6db984c3))
7. Informal Systems audit fixes IF-STRIDE-STAKEIBC-ARITHMETIC ([1c0a4b8](https://github.com/Stride-Labs/stride/commit/1c0a4b8eb795c19b58ce06bbf194a52ec1df649b))
8. Cleaned up epoch and callback logs ([d1938b9](https://github.com/Stride-Labs/stride/commit/d1938b9d381cdad627093b6b4adbf1e90ff5b9d0)), ([c5d137a](https://github.com/Stride-Labs/stride/commit/c5d137a32bc23a403f94ad37989c940da0715138)), ([ac45bc4](https://github.com/Stride-Labs/stride/commit/ac45bc4971aa3fbfaebc1be3e07aa71029616c98)), ([9a8b757](https://github.com/Stride-Labs/stride/commit/9a8b757eb3386997e8d589044b8acabde034410e))
9. Misc minor refactoring / cleanup ([af53f06](https://github.com/Stride-Labs/stride/commit/af53f06e923f6b2a546f003d345476496e44ff6f)), ([edb90e9](https://github.com/Stride-Labs/stride/commit/edb90e98875564fc21f94f9216f6bdc75e599176)), ([78790b0](https://github.com/Stride-Labs/stride/commit/78790b06a37bd50ba2ae993c4cb0bacfcc7c9ebf)), ([29dc4f2](https://github.com/Stride-Labs/stride/commit/29dc4f2e5f61db38714f0464b4ee17221fba88e3)), ([f3f607a](https://github.com/Stride-Labs/stride/commit/f3f607af8861f29ff2580c4c7f8814aa61828573)), ([3c3c54b](https://github.com/Stride-Labs/stride/commit/3c3c54b9539b9055f848f98e036221173ff8f58e))
10. Dependency bumps ([c1ff649](https://github.com/Stride-Labs/stride/commit/c1ff6495d01dcac519ee9643038c739b4265a4a1)), ([b06abe3](https://github.com/Stride-Labs/stride/commit/b06abe3ef6211bce1bbc1ee179866c24bad3dab5))

### Off-Chain changes
These changes do not affect any on-chain functionality, but have been implemented since `v4.0.3`.

1. Localstride upgrade support ([50eb2b4](https://github.com/Stride-Labs/stride/commit/50eb2b4ec541e94ccd46c02ea543ac9293094cef))
2. SetupConfig helper for unit tests ([4996b7b](https://github.com/Stride-Labs/stride/commit/4996b7b9549de7e2020b83e8e8559bfe28ec2a79))
3. Updated docs ([6816961](https://github.com/Stride-Labs/stride/commit/68169611c12bc609b4ecaa99c50f0e2166986eb9)), ([f306089](https://github.com/Stride-Labs/stride/commit/f3060896fdebc1cc911976daf37245db815fef98)), ([8ee30c2](https://github.com/Stride-Labs/stride/commit/8ee30c2ebc8d6965211ae0cb11ae10852aa93a51)), ([6889888](https://github.com/Stride-Labs/stride/commit/6889888e4acb0ce1b27872664a0e6c734ed041b3)), ([abcba0f](https://github.com/Stride-Labs/stride/commit/abcba0f6a2fc6e485fdcd1049952f73ad9ea9865))
4. Dockernet updates: improved docs, added v5 host zone, updated dependencies ([2a60eb5](https://github.com/Stride-Labs/stride/commit/2a60eb5222f5b17de0de28aa8e85ad9a257f7496)), ([e9351a2](https://github.com/Stride-Labs/stride/commit/e9351a2f5718e95c98b0fe1e329dda73be92beb8)), ([2e461a3](https://github.com/Stride-Labs/stride/commit/2e461a3fdc42de9f9007dfa8ec7f9b2ac8fa2dcb)), ([619f704](https://github.com/Stride-Labs/stride/commit/619f704c54bfdc8d3d77a280e8a325faf173ba9b)), ([39ad5b8](https://github.com/Stride-Labs/stride/commit/39ad5b887a922d385e04b3cbae24771ba3045fa3))
5. CI updates ([1c0a4b8](https://github.com/Stride-Labs/stride/commit/78225e2d18d50592f0a40d853419ddc596ad64a5)), ([5f73659](https://github.com/Stride-Labs/stride/commit/5f7365992782224c21fdabec5ade71be88080000))


## [v4.0.3](https://github.com/Stride-Labs/stride/releases/tag/v4.0.3) - 2022-12-10
1. Add list-pending-queries query to interchainquery module ([433dabc](https://github.com/Stride-Labs/stride/commit/433dabcc02c46d2ff038946fe2525a41b9e29b43)) 
2. Updated docs ([a37a56d](https://github.com/Stride-Labs/stride/commit/a37a56d0f17ff54763d77f2864c35a0c8141bd09), [87f7e6d](https://github.com/Stride-Labs/stride/commit/87f7e6d1c620633dca2286f2136dc4c4c855164d), [c913955](https://github.com/Stride-Labs/stride/commit/c913955cfef4bd23c238235c7592fd926427aaef))
3. Added Makefile commands to start dockernet with different host chains ([33cb844](https://github.com/Stride-Labs/stride/commit/33cb84438c3b605ae0e4a96ee2f73b6d7837b3d2))

## [v4.0.2](https://github.com/Stride-Labs/stride/releases/tag/v4.0.0) - 2022-12-05
1. Rebuild protos, fix v3 reference ((9d83474e7)[https://github.com/Stride-Labs/stride/commit/9d83474e741c4f79ba3a4132d10881362f28ab7e])

## [v4.0.0](https://github.com/Stride-Labs/stride/releases/tag/v4.0.0) - 2022-11-27
### On-Chain changes
1. Dependency bumps ([384178b2c](https://github.com/Stride-Labs/stride/commit/384178b2cf98e9af0815ffaf3c29649f41784f3e)), ([0a2297ea](https://github.com/Stride-Labs/stride/commit/0a2297eabe287d38723ab8213d5256ce34d2bb2d)), ([802a1dee3](https://github.com/Stride-Labs/stride/commit/802a1dee380e02916f1e89a231d36365a026bb46)), ([f1d90c6a](https://github.com/Stride-Labs/stride/commit/f1d90c6aa50daf75d9608095cc39cbbbe3fac16c))
2. Add max claimable tokens query ([613e8571](https://github.com/Stride-Labs/stride/commit/613e85711485d3bebeeb5777ba35e701cc795a43))
3. Interchain query proto cleanup ([9d5e1f6d](https://github.com/Stride-Labs/stride/commit/9d5e1f6d9e24113afa5b7f21e72a736bc8059b7f))
4. Add undelegation logging ([e74c34d12](https://github.com/Stride-Labs/stride/commit/e74c34d12a462e2d23463d717abfe01db9490d8f))
5. v4 upgrade changes ([928618b](https://github.com/Stride-Labs/stride/commit/928618b035610d1f5d24be3ab60628e10858aeff))
6. Revert HostZoneUnbonding status upon channel restoration ([730cf3d38](https://github.com/Stride-Labs/stride/commit/730cf3d38589887b57dfe3dd5de071273d5a9b73))
7. Informal Systems audit fixes IF-STRIDE-STAKEIBC-REDEEM_STAKE ([f8990a7e4f4](https://github.com/Stride-Labs/stride/commit/f8990a7e4f4f449984ff088c5fbe55908028cb14)), IF-STRIDE-STAKEIBC-UPDATEDELEGATIONBALANCES ([a8a1658ca](https://github.com/Stride-Labs/stride/commit/a8a1658ca4eb27ea7f8691219d748f07aad17964))

### Off-Chain changes

These changes do not affect any on-chain functionality, but have been implemented since `v4.0.0`.
1. Update Go Relayer to use Stride v3 ([faf3e7b2](https://github.com/Stride-Labs/stride/commit/faf3e7b21f4213b64a61bc2de5b400964cb61963))
2. Generalized Integration Tests ([80e8e2a4](https://github.com/Stride-Labs/stride/commit/80e8e2a49c3d63d8deabf4235e8e00151fcd8747))
3. Add localstride ([46a54f6c2](https://github.com/Stride-Labs/stride/commit/80e8e2a49c3d63d8deabf4235e8e00151fcd8747))
4. Dockernet updates ([b903866c64](https://github.com/Stride-Labs/stride/commit/b903866c6493e7f569d1975dc5f8efd34bdcc24e)), ([df2fcea21c](https://github.com/Stride-Labs/stride/commit/df2fcea21ce89e85d756d2cf8cbdf7e45bec2b0e))
5. Add ctx app test helper ([0e0238b3](https://github.com/Stride-Labs/stride/commit/0e0238b3dbddceaa7971e150649192e1f8be6982))


## [v3.0.0](https://github.com/Stride-Labs/stride/releases/tag/v3.0.0) - 2022-11-18
### On-Chain changes

1. Airdrop module ([24224f7](https://github.com/Stride-Labs/stride/commit/9be3314f7bca7e91f099d27ca11177639b76b468), [24224f7](https://github.com/Stride-Labs/stride/commit/24224f7386e7ee56781e7d254f9a48fab60a3bed)). Adds support for airdrop claims, including vesting. 
2. Proto reorganization ([8e3668a](https://github.com/Stride-Labs/stride/commit/8e3668a8e87381fb0f470ab60e4f0ba8590139cc)). This cleans up proto files to be more in-line with other Cosmos projects. 
3. Add Authz support ([e59c98e](https://github.com/Stride-Labs/stride/commit/e59c98e7bce574fa53e6e70222a80b974d84db3b)).
4. Cleanup ICQ Callbacks ([3ec6b8e](https://github.com/Stride-Labs/stride/commit/3ec6b8ebe9f4ba49aed3d671432a9d77e61b095a), [e747ac7](https://github.com/Stride-Labs/stride/commit/e747ac7bdd9385fdaa7d5cd6f2926f7efd519480)). Reorganizes ICQ Callbacks and errors self-heal faster. 
5. Versioning ([78fd819](https://github.com/Stride-Labs/stride/commit/78fd81918fe8f763f10525770eba1fee0a6dbe25), [0dbbbd8](https://github.com/Stride-Labs/stride/commit/0dbbbd867ffad5b331d09c155dca53a3f581ad5c), [dd6c26](https://github.com/Stride-Labs/stride/commit/dd6c264ea09448130484f7289eb085eb8bdb5766), [f77eac1](https://github.com/Stride-Labs/stride/commit/f77eac106291a59fd839c128f6aa9adb974eb7ef), [24f4b44](https://github.com/Stride-Labs/stride/commit/24f4b44e85518c0e800605265486af5f55f02693)). Updating versions to v3, as well as updating some Go modules.

### Off-Chain changes

These changes do not affect any on-chain functionality, but have been implemented since `v2.0.3`.

1. Testing flow to connect a local Stride chain to a production mainnet ([4cb9626](https://github.com/Stride-Labs/stride/commit/4cb9626a92b9cae5a970b3e4ddedf91bd44e8cef)). This is used to streamline onboarding a new Host Zone.  
2. Cleanup testing flow ([4133ccd](https://github.com/Stride-Labs/stride/commit/4133ccd3ef3f9b17c2602090078e3dae88e62e63), [1ba0b50](https://github.com/Stride-Labs/stride/commit/b18f483293ed6906b3f07ad5f6ab62e02130313d), [1ba0b50](https://github.com/Stride-Labs/stride/commit/1ba0b503ac4dec8fec167b680514dd367fc29bda), [fb03e0d](https://github.com/Stride-Labs/stride/commit/fb03e0d4cd7b7fd648e8b090d90a21cbb835a5d7)). There were a few deprecated testing scripts locally (e.g. testing outside of Docker, and spinning up a separate ICQ relayer), as well as some additional testing functionality (e.g. support for Linux, testing slashing)
3. Additional Docs ([c5cbb83](https://github.com/Stride-Labs/stride/commit/c5cbb83dfbc909f09e99a5633553fedeb0c0fd84)).

## [v2.0.3](https://github.com/Stride-Labs/stride/releases/tag/v2.0.3) - 2022-10-25

### On-Chain Changes
1. PENDING status for IBC/ICA function calls ([6660f60](https://github.com/Stride-Labs/stride/commit/6660f60094674b2e077f3775982ab4acc8a5ea96)). Added additional status field on internal accounting records to track when IBC/ICA calls are in flight and prevent re-submission. 
2. Add Validator through Governance ([c757364](https://github.com/Stride-Labs/stride/commit/c757364c4f532a8f7b9d17531f189c41cde90b14)). Added governance proposal type to enable adding validator's through governance. 
3. Validator Rebalancing ([725b991](https://github.com/Stride-Labs/stride/commit/725b9912073e4ff8c1fd5574ba4ebd68ec6aee88)). Added `rebalance-validators` transaction to redistribute delegations after validator weights are updated.

### Off-Chain changes

These changes do not affect any on-chain functionality, but have been implemented since `v1.0.4`.

1. Dockernet Multiple Host Zones ([f5c0b2c](https://github.com/Stride-Labs/stride/commit/f5c0b2cadcbc3995a6a91180a61fceb27afc4546)). Generalized host chain setup for dockernet to easily add and test the onboarding of new host zones.
2. Dockernet Upgrade Mode ([ceff377](https://github.com/Stride-Labs/stride/commit/ceff377d4dc3f4e8e5193c0eeb3b3ab94b74d91a)). Added functionality to dockernet to test upgrades locally.

## [v1.0.2](https://github.com/Stride-Labs/stride/releases/tag/v1.0.2) - 2022-09-06

* Fix stochastic sorting issue
* Add new query

## [v1.0.0](https://github.com/Stride-Labs/stride/releases/tag/v1.0.0) - 2022-09-04

Initial Release!