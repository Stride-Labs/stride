package app

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ sdk.AnteDecorator = BlockedMsgsDecorator{}

// BlockedMsgsDecorator rejects any tx containing a blocked message type URL, including ones
// wrapped in (possibly nested) authz MsgExec. After the v35 mass undelegation of STRD, the
// staking messages that could re-lock it must not be reachable by any entry path (authority
// spec §4). The same rule applies in check, simulate and deliver.
type BlockedMsgsDecorator struct {
	blockedTypeUrls map[string]bool
}

func NewBlockedMsgsDecorator(typeUrls []string) BlockedMsgsDecorator {
	blockedTypeUrls := make(map[string]bool, len(typeUrls))
	for _, typeUrl := range typeUrls {
		blockedTypeUrls[typeUrl] = true
	}
	return BlockedMsgsDecorator{blockedTypeUrls: blockedTypeUrls}
}

func (decorator BlockedMsgsDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	if err := decorator.checkMsgs(tx.GetMsgs()); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

// checkMsgs returns an unauthorized error for the first blocked message, recursing into MsgExec
func (decorator BlockedMsgsDecorator) checkMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		typeUrl := sdk.MsgTypeURL(msg)
		if decorator.blockedTypeUrls[typeUrl] {
			return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is disabled: the chain is winding down", typeUrl)
		}

		execMsg, isExec := msg.(*authz.MsgExec)
		if !isExec {
			continue
		}
		innerMsgs, err := execMsg.GetMessages()
		if err != nil {
			return err
		}
		if err := decorator.checkMsgs(innerMsgs); err != nil {
			return err
		}
	}
	return nil
}
