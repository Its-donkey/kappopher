---
layout: default
title: Extension JWT Authentication
description: Certain Twitch Extension API endpoints require a special **Extension JWT** (JSON Web Token) instead of a regular OAuth access token.
---

## How It Works

1. Your extension has a **secret** (obtained from the Twitch Developer Console or via `GetExtensionSecrets`)
2. Your Extension Backend Service (EBS) creates and signs a JWT using that secret
3. The JWT is passed as the Bearer token in API requests

## Creating an Extension JWT

`ExtensionJWT` signs tokens with your extension's secret. Pass the secret exactly as the Developer Console shows it (base64); it is decoded before signing.

```go
jwt, err := helix.NewExtensionJWT(
    "your-extension-id",
    "base64-encoded-secret",
    "extension-owner-user-id",
)
if err != nil {
    log.Fatal(err)
}

// EBS token scoped to a channel, allowed to send to its broadcast target
token, err := jwt.CreateExternalToken("67890", &helix.ExtensionPubSubPerms{
    Send: []string{helix.ExtensionPubSubTargetBroadcast},
}, 3*time.Minute)

// Shorthand for the same thing
token, err = jwt.CreatePubSubToken("67890", nil, []string{helix.ExtensionPubSubTargetBroadcast}, 3*time.Minute)

// A global message: channel "all", send target "global"
token, err = jwt.CreatePubSubToken(helix.ExtensionChannelAll, nil, []string{helix.ExtensionPubSubTargetGlobal}, 3*time.Minute)

// Whisper one viewer
token, err = jwt.CreatePubSubToken("67890", nil, []string{helix.ExtensionPubSubWhisperTarget("U12345")}, 3*time.Minute)
```

Use a token for one request with `helix.WithToken`, or have the client sign its own with `helix.NewExtensionClient` / `SetExtensionJWT` (see the [examples](examples/extension-jwt.md)):

```go
ctx = helix.WithToken(ctx, &helix.Token{AccessToken: token})
err = client.SendExtensionPubSubMessage(ctx, &helix.SendExtensionPubSubMessageParams{
    BroadcasterID: "67890",
    Message:       `{"event":"update"}`,
    Target:        []string{helix.ExtensionPubSubTargetBroadcast},
})
```

## Verifying Frontend JWTs

The extension frontend sends a JWT signed with the same secret to your EBS. Verify it and read the viewer's claims:

```go
claims, err := jwt.ParseAuthorizationHeader(r.Header.Get("Authorization"))
if err != nil {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
}
fmt.Println(claims.ChannelID, claims.Role, claims.OpaqueUserID, claims.UserID)
if claims.PubsubPerms != nil {
    fmt.Println(claims.PubsubPerms.Listen, claims.PubsubPerms.Send)
}
```

`jwt.ParseToken(token)` takes the bare token; `helix.ParseExtensionJWT(token, base64Secret)` works without an `ExtensionJWT`.

## JWT Claims

| Claim | Field | Description |
|-------|-------|-------------|
| `exp` | `Exp` | Expiration time (Unix timestamp, seconds). |
| `user_id` | `UserID` | The user ID making the request (the extension owner for EBS tokens). |
| `role` | `Role` | `external` for EBS calls; `broadcaster`, `moderator` or `viewer` in frontend tokens. |
| `channel_id` | `ChannelID` | The target channel ID; `all` for global PubSub messages. |
| `opaque_user_id` | `OpaqueUserID` | The viewer's opaque ID (frontend tokens). |
| `is_unlinked` | `IsUnlinked` | Whether the viewer has not shared their identity (frontend tokens). |
| `pubsub_perms` | `PubsubPerms` | PubSub targets the token may `listen` to and `send` to (required for PubSub messages). |

## Auth Type Comparison

| Auth Type | Used For |
|-----------|----------|
| Extension JWT | Extension-specific endpoints (config, PubSub, secrets, chat) |
| App Access Token | Public extension data (Bits products, transactions) |
| User Access Token | User-authorized actions |

## Example Usage

```go
// Create JWT
secret := "your-extension-secret"
token, err := CreateExtensionJWT(secret, "12345", "67890")
if err != nil {
    log.Fatal(err)
}

// Use with client
client := helix.NewClient("client-id", token)
resp, err := client.GetExtensionConfigurationSegment(ctx, &helix.GetExtensionConfigurationSegmentParams{
    ExtensionID:   "your-extension-id",
    Segment:       []string{"broadcaster"},
    BroadcasterID: "67890",
})
```

