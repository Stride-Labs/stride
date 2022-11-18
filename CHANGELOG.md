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

## Unreleased (v2.0.0)

This release includes...

* changes...

### Breaking Changes


#### Golang API breaks


### Features


### Bug Fixes
* [#XXX](https://github.com/Stride-Labs/stride) Link to fix

### Improvements

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

## [v1.0.2](https://github.com/Stride-Labs/stride/releases/tag/v1.0.2) - 2022-09-06

* Fix stochastic sorting issue
* Add new query

## [v1.0.0](https://github.com/Stride-Labs/stride/releases/tag/v1.0.0) - 2022-09-04

Initial Release!