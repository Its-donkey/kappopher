package helix

import (
	"context"
	"net/url"
	"time"
)

// BitsLeaderboard represents a bits leaderboard entry.
type BitsLeaderboard struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Rank      int    `json:"rank"`
	Score     int    `json:"score"`
}

// BitsLeaderboardResponse represents the response from GetBitsLeaderboard.
type BitsLeaderboardResponse struct {
	Data      []BitsLeaderboard `json:"data"`
	DateRange DateRange         `json:"date_range"`
	Total     int               `json:"total"`
}

// DateRange represents a date range.
type DateRange struct {
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// GetBitsLeaderboardParams contains parameters for GetBitsLeaderboard.
type GetBitsLeaderboardParams struct {
	Count     int       // Max entries (1-100)
	Period    string    // day, week, month, year, all
	StartedAt time.Time // Start of the date range
	UserID    string    // Filter to specific user
}

// GetBitsLeaderboard gets the bits leaderboard for a channel.
// Requires: bits:read scope.
func (c *Client) GetBitsLeaderboard(ctx context.Context, params *GetBitsLeaderboardParams) (*BitsLeaderboardResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Count > 0 {
			q.Set("count", url.QueryEscape(string(rune(params.Count))))
		}
		if params.Period != "" {
			q.Set("period", params.Period)
		}
		if !params.StartedAt.IsZero() {
			q.Set("started_at", params.StartedAt.Format(time.RFC3339))
		}
		if params.UserID != "" {
			q.Set("user_id", params.UserID)
		}
	}

	var resp BitsLeaderboardResponse
	if err := c.get(ctx, "/bits/leaderboard", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Cheermote represents a cheermote.
type Cheermote struct {
	Prefix       string          `json:"prefix"`
	Tiers        []CheermoteTier `json:"tiers"`
	Type         string          `json:"type"`
	Order        int             `json:"order"`
	LastUpdated  time.Time       `json:"last_updated"`
	IsCharitable bool            `json:"is_charitable"`
}

// CheermoteTier represents a tier of a cheermote.
type CheermoteTier struct {
	MinBits        int             `json:"min_bits"`
	ID             string          `json:"id"`
	Color          string          `json:"color"`
	Images         CheermoteImages `json:"images"`
	CanCheer       bool            `json:"can_cheer"`
	ShowInBitsCard bool            `json:"show_in_bits_card"`
}

// CheermoteImages contains the images for a cheermote tier.
type CheermoteImages struct {
	Dark  CheermoteTheme `json:"dark"`
	Light CheermoteTheme `json:"light"`
}

// CheermoteTheme contains themed cheermote images.
type CheermoteTheme struct {
	Animated map[string]string `json:"animated"`
	Static   map[string]string `json:"static"`
}

// GetCheermotes gets the cheermotes available to a broadcaster.
func (c *Client) GetCheermotes(ctx context.Context, broadcasterID string) (*Response[Cheermote], error) {
	q := url.Values{}
	if broadcasterID != "" {
		q.Set("broadcaster_id", broadcasterID)
	}

	var resp Response[Cheermote]
	if err := c.get(ctx, "/bits/cheermotes", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CustomPowerUp represents a broadcaster's custom Bits Power-up.
// The nested settings reuse the channel points reward types, which share the
// same JSON shape.
type CustomPowerUp struct {
	BroadcasterID                    string              `json:"broadcaster_id"`
	BroadcasterLogin                 string              `json:"broadcaster_login"`
	BroadcasterName                  string              `json:"broadcaster_name"`
	ID                               string              `json:"id"`
	Title                            string              `json:"title"`
	Prompt                           string              `json:"prompt"`
	Bits                             int                 `json:"bits"`
	Image                            *RewardImage        `json:"image,omitempty"` // null if the broadcaster didn't upload images
	DefaultImage                     RewardImage         `json:"default_image"`
	BackgroundColor                  string              `json:"background_color"`
	IsEnabled                        bool                `json:"is_enabled"`
	IsUserInputRequired              bool                `json:"is_user_input_required"`
	MaxPerStreamSetting              MaxPerStream        `json:"max_per_stream_setting"`
	MaxPerUserPerStreamSetting       MaxPerUserPerStream `json:"max_per_user_per_stream_setting"`
	GlobalCooldownSetting            GlobalCooldown      `json:"global_cooldown_setting"`
	IsPaused                         bool                `json:"is_paused"`
	IsInStock                        bool                `json:"is_in_stock"`
	RedemptionsRedeemedCurrentStream int                 `json:"redemptions_redeemed_current_stream,omitempty"` // null if not live or no per-stream limit
	CooldownExpiresAt                *time.Time          `json:"cooldown_expires_at,omitempty"`                 // null when not in cooldown
}

// GetCustomPowerUpParams contains parameters for GetCustomPowerUp.
type GetCustomPowerUpParams struct {
	BroadcasterID string   // Must match the user ID in the OAuth token
	IDs           []string // Filter by Power-up IDs (max 50)
}

// GetCustomPowerUp gets the broadcaster's custom Bits Power-ups.
// Requires: user access token with the bits:read scope.
func (c *Client) GetCustomPowerUp(ctx context.Context, params *GetCustomPowerUpParams) (*Response[CustomPowerUp], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}

	var resp Response[CustomPowerUp]
	if err := c.get(ctx, "/bits/custom_power_ups", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
