package helix

import (
	"context"
	"net/url"
	"time"
)

// ExtensionAnalytics represents extension analytics data.
type ExtensionAnalytics struct {
	ExtensionID string    `json:"extension_id"`
	URL         string    `json:"url"`
	Type        string    `json:"type"`
	DateRange   DateRange `json:"date_range"`
}

// GetExtensionAnalyticsParams contains parameters for GetExtensionAnalytics.
type GetExtensionAnalyticsParams struct {
	ExtensionID string
	Type        string // overview_v2
	StartedAt   time.Time
	EndedAt     time.Time
	*PaginationParams
}

// GetExtensionAnalytics gets analytics for extensions.
// Requires: analytics:read:extensions scope.
func (c *Client) GetExtensionAnalytics(ctx context.Context, params *GetExtensionAnalyticsParams) (*Response[ExtensionAnalytics], error) {
	q := url.Values{}
	if params != nil {
		if params.ExtensionID != "" {
			q.Set("extension_id", params.ExtensionID)
		}
		if params.Type != "" {
			q.Set("type", params.Type)
		}
		if !params.StartedAt.IsZero() {
			q.Set("started_at", params.StartedAt.Format(time.RFC3339))
		}
		if !params.EndedAt.IsZero() {
			q.Set("ended_at", params.EndedAt.Format(time.RFC3339))
		}
		addPaginationParams(q, params.PaginationParams)
	}

	var resp Response[ExtensionAnalytics]
	if err := c.get(ctx, "/analytics/extensions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GameAnalytics represents game analytics data.
type GameAnalytics struct {
	GameID    string    `json:"game_id"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	DateRange DateRange `json:"date_range"`
}

// GetGameAnalyticsParams contains parameters for GetGameAnalytics.
type GetGameAnalyticsParams struct {
	GameID    string
	Type      string // overview_v2
	StartedAt time.Time
	EndedAt   time.Time
	*PaginationParams
}

// GetGameAnalytics gets analytics for games.
// Requires: analytics:read:games scope.
func (c *Client) GetGameAnalytics(ctx context.Context, params *GetGameAnalyticsParams) (*Response[GameAnalytics], error) {
	q := url.Values{}
	if params != nil {
		if params.GameID != "" {
			q.Set("game_id", params.GameID)
		}
		if params.Type != "" {
			q.Set("type", params.Type)
		}
		if !params.StartedAt.IsZero() {
			q.Set("started_at", params.StartedAt.Format(time.RFC3339))
		}
		if !params.EndedAt.IsZero() {
			q.Set("ended_at", params.EndedAt.Format(time.RFC3339))
		}
		addPaginationParams(q, params.PaginationParams)
	}

	var resp Response[GameAnalytics]
	if err := c.get(ctx, "/analytics/games", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
