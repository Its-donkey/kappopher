---
layout: default
title: IRC Client Examples
description: Build chat bots using the low-level IRC/WebSocket client.
---

## Overview

The IRC client provides direct WebSocket connection to Twitch chat, offering:
- Lower latency than EventSub for chat messages
- Direct message sending without API rate limits
- Full IRC message parsing into typed structs (badges, emotes, etc.)
- Support for all TMI (Twitch Messaging Interface) events

## Basic IRC Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Create IRC client
    irc := helix.NewIRCClient(
        "your_bot_username",
        "oauth:your-oauth-token", // Get from twitchtokengenerator.com
        helix.WithMessageHandler(func(msg *helix.ChatMessage) {
            fmt.Printf("[%s] %s: %s\n", msg.Channel, msg.User, msg.Message)
        }),
    )

    // Connect
    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    // Join channels
    if err := irc.Join("channel1", "channel2"); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to chat!")

    // Keep running
    select {}
}
```

## ChatBotClient (High-Level)

`ChatBotClient` wraps the IRC client and takes its token from an `AuthClient`.
Instead of one handler per IRC command, you register typed callbacks such as
`OnMessage`, `OnSub` and `OnRaid`.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Auth client holding the bot account's user access token
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    authClient.SetToken(&helix.Token{AccessToken: "your-user-access-token"})

    // Create ChatBotClient (the nick is the bot's username)
    bot := helix.NewChatBotClient("your_bot_username", authClient)

    // Respond to commands
    bot.OnMessage(func(msg *helix.ChatMessage) {
        switch msg.Message {
        case "!ping":
            bot.Say(msg.Channel, "Pong!")
        case "!time":
            bot.Say(msg.Channel, fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC1123)))
        }
    })

    // Typed event callbacks
    bot.OnSub(func(n *helix.UserNotice) {
        bot.Say(n.Channel, fmt.Sprintf("Welcome to the family, %s!", n.DisplayName))
    })
    bot.OnRaid(func(n *helix.UserNotice) {
        bot.Say(n.Channel, fmt.Sprintf("Thanks for the raid, %s!", n.DisplayName))
    })
    bot.OnError(func(err error) {
        log.Printf("chat error: %v", err)
    })

    // Connect using the auth client's token
    if err := bot.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    // Join channel
    if err := bot.Join("channel_name"); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Bot is running!")
    select {}
}
```

## IRC Message Handling

Each kind of IRC message has its own handler option and is delivered already
parsed into a typed struct, so there is no need to switch on the IRC command or
read raw tags.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    irc := helix.NewIRCClient("bot_name", "oauth:token",
        helix.WithMessageHandler(handleChatMessage),
        helix.WithUserNoticeHandler(handleUserNotice),
        helix.WithRoomStateHandler(handleRoomState),
        helix.WithClearChatHandler(handleClearChat),
        helix.WithClearMessageHandler(handleClearMsg),
        helix.WithNoticeHandler(func(n *helix.Notice) {
            fmt.Printf("Notice (%s): %s\n", n.MsgID, n.Message)
        }),
        helix.WithUserStateHandler(func(s *helix.UserState) {
            fmt.Printf("User state updated in %s\n", s.Channel)
        }),
    )

    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    if err := irc.Join("channel"); err != nil {
        log.Fatal(err)
    }
    select {}
}

func handleChatMessage(msg *helix.ChatMessage) {
    // Parsed fields are available directly on the message
    displayName := msg.DisplayName
    badges := msg.Badges
    color := msg.Color
    emotes := msg.Emotes
    msgID := msg.ID

    // Check for bits
    if msg.Bits > 0 {
        fmt.Printf("💎 %s cheered %d bits: %s\n", displayName, msg.Bits, msg.Message)
        return
    }

    // Check user status
    prefix := ""
    if msg.IsMod {
        prefix = "🗡️"
    } else if msg.IsVIP {
        prefix = "💎"
    } else if msg.IsSubscriber {
        prefix = "⭐"
    }

    fmt.Printf("%s[%s] %s: %s\n", prefix, msg.Channel, displayName, msg.Message)

    // Store message ID for potential moderation
    _ = msgID
    _ = badges
    _ = color
    _ = emotes
}

