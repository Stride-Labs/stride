---
title: "RateLimit"
excerpt: ""
category: 6392913957c533007128548e
---
# RateLimit Module
## Overview
This `ratelimit` module is a native golang implementation, inspired by Osmosis's CosmWasm [`ibc-rate-limit`](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-rate-limit) module. The module is meant as a safety control in the event of a bug, attack, or economic failure of an external zone. It prevents massive inflows or outflows of IBC tokens to/from Stride in a short time frame. See [here](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-rate-limit#motivation) for an excellent summary by the Osmosis team on the motivation for rate limiting.

Each rate limit is applied at a ChannelID + Denom granularity and is evaluated in evenly spaced fixed windows. For instance, a rate limit might be specified on `uosmo` (denominated as `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` on Stride), on the Stride <-> Osmosis transfer channel (`channel-5`), with a 24 hour window. 

Each rate limit will also have a configurable threshold that dictates the max inflow/outflow along the channel. The threshold is represented as a percentage of the total value along the channel. The channel value is calculated by querying the total supply of the denom at the start of the time window, and it remains constant until the window expires. For instance, the rate limit described above might have a threshold of 10% for both inflow and outflow. If the total supply of `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` was 100, then any transfer that would cause a net inflow or outflow greater than 10 (i.e. greater than 10% the channel value) would be rejected. Once the time window expires, the net inflow and outflow are reset to 0 and the channel value is re-calculated. 

The *net* inflow and outflow is used (rather than the total inflow/outflow) to prevent DOS attacks where someone repeatedly sends the same token back and forth across the same channel, causing the rate limit to be reached.

The module is implemented as IBC Middleware around the transfer module. The epoch's module is leveraged to determine when each rate limit window has expired (each window is denominated in hours). This means all rate limit windows with the same window duration will start and end at the same time.

## Implementation
Each rate limit is defined by the following three components:
1. **Path**: Defines the `ChannelId` and `Denom`
2. **Quota**: Defines the rate limit time window (`DurationHours`) and the max threshold for inflows/outflows (`MaxPercentRecv` and `MaxPercentSend` respectively)
3. **Flow**: Stores the current `Inflow`, `Outflow` and `ChannelValue`. Each time a quota expires, the inflow and outflow get reset to 0 and the channel value gets recalculated. Throughout the window, the inflow and outflow each increase monotonically. The net flow is used when determining if a transfer would exceed the quota. 
    * For `Send` packets: 
    $$\text{Exceeds Quota if:} \left(\frac{\text{Outflow} - \text{Inflow} + \text{Packet Amount}}{\text{ChannelValue}}\right) > \text{MaxPercentSend}$$
    * For `Receive` packets: 
    $$\text{Exceeds Quota if:} \left(\frac{\text{Inflow} - \text{Outflow} + \text{Packet Amount}}{\text{ChannelValue}}\right) > \text{MaxPercentRecv}$$

