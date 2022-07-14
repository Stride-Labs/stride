<!--
order: 0
title: "Epochs Overview"
parent:
  title: "epochs"
-->

# `epochs`

## Abstract

Often in the SDK, we would like to run certain code periodically. The purpose of `epochs` module is to allow other modules to set that they would like to be signaled once every period. So another module can specify it wants to execute code once a week, starting at UTC-time = x. `epochs` creates a generalized epoch interface to other modules so that they can easily be signalled upon such events.

## Contents

1. **[Concepts](#concepts)**
2. **[State](#state)**
3. **[Events](03_events.md)**
4. **[Keeper](04_keeper.md)**  
5. **[Hooks](05_hooks.md)**  
6. **[Queries](06_queries.md)**  
7. **[Future improvements](07_future_improvements.md)**

# Concepts

The purpose of `epochs` module is to provide generalized epoch interface to other modules so that they can easily implement epochs without keeping own code for epochs.

# State

Epochs module keeps `EpochInfo` objects and modifies the information as epochs info changes.
Epochs are initialized as part of genesis initialization, and modified on begin blockers or end blockers.

### Epoch information type

```protobuf
message EpochInfo {
    string identifier = 1;
    google.protobuf.Timestamp start_time = 2 [
        (gogoproto.stdtime) = true,
        (gogoproto.nullable) = false,
        (gogoproto.moretags) = "yaml:\"start_time\""
    ];
    google.protobuf.Duration duration = 3 [
        (gogoproto.nullable) = false,
        (gogoproto.stdduration) = true,
        (gogoproto.jsontag) = "duration,omitempty",
        (gogoproto.moretags) = "yaml:\"duration\""
    ];
    int64 current_epoch = 4;
    google.protobuf.Timestamp current_epoch_start_time = 5 [
        (gogoproto.stdtime) = true,
        (gogoproto.nullable) = false,
        (gogoproto.moretags) = "yaml:\"current_epoch_start_time\""
    ];
    bool epoch_counting_started = 6;
    reserved 7;
    int64 current_epoch_start_height = 8;
}
```

EpochInfo keeps `identifier`, `start_time`,`duration`, `current_epoch`, `current_epoch_start_time`,  `epoch_counting_started`, `current_epoch_start_height`.

1. `identifier` keeps epoch identification string.
2. `start_time` keeps epoch counting start time, if block time passes `start_time`, `epoch_counting_started` is set.
3. `duration` keeps target epoch duration.
4. `current_epoch` keeps current active epoch number.
5. `current_epoch_start_time` keeps the start time of current epoch.
6. `epoch_number` is counted only when `epoch_counting_started` flag is set.
7. `current_epoch_start_height` keeps the start block height of current epoch.