package helix

import (
	"context"
	"net/url"
	"time"
)

// Poll represents a channel poll.
type Poll struct {
	ID                         string       `json:"id"`
	BroadcasterID              string       `json:"broadcaster_id"`
	BroadcasterName            string       `json:"broadcaster_name"`
	BroadcasterLogin           string       `json:"broadcaster_login"`
	Title                      string       `json:"title"`
	Choices                    []PollChoice `json:"choices"`
	BitsVotingEnabled          bool         `json:"bits_voting_enabled"`
	BitsPerVote                int          `json:"bits_per_vote"`
	ChannelPointsVotingEnabled bool         `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote       int          `json:"channel_points_per_vote"`
	Status                     string       `json:"status"` // ACTIVE, COMPLETED, TERMINATED, ARCHIVED, MODERATED, INVALID
	Duration                   int          `json:"duration"`
	StartedAt                  time.Time    `json:"started_at"`
	EndedAt                    time.Time    `json:"ended_at,omitempty"`
}

// PollChoice represents a choice in a poll.
type PollChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Votes              int    `json:"votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
	BitsVotes          int    `json:"bits_votes"`
}

// GetPollsParams contains parameters for GetPolls.
type GetPollsParams struct {
	BroadcasterID string
	IDs           []string // Poll IDs (max 100)
	*PaginationParams
}

// GetPolls gets polls for a channel.
// Requires: channel:read:polls scope.
func (c *Client) GetPolls(ctx context.Context, params *GetPollsParams) (*Response[Poll], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Poll]
	if err := c.get(ctx, "/polls", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePollParams contains parameters for CreatePoll.
type CreatePollParams struct {
	BroadcasterID              string             `json:"broadcaster_id"`
	Title                      string             `json:"title"`
	Choices                    []CreatePollChoice `json:"choices"`
	Duration                   int                `json:"duration"` // 15-1800 seconds
	ChannelPointsVotingEnabled bool               `json:"channel_points_voting_enabled,omitempty"`
	ChannelPointsPerVote       int                `json:"channel_points_per_vote,omitempty"`
}

// CreatePollChoice represents a choice when creating a poll.
type CreatePollChoice struct {
	Title string `json:"title"`
}

// CreatePoll creates a poll on a channel.
// Requires: channel:manage:polls scope.
func (c *Client) CreatePoll(ctx context.Context, params *CreatePollParams) (*Poll, error) {
	var resp Response[Poll]
	if err := c.post(ctx, "/polls", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// EndPollParams contains parameters for EndPoll.
type EndPollParams struct {
	BroadcasterID string `json:"broadcaster_id"`
	ID            string `json:"id"`
	Status        string `json:"status"` // TERMINATED or ARCHIVED
}

// EndPoll ends a poll.
// Requires: channel:manage:polls scope.
func (c *Client) EndPoll(ctx context.Context, params *EndPollParams) (*Poll, error) {
	var resp Response[Poll]
	if err := c.patch(ctx, "/polls", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}
