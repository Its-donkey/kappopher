---
layout: default
title: FAQ
description: Frequently asked questions about Kappopher.
---

## General

### What is Kappopher?

Kappopher is a comprehensive Go wrapper for the Twitch Helix API. It provides type-safe access to 147+ API endpoints, multiple OAuth flows, real-time event handling via EventSub and IRC, and advanced features like caching, batch operations, and middleware.

### Why "Kappopher"?

A playful combination of "Kappa" (Twitch's iconic emote) and "Gopher" (Go's mascot).

### Is this an official Twitch library?

No, Kappopher is a community-maintained library. It is not affiliated with or endorsed by Twitch.

### Which Go version is required?

Kappopher requires Go 1.21 or later. We use generics and other modern Go features.

---

## Authentication

### Which OAuth flow should I use?

| Use Case | Recommended Flow |
|----------|------------------|
| Server-side app | Authorization Code |
| CLI tool | Device Code |
| Server-to-server (no user) | Client Credentials |
| Client-side/SPA | Implicit (least secure) |

### How do I get a Client ID and Secret?

1. Go to the [Twitch Developer Console](https://dev.twitch.tv/console)
2. Register a new application
3. Copy your Client ID
4. Generate a Client Secret

### My token keeps expiring. What should I do?

Use the `AutoRefresh` feature to automatically refresh tokens before they expire:

```go
cancel := authClient.AutoRefresh(ctx)
defer cancel()
```

Or manually refresh using `RefreshToken()` when you get a 401 error.

### What scopes do I need?

It depends on the endpoints you're using. Check the [Twitch API Reference](https://dev.twitch.tv/docs/api/reference) for required scopes per endpoint. Common scope groups are available via `helix.CommonScopes`.

---

## EventSub

### WebSocket or Webhooks?

| Feature | WebSocket | Webhooks |
|---------|-----------|----------|
| Public endpoint required | No | Yes |
| Best for | Bots, local dev | Production servers |
| Max subscriptions | 300 per connection | 10,000 total |
| Delivery | Real-time push | HTTP POST |

### Why am I not receiving events?

1. **Check subscription status** - Use `GetEventSubSubscriptions()` to verify your subscriptions are `enabled`
2. **Verify scopes** - Some events require specific OAuth scopes
3. **Check user authorization** - User token events require the user to have authorized your app
4. **WebSocket connected?** - Ensure your WebSocket client is connected and listening

### How do I handle EventSub reconnection?

The WebSocket client handles reconnection automatically. Use the reconnect handler to be notified:

```go
helix.WithEventSubReconnectHandler(func(oldSessionID, newSessionID string) {
    log.Printf("Reconnected: %s -> %s", oldSessionID, newSessionID)
})
```

---

## IRC / Chat

### Should I use IRC or EventSub for chat?

| Feature | IRC | EventSub |
|---------|-----|----------|
| Latency | ~50ms | ~200ms |
| Send messages | Direct | Via API |
| Non-chat events | No | Yes |
| Best for | Chat bots | Dashboards |

Use IRC if you're building a chat bot that needs to send messages quickly. Use EventSub if you need other events alongside chat.

### Why can't I send messages?

1. **Verify your token** - Needs `chat:edit` scope
2. **Join the channel first** - Call `irc.Join(channel)` before sending
3. **Check rate limits** - Twitch limits messages to 20 per 30 seconds (100 for verified bots)
4. **Bot account verified?** - For higher limits, verify your bot at [Twitch Developer Console](https://dev.twitch.tv/console)

### How do I parse emotes from chat messages?

Emote positions are in the `emotes` tag of IRC messages:

```go
emotes := msg.Tags["emotes"]
// Format: "emote_id:start-end,start-end/emote_id:start-end"
```

---

## API Usage

### How do I handle pagination?

Use the cursor from the response:

```go
cursor := ""
for {
    resp, _ := client.GetStreams(ctx, &helix.GetStreamsParams{
        First: 100,
        After: cursor,
    })

    // Process resp.Data...

    if resp.Pagination == nil || resp.Pagination.Cursor == "" {
        break
    }
    cursor = resp.Pagination.Cursor
}
```

### How do I handle rate limits?

Kappopher tracks rate limits automatically. Check before making requests:

```go
remaining, reset := client.GetRateLimitInfo()
if remaining < 10 {
    time.Sleep(time.Until(reset))
}
```

### Can I make requests concurrently?

Yes! The client is thread-safe. For batch operations, use the batch helper:

```go
batcher := helix.NewBatcher(client, helix.BatchConfig{
    MaxConcurrent: 5,
})
results := batcher.GetUsers(ctx, userIDs)
```

---

## Troubleshooting

### I'm getting "invalid token" errors

1. Validate your token: `authClient.ValidateToken(ctx, token)`
2. Check if it's expired
3. Ensure you're using the correct token type (app vs user)
4. Verify the token has required scopes

### API calls return empty data

1. Check if the resource exists (user might have changed name, stream might be offline)
2. Verify your parameters are correct
3. Check the response's `Pagination` field - you might need to paginate

### WebSocket disconnects frequently

1. Handle PING/PONG keepalives (Kappopher does this automatically)
2. Check your network stability
3. Use the reconnect handler to recover gracefully

---

## Contributing

### How can I contribute?

See [CONTRIBUTING.md](https://github.com/Its-donkey/kappopher/blob/main/CONTRIBUTING.md) for guidelines. We welcome bug reports, feature requests, and pull requests!

### Where do I report bugs?

Open an issue on [GitHub](https://github.com/Its-donkey/kappopher/issues).
