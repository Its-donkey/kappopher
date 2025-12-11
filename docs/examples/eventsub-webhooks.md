# EventSub Webhooks

Handle EventSub webhook notifications with built-in signature verification.

## Basic Setup

```go
handler := helix.NewEventSubWebhookHandler(
    helix.WithWebhookSecret("your-webhook-secret"),
    helix.WithNotificationHandler(func(msg *helix.EventSubWebhookMessage) {
        switch msg.SubscriptionType {
        case helix.EventSubTypeStreamOnline:
            event, _ := helix.ParseEventSubEvent[helix.StreamOnlineEvent](msg)
            fmt.Printf("Stream online: %s\n", event.BroadcasterUserName)
        case helix.EventSubTypeChannelFollow:
            event, _ := helix.ParseEventSubEvent[helix.ChannelFollowEvent](msg)
            fmt.Printf("New follower: %s\n", event.UserName)
        }
    }),
    helix.WithVerificationHandler(func(msg *helix.EventSubWebhookMessage) bool {
        return true // Accept all subscription verifications
    }),
)

http.Handle("/webhook", handler)
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    handler := helix.NewEventSubWebhookHandler(
        helix.WithWebhookSecret("your-webhook-secret"),
        helix.WithNotificationHandler(handleNotification),
        helix.WithVerificationHandler(handleVerification),
        helix.WithRevocationHandler(handleRevocation),
    )

    http.Handle("/webhook", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleNotification(msg *helix.EventSubWebhookMessage) {
    switch msg.SubscriptionType {
    case helix.EventSubTypeStreamOnline:
        event, err := helix.ParseEventSubEvent[helix.StreamOnlineEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("Stream online: %s started streaming\n", event.BroadcasterUserName)

    case helix.EventSubTypeStreamOffline:
        event, err := helix.ParseEventSubEvent[helix.StreamOfflineEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("Stream offline: %s stopped streaming\n", event.BroadcasterUserName)

    case helix.EventSubTypeChannelFollow:
        event, err := helix.ParseEventSubEvent[helix.ChannelFollowEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("New follower: %s followed %s\n", event.UserName, event.BroadcasterUserName)

    case helix.EventSubTypeChannelSubscribe:
        event, err := helix.ParseEventSubEvent[helix.ChannelSubscribeEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("New subscriber: %s subscribed with tier %s\n", event.UserName, event.Tier)

    case helix.EventSubTypeChannelCheer:
        event, err := helix.ParseEventSubEvent[helix.ChannelCheerEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("Cheer: %s cheered %d bits\n", event.UserName, event.Bits)

    default:
        log.Printf("Unhandled event type: %s", msg.SubscriptionType)
    }
}

func handleVerification(msg *helix.EventSubWebhookMessage) bool {
    log.Printf("Verifying subscription: %s", msg.SubscriptionType)
    return true // Accept all verifications
}

func handleRevocation(msg *helix.EventSubWebhookMessage) {
    log.Printf("Subscription revoked: %s (reason: %s)",
        msg.SubscriptionType, msg.Subscription.Status)
}
```

## Creating Subscriptions

To receive webhook notifications, you must first create EventSub subscriptions:

```go
ctx := context.Background()

// Create subscription for stream online events
_, err := client.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
    Type:    helix.EventSubTypeStreamOnline,
    Version: "1",
    Condition: map[string]string{
        "broadcaster_user_id": "12345",
    },
    Transport: helix.CreateEventSubTransport{
        Method:   "webhook",
        Callback: "https://your-domain.com/webhook",
        Secret:   "your-webhook-secret",
    },
})
if err != nil {
    log.Fatal(err)
}
```

## Supported Event Types

See [EventSub documentation](../eventsub.md) for a complete list of event types.
