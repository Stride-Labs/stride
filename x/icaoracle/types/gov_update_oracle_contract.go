package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdateOracleContract = "UpdateOracleContract"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateOracleContract)
}

var (
	_ govtypes.Content = &UpdateOracleContractProposal{}
)

func NewUpdateOracleContractProposal(title, description, oracleChainId, contractAddress string) govtypes.Content {
	return &UpdateOracleContractProposal{
		Title:           title,
		Description:     description,
		OracleChainId:   oracleChainId,
		ContractAddress: contractAddress,
	}
}

func (p *UpdateOracleContractProposal) GetTitle() string { return p.Title }

func (p *UpdateOracleContractProposal) GetDescription() string { return p.Description }

func (p *UpdateOracleContractProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateOracleContractProposal) ProposalType() string {
	return ProposalTypeUpdateOracleContract
}

func (p *UpdateOracleContractProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if p.OracleChainId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-chain-id is required")
	}

	if p.ContractAddress == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "contract address is required")
	}

	return nil
}

func (p *UpdateOracleContractProposal) String() string {
	return fmt.Sprintf(`Update Oracle Contract Proposal:
	Title:         %s
	Description:   %s
	OracleChainId: %s
	ContractAddress: %s
  `, p.Title, p.Description, p.OracleChainId, p.ContractAddress)
}
