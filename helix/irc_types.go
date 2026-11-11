package helix

import "time"

// ChatMessage represents a PRIVMSG from Twitch IRC.
//
// During a shared chat session, a message sent in one channel is duplicated
// into every channel in the session. Each copy has its own ID and RoomID; the
// Source* fields identify the original message and the channel it was sent in
// (SourceRoomID == RoomID for the copy delivered to the source channel). The
// Source* fields are empty outside a shared chat session.
type ChatMessage struct {
	ID                         string            // Unique message ID
	Channel                    string            // Channel name (without #)
	RoomID                     string            // Channel (room) ID
	User                       string            // Username (login), from the message prefix
	UserID                     string            // User's Twitch ID
	UserType                   string            // "", admin, global_mod or staff
	Message                    string            // Message content
	Emotes                     []IRCEmote        // Emotes used in the message
	Badges                     map[string]string // User badges (badge-name -> version)
	BadgeInfo                  map[string]string // Additional badge info (e.g., subscriber months)
	Color                      string            // User's chat color (#RRGGBB)
	DisplayName                string            // User's display name
	IsMod                      bool              // Is channel moderator
	IsVIP                      bool              // Is VIP
	IsSubscriber               bool              // Is subscriber
	IsBroadcaster              bool              // Is the channel broadcaster
	Turbo                      bool              // Has site-wide commercial-free mode
	Bits                       int               // Bits cheered (0 if none)
	Gifs                       []IRCGif          // GIFs in the message
	ClientNonce                string            // Nonce supplied by the sending client
	FirstMessage               bool              // Is this user's first message in channel
	ReturningChatter           bool              // Is a returning chatter
	ReplyParentMsgID           string            // Parent message ID if this is a reply
	ReplyParentUserID          string            // Parent message user ID
	ReplyParentUserLogin       string            // Parent message username
	ReplyParentDisplayName     string            // Parent message display name
	ReplyParentMsgBody         string            // Parent message content
	ReplyThreadParentMsgID     string            // Top-level message ID of the reply thread
	ReplyThreadParentUserLogin string            // Top-level message sender's login
	SourceID                   string            // Shared chat: ID of the original message
	SourceRoomID               string            // Shared chat: ID of the channel the message was sent in
	SourceBadges               map[string]string // Shared chat: user badges in the source channel
	SourceBadgeInfo            map[string]string // Shared chat: badge info in the source channel
	SourceOnly                 bool              // Shared chat: message was sent only to the source channel
	Timestamp                  time.Time         // Server timestamp
	Raw                        string            // Raw IRC message
}

// IRCEmote represents an emote used in an IRC message.
type IRCEmote struct {
	ID    string // Emote ID
	Name  string // Emote name/code
	Start int    // Start position in message
	End   int    // End position in message
	Count int    // Number of times used
}

// IRCGif represents a GIF in an IRC chat message, parsed from the gifs tag.
// Applications rendering the GIF must use URL unmodified.
type IRCGif struct {
	ID    string // GIF ID
	URL   string // Full GIF URL
	Start int    // Start position in message (zero-based)
	End   int    // End position in message (zero-based)
}

// UserNotice represents a USERNOTICE message (subs, raids, etc.).
//
// During a shared chat session, a notice for an event in another channel of the
// session arrives with Type "sharedchatnotice"; SourceMsgID then holds the
// original type (e.g. "sub") and the Source* fields identify the original
// notice and channel. The Source* fields are empty outside a shared chat
// session.
type UserNotice struct {
	ID              string            // Unique message ID
	Type            string            // msg-id: sub, resub, subgift, raid, viewermilestone, etc.
	Channel         string            // Channel name
	RoomID          string            // Channel (room) ID
	User            string            // Username (login)
	UserID          string            // User's Twitch ID
	UserType        string            // "", admin, global_mod or staff
	DisplayName     string            // Display name
	Message         string            // Optional user message
	SystemMessage   string            // System-generated message
	MsgParams       map[string]string // Type-specific parameters (msg-param-* tags, prefix removed)
	Badges          map[string]string // User badges
	BadgeInfo       map[string]string // Badge info
	Color           string            // User color
	Emotes          []IRCEmote        // Emotes in message
	IsMod           bool              // Is channel moderator
	IsSubscriber    bool              // Is subscriber
	Turbo           bool              // Has site-wide commercial-free mode
	SourceID        string            // Shared chat: ID of the original notice
	SourceRoomID    string            // Shared chat: ID of the channel the notice was sent in
	SourceMsgID     string            // Shared chat: msg-id (Type) of the original notice
	SourceBadges    map[string]string // Shared chat: user badges in the source channel
	SourceBadgeInfo map[string]string // Shared chat: badge info in the source channel
	Timestamp       time.Time         // Server timestamp
	Raw             string            // Raw IRC message
}

