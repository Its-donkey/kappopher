package helix

import (
	"context"
	"net/url"
	"time"
)

// Raid represents a raid.
type Raid struct {
	CreatedAt time.Time `json:"created_at"`
	IsMature  bool      `json:"is_mature"`
}

// StartRaidParams contains parameters for StartRaid.
type StartRaidParams struct {
	FromBroadcasterID string // The broadcaster initiating the raid
	ToBroadcasterID   string // The broadcaster being raided
}

// StartRaid starts a raid on another channel.
// Requires: channel:manage:raids scope.
func (c *Client) StartRaid(ctx context.Context, params *StartRaidParams) (*Raid, error) {
	q := url.Values{}
	q.Set("from_broadcaster_id", params.FromBroadcasterID)
	q.Set("to_broadcaster_id", params.ToBroadcasterID)

	var resp Response[Raid]
	if err := c.post(ctx, "/raids", q, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// CancelRaid cancels a pending raid.
// Requires: channel:manage:raids scope.
func (c *Client) CancelRaid(ctx context.Context, broadcasterID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	return c.delete(ctx, "/raids", q, nil)
}
