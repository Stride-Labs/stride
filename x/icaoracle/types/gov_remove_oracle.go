package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeRemoveOracle = "RemoveOracle"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeRemoveOracle)
}

var (
	_ govtypes.Content = &RemoveOracleProposal{}
)

func NewRemoveOracleProposal(title, description, oracleChainId string) govtypes.Content {
	return &RemoveOracleProposal{
		Title:         title,
		Description:   description,
		OracleChainId: oracleChainId,
	}
}

func (p *RemoveOracleProposal) GetTitle() string { return p.Title }

func (p *RemoveOracleProposal) GetDescription() string { return p.Description }

func (p *RemoveOracleProposal) ProposalRoute() string { return RouterKey }

func (p *RemoveOracleProposal) ProposalType() string {
	return ProposalTypeRemoveOracle
}

func (p *RemoveOracleProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if p.OracleChainId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-chain-id is required")
	}

	return nil
}

func (p *RemoveOracleProposal) String() string {
	return fmt.Sprintf(`Remove Oracle Proposal:
	Title:         %s
	Description:   %s
	OracleChainId: %s
  `, p.Title, p.Description, p.OracleChainId)
}
