# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [1.5.0] - 2026-11-11

### Added
- `HypeTrainCurrent` and `HypeTrainRecord` types for the Get Hype Train Status response
- `HypeTrainTypeTreasure` (`treasure`) Hype Train type
- `AutomodMessageUpdateEvent` v2 fields `HeldAt`, `Reason`, `Automod` and `BlockedTerm`, and v1 fields `Level` and `Category`
- `BlockedTermFound.OwnerBroadcasterUser{ID,Login,Name}`: Twitch sends the term's owner on each term found
- Tests that decode the Twitch docs' own example payloads (Get Hype Train Status, automod message hold/update v2, charity campaign start/progress/stop)
- Extension JWT: `ExtensionPubSubPerms` and `ExtensionJWTClaims.PubsubPerms` (Twitch's nested `pubsub_perms` claim); `ExtensionJWT.CreateExternalToken`, `ParseToken` and `ParseAuthorizationHeader` (verify frontend JWTs); `ExtensionPubSubTarget{Broadcast,Global,All}`, `ExtensionChannelAll` and `ExtensionPubSubWhisperTarget`
- `CreateEventSubSubscriptionParams.IsBatchingEnabled` and `EventSubSubscription.IsBatchingEnabled` (`drop.entitlement.grant` requires batching)
- Batched EventSub notifications: `Events` on `EventSubWebhookPayload`, `EventSubWebhookMessage` and `WebSocketNotificationPayload`, and `ParseEventSubEvents[T]`; WebSocket handlers receive the `events` array when a notification has no `event`
- Condition helpers `BroadcasterUserCondition`, `ExtensionClientCondition` (`extension_client_id`), `ConduitShardDisabledCondition` and `DropCondition`
- EventSub fields: `StreamOfflineEvent.ID`; `ChannelPredictionLockEvent.LockedAt`; automatic reward v2 `ChannelPoints`, `Emote` and message `Fragments`; chat `gif` fragments (`ChatEventFragment.Gif`, also accepting `gif_id`); `Mention`/`Gif` on user message hold fragments; `GiftedDropsSummary` and `SharedChatGiftedDropsSummary` on chat notifications; `ChatNoticeType*`, `BitsUseType*`, `PowerUpType*` and `AutomaticRewardType*` constants
- `EventSubSharedChatParticipant` for `channel.shared_chat.begin`/`update`, which name participants `broadcaster_user_*`
- `MaxConduitShardsPerUpdate` (100); `UpdateConduitShards` returns an error before sending more
- IRC: `ChatMessage` room, user type, turbo, client nonce, reply thread, shared chat source and `Gifs` fields; `UserNotice` ID, room, mod/subscriber/turbo/user type and shared chat source fields; `ClearMessage.RoomID`, `Notice.TargetUserID`, `UserState.ID`; `UserNoticeType{ViewerMilestone,Modiversary,SharedChatNotice,RewardGift,AnonGiftPaidUpgrade}`; the documented NOTICE msg-ids missing before
- Helix fields and params: `Response.Template` (emote endpoints), `Emote.OwnerID`, `UserAuthorization.HasAuthorized`, `StreamMarker.URL`, `ExtensionViews.Config`/`LiveConfig` and component sizing fields, `ExtensionTransactionProductData.Domain`/`Broadcast`, `CreateClipParams.Title`/`Duration`, `GetUserBlockListParams.UserIDs`, `GetDropsEntitlementsParams.IDs`, `DeleteGuestStarSlotWithParams` (`should_reinvite_guest`), `SuspiciousUserTypeUnknown`
- `ModifyChannelInformationParams` sends an empty, non-nil `Tags` as `[]`, removing the channel's tags

### Changed
- CI: added a stable `Test Pass` aggregate job (gates on the `Test` matrix) so branch protection can require a version-independent status check. This prevents the required check from going stale whenever the Go version matrix changes (as happened when `Test (1.24)` was retired for `Test (1.26)`).
- Applied `go fix`: rewrote an in-loop string concatenation in `whispers_test.go` to use `strings.Builder` (Go 1.26 fixer). Test-only; no library behavior change.
- **BREAKING:** `HypeTrainStatus` now matches the Get Hype Train Status response: `Current` (*HypeTrainCurrent, nil when no train is active), `AllTimeHigh` and `SharedAllTimeHigh` (*HypeTrainRecord). The previous flat shape (`ID`, `BroadcasterID`, `LastContribution`, `CooldownEndTime`, ...) never matched what Twitch sends, so every field decoded to its zero value
- **BREAKING:** `AutomodMessageUpdateEvent.Message` is now `AutomodHeldMessage` (`text` and `fragments`) instead of `string`; Twitch sends an object, which failed to decode
- `NullableTime` also decodes Unix timestamps in seconds, with `0` as unset. Twitch sends Get Ad Schedule's `next_ad_at`, `last_ad_at` and `snooze_refresh_at`, and Snooze Next Ad's `snooze_refresh_at` and `next_ad_at`, this way
- Deprecated `HypeTrainTypeShared`: Twitch's types are `regular`, `golden_kappa` and `treasure`; a shared train is flagged by `IsSharedTrain`
- Deprecated `BlockedTermInfo.OwnerBroadcasterUser{ID,Login,Name}`: Twitch sends these on each `BlockedTermFound`
- CI: test matrix runs Go 1.26 and 1.27; lint, build, coverage and release run on Go 1.27. `go.mod` stays at `go 1.26` so the library doesn't force consumers onto 1.27
- CI: bumped `actions/checkout` to v7 and `golangci/golangci-lint-action` to v9
- Applied Go 1.27 `go fix`: `slices.Backward` for the reverse middleware loop in `client.go`; `atomic.Int64`/`atomic.Int32` types in `auth_test.go` and `batch_test.go`. No behavior change
- Changelog monitor reads Twitch's changelog RSS feed (`https://dev.twitch.tv/docs/rss/change-log.xml`) instead of scraping the HTML page, whose layout changed on 2026-11-09 (Twitch recommends the feed for monitoring). New entries are tracked by guid in `state/twitch-changelog` (`.github/twitch-changelog-seen.txt`) and posted as rendered Markdown; on first run only entries from the last 14 days are reported. Logic lives in `.github/scripts/twitch_changelog.py` (Python standard library only)
- **BREAKING:** `ChannelSubscriptionGiftEvent.CumulativeTotal`, `ChannelSubscriptionMessageEvent.StreakMonths`, `ChatNotificationResub.StreakMonths` and `ChannelChatSettingsUpdateEvent.SlowModeWaitTimeSeconds` are `*int`: Twitch sends null (an anonymous gift, an unshared streak, slow mode off), which is not 0
- **BREAKING:** `ChannelBitsUseEvent.PowerUp` and `CustomPowerUp` are typed (`*BitsUsePowerUp`, `*BitsUseCustomPowerUp`) instead of `*json.RawMessage`
- **BREAKING:** `DropEntitlementGrantEvent.Data` is a single `DropEntitlement`; Twitch batches grants as `events: [{id, data}]`
- **BREAKING:** `ChannelSharedChatBeginEvent.Participants` is `[]EventSubSharedChatParticipant` (its deprecated `BroadcasterID` is still filled); `ChannelPredictionLockEvent` is its own type embedding `ChannelPredictionBeginEvent`
- **BREAKING:** `DateRange` (bits leaderboard, extension analytics), `ShieldModeStatus.LastActivatedAt` and `ScheduleSegment.EndTime` are `NullableTime`: Twitch sends empty strings for them
- `ExtensionJWTClaims.PubsubPermsListen`/`PubsubPermsSend` are deprecated and no longer signed as flat claims; `CreateToken` signs them into `pubsub_perms`, and parsing fills them from it
- `UserActiveExtensions` always encodes `x` and `y` for an active component, so a position of 0 is sent
- Deprecated: `ConduitCondition` (use `ConduitShardDisabledCondition`), `HypeTrain*.LastContribution` (v1 only), automatic reward v1 fields, poll Bits voting fields, `GetChannelStreamScheduleParams.UTCOffset`, `CreateClipParams.HasDelay`, `UpdateGuestStarSlotSettingsParams.Volume`, `GuestStarGuest.UserLogin`/`UserDisplayName`, `GetDropsEntitlementsParams.ID`, IRC user notice and NOTICE ids Twitch no longer documents
- Doc comments list the token each endpoint accepts, including the app access tokens Twitch now accepts (with prior authorization) for chat, moderation, ads, clips and subscription endpoints
- `ChatBotClient` routes `sharedchatnotice` user notices by their source type
- Docs: extension JWT docs and examples use kappopher's signer instead of a third-party library; IRC docs and examples use the real client API; `ENDPOINTS.md` paths and methods corrected, missing endpoints listed

### Fixed
- `GetAdSchedule` and `SnoozeNextAd` failed to decode Twitch's Unix-second timestamps
- `ChannelCharityCampaignStartEvent`, `ChannelCharityCampaignProgressEvent` and `ChannelCharityCampaignStopEvent` left the broadcaster empty: these events name it `broadcaster_id`, `broadcaster_login` and `broadcaster_name`, which now fill the embedded `EventSubBroadcaster`
- `EventSubConflictError.ExistingSubscriptionID` was empty: Twitch puts the existing subscription's id in the 409 body's `message` field, which is now read (an `id` field is still honoured)
- `UpdateChannelGuestStarSettings` sent `PUT`; the endpoint is `PATCH`
- `GuestStarGuest.AudioSettings` and `VideoSettings` never decoded: Twitch's fields are `audio` and `video`, not `audio_settings` and `video_settings`
- `GetBitsLeaderboard` failed to decode when `started_at` was not sent (empty `date_range`), and sent `count` as a character instead of a number
- Get/Update Shield Mode Status failed to decode when Shield Mode had never been activated
- `DropsEntitlement.LastUpdated` never decoded (tag `updated_at`; Twitch sends `last_updated`)
- `drop.entitlement.grant` notifications never decoded
- `UpdateGuestStarSlotSettings` sent its settings in the body; they are query parameters. `UpdateGuestStarSlot` always sends `destination_slot_id`
- `SendExtensionChatMessage` now sends `broadcaster_id` as the required query parameter; `SendExtensionPubSubMessage` omits `broadcaster_id` for global messages
- `GetUnbanRequests` always sends the required `status`
- `GetChannelICalendar` returned error pages as calendar text; it now returns an `*APIError` for non-2xx responses
- Extension JWTs signed for PubSub used flat `pubsub_perms_listen`/`pubsub_perms_send` claims, which Twitch ignores
- Automatic reward v2 redemptions decoded with no cost or emote; shared chat participant IDs were empty; the prediction lock time was lost
- `ClientCondition` documented for `extension.bits_transaction.create`, which needs `extension_client_id` (use `ExtensionClientCondition`)
- IRC `ChatMessage.User` was always empty (PRIVMSG has no `login` tag); it now comes from the prefix


## [1.4.0] - 2026-06-13 ([#94](https://github.com/Its-donkey/kappopher/pull/94))

### Added
- `NullableTime` type that decodes Twitch's empty-string/null optional timestamps without erroring and encodes an unset value back to `null`

### Changed
- Bumped the Go version to 1.26 in `go.mod`, and updated the CI workflows (test matrix, lint, build, coverage, and release) to run on Go 1.26
- CI: upgraded the lint job to golangci-lint v2 (`golangci-lint-action@v8`); the previous v1 line was built with Go 1.24 and refused to run against a `go 1.26` module. Cleared the findings it then surfaced (unchecked `w.Write` results in the mock integration test, redundant embedded-field selectors in the `NullableTime` test)
- Modernized the codebase to Go 1.26 idioms: `interface{}` → `any`, `strings.SplitSeq`/`strings.Cut`/`strings.CutPrefix`, `maps.Copy`, `slices.Contains`, `sync.WaitGroup.Go`, and the `max` builtin. `LoggingMiddleware`'s logger parameter is now `func(format string, args ...any)` (identical to the previous `...interface{}`)
- **BREAKING:** Optional timestamp fields that Twitch may return empty now use `NullableTime` instead of `time.Time`: `AdSchedule.{NextAdAt,LastAdAt,SnoozeRefreshAt}`, `SnoozeNextAdResponse.{SnoozeRefreshAt,NextAdAt}`, `BannedUser.ExpiresAt`, `BanUserResponse.EndTime`, `BlockedTerm.ExpiresAt`, `Poll.EndedAt`, `Prediction.{EndedAt,LockedAt}`, `SearchChannel.StartedAt`
- Removed no-op `omitempty` from value-type `time.Time` JSON fields (a struct is never "empty" to the JSON encoder)

### Fixed

## [1.3.1] - 2026-06-13 ([#91](https://github.com/Its-donkey/kappopher/pull/91))

### Added
- `GetEventSubSubscriptionsParams.SubscriptionID` (`subscription_id`) and `GetEventSubSubscriptionsParams.ConduitID` (`conduit_id`) filters for `GetEventSubSubscriptions`
- `EventSubConflictError`, returned by `CreateEventSubSubscription` on a `409 Conflict`. Its `ExistingSubscriptionID` exposes the id of the already-existing subscription that Twitch reports in the 409 body; the underlying `*APIError` is wrapped (accessible via `errors.As`)
- `APIError.Body` field exposing the raw error response body, so callers can read endpoint-specific fields the standard error envelope does not surface

### Changed
- Deprecated `GetHypeTrainEvents`, `GetHypeTrainEventsParams`, `HypeTrainEvent`, and `HypeTrainEventData`: Twitch removed the "Get Hype Train Events" endpoint on 2026-02-05 (any live call now returns `404`). Use `GetHypeTrainStatus` instead. These will be removed in the next major release.

### Fixed

## [1.3.0] - 2026-06-13 ([#87](https://github.com/Its-donkey/kappopher/pull/87))

### Added
- `WithEventSubWSURL` option to point the high-level `EventSubWebSocket` at a custom WebSocket URL (useful for testing)
- Pinned Chat Message endpoints: `GetPinnedChatMessage`, `PinChatMessage`, `UpdatePinnedChatMessage`, `UnpinChatMessage` (`/chat/pins`), and the `PinnedChatMessage` type
- `SendChatMessageParams.ForSourceOnly` and `SendChatMessageParams.Pin`, and `SendChatAnnouncementParams.ForSourceOnly`, matching recent Twitch additions
- `channel.custom_power_up_redemption.add` EventSub subscription type (`EventSubTypeChannelCustomPowerUpRedemptionAdd`) and the `ChannelCustomPowerUpRedemptionAddEvent` / `EventSubCustomPowerUp` types
- `GetCustomPowerUp` (`GET /bits/custom_power_ups`) and the `CustomPowerUp` type, returning a broadcaster's configured custom Bits Power-ups
- `SuspiciousUserAction` and `SuspiciousUserType` types, and the `SuspiciousUserStatusActiveMonitoring`/`SuspiciousUserStatusNoTreatment` status constants
- EventSub event fields that were missing vs. the reference: `ChannelChatMessageEvent.IsSourceOnly`; `ChannelChatNotificationEvent.{IsSourceOnly,WatchStreak,Modiversary,SharedChatModiversary}` (+ `ChatNotificationWatchStreak`/`ChatNotificationModiversary` types); `ChannelModerateEvent.{SourceBroadcasterUserID,SourceBroadcasterUserLogin,SourceBroadcasterUserName,SharedChatUnban,SharedChatUntimeout}`; `ChannelGuestStarSessionEndEvent.{HostUserID,HostUserName,HostUserLogin}` and `ChannelGuestStarGuestUpdateEvent.{HostUserID,HostUserName,HostUserLogin}`
- `IngestServer.URLTemplateSecure` (`url_template_secure`) — the RTMPS ingest URL template returned by the ingest endpoint
- `AutomodSettingsUpdateEvent.Aggression` (`aggression`) — the documented AutoMod category level was missing from the `automod.settings.update` event

### Changed
- Expanded test coverage (cache context invalidation, IRC timestamp fallback, WebSocket close-error classification, and the high-level `EventSubWebSocket.Connect` success/reconnect/error paths)

### Fixed
- Corrected the suspicious-user endpoints to match the Twitch reference: `SuspiciousUserStatus` values are now `ACTIVE_MONITORING`/`RESTRICTED` (was `monitored`/`restricted`); renamed `SuspiciousUserStatusMonitored` → `SuspiciousUserStatusActiveMonitoring`. `AddSuspiciousStatusToChatUser` and `RemoveSuspiciousStatusFromChatUser` now return `(*SuspiciousUserAction, error)` (the documented response with `updated_at`, `status`, and `types`) instead of discarding the response body.
- Corrected `ChannelBitsUseEvent` to match the `channel.bits.use` payload: `BitsUsed` JSON tag fixed (`bits_used` → `bits`), removed the non-existent `UsedAt` field, changed `Message` from `*string` to `*ChatEventMessage` (the documented `{text, fragments[]}` object), and added `CustomPowerUp`. `PowerUp`/`CustomPowerUp` are `*json.RawMessage` since Twitch does not document their object fields. Removed the unused `PowerUp` type.
- Corrected `GetClipsDownload` to match the Twitch reference: endpoint path `/clips/download` → `/clips/downloads`; it now takes `*GetClipsDownloadParams` (`BroadcasterID`, `EditorID`, `ClipIDs`) and `ClipDownload` exposes `ClipID`/`LandscapeDownloadURL`/`PortraitDownloadURL` (was `ID`/`URL`/`ExpiresAt`)
- `CreateClipFromVOD` now sends its parameters as query parameters (the Twitch endpoint reads them from the query string; they were incorrectly sent as a JSON body)
- Corrected `GetAuthorizationByUser` to match the Twitch reference: endpoint path `/users/authorization` → `/authorization/users`; it now takes `UserIDs []string` (up to 10) and `UserAuthorization` exposes `UserID`/`UserName`/`UserLogin`/`Scopes` (removed the non-existent `ClientID`, renamed `Login` → `UserLogin`)
- Corrected the Guest Star channel-settings endpoint path `/channels/guest_star_settings` → `/guest_star/channel_settings` (`GetChannelGuestStarSettings` and `UpdateChannelGuestStarSettings`)
- `HypeTrainParticipant` JSON tags corrected to `broadcaster_user_id`/`broadcaster_user_login`/`broadcaster_user_name` (were `broadcaster_id`/`broadcaster_login`/`broadcaster_name`), matching the shared-train-participants payload
- Fixed `GuestStarAudioSettings` and `GuestStarVideoSettings`: the second flag is `is_guest_enabled`, not `is_self_muted` (field renamed `IsSelfMuted` → `IsGuestEnabled`)

## [1.2.2] - 2026-04-23 ([#75](https://github.com/Its-donkey/kappopher/pull/75))

### Added

### Changed
- CI: Twitch changelog monitor now lists the specific added entries inline in the issue body (and in comments when the issue is already open) rather than just pointing at the changelog URL. A full entries snapshot is kept on `state/twitch-changelog` alongside the hash so each run can diff against the previous one.

### Fixed

## [1.2.1] - 2026-04-23 ([#67](https://github.com/Its-donkey/kappopher/pull/67))

### Added
- CI: `Require CHANGELOG Update` workflow that fails PRs to `test` when `CHANGELOG.md` is not modified (override with a `skip-changelog` label for docs-only or internal changes)

### Changed
- CI: moved Twitch changelog monitor state to a dedicated `state/twitch-changelog` orphan branch so the bot no longer fights `main` branch protection, and bumped `actions/checkout` and `actions/setup-go` to v6

### Fixed

## [1.2.0] - 2026-02-10 ([#59](https://github.com/Its-donkey/kappopher/pull/59))

### Added
- `GetCharityCampaignParams` struct with pagination support for `GetCharityCampaign`
- Custom `Pagination.UnmarshalJSON` to handle Twitch endpoints that return pagination as a string instead of an object (e.g. Get Extension Live Channels)

### Changed
- Renamed `GetCustomRewards` → `GetCustomReward` (aligns with Twitch API reference)
- Renamed `GetCustomRewardRedemptions` → `GetCustomRewardRedemption` (aligns with Twitch API reference)
- Renamed `GetCharityDonations` → `GetCharityCampaignDonations` (aligns with Twitch API reference)
- Renamed `AddSuspiciousUserStatus` → `AddSuspiciousStatusToChatUser` (aligns with Twitch API reference)
- Renamed `RemoveSuspiciousUserStatus` → `RemoveSuspiciousStatusFromChatUser` (aligns with Twitch API reference)
- `GetCharityCampaign` now takes `*GetCharityCampaignParams` instead of a bare `string`, and returns `*Response[CharityCampaign]` instead of `*CharityCampaign`
- All params types renamed to match their function names (`GetCustomRewardsParams` → `GetCustomRewardParams`, etc.)

### Fixed
- Pagination deserialization for endpoints where Twitch returns `"pagination": ""` instead of `{"cursor": "..."}`

## [1.1.1] - 2026-02-06 ([#56](https://github.com/Its-donkey/kappopher/pull/56))

### Added
- `WithToken(ctx, token)` context override for per-request authentication, enabling concurrent requests with different user tokens through a single client instance

### Changed

### Fixed

## [1.1.0] - 2026-01-20

### Added
- `AddSuspiciousUserStatus` - Add a suspicious status to a chat user (POST /moderation/suspicious_users)
- `RemoveSuspiciousUserStatus` - Remove a suspicious status from a chat user (DELETE /moderation/suspicious_users)
- `ScopeModeratorManageSuspiciousUsers` scope constant for suspicious user management
- GitHub Action workflow to monitor Twitch changelog for API updates

### Fixed
- `GetChannelICalendar` now correctly reads the response body using `io.ReadAll`

## [1.0.2] - 2026-01-19 ([#45](https://github.com/Its-donkey/kappopher/pull/45))

### Added
- High-level EventSub WebSocket handlers: `WithEventSubRevocationHandler`, `WithEventSubReconnectHandler`, `WithEventSubErrorHandler`
- Automatic reconnection handling for EventSub WebSocket when Twitch sends reconnect message
- `CacheKeyWithContext()` function for cache key generation with base URL and token isolation
- `TokenHash()` function to generate hashed tokens for cache key isolation
- Body size limit (1MB) for EventSub webhook handlers to prevent memory exhaustion attacks
- Future timestamp rejection in EventSub webhook handlers (with 1-minute clock skew tolerance)
- `InvalidateCacheWithContext()` method for context-aware cache invalidation with token isolation

### Changed
- `MemoryCache.Get()` and `Set()` now copy byte slices to prevent mutation of cached data
- `MessageDeduplicator` now properly enforces `maxSize` by evicting oldest entries when at capacity
- `MessageDeduplicator` now treats `maxSize <= 0` as unlimited (previously broke deduplication)
- `SetExtensionRequiredConfigurationParams` now includes `ExtensionID` field as required by Twitch API
- Removed duplicate `irc/` package (functionality consolidated in `helix/` IRC implementation)

### Fixed
- IRC `Close()` now uses `sync.Once` to prevent panic from closing `stopChan` multiple times
- IRC `Close()` now waits for `readLoop()` to finish before returning
- IRC `Close()` now properly closes in-progress connections (not just fully connected ones)
- IRC `Connect()` is now serialized - concurrent calls return `ErrAlreadyConnected`
- IRC `waitForAuth()` now respects context deadline with a 30-second default timeout
- IRC handler panics are now recovered to prevent crashing the connection loop
- IRC messages now sanitize CR/LF characters to prevent injection attacks
- EventSub WebSocket `waitForWelcome()` now respects context deadline (uses whichever is sooner between context deadline and 10-second default)
- EventSub WebSocket `Close()` now properly handles in-progress connections
- EventSub WebSocket `Connect()` now closes existing connection before reconnecting
- `ChatBotClient.Connect()` now returns an error if no authentication token is available instead of panicking
- `extensionTokenProvider.GetToken()` is now thread-safe with proper mutex synchronization
- `PubSubClient.Connect()` no longer holds the lock when calling the `onConnect` handler, preventing deadlocks
- `PubSubClient.Close()` now returns combined errors using `errors.Join`
- `PubSubClient` goroutines are now tracked with `sync.WaitGroup` to prevent leaks
- `NewPubSubClient()` now validates that `helixClient` is not nil
- `NewEventSubWebSocket()` now validates that `helixClient` is not nil
- `WaitForDeviceToken()` now validates that `deviceCode` is not nil and `Interval` > 0
- `CreateToken()` now validates that `claims` is not nil
- `PubSubClient.Listen()` now handles empty response from `CreateEventSubSubscription` instead of panicking
- `PubSubClient.handleReconnect()` now copies `ws` under lock to prevent race conditions
- `EventSubWebSocket.handleReconnect()` now copies `ws` under lock to prevent race conditions
- IRC `Join`/`Part`/`Say`/`Reply` now sanitize channel names to prevent IRC command injection

## [1.0.1] - 2026-01-18 ([#37](https://github.com/Its-donkey/kappopher/pull/37))

### Added
- Separate `docs/hype-train.md` documentation for Hype Train API
- Documentation for `GetHypeTrainStatus` endpoint (previously undocumented)

### Changed
- Separated Hype Train API into its own file (`helix/hype_train.go`) to match Twitch API structure
- Updated Goals and Hype Train tests to use official Twitch API response samples
- Updated `docs/goals.md` to focus on Goals API only
- Updated `docs/README.md` to list Goals and Hype Train as separate entries

### Fixed
- WebSocket `Close()` and `Reconnect()` now set an immediate read deadline before closing to ensure `ReadMessage` unblocks quickly
- WebSocket error handler no longer reports expected close errors (connection closed, timeout during shutdown)
- IRC `Reconnect()` now properly triggers auto-reconnect when Twitch sends a RECONNECT command
- IRC reconnect loop now checks if auto-reconnect was disabled during the delay period

## [1.0.0] - 2026-01-15 ([#34](https://github.com/Its-donkey/kappopher/pull/34))

### Added
- Hype Train EventSub v2 support (v1 is deprecated by Twitch):
  - `Type` field indicating hype train type (`regular`, `golden_kappa`, `shared`)
  - `IsSharedTrain` flag for shared hype trains
  - `SharedTrainParticipants` list for multi-channel shared trains
  - `AllTimeHighLevel` and `AllTimeHighTotal` for channel records
- Automatic v1→v2 field conversion during JSON unmarshaling to ease migration from deprecated v1

### Changed
- Default Hype Train EventSub version changed from v1 to v2

### Deprecated
- `IsGoldenKappaTrain` field - use `Type == "golden_kappa"` instead
- `EventSubVersionHypeTrainV1` constant - v1 is deprecated by Twitch

### Fixed

## [0.4.0] - 2025-12-23 ([#30](https://github.com/Its-donkey/kappopher/pull/30))

### Added
- `CreateClipFromVOD` - Create clips from existing VODs (POST /videos/clips)
  - Requires `editor:manage:clips` or `channel:manage:clips` scope
  - Supports custom duration (5-60 seconds)
- IRC/TMI chat client for building chat bots
  - WebSocket-based connection to Twitch IRC
  - Message parsing for PRIVMSG, USERNOTICE, ROOMSTATE, etc.
  - Event handlers for messages, subscriptions, raids, and moderation
  - Support for sending messages, whispers, and chat commands

### Changed

### Fixed

## [0.3.1] - 2025-12-11

### Fixed
- Fixed errcheck lint issues (unchecked error returns) across all source and test files
- Fixed staticcheck lint issues (unnecessary embedded field selectors) in auth.go and eventsub.go

## [0.3.0] - 2025-12-11

### Added
- Caching layer
- Batch request support
- Webhook signature validation

## [0.2.0] - 2025-12-09

### Added
- WebSocket support for EventSub
- Retry logic with exponential backoff
- Request middleware support

## [0.1.0] - 2025-12-08

### Added

#### Authentication (`auth` package)
- Full OAuth 2.0 support for all Twitch authentication flows:
  - Implicit Grant Flow
  - Authorization Code Grant Flow
  - Client Credentials Grant Flow
  - Device Code Grant Flow
- Token management:
  - Token validation via `/validate` endpoint
  - Token refresh with automatic retry
  - Token revocation
  - Automatic token refresh background goroutine
- Comprehensive scope constants for all Twitch API scopes
- Pre-defined common scope combinations (Chat, Bot, Moderation, Channel, Broadcaster)
- Thread-safe token storage

#### Helix API Client (`helix` package)
- Full Helix API client with 100+ endpoints

**Ads**
- Start Commercial
- Get Ad Schedule
- Snooze Next Ad

**Analytics**
- Get Extension Analytics
- Get Game Analytics

**Bits**
- Get Bits Leaderboard
- Get Cheermotes

**Channel Points**
- Create/Get/Update/Delete Custom Rewards
- Get/Update Custom Reward Redemptions

**Channels**
- Get/Modify Channel Information
- Get Channel Editors
- Get Followed Channels
- Get Channel Followers
- Get/Add/Remove VIPs

**Chat**
- Get Chatters
- Get Channel/Global Emotes
- Get Emote Sets
- Get Channel/Global Chat Badges
- Get/Update Chat Settings
- Send Chat Announcement
- Send Shoutout
- Get/Update User Chat Color
- Send Chat Message

**Clips**
- Create Clip
- Get Clips

**EventSub**
- Create/Get/Delete EventSub Subscriptions

**Games**
- Get Games
- Get Top Games

**Goals**
- Get Creator Goals

**Hype Train**
- Get Hype Train Events

**Moderation**
- Get Banned Users
- Ban/Unban User
- Get/Add/Remove Moderators
- Delete Chat Messages
- Get/Add/Remove Blocked Terms
- Get/Update Shield Mode Status
- Warn Chat User

**Polls**
- Get/Create/End Polls

**Predictions**
- Get/Create/End Predictions

**Raids**
- Start/Cancel Raid

**Schedule**
- Get/Update Channel Stream Schedule
- Get Channel iCalendar
- Create/Update/Delete Schedule Segments

**Search**
- Search Categories
- Search Channels

**Streams**
- Get Streams
- Get Followed Streams
- Get Stream Key
- Create/Get Stream Markers

**Subscriptions**
- Get Broadcaster Subscriptions
- Check User Subscription

**Teams**
- Get Channel Teams
- Get Teams

**Users**
- Get Users
- Update User
- Get/Block/Unblock Users
- Get User Extensions

**Videos**
- Get/Delete Videos

**Whispers**
- Send Whisper

#### Features
- Generic response types with Go generics
- Automatic rate limit tracking from response headers
- Customizable HTTP client
- Configurable base URL for testing
- Pagination helpers

#### Testing
- Comprehensive unit tests for auth package
- Comprehensive unit tests for helix package
- Mock server support for testing

#### Documentation
- Full README with quick start guide
- Detailed documentation in `docs/` folder
- Usage examples for common scenarios
- Endpoint support matrix

### Security
- Secure token handling with mutex protection
- CSRF state parameter support
- Scope validation

---