func handleUserNotice(n *helix.UserNotice) {
    // Type-specific details live in MsgParams (the msg-param- prefix is removed)
    switch n.Type {
    case helix.UserNoticeTypeSub:
        plan := n.MsgParams["sub-plan"]
        fmt.Printf("🎉 %s subscribed! (Tier %s)\n", n.DisplayName, tierName(plan))

    case helix.UserNoticeTypeResub:
        months := n.MsgParams["cumulative-months"]
        plan := n.MsgParams["sub-plan"]
        fmt.Printf("🎉 %s resubscribed for %s months! (Tier %s)\n", n.DisplayName, months, tierName(plan))
        if n.Message != "" {
            fmt.Printf("   Message: %s\n", n.Message)
        }

    case helix.UserNoticeTypeSubGift:
        recipient := n.MsgParams["recipient-display-name"]
        plan := n.MsgParams["sub-plan"]
        fmt.Printf("🎁 %s gifted a sub to %s! (Tier %s)\n", n.DisplayName, recipient, tierName(plan))

    case helix.UserNoticeTypeSubMysteryGift:
        count := n.MsgParams["mass-gift-count"]
        fmt.Printf("🎁 %s gifted %s subs to the community!\n", n.DisplayName, count)

    case helix.UserNoticeTypeRaid:
        viewers := n.MsgParams["viewerCount"]
        fmt.Printf("🚀 %s is raiding with %s viewers!\n", n.DisplayName, viewers)

    case helix.UserNoticeTypeViewerMilestone:
        // e.g. a watch streak
        fmt.Printf("🏆 %s reached a %s milestone: %s\n",
            n.DisplayName, n.MsgParams["category"], n.MsgParams["value"])

    case helix.UserNoticeTypeModiversary:
        fmt.Printf("🗡️ %s has been a moderator for %s months!\n", n.DisplayName, n.MsgParams["months"])

    case helix.UserNoticeTypeSharedChatNotice:
        // An event in another channel of a shared chat session.
        // SourceMsgID holds the original type (sub, raid, ...).
        fmt.Printf("🔗 Shared chat %s: %s\n", n.SourceMsgID, n.SystemMessage)

    case helix.UserNoticeTypeAnnouncement:
        fmt.Printf("📢 Announcement from %s: %s\n", n.DisplayName, n.Message)
    }
}

func handleRoomState(state *helix.RoomState) {
    fmt.Printf("Room state for %s:\n", state.Channel)

    if state.EmoteOnly {
        fmt.Println("  - Emote-only mode ON")
    }
    if state.FollowersOnly != -1 {
        fmt.Printf("  - Followers-only: %d minutes\n", state.FollowersOnly)
    }
    if state.Slow != 0 {
        fmt.Printf("  - Slow mode: %d seconds\n", state.Slow)
    }
    if state.SubsOnly {
        fmt.Println("  - Subscribers-only mode ON")
    }
    if state.R9K {
        fmt.Println("  - R9K mode ON")
    }
}

func handleClearChat(c *helix.ClearChat) {
    switch {
    case c.User == "":
        fmt.Printf("🧹 Chat was cleared in %s\n", c.Channel)
    case c.BanDuration == 0:
        fmt.Printf("🔨 %s was permanently banned\n", c.User)
    default:
        fmt.Printf("⏰ %s was timed out for %d seconds\n", c.User, c.BanDuration)
    }
}

func handleClearMsg(m *helix.ClearMessage) {
    fmt.Printf("🗑️ Message from %s deleted (ID: %s)\n", m.User, m.TargetMsgID)
}

func tierName(plan string) string {
    switch plan {
    case "Prime":
        return "Prime"
    case "1000":
        return "1"
    case "2000":
        return "2"
    case "3000":
        return "3"
    default:
        return plan
    }
}
```

### Handling Notices

`Notice.MsgID` identifies the kind of notice, and the `NoticeMsgID*` constants cover the common ones.

```go
func handleNotice(n *helix.Notice) {
    switch n.MsgID {
    case helix.NoticeMsgIDSlowOn:
        fmt.Printf("Slow mode enabled in %s\n", n.Channel)
    case helix.NoticeMsgIDMsgRateLimit:
        fmt.Println("Slow down, the bot is being rate limited")
    case helix.NoticeMsgIDMsgDuplicate:
        fmt.Println("Duplicate message was rejected")
    case helix.NoticeMsgIDMsgBanned:
        log.Printf("Bot is banned from %s", n.Channel)
    default:
        fmt.Printf("Notice: %s\n", n.Message)
    }
}
```

Register it with `helix.WithNoticeHandler(handleNotice)`.

## Sending Messages

```go
package main

import (
    "context"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    irc := helix.NewIRCClient("bot_name", "oauth:token")

    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    channel := "channel_name"
    if err := irc.Join(channel); err != nil {
        log.Fatal(err)
    }

    // Send a message
    if err := irc.Say(channel, "Hello, chat!"); err != nil {
        log.Printf("Failed to send message: %v", err)
    }

    // Reply to a message
    if err := irc.Reply(channel, "message-id-here", "This is a reply!"); err != nil {
        log.Printf("Failed to send reply: %v", err)
    }

    // Send /me action
    if err := irc.Say(channel, "/me waves at chat"); err != nil {
        log.Printf("Failed to send action: %v", err)
    }
}
```

## Auto-Reconnect

When auto-reconnect is enabled the client reconnects after a dropped connection
and rejoins the channels it had joined, so the reconnect handler is only needed
for logging or resetting your own state.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    irc := helix.NewIRCClient("bot_name", "oauth:token",
        helix.WithAutoReconnect(true),
        helix.WithReconnectDelay(5*time.Second),
        helix.WithDisconnectHandler(func() {
            fmt.Println("Disconnected from chat")
        }),
        helix.WithReconnectHandler(func() {
            fmt.Println("Reconnecting... channels are rejoined automatically")
        }),
        helix.WithIRCErrorHandler(func(err error) {
            log.Printf("IRC error: %v", err)
        }),
    )

    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    if err := irc.Join("channel1", "channel2"); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Bot running with auto-reconnect...")
    select {}
}
```