## Example Walk-Through
Using the example above, let's say we created a 24 hour rate limit on `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` ("`ibc/uosmo`"), `channel-5`, with a 10% send and receive threshold. 
1. At the start of the window, the supply will be queried, to determine the channel value. Let's say the total supply was 100
2. If someone transferred `8uosmo` from `Osmosis -> Stride`, the `Inflow` would increment by 8
3. If someone tried to transfer another `8uosmo` from `Osmosis -> Stride`, it would exceed the quota since `(8+8)/100 = 16%` (which is greater than 10%) and thus, the transfer would be rejected.
4. If someone tried to transfer `12ibc/uosmo` from Stride -> Osmosis, the `Outflow` would increment by 12. Notice, even though 12 is greater than 10% the total channel value, the *net* outflow is only `4uatom` (since it's offset by the `8uatom` `Inflow`). As a result, this transaction would succeed.
5. Now if the person in (3) attempted to retry their transfer of`8uosmo` from `Osmosis -> Stride`, the `Inflow` would increment by 8 and the transaction would succeed (leaving a net inflow of 4).
6. Finally, at the end of the 24 hours, the `Inflow` and `Outflow` would get reset to 0 and the `ChannelValue` would be re-calculated. In this example, the new channel value would be 104 (since more `uosmo` was sent to Stride, and thus more `ibc/uosmo` was minted)

| Step |            Description           | Transfer Status | Inflow | Outflow | Net Inflow | Net Outflow | Channel Value |
|:----:|:--------------------------------:|:---------------:|:------:|:-------:|:----------:|:-----------:|:-------------:|
|   1  |        Rate limit created        |                 |    0   |    0    |            |             |      100      |
|   2  |      8usomo Osmosis → Stride     |    Successful   |    8   |    0    |     8%     |             |      100      |
|   3  |      8usomo Osmosis → Stride     |     Rejected    |   16   |    0    | 16% (>10%) |             |      100      |
|   3  | State reverted after rejected Tx |                 |    8   |    0    |     8%     |             |      100      |
|   4  |   12ibc/uosmo Stride → Osmosis   |    Successful   |    8   |    12   |            |      4%     |      100      |
|   5  |      8usomo Osmosis → Stride     |    Successful   |   16   |    12   |     4%     |             |      100      |
|   6  |            Quota Reset           |                 |    0   |    0    |            |             |      104      |

## Denom Blacklist
The module also contains a blacklist to completely halt all IBC transfers for a given denom. There are keeper functions to add or remove denoms from the blacklist; however, these functions are not exposed externally through transactions or governance, and they should only be leveraged internally from the protocol in extreme scenarios.

## Denoms
We always want to refer to the channel ID and denom as they appear on Stride. For instance, in the example above, we would store the rate limit with denom `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` and `channel-5`, instead of `uosmo` and `channel-326` (the ChannelID on Osmosis).

However, since the ratelimit module acts as middleware to the transfer module, the respective denoms need to be interpreted using the denom trace associated with each packet. There are a few scenarios at play here...

### Send Packets
The denom that the rate limiter will use for a send packet depends on whether it was a native token (e.g. ustrd, stuatom, etc.) or non-native token (e.g. ibc/...)...
#### Native vs Non-Native
* We can identify if the token is native or not by parsing the denom trace from the packet
    * If the token is **native**, it **will not** have a prefix (e.g. `ustrd`)
    * If the token is **non-native**, it **will** have a prefix (e.g. `transfer/channel-X/uosmo`)
#### Determining the denom in the rate limit
* For **native** tokens, return as is (e.g. `ustrd`)
* For **non-native** tokens, take the ibc hash (e.g. hash `transfer/channel-X/uosmo` into `ibc/...`)

### Receive Packets
The denom that the rate limiter will use for a receive packet depends on whether it was a source or sink.

#### Source vs Sink
As a token travels across IBC chains, its path is recorded in the denom trace. 

* **Sink**: If the token moves **forward**, to a chain different than its previous hop, the destination chain acts as a **sink zone**, and the new port and channel are **appended** to the denom trace.
    * Ex1: `uatom` is sent from Cosmoshub to Stride 
      * Stride is the first destination for `uatom`, and acts as a sink zone
      * The IBC denom becomes the hash of: `/{stride-port)/{stride-channel}/uatom`
    * Ex2: `uatom` is sent from Cosmoshub to Osmosis then to Stride
      * Here the receiving chain (Stride) is not the same as the previous hop (Cosmoshub), so Stride, once again, is acting as a sink zone
      *  The IBC denom becomes the hash of: `/{stride-port)/{stride-channel}/{osmosis-port}/{osmosis-channel}/uatom`
   
* **Source**: If the token moves **backwards** (i.e. revisits the last chain it was sent from), the destination chain is acting as a **source zone**, and the port and channel are **removed** from the denom trace - undoing the last hop. Should a token reverse its course completely and head back along the same path to its native chain, the denom trace will unwind and reduce back down to the original base denom.
    * Ex1: `ustrd` is sent from Stride to Osmosis, and then back to Stride 
      * Here the trace reduces from `/{osmosis-port}/{osmosis-channel}/ustrd` to simply `ustrd`
    * Ex2: `ujuno` is sent to Stride, then to Osmosis, then back to Stride 
      * Here the trace reduces from `/{osmosis-port}/{osmosis-channel}/{stride-port}/{stride-channel}/ujuno` to just `/{stride-port}/{stride-channel}/ujuno` (the Osmosis hop is removed)
    * Stride is the source in the examples above because the token went back and forth from Stride -> Osmosis -> Stride

For a more detailed explanation, see the[ ICS-20 ADR](https://github.com/cosmos/ibc-go/blob/main/docs/architecture/adr-001-coin-source-tracing.md#example) and [spec](https://github.com/cosmos/ibc/tree/main/spec/app/ics-020-fungible-token-transfer).

#### Determining the denom in the rate limit
* If the chain is acting as a **Sink**: Add on the Stride port and channel and hash it
    * Ex1: `uosmo` sent from Osmosis to Stride
        * Packet Denom Trace: `uosmo`
        * (1) Add Stride Channel as Prefix:  `transfer/channel-X/uosmo`
        * (2) Hash: `ibc/...`

    * Ex2: `ujuno` sent from Osmosis to Stride
        * Packet Denom Trace: `transfer/channel-Y/ujuno` (where channel-Y is the Juno <> Osmosis channel)
        * (1) Add Stride Channel as Prefix:  `transfer/channel-X/transfer/channel-Y/ujuno`
        * (2) Hash: `ibc/...`

* If the chain is acting as a **Source**: First, remove the prefix. Then if there is still a trace prefix, hash it
    * Ex1: `ustrd` sent back to Stride from Osmosis
        * Packet Denom: `transfer/channel-X/ustrd`
        * (1) Remove Prefix: `ustrd`
        * (2) No trace remaining, leave as is: `ustrd`
    * Ex2: juno was sent to Stride, then to Osmosis, then back to Stride
        * Packet Denom: `transfer/channel-X/transfer/channel-Z/ujuno`
        * (1) Remove Prefix: `transfer/channel-Z/ujuno`
        * (2) Hash: `ibc/...`

## State
```go
RateLimit
    Path
        Denom string
        ChannelId string
    Quota
        MaxPercentSend sdkmath.Int
        MaxPercentRecv sdkmath.Int
        DurationHours uint64
    Flow
        Inflow sdkmath.Int
        Outflow sdkmath.Int
        ChannelValue sdkmath.Int
```

## Keeper functions
```go
// Stores a RateLimit object in the store
SetRateLimit(rateLimit types.RateLimit)

// Removes a RateLimit object from the store
RemoveRateLimit(denom string, channelId string)

// Reads a RateLimit object from the store
GetRateLimit(denom string, channelId string)

// Gets a list of all RateLimit objects
GetAllRateLimits()

// Resets the Inflow and Outflow of a RateLimit and re-calculates the ChannelValue
ResetRateLimit(denom string, channelId string) 

// Checks whether a packet will exceed a rate limit quota
// If it does not exceed the quota, it updates the `Inflow` or `Outflow`
// If it exceeds the quota, it returns an error
CheckRateLimitAndUpdateFlow(direction types.PacketDirection, denom string, channelId string, amount sdkmath.Int)
```

## Middleware Functions
```go
SendRateLimitedPacket (ICS4Wrapper SendPacket)
ReceiveRateLimitedPacket (IBCModule OnRecvPacket)
```

## Transactions (via Governance)
```go
// Adds a new rate limit
// Errors if:
//   - `ChannelValue` is 0 (meaning supply of the denom is 0)
//   - Rate limit already exists (as identified by the `channel_id` and `denom`)
//   - Channel does not exist
AddRateLimit()
{"denom": string, "channel_id": string, "duration_hours": string, "max_percent_send": string, "max_percent_recv": string}

// Updates a rate limit quota, and resets the rate limit
// Errors if: 
//   - Rate limit does not exist (as identified by the `channel_id` and `denom`)
UpdateRateLimit()
{"denom": string, "channel_id": string, "duration_hours": string, "max_percent_send": string, "max_percent_recv": string}

// Resets the `Inflow` and `Outflow` of a rate limit to 0, and re-calculates the `ChannelValue`
// Errors if: 
//   - Rate limit does not exist (as identified by the `channel_id` and `denom`)
ResetRateLimit() 
{"denom": string, "channel_id": string}

// Removes the rate limit from the store
// Errors if: 
//   - Rate limit does not exist (as identified by the `channel_id` and `denom`)
RemoveRateLimit()
{"denom": string, "channel_id": string}
```

## Queries
```go
// Queries all rate limits
//   CLI: 
//      strided q ratelimit list-rate-limits
//   API: 
//      /Stride-Labs/stride/ratelimit/ratelimits
QueryRateLimits()

// Queries a specific rate limit given a ChannelID and Denom
//   CLI:
//      strided q ratelimit rate-limit [denom] [channel-id]
//   API:
//      /Stride-Labs/stride/ratelimit/ratelimit/{denom}/{channel_id}
QueryRateLimit(denom string, channelId string)

// Queries all rate limits associated with a given host chain
//   CLI: 
//      strided q ratelimit rate-limits-by-chain [chain-id]
//   API: 
//      /Stride-Labs/stride/ratelimit/ratelimits/{chain_id}
QueryRateLimitsByChainId(chainId string)
```
