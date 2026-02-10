package helix

import (
	"context"
	"net/url"
)

// CharityCampaign represents a charity campaign.
type CharityCampaign struct {
	ID                 string        `json:"id"`
	BroadcasterID      string        `json:"broadcaster_id"`
	BroadcasterLogin   string        `json:"broadcaster_login"`
	BroadcasterName    string        `json:"broadcaster_name"`
	CharityName        string        `json:"charity_name"`
	CharityDescription string        `json:"charity_description"`
	CharityLogo        string        `json:"charity_logo"`
	CharityWebsite     string        `json:"charity_website"`
	CurrentAmount      CharityAmount `json:"current_amount"`
	TargetAmount       CharityAmount `json:"target_amount"`
}

// CharityAmount represents a monetary amount for charity.
type CharityAmount struct {
	Value         int    `json:"value"`
	DecimalPlaces int    `json:"decimal_places"`
	Currency      string `json:"currency"`
}

// GetCharityCampaignParams contains parameters for GetCharityCampaign.
type GetCharityCampaignParams struct {
	BroadcasterID string
	*PaginationParams
}

// GetCharityCampaign gets the active charity campaign for a channel.
// Requires: channel:read:charity scope.
func (c *Client) GetCharityCampaign(ctx context.Context, params *GetCharityCampaignParams) (*Response[CharityCampaign], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[CharityCampaign]
	if err := c.get(ctx, "/charity/campaigns", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CharityDonation represents a donation to a charity campaign.
type CharityDonation struct {
	ID         string        `json:"id"`
	CampaignID string        `json:"campaign_id"`
	UserID     string        `json:"user_id"`
	UserLogin  string        `json:"user_login"`
	UserName   string        `json:"user_name"`
	Amount     CharityAmount `json:"amount"`
}

// GetCharityCampaignDonationsParams contains parameters for GetCharityCampaignDonations.
type GetCharityCampaignDonationsParams struct {
	BroadcasterID string
	*PaginationParams
}

// GetCharityCampaignDonations gets the donations for a charity campaign.
// Requires: channel:read:charity scope.
func (c *Client) GetCharityCampaignDonations(ctx context.Context, params *GetCharityCampaignDonationsParams) (*Response[CharityDonation], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[CharityDonation]
	if err := c.get(ctx, "/charity/donations", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
