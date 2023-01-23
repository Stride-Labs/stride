package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

func TestGovRemoveRateLimit(t *testing.T) {
	apptesting.SetupConfig()

	validTitle := "RemoveRateLimit"
	validDescription := "Removing a rate limit"
	validDenom := "denom"
	validChannelId := "channel-0"

	tests := []struct {
		name     string
		proposal types.RemoveRateLimitProposal
		err      string
	}{
		{
			name: "successful message",
			proposal: types.RemoveRateLimitProposal{
				Title:       validTitle,
				Description: validDescription,
				Denom:       validDenom,
				ChannelId:   validChannelId,
			},
		},
		{
			name: "invalid title",
			proposal: types.RemoveRateLimitProposal{
				Title:       "",
				Description: validDescription,
				Denom:       validDenom,
				ChannelId:   validChannelId,
			},
			err: "title cannot be blank",
		},
		{
			name: "invalid description",
			proposal: types.RemoveRateLimitProposal{
				Title:       validTitle,
				Description: "",
				Denom:       validDenom,
				ChannelId:   validChannelId,
			},
			err: "description cannot be blank",
		},
		{
			name: "invalid denom",
			proposal: types.RemoveRateLimitProposal{
				Title:       validTitle,
				Description: validDescription,
				Denom:       "",
				ChannelId:   validChannelId,
			},
			err: "invalid denom",
		},
		{
			name: "invalid channel-id",
			proposal: types.RemoveRateLimitProposal{
				Title:       validTitle,
				Description: validDescription,
				Denom:       validDenom,
				ChannelId:   "chan-1",
			},
			err: "invalid channel-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.proposal.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.proposal.Denom, validDenom, "denom")
				require.Equal(t, test.proposal.ChannelId, validChannelId, "channelId")
			} else {
				require.ErrorContains(t, test.proposal.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
