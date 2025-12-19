package gql

import (
	"sync"
	"time"
)

// OperationsRegistry stores known GraphQL operations.
type OperationsRegistry struct {
	operations map[string]Operation
	mu         sync.RWMutex
}

// NewOperationsRegistry creates a new registry with known operations.
func NewOperationsRegistry() *OperationsRegistry {
	r := &OperationsRegistry{
		operations: make(map[string]Operation),
	}
	r.loadKnownOperations()
	return r
}

// Get returns an operation by name.
func (r *OperationsRegistry) Get(name string) (Operation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	op, ok := r.operations[name]
	return op, ok
}

// Add adds or updates an operation.
func (r *OperationsRegistry) Add(op Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.operations[op.Name] = op
}

// List returns all operations.
func (r *OperationsRegistry) List() []Operation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ops := make([]Operation, 0, len(r.operations))
	for _, op := range r.operations {
		ops = append(ops, op)
	}
	return ops
}

// ListByType returns operations of a specific type.
func (r *OperationsRegistry) ListByType(opType OperationType) []Operation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ops []Operation
	for _, op := range r.operations {
		if op.Type == opType {
			ops = append(ops, op)
		}
	}
	return ops
}

// Count returns the number of registered operations.
func (r *OperationsRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.operations)
}

// loadKnownOperations populates the registry with known Twitch GQL operations.
func (r *OperationsRegistry) loadKnownOperations() {
	now := time.Now()

	// Channel and User Queries
	knownOps := []Operation{
		// User/Channel queries
		{
			Name:        "ChannelPage_Query",
			Type:        OperationQuery,
			Description: "Get channel page information including user details",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ChannelRoot_Query",
			Type:        OperationQuery,
			Description: "Root channel query for channel data",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ChannelShell",
			Type:        OperationQuery,
			Description: "Channel shell data for layout",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "UserByLogin",
			Type:        OperationQuery,
			Description: "Get user by login name",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "UserRolesCacheQuery",
			Type:        OperationQuery,
			Description: "Get all user roles for a channel (editors, mods, VIPs, artists, lead mods)",
			SHA256Hash:  "5dc1290081dfd59b1a481cbe961d23f9b0636ddfd439d1f13f3a0f35bf818bb5",
			Variables: []VariableDefinition{
				{Name: "channelID", Type: "ID!", Required: true},
				{Name: "includeEditors", Type: "Boolean", Required: false},
				{Name: "includeMods", Type: "Boolean", Required: false},
				{Name: "includeVIPs", Type: "Boolean", Required: false},
				{Name: "includeArtists", Type: "Boolean", Required: false},
				{Name: "includeLeadMods", Type: "Boolean", Required: false},
			},
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "UserByID",
			Type:        OperationQuery,
			Description: "Get user by ID",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Stream queries
		{
			Name:        "StreamMetadata",
			Type:        OperationQuery,
			Description: "Get stream metadata including viewer count",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "UseLive",
			Type:        OperationQuery,
			Description: "Check if a channel is currently live",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "VideoPlayerStreamInfoOverlayChannel",
			Type:        OperationQuery,
			Description: "Get stream info for video player overlay",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "PlaybackAccessToken",
			Type:        OperationQuery,
			Description: "Get playback access token for stream",
			SHA256Hash:  "0828119ded1c13477966434e15800ff57ddacf13ba1911c129dc2200705b0712",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "PlaybackAccessToken_Template",
			Type:        OperationQuery,
			Description: "Template for playback access token",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "VideoAccessToken_Clip",
			Type:        OperationQuery,
			Description: "Get access token for clip playback",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Search queries
		{
			Name:        "SearchResultsPage_SearchResults",
			Type:        OperationQuery,
			Description: "Search for channels, games, and videos",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "SearchTray_SearchSuggestions",
			Type:        OperationQuery,
			Description: "Get search suggestions/autocomplete",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "SearchResults_Games",
			Type:        OperationQuery,
			Description: "Search for games/categories",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "SearchResults_Channels",
			Type:        OperationQuery,
			Description: "Search for channels",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Chat queries
		{
			Name:        "ChatRoomState",
			Type:        OperationQuery,
			Description: "Get chat room configuration and state",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ChatList_Badges",
			Type:        OperationQuery,
			Description: "Get chat badges for a channel",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "EmotePicker_EmotePicker_UserSubscriptionProducts",
			Type:        OperationQuery,
			Description: "Get emotes available to user",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ChannelPointsContext",
			Type:        OperationQuery,
			Description: "Get channel points context and rewards",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Video/VOD queries
		{
			Name:        "VideoMetadata",
			Type:        OperationQuery,
			Description: "Get video/VOD metadata",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "VideoPlayer_ChapterSelectButtonVideo",
			Type:        OperationQuery,
			Description: "Get video chapters",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "VideoPlayer_VODSeekbarPreviewVideo",
			Type:        OperationQuery,
			Description: "Get VOD seekbar preview data",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ClipsCards__User",
			Type:        OperationQuery,
			Description: "Get clips for a user",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ClipsManagerTable_User",
			Type:        OperationQuery,
			Description: "Get user's clip manager data",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Directory/Browse queries
		{
			Name:        "DirectoryPage_Game",
			Type:        OperationQuery,
			Description: "Get game/category directory page",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "BrowsePage_AllDirectories",
			Type:        OperationQuery,
			Description: "Get all directories/categories",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "Directory_DirectoryBanner",
			Type:        OperationQuery,
			Description: "Get directory banner information",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "TopNav_Games",
			Type:        OperationQuery,
			Description: "Get games for top navigation",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Following queries
		{
			Name:        "FollowedChannels",
			Type:        OperationQuery,
			Description: "Get followed channels",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "FollowingLive_CurrentUser",
			Type:        OperationQuery,
			Description: "Get live followed channels",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "FollowedStreams",
			Type:        OperationQuery,
			Description: "Get streams from followed channels",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "FollowingHosting_CurrentUser",
			Type:        OperationQuery,
			Description: "Get hosting channels from following",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Subscription queries
		{
			Name:        "PersonalSections",
			Type:        OperationQuery,
			Description: "Get personalized content sections",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "SubProducts_User",
			Type:        OperationQuery,
			Description: "Get subscription products for a channel",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Notifications
		{
			Name:        "OnsiteNotifications_Notifications",
			Type:        OperationQuery,
			Description: "Get onsite notifications",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "NotificationSettings",
			Type:        OperationQuery,
			Description: "Get notification settings",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Recommendations
		{
			Name:        "RecommendedStreams",
			Type:        OperationQuery,
			Description: "Get recommended streams",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "Shelves",
			Type:        OperationQuery,
			Description: "Get content shelves for home page",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// MUTATIONS
		{
			Name:        "FollowButton_FollowUser",
			Type:        OperationMutation,
			Description: "Follow a user",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "FollowButton_UnfollowUser",
			Type:        OperationMutation,
			Description: "Unfollow a user",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "BlockUserFromChat",
			Type:        OperationMutation,
			Description: "Block a user from chat",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "UnblockUserFromChat",
			Type:        OperationMutation,
			Description: "Unblock a user from chat",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ReportContent",
			Type:        OperationMutation,
			Description: "Report content",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ClaimCommunityPoints",
			Type:        OperationMutation,
			Description: "Claim channel points",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "RedeemCommunityPointsCustomReward",
			Type:        OperationMutation,
			Description: "Redeem a custom channel points reward",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "SendChatMessage",
			Type:        OperationMutation,
			Description: "Send a chat message",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "JoinRaid",
			Type:        OperationMutation,
			Description: "Join a raid",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "LeaveRaid",
			Type:        OperationMutation,
			Description: "Leave a raid",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "CreateClip",
			Type:        OperationMutation,
			Description: "Create a clip",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "DeleteClip",
			Type:        OperationMutation,
			Description: "Delete a clip",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "UpdateNotificationSettings",
			Type:        OperationMutation,
			Description: "Update notification settings",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "MarkNotificationsRead",
			Type:        OperationMutation,
			Description: "Mark notifications as read",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Analytics/Telemetry mutations
		{
			Name:        "SendEvents",
			Type:        OperationMutation,
			Description: "Send analytics/telemetry events to Spade",
			Query: `mutation SendEvents($input: SendSpadeEventsInput!) {
    sendSpadeEvents(input: $input) {
      statusCode
    }
  }`,
			Variables: []VariableDefinition{
				{Name: "input", Type: "SendSpadeEventsInput!", Required: true},
			},
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},

		// Moderation mutations
		{
			Name:        "Chat_BanUserFromChatRoom",
			Type:        OperationMutation,
			Description: "Ban a user from chat",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "Chat_UnbanUserFromChatRoom",
			Type:        OperationMutation,
			Description: "Unban a user from chat",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "Chat_TimeoutUserFromChatRoom",
			Type:        OperationMutation,
			Description: "Timeout a user from chat",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
		{
			Name:        "ModeratorActions_ClearChatRoom",
			Type:        OperationMutation,
			Description: "Clear all messages in chat",
			Source:      SourceKnownList,
			DiscoveredAt: now,
		},
	}

	for _, op := range knownOps {
		r.operations[op.Name] = op
	}
}

// Merge adds operations from another registry, not overwriting existing.
func (r *OperationsRegistry) Merge(other *OperationsRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	for name, op := range other.operations {
		if _, exists := r.operations[name]; !exists {
			r.operations[name] = op
		}
	}
}

// MergeOperations adds operations from a slice, not overwriting existing.
func (r *OperationsRegistry) MergeOperations(ops []Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, op := range ops {
		if _, exists := r.operations[op.Name]; !exists {
			r.operations[op.Name] = op
		}
	}
}

// Clear removes all operations from the registry.
func (r *OperationsRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.operations = make(map[string]Operation)
}

// GetQueries returns all query operations.
func (r *OperationsRegistry) GetQueries() []Operation {
	return r.ListByType(OperationQuery)
}

// GetMutations returns all mutation operations.
func (r *OperationsRegistry) GetMutations() []Operation {
	return r.ListByType(OperationMutation)
}

// GetSubscriptions returns all subscription operations.
func (r *OperationsRegistry) GetSubscriptions() []Operation {
	return r.ListByType(OperationSubscription)
}
