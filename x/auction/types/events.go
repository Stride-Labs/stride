package types

// Event types and attribute keys for auction module
const (
	EventTypeBidPlaced   = "bid_placed"
	EventTypeBidAccepted = "bid_accepted"

	AttributeKeyAuctionName   = "auction_name"
	AttributeKeyBidder        = "bidder"
	AttributeKeyPaymentAmount = "payment_amount"
	AttributeKeyPaymentDenom  = "payment_denom"
	AttributeKeySellingAmount = "selling_amount"
	AttributeKeySellingDenom  = "selling_denom"
	AttributeKeyPrice         = "discounted_price"
)
