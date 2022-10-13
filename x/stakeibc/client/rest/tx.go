package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type (
	addValidatorReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
		Amount  sdk.Coins    `json:"amount" yaml:"amount"`
	}
)

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	// Add validator proposal
	r.HandleFunc(
		"/stakeibc.add_validator",
		newAddValidatorHandlerFn(clientCtx),
	).Methods("POST")
}

func newAddValidatorHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// return nil
		// var req AddValidatorProposalReq
		// if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
		// 	return
		// }

		// fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		// if rest.CheckBadRequestError(w, err) {
		// 	return
		// }

		// msg := types.NewAddValidatorProposal(req.Title, req.Description, req.HostZone, req.ValidatorName, req.ValidatorAddress)
		// if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
		// 	return
		// }

		// tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

// Auxiliary

// func checkDelegatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.AccAddress, bool) {
// 	addr, err := sdk.AccAddressFromBech32(mux.Vars(r)["delegatorAddr"])
// 	if rest.CheckBadRequestError(w, err) {
// 		return nil, false
// 	}

// 	return addr, true
// }

// func checkValidatorAddressVar(w http.ResponseWriter, r *http.Request) (sdk.ValAddress, bool) {
// 	addr, err := sdk.ValAddressFromBech32(mux.Vars(r)["validatorAddr"])
// 	if rest.CheckBadRequestError(w, err) {
// 		return nil, false
// 	}

// 	return addr, true
// }
