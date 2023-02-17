package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeToggleOracle = "ToggleOracle"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeToggleOracle)
}

var (
	_ govtypes.Content = &ToggleOracleProposal{}
)

func NewToggleOracleProposal(title, description, oracleChainId string) govtypes.Content {
	return &ToggleOracleProposal{
		Title:         title,
		Description:   description,
		OracleChainId: oracleChainId,
	}
}

func (p *ToggleOracleProposal) GetTitle() string { return p.Title }

func (p *ToggleOracleProposal) GetDescription() string { return p.Description }

func (p *ToggleOracleProposal) ProposalRoute() string { return RouterKey }

func (p *ToggleOracleProposal) ProposalType() string {
	return ProposalTypeToggleOracle
}

func (p *ToggleOracleProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if p.OracleChainId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-chain-id is required")
	}

	return nil
}

func (p *ToggleOracleProposal) String() string {
	return fmt.Sprintf(`Toggle Oracle Proposal:
	Title:         %s
	Description:   %s
	OracleChainId: %s
  `, p.Title, p.Description, p.OracleChainId)
}
