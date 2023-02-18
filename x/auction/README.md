# The Auction Module

This module allows assets to be sold to bidders in a variety of different ways.  An auction pool is a potential auction which can be run as well as the restrictions on that auction.  For example an address with which assets to auction, which auction algorithms are allowed, and various other constraints such as supply, default parameters etc.  In contrast to the pool, an auction is a specific instance of a running or finished auction with variables for changing state. Each pool may only have a single auction running at a time.

For testing now, 3 default auction pools are created which will allow for different types of auctions to be run. In the future, a different module such as stakeibc will likely call the CreateAuctionPool method to specify only the pools needed.

You can see what pools are available by running
`strided q auction list-auction-pool`

Notice in the allowed algorithms of these pools, the first is a SealedBidAuction type, the second is an AscendingAuction, the third is a Descending auction.

# Auction Algorithms

Part of the purpose of this project is to create a general purpose module which might be used in other projects one day. Part of the purpose of this project is for Stride to use an auction to distribute a finite amount of liquid interest in 'instant unbonding' which generates buying pressure and upholds a token - staked token peg.  To satisfy the first goal, the code cleanly separates several algorithms into simple interfaces in `types/auction.go` with structure such that adding future algorithms should be simple.  To satisfy the second part, the majority of the testing focus and several defaults have been set to SealedBid auction algorithm which is what Stride would use for the instant unbonding purpose.

Ascending (English) auctions are the simplest. They are set to run for a known auctionDuration after which they will end and payout.  However, each time a new bid comes in, if the auction is close to ending the time will be extended by a known extensionDuration.  This is the familiar ebay style auction.

Descending (Dutch) auctions start at a high bid which slowly ticks down over time.  When a bidder signals they accept the current price, they instantly make the sale.  If their requested volume doesn't cover the entire supply then the auction continues until other bidders jump in at better and better prices.  

SealedBid is the most complicated.  Bidders submit their bids in a way that no one else is aware of their bid until the end of the auction.  This is implemented in two phases, a submit sealed bid phase followed by a reveal phase.  The sealed bid is a hash of the bid (and a salt to avoid guessing) and the reveal is the original bid which can be verified to not have changed by running the hash and comparing to what they submitted in the first phase.  To raise the cost of a sybil attack, collateral is taken at the time of the sealed bid and returned to the user at the end of the auction only if they actually revealed their bid.

Somewhat specific to Stride's usecase is the redemption rate.  Instant unbonding will take in some stToken but the amount of Token returned has to be determined by the redemption rate -- thus the price represents an additional marginal amount which the bidder is willing to pay for this instant redemption.  In the more general auction setting, the price represents the ratio of what they would pay to what they would get.

# Testing Auctions

Locally create bid JSON as an env var.  Orders object has list of one or more (Price, Volume) pairs
`bid_data='{"Orders":[{"Price":102000,"Volume":4000000000},{"Price":93000,"Volume":3000000000]}'`
Create a salt value which can be anything you want, keep it secret until the reveal phase of the auction
`salt_data='seasalt'`
Create a sha256 of the bid_data and salt_data appended.  The -n is important to avoid unwanted new lines
`hash_data=$(echo -n "$bid_data$salt_data" | sha256sum | cut -c 1-64)`

With Stride, Gaia, and relayers running, in another shell window run
`tail -f ./dockernet/logs/stride.log | grep --line-buffered "\[auction\]"`
This will stream the stride log and filter for only events coming from the auction

To see the available auction pools run
`strided q auction list-auction-pool`
Notice that in the defaults right now id 0 will be a SealedBid, id 1 will be Ascending, id 2 will be Descending

Start a sealed bid auction with
`strided tx auction start-auction cosmoshub-4 0 --from stride1uk4ze...`
During the first phase, submit a sealed bid
`strided tx auction submit-sealed-bid cosmoshub-4 0 $hash_data --from stride1uk4ze...`
When you see the state change to REVEAL in the logs, reveal the bid with
`strided tx auction reveal-bid cosmoshub-4 0 "$bid_data" $salt_data --from stride1uk4ze...`
Shortly after the state should change to Payout and the auction will resolve printing out who would be paid how much.  You can do this whole process in several different shells with different data and from addresses to see a more complicated resolution.


Start an ascending auction with
`strided tx auction start-auction cosmoshub-4 1 --from stride1uk4ze...`
While running, submit a bid
`strided tx auction submit-open-bid cosmoshub-4 1 "$bid_data" --from stride1uk4ze...`
When the auction reaches its time limit with no bids coming in recently the auction will end.  If you submit bids near to the end of the auction you will see the ending time get extended with this info printed in the logs.


Start a descending auction with
`strided tx auction start-auction cosmoshub-4 2 --from stride1uk4ze...`
When the auction starts, the logs will print out the current bid which will move down every few blocks.  This bid is the price in the bid object, when you send in bids all that matters here is the volume. Submit a bid
`strided tx auction submit-open-bid cosmoshub-4 2 "$bid_data" --from stride1uk4ze...`
You should see in the logs instant "payout" the volume requested.  This will either end the auction if your requested volume covered the supply or it will continue if there is supply remaining.  The auction ends when supply is gone, or the bid gets below the minimum bid setting.


# Auction TODO

Right now the resolve methods will calculate who it will send how much to, but the bankKeeper is not yet being used to actually make these sends (or to refund the collateral in the case of the SealedBidAuction).  Perhaps these methods should not be on the auction types but on the keeper to facilitate the payouts.

There are several params in the pool settings which are being ignored currently like the minSupply which would prevent an auction from starting if the poolAddress had to few funds in it, etc.

Eventually the create stake pool method should be called from an external module like stakeibc instead of initialized from defaults in genesis.

Currently, when stake pools are made, it looks at how much and what denom coins are already in the pool address to set the denoms in the pool properties.  In reality, we want to specify the denoms and amount for the pool when created and have it find the correct coins if they exist in that pool address.

An end to end script which goes through this whole process with multiple accounts bidding would be nice for testing.