// Common MsgParams keys for UserNotice. The msg-param- prefix is removed from
// the keys stored in UserNotice.MsgParams (e.g. MsgParams["sub-plan"]).
// - msg-param-cumulative-months: Total months subscribed (sub, resub)
// - msg-param-months: Total months subscribed (subgift), or months as a moderator (modiversary)
// - msg-param-multimonth-duration: Multi-month gift duration
// - msg-param-multimonth-tenure: Tenure of multi-month
// - msg-param-should-share-streak: Whether to share streak (sub, resub)
// - msg-param-streak-months: Consecutive months subscribed (sub, resub)
// - msg-param-sub-plan: Subscription plan (1000, 2000, 3000, Prime) (sub, resub, subgift)
// - msg-param-sub-plan-name: Plan display name (sub, resub, subgift)
// - msg-param-gift-months: Months gifted in a single multi-month gift (subgift)
// - msg-param-recipient-id: Gift recipient ID (subgift)
// - msg-param-recipient-user-name: Gift recipient username (subgift; Twitch's examples use msg-param-recipient-name)
// - msg-param-recipient-display-name: Gift recipient display name (subgift)
// - msg-param-sender-count: Sender's total gift count
// - msg-param-sender-login: Gifter's login (giftpaidupgrade)
// - msg-param-sender-name: Gifter's display name (giftpaidupgrade)
// - msg-param-promo-gift-total: Gifts given during the promo (giftpaidupgrade, anongiftpaidupgrade)
// - msg-param-promo-name: Ongoing subscription promo name (giftpaidupgrade, anongiftpaidupgrade)
// - msg-param-viewerCount: Raid viewer count (raid)
// - msg-param-displayName: Raiding broadcaster's display name (raid)
// - msg-param-login: Raiding broadcaster's login (raid)
// - msg-param-threshold: Bits badge tier earned, e.g. 100, 1000 (bitsbadgetier)
// - msg-param-category: Viewer milestone category, e.g. watch-streak (viewermilestone)
// - msg-param-id: Viewer milestone event ID, not the chat message ID (viewermilestone)
// - msg-param-value: Consecutive streams watched for the watch streak (viewermilestone)

// UserNotice types (the msg-id tag).
const (
	UserNoticeTypeSub                 = "sub"
	UserNoticeTypeResub               = "resub"
	UserNoticeTypeSubGift             = "subgift"
	UserNoticeTypeSubMysteryGift      = "submysterygift"
	UserNoticeTypeGiftPaidUpgrade     = "giftpaidupgrade"
	UserNoticeTypeRewardGift          = "rewardgift"
	UserNoticeTypeAnonGiftPaidUpgrade = "anongiftpaidupgrade"
	UserNoticeTypeRaid                = "raid"
	UserNoticeTypeUnraid              = "unraid"
	UserNoticeTypeBitsBadgeTier       = "bitsbadgetier"
	UserNoticeTypeSharedChatNotice    = "sharedchatnotice"
	UserNoticeTypeModiversary         = "modiversary"
	UserNoticeTypeViewerMilestone     = "viewermilestone"

	// UserNoticeTypeAnnouncement is not listed in Twitch's IRC documentation
	// but may still be received for /announce messages.
	UserNoticeTypeAnnouncement = "announcement"

	// Deprecated: anonsubgift is no longer documented by Twitch.
	UserNoticeTypeAnonSubGift = "anonsubgift"
	// Deprecated: primepaidupgrade is no longer documented by Twitch.
	UserNoticeTypePrimePaidUpgrade = "primepaidupgrade"
	// Deprecated: ritual is no longer documented by Twitch.
	UserNoticeTypeRitual = "ritual"
	// Deprecated: communitypayforward is no longer documented by Twitch.
	UserNoticeTypeCommunityPayForward = "communitypayforward"
	// Deprecated: standardpayforward is no longer documented by Twitch.
	UserNoticeTypeStandardPayForward = "standardpayforward"
)

