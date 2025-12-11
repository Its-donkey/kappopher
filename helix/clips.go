package helix

import (
	"context"
	"net/url"
	"time"
)

// Clip represents a Twitch clip.
type Clip struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	EmbedURL        string    `json:"embed_url"`
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name"`
	CreatorID       string    `json:"creator_id"`
	CreatorName     string    `json:"creator_name"`
	VideoID         string    `json:"video_id"`
	GameID          string    `json:"game_id"`
	Language        string    `json:"language"`
	Title           string    `json:"title"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	Duration        float64   `json:"duration"`
	VODOffset       int       `json:"vod_offset,omitempty"`
	IsFeatured      bool      `json:"is_featured"`
}

// CreateClipParams contains parameters for CreateClip.
type CreateClipParams struct {
	BroadcasterID string
	HasDelay      bool // Add delay before clip creation (for live streams)
}

// CreateClipResponse represents the response from CreateClip.
type CreateClipResponse struct {
	ID      string `json:"id"`
	EditURL string `json:"edit_url"`
}

// CreateClip creates a clip from the broadcaster's stream.
// Requires: clips:edit scope.
func (c *Client) CreateClip(ctx context.Context, params *CreateClipParams) (*CreateClipResponse, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	if params.HasDelay {
		q.Set("has_delay", "true")
	}

	var resp Response[CreateClipResponse]
	if err := c.post(ctx, "/clips", q, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// GetClipsParams contains parameters for GetClips.
type GetClipsParams struct {
	BroadcasterID string
	GameID        string
	IDs           []string // Clip IDs (max 100)
	StartedAt     time.Time
	EndedAt       time.Time
	IsFeatured    *bool
	*PaginationParams
}

// GetClips gets clips for a broadcaster or game.
func (c *Client) GetClips(ctx context.Context, params *GetClipsParams) (*Response[Clip], error) {
	q := url.Values{}
	if params.BroadcasterID != "" {
		q.Set("broadcaster_id", params.BroadcasterID)
	}
	if params.GameID != "" {
		q.Set("game_id", params.GameID)
	}
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	if !params.StartedAt.IsZero() {
		q.Set("started_at", params.StartedAt.Format(time.RFC3339))
	}
	if !params.EndedAt.IsZero() {
		q.Set("ended_at", params.EndedAt.Format(time.RFC3339))
	}
	if params.IsFeatured != nil {
		if *params.IsFeatured {
			q.Set("is_featured", "true")
		} else {
			q.Set("is_featured", "false")
		}
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Clip]
	if err := c.get(ctx, "/clips", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ClipDownload represents a clip download URL.
type ClipDownload struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

// GetClipsDownload gets a download URL for clips.
// Requires: clips:edit scope for clips the user created, or the broadcaster's clips.
func (c *Client) GetClipsDownload(ctx context.Context, clipIDs []string) (*Response[ClipDownload], error) {
	q := url.Values{}
	for _, id := range clipIDs {
		q.Add("id", id)
	}

	var resp Response[ClipDownload]
	if err := c.get(ctx, "/clips/download", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
