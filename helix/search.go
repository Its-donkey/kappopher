package helix

import (
	"context"
	"net/url"
	"time"
)

// SearchCategory represents a search result for categories.
type SearchCategory struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

// SearchCategoriesParams contains parameters for SearchCategories.
type SearchCategoriesParams struct {
	Query string
	*PaginationParams
}

// SearchCategories searches for categories/games.
func (c *Client) SearchCategories(ctx context.Context, params *SearchCategoriesParams) (*Response[SearchCategory], error) {
	q := url.Values{}
	q.Set("query", params.Query)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[SearchCategory]
	if err := c.get(ctx, "/search/categories", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchChannel represents a search result for channels.
type SearchChannel struct {
	BroadcasterLanguage string    `json:"broadcaster_language"`
	BroadcasterLogin    string    `json:"broadcaster_login"`
	DisplayName         string    `json:"display_name"`
	GameID              string    `json:"game_id"`
	GameName            string    `json:"game_name"`
	ID                  string    `json:"id"`
	IsLive              bool      `json:"is_live"`
	Tags                []string  `json:"tags"`
	ThumbnailURL        string    `json:"thumbnail_url"`
	Title               string    `json:"title"`
	StartedAt           time.Time `json:"started_at,omitempty"`
}

// SearchChannelsParams contains parameters for SearchChannels.
type SearchChannelsParams struct {
	Query    string
	LiveOnly bool
	*PaginationParams
}

// SearchChannels searches for channels.
func (c *Client) SearchChannels(ctx context.Context, params *SearchChannelsParams) (*Response[SearchChannel], error) {
	q := url.Values{}
	q.Set("query", params.Query)
	if params.LiveOnly {
		q.Set("live_only", "true")
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[SearchChannel]
	if err := c.get(ctx, "/search/channels", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