// Viewer milestone categories (MsgParams["category"] on viewermilestone notices).
const (
	ViewerMilestoneCategoryWatchStreak = "watch-streak"
)

// RoomState represents a ROOMSTATE message.
type RoomState struct {
	Channel       string // Channel name
	EmoteOnly     bool   // Emote-only mode
	FollowersOnly int    // Followers-only mode (-1 = off, 0+ = minutes required)
	R9K           bool   // R9K/unique mode
	Slow          int    // Slow mode (seconds between messages, 0 = off)
	SubsOnly      bool   // Subscribers-only mode
	RoomID        string // Channel/room ID
	Raw           string // Raw IRC message
}

// Notice represents a NOTICE message from the server.
type Notice struct {
	Channel      string // Channel name (may be empty for global notices)
	Message      string // Notice message
	MsgID        string // Notice type identifier
	TargetUserID string // ID of the user the action targeted, if any
	Raw          string // Raw IRC message
}

// Common Notice MsgIDs
const (
	NoticeMsgIDSubsOn                         = "subs_on"
	NoticeMsgIDSubsOff                        = "subs_off"
	NoticeMsgIDEmoteOnlyOn                    = "emote_only_on"
	NoticeMsgIDEmoteOnlyOff                   = "emote_only_off"
	NoticeMsgIDSlowOn                         = "slow_on"
	NoticeMsgIDSlowOff                        = "slow_off"
	NoticeMsgIDFollowersOn                    = "followers_on"
	NoticeMsgIDFollowersOnZero                = "followers_on_zero"
	NoticeMsgIDFollowersOff                   = "followers_off"
	NoticeMsgIDMsgChannelSuspended            = "msg_channel_suspended"
	NoticeMsgIDMsgBanned                      = "msg_banned"
	NoticeMsgIDMsgBadCharacters               = "msg_bad_characters"
	NoticeMsgIDMsgChannelBlocked              = "msg_channel_blocked"
	NoticeMsgIDMsgRateLimit                   = "msg_ratelimit"
	NoticeMsgIDMsgDuplicate                   = "msg_duplicate"
	NoticeMsgIDMsgFollowersOnly               = "msg_followersonly"
	NoticeMsgIDMsgFollowersOnlyFollowed       = "msg_followersonly_followed"
	NoticeMsgIDMsgFollowersOnlyZero           = "msg_followersonly_zero"
	NoticeMsgIDMsgSubsOnly                    = "msg_subsonly"
	NoticeMsgIDMsgEmoteOnly                   = "msg_emoteonly"
	NoticeMsgIDMsgSlowMode                    = "msg_slowmode"
	NoticeMsgIDMsgR9K                         = "msg_r9k"
	NoticeMsgIDMsgRejected                    = "msg_rejected"
	NoticeMsgIDMsgRejectedMandatory           = "msg_rejected_mandatory"
	NoticeMsgIDMsgRequiresVerifiedPhoneNumber = "msg_requires_verified_phone_number"
	NoticeMsgIDMsgSuspended                   = "msg_suspended"
	NoticeMsgIDMsgTimedOut                    = "msg_timedout"
	NoticeMsgIDMsgVerifiedEmail               = "msg_verified_email"
	NoticeMsgIDTOSBan                         = "tos_ban"
	NoticeMsgIDUnrecognizedCmd                = "unrecognized_cmd"
	NoticeMsgIDDeleteMessageSuccess           = "delete_message_success"
	NoticeMsgIDWhisperRestricted              = "whisper_restricted"

	// Deprecated: r9k_on is no longer documented by Twitch; unique-chat
	// changes are reported by ROOMSTATE (RoomState.R9K).
	NoticeMsgIDR9KOn = "r9k_on"
	// Deprecated: r9k_off is no longer documented by Twitch; unique-chat
	// changes are reported by ROOMSTATE (RoomState.R9K).
	NoticeMsgIDR9KOff = "r9k_off"
	// Deprecated: Twitch removed hosting; host_on is no longer sent.
	NoticeMsgIDHostOn = "host_on"
	// Deprecated: Twitch removed hosting; host_off is no longer sent.
	NoticeMsgIDHostOff = "host_off"
	// Deprecated: no_permission is no longer documented by Twitch; see
	// NoticeMsgIDMsgSuspended.
	NoticeMsgIDNoPermission = "no_permission"
)

