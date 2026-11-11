---
layout: default
title: Hype Train API
description: Manage hype train events for Twitch channels.
---

> **Note:** Twitch **removed** the Get Hype Train Events endpoint on 2026-02-05 in favor of [Get Hype Train Status](#gethypetrainstatus). Any live call now returns `404`. Use `GetHypeTrainStatus` (below) or [EventSub](eventsub.md) for real-time hype train events. `GetHypeTrainEvents` and its types remain for one release cycle and will be removed in the next major release.

## GetHypeTrainEvents

Get the broadcaster's hype train events.

**Requires:** `channel:read:hype_train` scope

**REMOVED BY TWITCH (2026-02-05):** This endpoint no longer exists and returns `404`. Use [`GetHypeTrainStatus`](#gethypetrainstatus) instead. Retained as deprecated for one release cycle.

```go
resp, err := client.GetHypeTrainEvents(ctx, &helix.GetHypeTrainEventsParams{
    BroadcasterID: "270954519",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})
if err != nil {
    log.Fatal(err)
}
for _, event := range resp.Data {
    fmt.Printf("Event: %s (%s)\n", event.ID, event.EventType)
    fmt.Printf("Hype Train Level: %d\n", event.EventData.Level)
    fmt.Printf("Progress: %d / %d\n", event.EventData.Total, event.EventData.Goal)
    fmt.Printf("Started: %s, Expires: %s\n", event.EventData.StartedAt, event.EventData.ExpiresAt)
}

// Paginate through more results
if resp.Pagination != nil && resp.Pagination.Cursor != "" {
    resp, err = client.GetHypeTrainEvents(ctx, &helix.GetHypeTrainEventsParams{
        BroadcasterID: "270954519",
        PaginationParams: &helix.PaginationParams{
            After: resp.Pagination.Cursor,
        },
    })
}
```

**Parameters:**
- `BroadcasterID` (string): The ID of the broadcaster
- `PaginationParams.First` (int, optional): Maximum number of results (1-100)
- `PaginationParams.After` (string, optional): Cursor for forward pagination

**Response Fields:**
- `ID` (string): Event ID
- `EventType` (string): Type of event (e.g., `hypetrain.progression`)
- `EventTimestamp` (time.Time): When the event occurred
- `Version` (string): Event version
- `EventData.ID` (string): Hype train ID
- `EventData.BroadcasterID` (string): Broadcaster's user ID
- `EventData.Level` (int): Current level of the hype train
- `EventData.Total` (int): Total points contributed
- `EventData.Goal` (int): Points needed to reach the next level
- `EventData.StartedAt` (time.Time): When the hype train started
- `EventData.ExpiresAt` (time.Time): When the hype train expires
- `EventData.CooldownEndTime` (time.Time): When a new hype train can start
- `EventData.LastContribution` (HypeTrainContribution): Most recent contribution
- `EventData.TopContributions` ([]HypeTrainContribution): Top contributors

**Sample Response (from Twitch docs):**
```json
{
  "data": [
    {
      "id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
      "event_type": "hypetrain.progression",
      "event_timestamp": "2020-04-24T20:07:24Z",
      "version": "1.0",
      "event_data": {
        "broadcaster_id": "270954519",
        "cooldown_end_time": "2020-04-24T20:13:21.003802269Z",
        "expires_at": "2020-04-24T20:12:21.003802269Z",
        "goal": 1800,
        "id": "70f0c7d8-ff60-4c50-b138-f3a352833b50",
        "last_contribution": {
          "total": 200,
          "type": "BITS",
          "user": "134247454"
        },
        "level": 2,
        "started_at": "2020-04-24T20:05:47.30473127Z",
        "top_contributions": [
          {
            "total": 600,
            "type": "BITS",
            "user": "134247450"
          }
        ],
        "total": 600
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjI3MDk1NDUxOToxNTg3NzU4ODQ0OjFiMEFzYkluQ0haVzJTUUZRa0N6cU4wN0liMiJ9fQ"
  }
}
```

## GetHypeTrainStatus

Get the Hype Train status for a broadcaster's channel: the train in progress (if any) and the channel's record trains. `Current` is `nil` when no train is active.

**Requires:** user access token with the `channel:read:hype_train` scope, for the broadcaster

```go
status, err := client.GetHypeTrainStatus(ctx, "1337")
if err != nil {
    log.Fatal(err)
}
if status == nil || status.Current == nil {
    fmt.Println("No active hype train")
    return
}

cur := status.Current
fmt.Printf("Hype Train (%s) level %d: %d / %d at this level, %d total\n",
    cur.Type, cur.Level, cur.Progress, cur.Goal, cur.Total)
fmt.Printf("Started: %s, Expires: %s\n", cur.StartedAt, cur.ExpiresAt)

fmt.Println("Top Contributors:")
for _, contrib := range cur.TopContributions {
    fmt.Printf("  %s: %d (%s)\n", contrib.UserName, contrib.Total, contrib.Type)
}
if status.AllTimeHigh != nil {
    fmt.Printf("Record: level %d, %d points\n", status.AllTimeHigh.Level, status.AllTimeHigh.Total)
}
```

**Parameters:**
- `broadcasterID` (string): The ID of the broadcaster

**Response Fields (`HypeTrainStatus`):**
- `Current` (*HypeTrainCurrent): The Hype Train in progress; `nil` if none is active
- `AllTimeHigh` (*HypeTrainRecord): The channel's record Hype Train; `nil` if it has never had one
- `SharedAllTimeHigh` (*HypeTrainRecord): The channel's record shared Hype Train; `nil` if it has never taken part in one

**HypeTrainCurrent Fields:**
- `ID` (string): Hype Train ID
- `BroadcasterUserID`, `BroadcasterUserLogin`, `BroadcasterUserName` (string): The broadcaster
- `Level` (int): Current level
- `Total` (int): Total points contributed
- `Progress` (int): Points contributed at the current level
- `Goal` (int): Points needed to reach the next level
- `TopContributions` ([]EventSubContribution): Top contributors (`UserID`, `UserLogin`, `UserName`, `Type` of `bits`/`subscription`/`other`, `Total`)
- `SharedTrainParticipants` ([]HypeTrainParticipant): Broadcasters in a shared train; `nil` if not shared
- `StartedAt`, `ExpiresAt` (time.Time): When the train started and when it expires (extended on each new level)
- `Type` (HypeTrainType): `regular`, `golden_kappa` or `treasure`
- `IsSharedTrain` (bool): Whether the train is shared

**HypeTrainRecord Fields:**
- `Level` (int), `Total` (int), `AchievedAt` (time.Time)

**Sample Response (from the Twitch docs):**
```json
{
  "data": [
    {
      "current": {
        "id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
        "broadcaster_user_id": "1337",
        "broadcaster_user_login": "cool_user",
        "broadcaster_user_name": "Cool_User",
        "level": 2,
        "total": 700,
        "progress": 200,
        "goal": 1000,
        "top_contributions": [
          {
            "user_id": "123",
            "user_login": "pogchamp",
            "user_name": "PogChamp",
            "type": "bits",
            "total": 50
          },
          {
            "user_id": "456",
            "user_login": "kappa",
            "user_name": "Kappa",
            "type": "subscription",
            "total": 45
          }
        ],
        "shared_train_participants": [
          {
            "broadcaster_user_id": "456",
            "broadcaster_user_login": "pogchamp",
            "broadcaster_user_name": "PogChamp"
          },
          {
            "broadcaster_user_id": "321",
            "broadcaster_user_login": "pogchamp",
            "broadcaster_user_name": "PogChamp"
          }
        ],
        "started_at": "2020-07-15T17:16:03.17106713Z",
        "expires_at": "2020-07-15T17:16:11.17106713Z",
        "type": "golden_kappa",
        "is_shared_train": true
      },
      "all_time_high": {
        "level": 6,
        "total": 2850,
        "achieved_at": "2020-04-24T20:12:21.003802269Z"
      },
      "shared_all_time_high": {
        "level": 16,
        "total": 23850,
        "achieved_at": "2020-04-27T20:12:21.003802269Z"
      }
    }
  ]
}
```

## EventSub Alternative

For real-time hype train events, use EventSub subscriptions instead of polling the deprecated API:

- `channel.hype_train.begin` - Hype train starts
- `channel.hype_train.progress` - Hype train level increases
- `channel.hype_train.end` - Hype train ends
See [EventSub documentation](eventsub.md) for more details.

