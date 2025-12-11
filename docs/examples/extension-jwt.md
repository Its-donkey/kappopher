# Extension JWT Authentication

For Twitch Extensions that require JWT authentication.

## Creating JWT Tokens

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    // Create JWT handler
    jwt, err := helix.NewExtensionJWT(
        "your-extension-id",
        "base64-encoded-secret", // From Extension settings
        "extension-owner-user-id",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create EBS (Extension Backend Service) token
    ebsToken, err := jwt.CreateEBSToken(time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("EBS Token: %s\n", ebsToken)

    // Create broadcaster token for specific channel
    broadcasterToken, err := jwt.CreateBroadcasterToken("channel-id", time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Broadcaster Token: %s\n", broadcasterToken)
}
```

## Using with Extension Client

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    ctx := context.Background()

    // Create JWT handler
    jwt, err := helix.NewExtensionJWT(
        "your-extension-id",
        "base64-encoded-secret",
        "extension-owner-user-id",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create a client with extension JWT auth
    client := helix.NewExtensionClient("your-client-id", jwt)

    // Make extension API calls
    resp, err := client.GetExtensionConfigurationSegment(ctx, &helix.GetExtensionConfigurationSegmentParams{
        ExtensionID: "your-extension-id",
        Segment:     []string{"broadcaster"},
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, config := range resp.Data {
        fmt.Printf("Segment: %s, Content: %s\n", config.Segment, config.Content)
    }
}
```

## Adding JWT to Existing Client

```go
// Create standard auth client first
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
})

// Create Helix client
client := helix.NewClient("your-client-id", authClient)

// Create JWT handler
jwt, _ := helix.NewExtensionJWT(
    "your-extension-id",
    "base64-encoded-secret",
    "extension-owner-user-id",
)

// Add JWT to existing client for extension endpoints
client.SetExtensionJWT(jwt)
```

## Token Types

### EBS Token

Used for Extension Backend Service calls:

```go
ebsToken, _ := jwt.CreateEBSToken(time.Hour)
```

### Broadcaster Token

Used for calls on behalf of a specific broadcaster:

```go
broadcasterToken, _ := jwt.CreateBroadcasterToken("channel-id", time.Hour)
```

### Custom Claims

For advanced use cases with custom claims:

```go
token, _ := jwt.CreateToken(helix.ExtensionJWTClaims{
    UserID:    "user-id",
    ChannelID: "channel-id",
    Role:      "broadcaster",
    Pubsub: &helix.PubsubClaims{
        Listen: []string{"broadcast"},
        Send:   []string{"broadcast"},
    },
}, time.Hour)
```

## Extension API Endpoints

See [Extensions documentation](../extensions.md) for all available extension endpoints.