// ClearChat represents a CLEARCHAT message (timeout/ban).
type ClearChat struct {
	Channel      string    // Channel name
	User         string    // User being cleared (empty = chat cleared)
	BanDuration  int       // Duration in seconds (0 = permanent ban)
	RoomID       string    // Channel ID
	TargetUserID string    // Banned user's ID
	Timestamp    time.Time // Server timestamp
	Raw          string    // Raw IRC message
}

// ClearMessage represents a CLEARMSG message (single message deletion).
type ClearMessage struct {
	Channel     string    // Channel name
	RoomID      string    // Channel ID (may be empty)
	User        string    // Message author
	Message     string    // Deleted message content
	TargetMsgID string    // ID of deleted message
	Timestamp   time.Time // Server timestamp
	Raw         string    // Raw IRC message
}

// Whisper represents a WHISPER message (direct message).
type Whisper struct {
	From        string            // Sender username
	FromID      string            // Sender user ID
	To          string            // Recipient username
	Message     string            // Message content
	DisplayName string            // Sender display name
	Color       string            // Sender color
	Badges      map[string]string // Sender badges
	Emotes      []IRCEmote        // Emotes in message
	MessageID   string            // Unique message ID
	ThreadID    string            // Thread ID
	Raw         string            // Raw IRC message
}

// GlobalUserState represents a GLOBALUSERSTATE message.
type GlobalUserState struct {
	UserID      string            // User's Twitch ID
	UserType    string            // "", admin, global_mod or staff
	DisplayName string            // Display name
	Color       string            // Chat color
	Badges      map[string]string // Global badges
	BadgeInfo   map[string]string // Badge info
	EmoteSets   []string          // Available emote set IDs
	Turbo       bool              // Has site-wide commercial-free mode
	Raw         string            // Raw IRC message
}

// UserState represents a USERSTATE message.
type UserState struct {
	Channel      string            // Channel name
	ID           string            // ID of the message just sent (set when sent after a PRIVMSG)
	DisplayName  string            // Display name
	UserType     string            // "", admin, global_mod or staff
	Color        string            // Chat color
	Badges       map[string]string // Channel-specific badges
	BadgeInfo    map[string]string // Badge info
	EmoteSets    []string          // Available emote set IDs
	IsMod        bool              // Is moderator
	IsSubscriber bool              // Is subscriber
	Turbo        bool              // Has site-wide commercial-free mode
	Raw          string            // Raw IRC message
}

// IRCMessage represents a parsed IRC message.
type IRCMessage struct {
	Raw      string            // Raw message
	Tags     map[string]string // IRCv3 tags
	Prefix   string            // Source prefix (nick!user@host)
	Command  string            // IRC command
	Params   []string          // Command parameters
	Trailing string            // Trailing parameter (after :)
}
