package helix

import (
	"context"
	"net/url"
)

// ContentClassificationLabel represents a content classification label.
type ContentClassificationLabel struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// GetContentClassificationLabelsParams contains parameters for GetContentClassificationLabels.
type GetContentClassificationLabelsParams struct {
	Locale string // Locale for the labels (e.g., "en-US")
}

// GetContentClassificationLabels gets the list of content classification labels.
func (c *Client) GetContentClassificationLabels(ctx context.Context, params *GetContentClassificationLabelsParams) (*Response[ContentClassificationLabel], error) {
	q := url.Values{}
	if params != nil && params.Locale != "" {
		q.Set("locale", params.Locale)
	}

	var resp Response[ContentClassificationLabel]
	if err := c.get(ctx, "/content_classification_labels", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
