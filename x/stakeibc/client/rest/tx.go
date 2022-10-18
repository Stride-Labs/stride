package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

// TODO: Do we need to implement these?
func ProposalAddValidatorRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "add-validator",
		Handler:  newAddValidatorProposalHandler(clientCtx),
	}
}

func newAddValidatorProposalHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func ProposalDeleteValidatorRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "delete-validator",
		Handler:  newDeleteValidatorProposalHandler(clientCtx),
	}
}

func newDeleteValidatorProposalHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
