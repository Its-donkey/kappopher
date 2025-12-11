# API Usage Examples

Common API usage patterns and examples.

## Users

```go
// Get users by ID or login
users, _ := client.GetUsers(ctx, &helix.GetUsersParams{
    IDs:    []string{"12345"},
    Logins: []string{"username"},
})

// Get current authenticated user
user, _ := client.GetCurrentUser(ctx)

// Block/unblock users
client.BlockUser(ctx, &helix.BlockUserParams{TargetUserID: "12345"})
client.UnblockUser(ctx, "12345")
```

## Channels

```go
// Get channel information
channels, _ := client.GetChannelInformation(ctx, &helix.GetChannelInformationParams{
    BroadcasterIDs: []string{"12345"},
})

// Modify channel
client.ModifyChannelInformation(ctx, &helix.ModifyChannelInformationParams{
    BroadcasterID: "12345",
    Title:         "New Stream Title",
    GameID:        "game-id",
})

// Get followers
followers, _ := client.GetChannelFollowers(ctx, &helix.GetChannelFollowersParams{
    BroadcasterID: "12345",
})
```

## Streams

```go
// Get live streams
streams, _ := client.GetStreams(ctx, &helix.GetStreamsParams{
    UserLogins: []string{"streamer1", "streamer2"},
})

// Get stream key
key, _ := client.GetStreamKey(ctx, "broadcaster-id")

// Create stream marker
marker, _ := client.CreateStreamMarker(ctx, &helix.CreateStreamMarkerParams{
    UserID:      "12345",
    Description: "Highlight moment",
})
```

## Chat

```go
// Send chat message
resp, _ := client.SendChatMessage(ctx, &helix.SendChatMessageParams{
    BroadcasterID: "12345",
    SenderID:      "67890",
    Message:       "Hello, chat!",
})

// Send announcement
client.SendChatAnnouncement(ctx, &helix.SendChatAnnouncementParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Message:       "Important announcement!",
    Color:         "blue",
})

// Get/update chat settings
settings, _ := client.GetChatSettings(ctx, "12345", "")
client.UpdateChatSettings(ctx, &helix.UpdateChatSettingsParams{
    BroadcasterID:    "12345",
    ModeratorID:      "67890",
    SlowMode:         boolPtr(true),
    SlowModeWaitTime: intPtr(30),
})
```

## Moderation

```go
// Ban user
client.BanUser(ctx, &helix.BanUserParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Data: helix.BanUserData{
        UserID:   "banned-user-id",
        Duration: 300, // 5 minutes, 0 for permanent
        Reason:   "Violation of rules",
    },
})

// Unban user
client.UnbanUser(ctx, "12345", "67890", "banned-user-id")

// Get banned users
banned, _ := client.GetBannedUsers(ctx, &helix.GetBannedUsersParams{
    BroadcasterID: "12345",
})

// Manage moderators
client.AddChannelModerator(ctx, "12345", "new-mod-id")
client.RemoveChannelModerator(ctx, "12345", "mod-id")
```

## Polls & Predictions

```go
// Create poll
poll, _ := client.CreatePoll(ctx, &helix.CreatePollParams{
    BroadcasterID: "12345",
    Title:         "What game next?",
    Choices: []helix.CreatePollChoice{
        {Title: "Game A"},
        {Title: "Game B"},
    },
    Duration: 300,
})

// Create prediction
prediction, _ := client.CreatePrediction(ctx, &helix.CreatePredictionParams{
    BroadcasterID: "12345",
    Title:         "Will I win?",
    Outcomes: []helix.CreatePredictionOutcome{
        {Title: "Yes"},
        {Title: "No"},
    },
    PredictionWindow: 120,
})
```

## Clips

```go
// Create clip
clip, _ := client.CreateClip(ctx, &helix.CreateClipParams{
    BroadcasterID: "12345",
})

// Get clips
clips, _ := client.GetClips(ctx, &helix.GetClipsParams{
    BroadcasterID: "12345",
})
```

## Common Scopes

The library provides pre-defined scope combinations:

```go
helix.CommonScopes.Chat        // Chat read/write
helix.CommonScopes.Bot         // Bot functionality
helix.CommonScopes.Moderation  // Moderation tools
helix.CommonScopes.Channel     // Channel management
helix.CommonScopes.Broadcaster // Full broadcaster access
```
