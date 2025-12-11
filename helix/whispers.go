package helix

import (
	"context"
	"net/url"
)

// SendWhisperParams contains parameters for SendWhisper.
type SendWhisperParams struct {
	FromUserID string `json:"-"`
	ToUserID   string `json:"-"`
	Message    string `json:"message"`
}

// SendWhisper sends a whisper message to another user.
// Requires: user:manage:whispers scope.
func (c *Client) SendWhisper(ctx context.Context, params *SendWhisperParams) error {
	q := url.Values{}
	q.Set("from_user_id", params.FromUserID)
	q.Set("to_user_id", params.ToUserID)

	return c.post(ctx, "/whispers", q, params, nil)
}