## Complete Chat Bot Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "strings"
    "sync"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

type ChatBot struct {
    irc      *helix.IRCClient
    commands map[string]CommandFunc
    cooldown map[string]time.Time
    cdMu     sync.Mutex
}

type CommandFunc func(msg *helix.ChatMessage, args []string) string

func NewChatBot(nick, token string) *ChatBot {
    bot := &ChatBot{
        commands: make(map[string]CommandFunc),
        cooldown: make(map[string]time.Time),
    }

    bot.irc = helix.NewIRCClient(nick, token,
        helix.WithAutoReconnect(true),
        helix.WithMessageHandler(bot.handleMessage),
    )

    // Register commands
    bot.commands["!ping"] = func(msg *helix.ChatMessage, args []string) string {
        return "Pong!"
    }

    bot.commands["!dice"] = func(msg *helix.ChatMessage, args []string) string {
        return fmt.Sprintf("@%s rolled a %d!", msg.DisplayName, rand.Intn(6)+1)
    }

    bot.commands["!uptime"] = func(msg *helix.ChatMessage, args []string) string {
        // Implement actual uptime check
        return "Stream has been live for 2 hours!"
    }

    bot.commands["!commands"] = func(msg *helix.ChatMessage, args []string) string {
        cmds := make([]string, 0, len(bot.commands))
        for cmd := range bot.commands {
            cmds = append(cmds, cmd)
        }
        return "Commands: " + strings.Join(cmds, ", ")
    }

    bot.commands["!8ball"] = func(msg *helix.ChatMessage, args []string) string {
        answers := []string{
            "Yes!", "No!", "Maybe...", "Ask again later",
            "Definitely!", "I don't think so", "It is certain",
        }
        return fmt.Sprintf("🎱 %s", answers[rand.Intn(len(answers))])
    }

    return bot
}

func (b *ChatBot) handleMessage(msg *helix.ChatMessage) {
    text := strings.TrimSpace(msg.Message)
    if !strings.HasPrefix(text, "!") {
        return
    }

    parts := strings.Fields(text)
    if len(parts) == 0 {
        return
    }

    cmdName := strings.ToLower(parts[0])
    args := parts[1:]

    // Check cooldown (5 seconds per user per command)
    cdKey := fmt.Sprintf("%s:%s:%s", msg.Channel, msg.User, cmdName)
    b.cdMu.Lock()
    if lastUse, ok := b.cooldown[cdKey]; ok && time.Since(lastUse) < 5*time.Second {
        b.cdMu.Unlock()
        return
    }
    b.cooldown[cdKey] = time.Now()
    b.cdMu.Unlock()

    // Execute command
    if handler, ok := b.commands[cmdName]; ok {
        response := handler(msg, args)
        if response != "" {
            b.irc.Say(msg.Channel, response)
        }
    }
}

func (b *ChatBot) Connect(ctx context.Context) error {
    return b.irc.Connect(ctx)
}

func (b *ChatBot) Join(channel string) error {
    return b.irc.Join(channel)
}

func (b *ChatBot) Close() error {
    return b.irc.Close()
}

func main() {
    ctx := context.Background()

    bot := NewChatBot("your_bot_name", "oauth:your-token")

    if err := bot.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    // Join channels
    channels := []string{"channel1", "channel2"}
    for _, ch := range channels {
        if err := bot.Join(ch); err != nil {
            log.Printf("Failed to join %s: %v", ch, err)
        }
    }

    fmt.Println("Bot is running!")
    select {}
}
```

## IRC vs EventSub Comparison

| Feature | IRC Client | EventSub |
|---------|-----------|----------|
| Chat Messages | ✅ Real-time | ✅ Real-time |
| Latency | Lower (~50ms) | Higher (~200ms) |
| Message Sending | Direct | Via API |
| Rate Limits | IRC limits (20/30s) | API limits |
| Badges/Emotes | In tags | In event data |
| Subscriptions | USERNOTICE | EventSub events |
| Bits | In message tags | EventSub events |
| Raids | USERNOTICE | EventSub events |
| Moderation | CLEARCHAT/CLEARMSG | EventSub events |
| Connection | WebSocket to TMI | WebSocket to EventSub |

**Use IRC when:**
- You need lowest latency chat
- Building a chat-focused bot
- You need to send many messages

**Use EventSub when:**
- You need non-chat events (follows, subs, etc.)
- Building a dashboard/overlay
- You need reliable delivery with acknowledgment

