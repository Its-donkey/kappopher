---
layout: default
title: Troubleshooting
description: Solutions to common issues when using Kappopher.
---

## Authentication Issues

### "Invalid OAuth token" or 401 Unauthorized

**Symptoms:** API calls return 401 errors or "Invalid OAuth token" messages.

**Solutions:**

1. **Validate your token**
   ```go
   validation, err := authClient.ValidateToken(ctx, token.AccessToken)
   if err != nil {
       // Token is invalid - refresh or re-authenticate
   }
   ```

2. **Check token expiry** - Access tokens expire after ~4 hours. Use refresh tokens:
   ```go
   newToken, err := authClient.RefreshToken(ctx, token.RefreshToken)
   ```

3. **Verify token type** - Some endpoints need user tokens, others need app tokens:
   - User token: `GetUsers` with `IDs` of other users
   - App token: Most read-only public data endpoints

4. **Enable auto-refresh**
   ```go
   cancel := authClient.AutoRefresh(ctx)
   defer cancel()
   ```

### "Missing required scope"

**Symptoms:** API returns 403 with scope-related error message.

**Solutions:**

1. Check the [Twitch API docs](https://dev.twitch.tv/docs/api/reference) for required scopes
2. Re-authenticate with the needed scopes:
   ```go
   authClient := helix.NewAuthClient(helix.AuthConfig{
       Scopes: []string{"channel:read:subscriptions", "channel:manage:broadcast"},
       // ...
   })
   ```

3. Validate your token to see current scopes:
   ```go
   validation, _ := authClient.ValidateToken(ctx, token)
   fmt.Println("Scopes:", validation.Scopes)
   ```

### "Client ID and token mismatch"

**Symptoms:** 401 error mentioning client ID mismatch.

**Solution:** The token was created with a different Client ID. Either:
- Use the correct Client ID that matches the token
- Generate a new token with your Client ID

---

## EventSub Issues

### Subscriptions stuck in "webhook_callback_verification_pending"

**Symptoms:** Webhook subscriptions never become `enabled`.

**Solutions:**

1. **Verify your endpoint is reachable** - Twitch sends a challenge request that must be answered
2. **Check your callback URL** - Must be HTTPS with valid certificate
3. **Return the challenge correctly**:
   ```go
   handler := helix.NewEventSubWebhookHandler(secret,
       helix.WithEventSubChallengeHandler(func(challenge string) {
           // Handler returns challenge automatically
       }),
   )
   ```
4. **Check firewall/proxy settings** - Ensure Twitch can reach your server

### Not receiving WebSocket events

**Symptoms:** WebSocket connected but no events arrive.

**Solutions:**

1. **Verify subscription is active**:
   ```go
   subs, _ := client.GetEventSubSubscriptions(ctx, nil)
   for _, sub := range subs.Data {
       fmt.Printf("%s: %s\n", sub.Type, sub.Status)
   }
   ```

2. **Check you're subscribed to the right events** - Use `SubscribeTo*` methods after connecting

3. **Trigger test events** - Use the Twitch CLI to send test events:
   ```bash
   twitch event trigger channel.follow -F wss://localhost:8080/eventsub
   ```

4. **Verify your session ID** - Subscriptions must use the session ID from the welcome message

### "subscription limit reached"

**Symptoms:** Cannot create more EventSub subscriptions.

**Solutions:**

1. **WebSocket limit**: 300 subscriptions per connection. Use multiple connections or webhooks.
2. **Total limit**: 10,000 subscriptions total. Delete unused subscriptions:
   ```go
   client.DeleteEventSubSubscription(ctx, subscriptionID)
   ```
3. **List and clean up**:
   ```go
   subs, _ := client.GetEventSubSubscriptions(ctx, nil)
   for _, sub := range subs.Data {
       if sub.Status != "enabled" {
           client.DeleteEventSubSubscription(ctx, sub.ID)
       }
   }
   ```

---

## IRC/Chat Issues

### "Login authentication failed"

**Symptoms:** IRC connection fails with NOTICE about authentication.

**Solutions:**

1. **Check token format** - Must include `oauth:` prefix:
   ```go
   helix.WithIRCToken("oauth:your-token-here")
   ```

2. **Verify token scopes** - Need `chat:read` to read, `chat:edit` to send

3. **Check username matches token** - The nick must match the token owner:
   ```go
   helix.WithIRCNick("your_username")
   ```

### Messages not sending

**Symptoms:** `Say()` returns no error but messages don't appear.

**Solutions:**

1. **Join the channel first**:
   ```go
   irc.Join("channelname")
   time.Sleep(time.Second) // Wait for join confirmation
   irc.Say("channelname", "Hello!")
   ```

2. **Check rate limits** - 20 messages per 30 seconds for regular users

3. **Verify you're not banned/timed out** in that channel

4. **Check for shadowban** - Your messages might be hidden from others

### Not receiving messages

**Symptoms:** Connected and joined but no messages received.

**Solutions:**

1. **Verify message handler is set**:
   ```go
   helix.WithIRCMessageHandler(func(msg *helix.IRCMessage) {
       fmt.Printf("Received: %+v\n", msg)
   })
   ```

2. **Request capabilities** - Kappopher requests these automatically, but verify:
   - `twitch.tv/membership` - JOIN/PART messages
   - `twitch.tv/tags` - User badges, emotes
   - `twitch.tv/commands` - USERNOTICE, etc.

3. **Check the channel has activity** - Try a popular channel to verify connection

---

## API Issues

### Empty response data

**Symptoms:** API returns successfully but `Data` is empty.

**Solutions:**

1. **Resource doesn't exist** - User deleted, stream offline, etc.

2. **Check parameters**:
   ```go
   // Wrong: searching by ID when you have login
   client.GetUsers(ctx, &helix.GetUsersParams{IDs: []string{"shroud"}})

   // Correct: use the right parameter
   client.GetUsers(ctx, &helix.GetUsersParams{Logins: []string{"shroud"}})
   ```

3. **Paginate if needed** - First page might be empty, check cursor:
   ```go
   if resp.Pagination != nil && resp.Pagination.Cursor != "" {
       // More data available
   }
   ```

### "too many requests" / 429 errors

**Symptoms:** API returns 429 Too Many Requests.

**Solutions:**

1. **Check rate limit info**:
   ```go
   remaining, reset := client.GetRateLimitInfo()
   if remaining == 0 {
       time.Sleep(time.Until(reset))
   }
   ```

2. **Use caching** to reduce API calls:
   ```go
   client := helix.NewClient(clientID, authClient,
       helix.WithCache(helix.CacheConfig{
           DefaultTTL: 5 * time.Minute,
       }),
   )
   ```

3. **Batch requests** where possible:
   ```go
   // Instead of 100 individual calls
   client.GetUsers(ctx, &helix.GetUsersParams{
       IDs: userIDs[:100], // Up to 100 at once
   })
   ```

### Request timeout

**Symptoms:** Context deadline exceeded.

**Solutions:**

1. **Increase timeout**:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **Check network connectivity** to Twitch servers

3. **Use a custom HTTP client** with longer timeouts:
   ```go
   httpClient := &http.Client{Timeout: 60 * time.Second}
   client := helix.NewClient(clientID, authClient,
       helix.WithHTTPClient(httpClient),
   )
   ```

---

## Connection Issues

### WebSocket keeps disconnecting

**Symptoms:** Frequent disconnections from EventSub or IRC.

**Solutions:**

1. **Enable auto-reconnect**:
   ```go
   helix.WithEventSubAutoReconnect(true)
   // or for IRC
   helix.WithIRCAutoReconnect(true)
   ```

2. **Handle reconnection**:
   ```go
   helix.WithEventSubReconnectHandler(func(oldID, newID string) {
       log.Println("Reconnected")
   })
   ```

3. **Check for network issues** - Unstable connections, proxies, firewalls

### "connection refused"

**Symptoms:** Cannot connect to Twitch servers.

**Solutions:**

1. **Check firewall** - Allow outbound connections to:
   - `api.twitch.tv` (HTTPS/443)
   - `id.twitch.tv` (HTTPS/443)
   - `eventsub.wss.twitch.tv` (WSS/443)
   - `irc-ws.chat.twitch.tv` (WSS/443)

2. **Check proxy settings** - Configure HTTP client if needed

3. **Verify DNS resolution** - Try `ping api.twitch.tv`

---

## Still Having Issues?

1. **Enable debug logging** to see raw requests/responses
2. **Check [Twitch Status](https://status.twitch.tv/)** for outages
3. **Search [existing issues](https://github.com/Its-donkey/kappopher/issues)**
4. **Open a new issue** with:
   - Go version
   - Kappopher version
   - Minimal reproducible example
   - Error messages/logs
