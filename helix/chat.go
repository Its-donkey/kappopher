package helix

import (
	"context"
	"net/url"
)

// Chatter represents a user in a channel's chat.
type Chatter struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// GetChattersParams contains parameters for GetChatters.
type GetChattersParams struct {
	BroadcasterID string
	ModeratorID   string
	*PaginationParams
}

// GetChatters gets the list of users in a channel's chat.
// Requires: moderator:read:chatters scope.
func (c *Client) GetChatters(ctx context.Context, params *GetChattersParams) (*Response[Chatter], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Chatter]
	if err := c.get(ctx, "/chat/chatters", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Emote represents a Twitch emote.
type Emote struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Images     EmoteImages `json:"images"`
	Tier       string      `json:"tier,omitempty"`
	EmoteType  string      `json:"emote_type,omitempty"`
	EmoteSetID string      `json:"emote_set_id,omitempty"`
	Format     []string    `json:"format"`
	Scale      []string    `json:"scale"`
	ThemeMode  []string    `json:"theme_mode"`
}

// EmoteImages contains emote image URLs.
type EmoteImages struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

// GetChannelEmotes gets the emotes for a channel.
func (c *Client) GetChannelEmotes(ctx context.Context, broadcasterID string) (*Response[Emote], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[Emote]
	if err := c.get(ctx, "/chat/emotes", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGlobalEmotes gets all global emotes.
func (c *Client) GetGlobalEmotes(ctx context.Context) (*Response[Emote], error) {
	var resp Response[Emote]
	if err := c.get(ctx, "/chat/emotes/global", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEmoteSets gets emotes from specific emote sets.
func (c *Client) GetEmoteSets(ctx context.Context, emoteSetIDs []string) (*Response[Emote], error) {
	q := url.Values{}
	for _, id := range emoteSetIDs {
		q.Add("emote_set_id", id)
	}

	var resp Response[Emote]
	if err := c.get(ctx, "/chat/emotes/set", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChatBadge represents a chat badge.
type ChatBadge struct {
	SetID    string         `json:"set_id"`
	Versions []BadgeVersion `json:"versions"`
}

// BadgeVersion represents a version of a chat badge.
type BadgeVersion struct {
	ID          string `json:"id"`
	ImageURL1x  string `json:"image_url_1x"`
	ImageURL2x  string `json:"image_url_2x"`
	ImageURL4x  string `json:"image_url_4x"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ClickAction string `json:"click_action,omitempty"`
	ClickURL    string `json:"click_url,omitempty"`
}

// GetChannelChatBadges gets the chat badges for a channel.
func (c *Client) GetChannelChatBadges(ctx context.Context, broadcasterID string) (*Response[ChatBadge], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[ChatBadge]
	if err := c.get(ctx, "/chat/badges", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGlobalChatBadges gets all global chat badges.
func (c *Client) GetGlobalChatBadges(ctx context.Context) (*Response[ChatBadge], error) {
	var resp Response[ChatBadge]
	if err := c.get(ctx, "/chat/badges/global", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChatSettings represents chat room settings.
type ChatSettings struct {
	BroadcasterID                 string `json:"broadcaster_id"`
	SlowMode                      bool   `json:"slow_mode"`
	SlowModeWaitTime              int    `json:"slow_mode_wait_time,omitempty"`
	FollowerMode                  bool   `json:"follower_mode"`
	FollowerModeDuration          int    `json:"follower_mode_duration,omitempty"`
	SubscriberMode                bool   `json:"subscriber_mode"`
	EmoteMode                     bool   `json:"emote_mode"`
	UniqueChatMode                bool   `json:"unique_chat_mode"`
	NonModeratorChatDelay         bool   `json:"non_moderator_chat_delay"`
	NonModeratorChatDelayDuration int    `json:"non_moderator_chat_delay_duration,omitempty"`
}

// GetChatSettings gets the chat settings for a channel.
func (c *Client) GetChatSettings(ctx context.Context, broadcasterID, moderatorID string) (*Response[ChatSettings], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	if moderatorID != "" {
		q.Set("moderator_id", moderatorID)
	}

	var resp Response[ChatSettings]
	if err := c.get(ctx, "/chat/settings", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateChatSettingsParams contains parameters for UpdateChatSettings.
type UpdateChatSettingsParams struct {
	BroadcasterID                 string `json:"-"`
	ModeratorID                   string `json:"-"`
	EmoteMode                     *bool  `json:"emote_mode,omitempty"`
	FollowerMode                  *bool  `json:"follower_mode,omitempty"`
	FollowerModeDuration          *int   `json:"follower_mode_duration,omitempty"`
	NonModeratorChatDelay         *bool  `json:"non_moderator_chat_delay,omitempty"`
	NonModeratorChatDelayDuration *int   `json:"non_moderator_chat_delay_duration,omitempty"`
	SlowMode                      *bool  `json:"slow_mode,omitempty"`
	SlowModeWaitTime              *int   `json:"slow_mode_wait_time,omitempty"`
	SubscriberMode                *bool  `json:"subscriber_mode,omitempty"`
	UniqueChatMode                *bool  `json:"unique_chat_mode,omitempty"`
}

// UpdateChatSettings updates the chat settings for a channel.
// Requires: moderator:manage:chat_settings scope.
func (c *Client) UpdateChatSettings(ctx context.Context, params *UpdateChatSettingsParams) (*Response[ChatSettings], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	var resp Response[ChatSettings]
	if err := c.patch(ctx, "/chat/settings", q, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendChatAnnouncementParams contains parameters for SendChatAnnouncement.
type SendChatAnnouncementParams struct {
	BroadcasterID string `json:"-"`
	ModeratorID   string `json:"-"`
	Message       string `json:"message"`
	Color         string `json:"color,omitempty"` // blue, green, orange, purple, primary
}

// SendChatAnnouncement sends an announcement to a channel's chat.
// Requires: moderator:manage:announcements scope.
func (c *Client) SendChatAnnouncement(ctx context.Context, params *SendChatAnnouncementParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	return c.post(ctx, "/chat/announcements", q, params, nil)
}

// SendShoutoutParams contains parameters for SendShoutout.
type SendShoutoutParams struct {
	FromBroadcasterID string
	ToBroadcasterID   string
	ModeratorID       string
}

// SendShoutout sends a shoutout to another channel.
// Requires: moderator:manage:shoutouts scope.
func (c *Client) SendShoutout(ctx context.Context, params *SendShoutoutParams) error {
	q := url.Values{}
	q.Set("from_broadcaster_id", params.FromBroadcasterID)
	q.Set("to_broadcaster_id", params.ToBroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	return c.post(ctx, "/chat/shoutouts", q, nil, nil)
}

// UserChatColor represents a user's chat color.
type UserChatColor struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Color     string `json:"color"`
}

// GetUserChatColor gets the chat color for one or more users.
func (c *Client) GetUserChatColor(ctx context.Context, userIDs []string) (*Response[UserChatColor], error) {
	q := url.Values{}
	for _, id := range userIDs {
		q.Add("user_id", id)
	}

	var resp Response[UserChatColor]
	if err := c.get(ctx, "/chat/color", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserChatColor updates the authenticated user's chat color.
// Requires: user:manage:chat_color scope.
func (c *Client) UpdateUserChatColor(ctx context.Context, userID, color string) error {
	q := url.Values{}
	q.Set("user_id", userID)
	q.Set("color", color)

	return c.put(ctx, "/chat/color", q, nil, nil)
}

// SendChatMessageParams contains parameters for SendChatMessage.
type SendChatMessageParams struct {
	BroadcasterID        string `json:"broadcaster_id"`
	SenderID             string `json:"sender_id"`
	Message              string `json:"message"`
	ReplyParentMessageID string `json:"reply_parent_message_id,omitempty"`
}

// SendChatMessageResponse represents the response from SendChatMessage.
type SendChatMessageResponse struct {
	MessageID  string `json:"message_id"`
	IsSent     bool   `json:"is_sent"`
	DropReason *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"drop_reason,omitempty"`
}

// SendChatMessage sends a chat message to a channel.
// Requires: user:write:chat scope.
func (c *Client) SendChatMessage(ctx context.Context, params *SendChatMessageParams) (*SendChatMessageResponse, error) {
	var resp Response[SendChatMessageResponse]
	if err := c.post(ctx, "/chat/messages", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// SharedChatSession represents a shared chat session.
type SharedChatSession struct {
	SessionID         string                  `json:"session_id"`
	HostBroadcasterID string                  `json:"host_broadcaster_id"`
	Participants      []SharedChatParticipant `json:"participants"`
	CreatedAt         string                  `json:"created_at"`
	UpdatedAt         string                  `json:"updated_at"`
}

// SharedChatParticipant represents a participant in a shared chat session.
type SharedChatParticipant struct {
	BroadcasterID string `json:"broadcaster_id"`
}

// GetSharedChatSession gets the active shared chat session for a channel.
// Requires: No authentication required.
func (c *Client) GetSharedChatSession(ctx context.Context, broadcasterID string) (*SharedChatSession, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[SharedChatSession]
	if err := c.get(ctx, "/shared_chat/session", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UserEmote represents a user's emote.
type UserEmote struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	EmoteType  string   `json:"emote_type"`
	EmoteSetID string   `json:"emote_set_id"`
	OwnerID    string   `json:"owner_id"`
	Format     []string `json:"format"`
	Scale      []string `json:"scale"`
	ThemeMode  []string `json:"theme_mode"`
}

// GetUserEmotesParams contains parameters for GetUserEmotes.
type GetUserEmotesParams struct {
	UserID        string
	BroadcasterID string // Optional: filter to specific channel
	*PaginationParams
}

// UserEmotesResponse represents the response from GetUserEmotes.
type UserEmotesResponse struct {
	Data       []UserEmote `json:"data"`
	Template   string      `json:"template"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// GetUserEmotes gets emotes available to a user.
// Requires: user:read:emotes scope.
func (c *Client) GetUserEmotes(ctx context.Context, params *GetUserEmotesParams) (*UserEmotesResponse, error) {
	q := url.Values{}
	q.Set("user_id", params.UserID)
	if params.BroadcasterID != "" {
		q.Set("broadcaster_id", params.BroadcasterID)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp UserEmotesResponse
	if err := c.get(ctx, "/chat/emotes/user", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
