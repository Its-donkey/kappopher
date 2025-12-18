package helix

// Badge SetID constants for common Twitch chat badges.
// These constants represent the SetID field in ChatEventBadge and ChatBadge types.
const (
	// BadgeBroadcaster is the badge for the channel owner.
	BadgeBroadcaster = "broadcaster"

	// BadgeModerator is the badge for channel moderators.
	// Note: Users with the Lead Moderator role may display either
	// BadgeModerator or BadgeLeadModerator depending on their preference.
	BadgeModerator = "moderator"

	// BadgeLeadModerator is the badge for Lead Moderators.
	// Lead Moderators have additional privileges to help streamers manage
	// their mod teams. They can choose to display either this badge or
	// the regular moderator badge.
	BadgeLeadModerator = "lead_moderator"

	// BadgeVIP is the badge for channel VIPs.
	BadgeVIP = "vip"

	// BadgeSubscriber is the badge for channel subscribers.
	BadgeSubscriber = "subscriber"

	// BadgeFounder is the badge for channel founders (first subscribers).
	BadgeFounder = "founder"

	// BadgeSubGifter is the badge for users who have gifted subscriptions.
	BadgeSubGifter = "sub-gifter"

	// BadgeBitsLeader is the badge for bits leaderboard leaders.
	BadgeBitsLeader = "bits-leader"

	// BadgeBits is the badge showing bits tier.
	BadgeBits = "bits"

	// BadgePremium is the badge for Twitch Prime/Turbo users.
	BadgePremium = "premium"

	// BadgePartner is the badge for Twitch Partners.
	BadgePartner = "partner"

	// BadgeStaff is the badge for Twitch Staff.
	BadgeStaff = "staff"

	// BadgeAdmin is the badge for Twitch Admins.
	BadgeAdmin = "admin"

	// BadgeGlobalMod is the badge for Twitch Global Moderators.
	BadgeGlobalMod = "global_mod"

	// BadgeArtist is the badge for channel artists.
	BadgeArtist = "artist"

	// BadgeNoAudio is the badge indicating the user has no audio.
	BadgeNoAudio = "no_audio"

	// BadgeNoVideo is the badge indicating the user has no video.
	BadgeNoVideo = "no_video"

	// BadgePredictions is the badge for prediction participation.
	BadgePredictions = "predictions"

	// BadgeHypeTrain is the badge for Hype Train participation.
	BadgeHypeTrain = "hype-train"
)

// ChatEventBadges is a slice of ChatEventBadge with helper methods.
type ChatEventBadges []ChatEventBadge

// HasBadge checks if the badges contain a specific badge SetID.
func (badges ChatEventBadges) HasBadge(setID string) bool {
	for _, badge := range badges {
		if badge.SetID == setID {
			return true
		}
	}
	return false
}

// HasAnyBadge checks if the badges contain any of the specified badge SetIDs.
func (badges ChatEventBadges) HasAnyBadge(setIDs ...string) bool {
	setIDMap := make(map[string]struct{}, len(setIDs))
	for _, id := range setIDs {
		setIDMap[id] = struct{}{}
	}
	for _, badge := range badges {
		if _, ok := setIDMap[badge.SetID]; ok {
			return true
		}
	}
	return false
}

// HasModeratorPrivileges checks if the user has moderator privileges.
// This returns true if the user has either the "moderator" or "lead_moderator" badge.
// Use this method instead of checking only for BadgeModerator to properly support
// Lead Moderators who may display either badge.
func (badges ChatEventBadges) HasModeratorPrivileges() bool {
	return badges.HasAnyBadge(BadgeModerator, BadgeLeadModerator)
}

// HasBroadcasterPrivileges checks if the user is the broadcaster.
func (badges ChatEventBadges) HasBroadcasterPrivileges() bool {
	return badges.HasBadge(BadgeBroadcaster)
}

// HasVIPStatus checks if the user is a VIP.
func (badges ChatEventBadges) HasVIPStatus() bool {
	return badges.HasBadge(BadgeVIP)
}

// IsSubscriber checks if the user is a subscriber.
func (badges ChatEventBadges) IsSubscriber() bool {
	return badges.HasAnyBadge(BadgeSubscriber, BadgeFounder)
}

// IsStaff checks if the user is Twitch staff (staff, admin, or global mod).
func (badges ChatEventBadges) IsStaff() bool {
	return badges.HasAnyBadge(BadgeStaff, BadgeAdmin, BadgeGlobalMod)
}

// GetBadge returns the badge with the specified SetID, or nil if not found.
func (badges ChatEventBadges) GetBadge(setID string) *ChatEventBadge {
	for i := range badges {
		if badges[i].SetID == setID {
			return &badges[i]
		}
	}
	return nil
}

// ToChatEventBadges converts a []ChatEventBadge to ChatEventBadges for use with helper methods.
func ToChatEventBadges(badges []ChatEventBadge) ChatEventBadges {
	return ChatEventBadges(badges)
}
