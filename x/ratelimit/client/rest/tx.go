package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

func ProposalAddRateLimitRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "add-rate-limit",
		Handler:  func(w http.ResponseWriter, r *http.Request) {},
	}
}

func ProposalUpdateRateLimitRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "update-rate-limit",
		Handler:  func(w http.ResponseWriter, r *http.Request) {},
	}
}

func ProposalRemoveRateLimitRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "remove-rate-limit",
		Handler:  func(w http.ResponseWriter, r *http.Request) {},
	}
}

func ProposalResetRateLimitRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "reset-rate-limit",
		Handler:  func(w http.ResponseWriter, r *http.Request) {},
	}
}
