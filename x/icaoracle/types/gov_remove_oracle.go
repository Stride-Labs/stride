package types

import (
	"fmt"
	"strings"

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

func NewRemoveOracleProposal(title, description, oracleMoniker string) govtypes.Content {
	return &RemoveOracleProposal{
		Title:         title,
		Description:   description,
		OracleMoniker: oracleMoniker,
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

	if p.OracleMoniker == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-moniker is required")
	}
	if strings.Contains(p.OracleMoniker, " ") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-moniker cannot contain any spaces")
	}

	return nil
}

func (p *RemoveOracleProposal) String() string {
	return fmt.Sprintf(`Remove Oracle Proposal:
	Title:         %s
	Description:   %s
	OracleMoniker: %s
  `, p.Title, p.Description, p.OracleMoniker)
}
