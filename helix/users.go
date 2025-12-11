package helix

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// User represents a Twitch user.
type User struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"` // Deprecated
	Email           string    `json:"email,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetUsersParams contains parameters for GetUsers.
type GetUsersParams struct {
	IDs    []string // User IDs (max 100)
	Logins []string // User login names (max 100)
}

// GetUsers gets information about one or more Twitch users.
// Requires: No scope for public data, user:read:email for email.
func (c *Client) GetUsers(ctx context.Context, params *GetUsersParams) (*Response[User], error) {
	q := url.Values{}
	if params != nil {
		for _, id := range params.IDs {
			q.Add("id", id)
		}
		for _, login := range params.Logins {
			q.Add("login", login)
		}
	}

	var resp Response[User]
	if err := c.get(ctx, "/users", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCurrentUser gets information about the authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.GetUsers(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UpdateUserParams contains parameters for UpdateUser.
type UpdateUserParams struct {
	Description string `json:"description,omitempty"`
}

// UpdateUser updates the authenticated user's information.
// Requires: user:edit scope.
func (c *Client) UpdateUser(ctx context.Context, params *UpdateUserParams) (*User, error) {
	q := url.Values{}
	if params != nil && params.Description != "" {
		q.Set("description", params.Description)
	}

	var resp Response[User]
	if err := c.put(ctx, "/users", q, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// BlockedUser represents a blocked user.
type BlockedUser struct {
	UserID      string `json:"user_id"`
	UserLogin   string `json:"user_login"`
	DisplayName string `json:"display_name"`
}

// GetUserBlockListParams contains parameters for GetUserBlockList.
type GetUserBlockListParams struct {
	BroadcasterID string
	*PaginationParams
}

// GetUserBlockList gets the authenticated user's block list.
// Requires: user:read:blocked_users scope.
func (c *Client) GetUserBlockList(ctx context.Context, params *GetUserBlockListParams) (*Response[BlockedUser], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[BlockedUser]
	if err := c.get(ctx, "/users/blocks", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BlockUserParams contains parameters for BlockUser.
type BlockUserParams struct {
	TargetUserID  string
	SourceContext string // "chat" or "whisper"
	Reason        string // "spam", "harassment", or "other"
}

// BlockUser blocks a user.
// Requires: user:manage:blocked_users scope.
func (c *Client) BlockUser(ctx context.Context, params *BlockUserParams) error {
	q := url.Values{}
	q.Set("target_user_id", params.TargetUserID)
	if params.SourceContext != "" {
		q.Set("source_context", params.SourceContext)
	}
	if params.Reason != "" {
		q.Set("reason", params.Reason)
	}

	return c.put(ctx, "/users/blocks", q, nil, nil)
}

// UnblockUser unblocks a user.
// Requires: user:manage:blocked_users scope.
func (c *Client) UnblockUser(ctx context.Context, targetUserID string) error {
	q := url.Values{}
	q.Set("target_user_id", targetUserID)

	return c.delete(ctx, "/users/blocks", q, nil)
}

// UserExtension represents a user's extension.
type UserExtension struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	CanActivate bool     `json:"can_activate"`
	Type        []string `json:"type"`
}

// GetUserExtensions gets extensions installed by the authenticated user.
// Requires: user:read:broadcast scope.
func (c *Client) GetUserExtensions(ctx context.Context) (*Response[UserExtension], error) {
	var resp Response[UserExtension]
	if err := c.get(ctx, "/users/extensions/list", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserActiveExtensions represents a user's active extensions.
type UserActiveExtensions struct {
	Panel     map[string]ActiveExtension `json:"panel"`
	Overlay   map[string]ActiveExtension `json:"overlay"`
	Component map[string]ActiveExtension `json:"component"`
}

// ActiveExtension represents an active extension in a slot.
type ActiveExtension struct {
	Active  bool   `json:"active"`
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
	Name    string `json:"name,omitempty"`
	X       int    `json:"x,omitempty"`
	Y       int    `json:"y,omitempty"`
}

// GetUserActiveExtensions gets the active extensions for a user.
// Requires: user:read:broadcast or user:edit:broadcast scope.
func (c *Client) GetUserActiveExtensions(ctx context.Context, userID string) (*UserActiveExtensions, error) {
	q := url.Values{}
	if userID != "" {
		q.Set("user_id", userID)
	}

	var resp struct {
		Data UserActiveExtensions `json:"data"`
	}
	if err := c.get(ctx, "/users/extensions", q, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// UpdateUserExtensionsParams contains parameters for UpdateUserExtensions.
type UpdateUserExtensionsParams struct {
	Data UserActiveExtensions `json:"data"`
}

// UpdateUserExtensions updates the active extensions for the authenticated user.
// Requires: user:edit:broadcast scope.
func (c *Client) UpdateUserExtensions(ctx context.Context, params *UpdateUserExtensionsParams) (*UserActiveExtensions, error) {
	var resp struct {
		Data UserActiveExtensions `json:"data"`
	}
	if err := c.put(ctx, "/users/extensions", nil, params, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// UserAuthorization represents the authorization scopes granted by a user.
type UserAuthorization struct {
	ClientID string   `json:"client_id"`
	UserID   string   `json:"user_id"`
	Login    string   `json:"login"`
	Scopes   []string `json:"scopes"`
}

// GetAuthorizationByUserParams contains parameters for GetAuthorizationByUser.
type GetAuthorizationByUserParams struct {
	UserID string // Required: The ID of a user that granted the application OAuth permissions
}

// GetAuthorizationByUser gets the authorization scopes that the specified user has granted the application.
// Requires: App access token.
func (c *Client) GetAuthorizationByUser(ctx context.Context, params *GetAuthorizationByUserParams) (*Response[UserAuthorization], error) {
	if params == nil || params.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	q := url.Values{}
	q.Set("user_id", params.UserID)

	var resp Response[UserAuthorization]
	if err := c.get(ctx, "/users/authorization", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
