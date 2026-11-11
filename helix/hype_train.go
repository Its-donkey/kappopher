package helix

import (
	"context"
	"net/url"
	"time"
)

// HypeTrainEvent represents a hype train event.
//
// Deprecated: Twitch removed the "Get Hype Train Events" endpoint (changelog
// 2026-02-05). Use [Client.GetHypeTrainStatus] and the [HypeTrainStatus] type
// instead.
type HypeTrainEvent struct {
	ID             string             `json:"id"`
	EventType      string             `json:"event_type"`
	EventTimestamp time.Time          `json:"event_timestamp"`
	Version        string             `json:"version"`
	EventData      HypeTrainEventData `json:"event_data"`
}

// HypeTrainEventData contains the hype train event data.
//
// Deprecated: Twitch removed the "Get Hype Train Events" endpoint (changelog
// 2026-02-05). Use [Client.GetHypeTrainStatus] and the [HypeTrainStatus] type
// instead.
type HypeTrainEventData struct {
	ID               string                  `json:"id"`
	BroadcasterID    string                  `json:"broadcaster_id"`
	CooldownEndTime  time.Time               `json:"cooldown_end_time"`
	ExpiresAt        time.Time               `json:"expires_at"`
	Goal             int                     `json:"goal"`
	LastContribution HypeTrainContribution   `json:"last_contribution"`
	Level            int                     `json:"level"`
	StartedAt        time.Time               `json:"started_at"`
	TopContributions []HypeTrainContribution `json:"top_contributions"`
	Total            int                     `json:"total"`
}

// HypeTrainContribution represents a contribution to a hype train.
type HypeTrainContribution struct {
	Total int    `json:"total"`
	Type  string `json:"type"` // BITS, SUBS, OTHER
	User  string `json:"user"`
}

// GetHypeTrainEventsParams contains parameters for GetHypeTrainEvents.
//
// Deprecated: Twitch removed the "Get Hype Train Events" endpoint (changelog
// 2026-02-05). Use [Client.GetHypeTrainStatus] instead.
type GetHypeTrainEventsParams struct {
	BroadcasterID string
	*PaginationParams
}

// GetHypeTrainEvents gets hype train events for a channel.
// Requires: channel:read:hype_train scope.
//
// Deprecated: Twitch removed the "Get Hype Train Events" endpoint (changelog
// 2026-02-05) in favor of Get Hype Train Status. Any live call now returns 404.
// Use [Client.GetHypeTrainStatus] instead. This will be removed in the next
// major release.
func (c *Client) GetHypeTrainEvents(ctx context.Context, params *GetHypeTrainEventsParams) (*Response[HypeTrainEvent], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[HypeTrainEvent]
	if err := c.get(ctx, "/hypetrain/events", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// HypeTrainStatus is a channel's Hype Train status: the train in progress, if
// any, and the channel's records.
type HypeTrainStatus struct {
	// Current is the Hype Train in progress; nil when none is active.
	Current *HypeTrainCurrent `json:"current"`
	// AllTimeHigh is the channel's record Hype Train; nil if it has never had one.
	AllTimeHigh *HypeTrainRecord `json:"all_time_high"`
	// SharedAllTimeHigh is the channel's record shared Hype Train; nil if it has
	// never taken part in one.
	SharedAllTimeHigh *HypeTrainRecord `json:"shared_all_time_high"`
}

// HypeTrainCurrent is a Hype Train in progress.
type HypeTrainCurrent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Level            int                    `json:"level"`
	Total            int                    `json:"total"`
	Progress         int                    `json:"progress"` // points at the current level
	Goal             int                    `json:"goal"`     // points needed for the next level
	TopContributions []EventSubContribution `json:"top_contributions"`
	// SharedTrainParticipants lists the broadcasters in a shared train; nil
	// when it is not shared.
	SharedTrainParticipants []HypeTrainParticipant `json:"shared_train_participants"`
	StartedAt               time.Time              `json:"started_at"`
	ExpiresAt               time.Time              `json:"expires_at"`
	Type                    HypeTrainType          `json:"type"`
	IsSharedTrain           bool                   `json:"is_shared_train"`
}

// HypeTrainRecord is a channel's record Hype Train.
type HypeTrainRecord struct {
	Level      int       `json:"level"`
	Total      int       `json:"total"`
	AchievedAt time.Time `json:"achieved_at"`
}

// GetHypeTrainStatus gets the Hype Train status for a channel. It returns nil
// when Twitch sends no data; a status with a nil Current means no train is
// active.
// Requires: user access token with channel:read:hype_train scope, for the
// broadcaster.
func (c *Client) GetHypeTrainStatus(ctx context.Context, broadcasterID string) (*HypeTrainStatus, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[HypeTrainStatus]
	if err := c.get(ctx, "/hypetrain/status", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}
