# Basic Usage

Simple examples for getting started with the Twitch Helix API.

## Setup

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    ctx := context.Background()

    // Create auth client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Get app access token
    token, err := authClient.GetAppAccessToken(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Token expires in: %d seconds\n", token.ExpiresIn)

    // Create Helix client
    client := helix.NewClient("your-client-id", authClient)

    // Now you can make API calls
    // ...
}
```

## Get User Information

```go
// Get user by login name
users, err := client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"twitchdev"},
})
if err != nil {
    log.Fatal(err)
}

for _, user := range users.Data {
    fmt.Printf("User: %s\n", user.DisplayName)
    fmt.Printf("  ID: %s\n", user.ID)
    fmt.Printf("  Description: %s\n", user.Description)
    fmt.Printf("  Profile Image: %s\n", user.ProfileImageURL)
    fmt.Printf("  Created: %s\n", user.CreatedAt)
}
```

## Get Live Streams

```go
// Get live streams for specific users
streams, err := client.GetStreams(ctx, &helix.GetStreamsParams{
    UserLogins: []string{"streamer1", "streamer2"},
})
if err != nil {
    log.Fatal(err)
}

if len(streams.Data) == 0 {
    fmt.Println("No streams are live")
} else {
    for _, stream := range streams.Data {
        fmt.Printf("%s is live!\n", stream.UserName)
        fmt.Printf("  Title: %s\n", stream.Title)
        fmt.Printf("  Game: %s\n", stream.GameName)
        fmt.Printf("  Viewers: %d\n", stream.ViewerCount)
    }
}
```

## Get Channel Information

```go
channels, err := client.GetChannelInformation(ctx, &helix.GetChannelInformationParams{
    BroadcasterIDs: []string{"12345"},
})
if err != nil {
    log.Fatal(err)
}

for _, channel := range channels.Data {
    fmt.Printf("Channel: %s\n", channel.BroadcasterName)
    fmt.Printf("  Title: %s\n", channel.Title)
    fmt.Printf("  Game: %s\n", channel.GameName)
    fmt.Printf("  Language: %s\n", channel.BroadcasterLanguage)
}
```

## Search for Channels

```go
results, err := client.SearchChannels(ctx, &helix.SearchChannelsParams{
    Query: "programming",
    First: 10,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d channels:\n", len(results.Data))
for _, channel := range results.Data {
    status := "offline"
    if channel.IsLive {
        status = "LIVE"
    }
    fmt.Printf("  [%s] %s - %s\n", status, channel.DisplayName, channel.Title)
}
```

## Search for Games/Categories

```go
games, err := client.SearchCategories(ctx, &helix.SearchCategoriesParams{
    Query: "minecraft",
    First: 5,
})
if err != nil {
    log.Fatal(err)
}

for _, game := range games.Data {
    fmt.Printf("Game: %s (ID: %s)\n", game.Name, game.ID)
}
```

## Get Top Games

```go
topGames, err := client.GetTopGames(ctx, &helix.GetTopGamesParams{
    First: 10,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println("Top 10 Games:")
for i, game := range topGames.Data {
    fmt.Printf("  %d. %s\n", i+1, game.Name)
}
```

## Pagination Example

```go
var allFollowers []helix.ChannelFollower
cursor := ""

for {
    resp, err := client.GetChannelFollowers(ctx, &helix.GetChannelFollowersParams{
        BroadcasterID: "12345",
        First:         100,
        After:         cursor,
    })
    if err != nil {
        log.Fatal(err)
    }

    allFollowers = append(allFollowers, resp.Data...)

    // Check if there are more pages
    if resp.Pagination == nil || resp.Pagination.Cursor == "" {
        break
    }
    cursor = resp.Pagination.Cursor
}

fmt.Printf("Total followers: %d\n", len(allFollowers))
```

## Error Handling

```go
users, err := client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"nonexistent_user_12345"},
})
if err != nil {
    if apiErr, ok := err.(*helix.APIError); ok {
        fmt.Printf("API Error: %d - %s\n", apiErr.StatusCode, apiErr.Message)
        switch apiErr.StatusCode {
        case 401:
            fmt.Println("Token expired, need to refresh")
        case 429:
            fmt.Println("Rate limited, slow down requests")
        case 404:
            fmt.Println("Resource not found")
        }
    } else {
        fmt.Printf("Network error: %v\n", err)
    }
    return
}

if len(users.Data) == 0 {
    fmt.Println("User not found")
}
```
