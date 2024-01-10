package app

import (
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// This is used for non-legacy gov transactions
// Returning true cause all txs are whitelisted
func IsModuleWhiteList(typeUrl string) bool {
	return true
}

// This is used for legacy gov transactions
// Returning true cause all txs are whitelisted
func IsProposalWhitelisted(content govv1beta1.Content) bool {
	return true
}
