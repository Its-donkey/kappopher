package helix

import (
	"context"
	"net/url"
)

// Subscription represents a channel subscription.
type Subscription struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	GifterID         string `json:"gifter_id,omitempty"`
	GifterLogin      string `json:"gifter_login,omitempty"`
	GifterName       string `json:"gifter_name,omitempty"`
	IsGift           bool   `json:"is_gift"`
	PlanName         string `json:"plan_name"`
	Tier             string `json:"tier"` // 1000, 2000, 3000
	UserID           string `json:"user_id"`
	UserLogin        string `json:"user_login"`
	UserName         string `json:"user_name"`
}

// GetBroadcasterSubscriptionsParams contains parameters for GetBroadcasterSubscriptions.
type GetBroadcasterSubscriptionsParams struct {
	BroadcasterID string
	UserIDs       []string // Filter by user IDs (max 100)
	*PaginationParams
}

// SubscriptionsResponse represents the response from GetBroadcasterSubscriptions.
type SubscriptionsResponse struct {
	Data       []Subscription `json:"data"`
	Pagination *Pagination    `json:"pagination,omitempty"`
	Total      int            `json:"total"`
	Points     int            `json:"points"` // Subscriber points (based on tiers)
}

// GetBroadcasterSubscriptions gets the list of subscribers for a channel.
// Requires: channel:read:subscriptions scope.
func (c *Client) GetBroadcasterSubscriptions(ctx context.Context, params *GetBroadcasterSubscriptionsParams) (*SubscriptionsResponse, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.UserIDs {
		q.Add("user_id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp SubscriptionsResponse
	if err := c.get(ctx, "/subscriptions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserSubscription represents a user's subscription to a channel.
type UserSubscription struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	GifterID         string `json:"gifter_id,omitempty"`
	GifterLogin      string `json:"gifter_login,omitempty"`
	GifterName       string `json:"gifter_name,omitempty"`
	IsGift           bool   `json:"is_gift"`
	Tier             string `json:"tier"`
}

// CheckUserSubscription checks if a user is subscribed to a channel.
// Requires: user:read:subscriptions scope.
func (c *Client) CheckUserSubscription(ctx context.Context, broadcasterID, userID string) (*UserSubscription, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("user_id", userID)

	var resp Response[UserSubscription]
	if err := c.get(ctx, "/subscriptions/user", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}
