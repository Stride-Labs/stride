package types

import (
	sdk_types "github.com/cosmos/cosmos-sdk/types"
	gov_types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	oth_types "github.com/cosmos/gogoproto/types"
)

func NewVote(creator string, hostZoneId string, proposalId uint64, amount sdk_types.Int, option gov_types.VoteOption, available *oth_types.Timestamp) (Vote, error) {
	p := Vote{
		Creator: creator,
		HostZoneId:  hostZoneId,
		ProposalId: proposalId,
		Amount: amount,
		Option: option,
		TimeAvailable: available,
	}

	return p, nil
}

// Proposals is an array of proposal
type Votes []Vote
