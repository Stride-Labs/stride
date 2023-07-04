package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakeibckeeper "github.com/Stride-Labs/stride/v11/x/stakeibc/keeper"
)

type HandlerOptions struct {
	AccountKeeper          ante.AccountKeeper
	BankKeeper             bankkeeper.Keeper
	FeegrantKeeper         feegrantkeeper.Keeper
	SignModeHandler        authsigning.SignModeHandler
	SigGasConsumer         func(meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	TxFeeChecker           ante.TxFeeChecker
	StakeIbcKeeper 		   stakeibckeeper.Keeper
}

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		stakeibckeeper.NewStakeIbcAnteDecorator(options.StakeIbcKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
