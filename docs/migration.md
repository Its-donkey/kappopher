---
layout: default
title: Migration Guide
description: How to migrate to Kappopher from other libraries or Twitch PubSub.
---

## Migrating from Twitch PubSub

> **Note:** Twitch PubSub was fully decommissioned on April 14, 2025. Kappopher provides a compatibility layer that uses EventSub under the hood.

### Using the PubSub Compatibility Layer

If you want to minimize code changes, use our PubSub compatibility layer:

```go
// Old PubSub code (no longer works)
pubsub.Listen("channel-points-channel-v1.12345", token)

// New Kappopher code (uses EventSub internally)
pubsub := helix.NewPubSubClient(helixClient,
    helix.WithPubSubMessageHandler(func(topic string, message json.RawMessage) {
        // Handle messages
    }),
)
pubsub.Connect(ctx)
pubsub.Listen(ctx, "channel-points-channel-v1.12345")
```

### Supported Topic Mappings

| PubSub Topic | EventSub Type(s) |
|--------------|------------------|
| `channel-bits-events-v1.<channel_id>` | `channel.cheer` |
| `channel-bits-events-v2.<channel_id>` | `channel.cheer` |
| `channel-points-channel-v1.<channel_id>` | `channel.channel_points_custom_reward_redemption.add` |
| `channel-subscribe-events-v1.<channel_id>` | `channel.subscribe`, `channel.subscription.gift`, `channel.subscription.message` |
| `chat_moderator_actions.<user_id>.<channel_id>` | `channel.moderate` |
| `whispers.<user_id>` | `user.whisper.message` |

### Key Differences

1. **Authentication** - Requires a Helix client for creating subscriptions
2. **Context** - All operations use `context.Context`
3. **Message format** - Events use EventSub format, not old PubSub format
4. **Connection** - Must call `Connect()` before `Listen()`

### Migrating Directly to EventSub

For new code or major refactors, we recommend using EventSub directly:

```go
// EventSub WebSocket (recommended for most use cases)
es := helix.NewEventSubClient(helixClient,
    helix.WithEventSubMessageHandler(func(event helix.EventSubMessage) {
        switch e := event.Event.(type) {
        case *helix.ChannelPointsRedemptionAddEvent:
            fmt.Printf("%s redeemed %s\n", e.UserName, e.Reward.Title)
        }
    }),
)
es.Connect(ctx)
es.SubscribeToChannelPointsRedemption(ctx, broadcasterID)
```

See [EventSub documentation](eventsub.md) for full details.

---

## Migrating from go-twitch-irc

If you're using [gempir/go-twitch-irc](https://github.com/gempir/go-twitch-irc), here's how to migrate:

### Basic Connection

**go-twitch-irc:**
```go
client := twitch.NewClient("username", "oauth:token")
client.OnPrivateMessage(func(message twitch.PrivateMessage) {
    fmt.Println(message.Message)
})
client.Join("channel")
client.Connect()
```

**Kappopher:**
```go
irc := helix.NewIRCClient(
    helix.WithIRCNick("username"),
    helix.WithIRCToken("oauth:token"),
    helix.WithIRCMessageHandler(func(msg *helix.IRCMessage) {
        if msg.Command == "PRIVMSG" {
            fmt.Println(msg.Text)
        }
    }),
)
irc.Connect(ctx)
irc.Join("channel")
```

### Message Structure

**go-twitch-irc:**
```go
message.User.DisplayName
message.User.ID
message.Message
message.Channel
message.Tags["emotes"]
```

**Kappopher:**
```go
msg.Tags["display-name"]
msg.Tags["user-id"]
msg.Text
msg.Channel
msg.Tags["emotes"]
```

### Sending Messages

**go-twitch-irc:**
```go
client.Say("channel", "Hello!")
client.Reply("channel", "parent-msg-id", "Reply text")
```

**Kappopher:**
```go
irc.Say("channel", "Hello!")
irc.Reply("channel", "parent-msg-id", "Reply text")
```

### Event Handlers

| go-twitch-irc | Kappopher |
|---------------|-----------|
| `OnPrivateMessage` | Check `msg.Command == "PRIVMSG"` |
| `OnUserNoticeMessage` | Check `msg.Command == "USERNOTICE"` |
| `OnClearChatMessage` | Check `msg.Command == "CLEARCHAT"` |
| `OnClearMessage` | Check `msg.Command == "CLEARMSG"` |
| `OnRoomStateMessage` | Check `msg.Command == "ROOMSTATE"` |
| `OnConnect` | `WithIRCConnectHandler` |
| `OnReconnect` | `WithIRCReconnectHandler` |

---

## Migrating from nicklaw5/helix

If you're using [nicklaw5/helix](https://github.com/nicklaw5/helix), here's how to migrate:

### Client Setup

**nicklaw5/helix:**
```go
client, _ := helix.NewClient(&helix.Options{
    ClientID:        "client-id",
    ClientSecret:    "client-secret",
    AppAccessToken:  "token",
})
```

**Kappopher:**
```go
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
})
token, _ := authClient.GetAppAccessToken(ctx)
authClient.SetToken(token)

client := helix.NewClient("client-id", authClient)
```

### API Calls

**nicklaw5/helix:**
```go
resp, err := client.GetUsers(&helix.UsersParams{
    Logins: []string{"shroud"},
})
user := resp.Data.Users[0]
```

**Kappopher:**
```go
resp, err := client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"shroud"},
})
user := resp.Data[0]
```

### Key Differences

1. **Context required** - All API methods require `context.Context`
2. **Separate auth client** - Authentication is handled by `AuthClient`
3. **Response structure** - `resp.Data` is the slice directly, not wrapped
4. **Generics** - Kappopher uses Go generics for type safety

### Error Handling

**nicklaw5/helix:**
```go
if resp.ErrorMessage != "" {
    // Handle error
}
```

**Kappopher:**
```go
if err != nil {
    if apiErr, ok := err.(*helix.APIError); ok {
        fmt.Printf("Status %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

---

## Migrating from Direct API Calls

If you're making direct HTTP calls to the Twitch API:

### Before (raw HTTP)

```go
req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/users?login=shroud", nil)
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Client-Id", clientID)
resp, _ := http.DefaultClient.Do(req)
// Parse JSON manually...
```

### After (Kappopher)

```go
resp, err := client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"shroud"},
})
if err != nil {
    return err
}
user := resp.Data[0]
```

### Benefits

- **Type safety** - Responses are properly typed structs
- **Auto-retry** - Handles rate limits and transient failures
- **Token refresh** - Automatic token management
- **Pagination** - Built-in cursor handling
- **Caching** - Optional response caching
- **Middleware** - Extensible request/response processing

---

## Feature Comparison

| Feature | Kappopher | nicklaw5/helix | go-twitch-irc |
|---------|-----------|----------------|---------------|
| Helix API | 147+ endpoints | ~100 endpoints | - |
| IRC Chat | Yes | No | Yes |
| EventSub WebSocket | Yes | No | No |
| EventSub Webhooks | Yes | Partial | No |
| PubSub Compat | Yes | No | No |
| Auto Token Refresh | Yes | Manual | - |
| Caching | Yes | No | No |
| Batch Operations | Yes | No | No |
| Middleware | Yes | No | No |
| Go Generics | Yes | No | No |

---

## Need Help?

- [Full Documentation](/)
- [API Reference](api-reference.md)
- [Examples](cookbook.md)
- [GitHub Issues](https://github.com/Its-donkey/kappopher/issues)
