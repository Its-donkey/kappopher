package helix

import "testing"

func TestChatEventBadges_HasBadge(t *testing.T) {
	badges := ChatEventBadges{
		{SetID: BadgeModerator, ID: "1", Info: ""},
		{SetID: BadgeSubscriber, ID: "12", Info: "12"},
	}

	tests := []struct {
		name   string
		setID  string
		expect bool
	}{
		{"has moderator badge", BadgeModerator, true},
		{"has subscriber badge", BadgeSubscriber, true},
		{"does not have broadcaster badge", BadgeBroadcaster, false},
		{"does not have VIP badge", BadgeVIP, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := badges.HasBadge(tt.setID); got != tt.expect {
				t.Errorf("HasBadge(%q) = %v, want %v", tt.setID, got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_HasAnyBadge(t *testing.T) {
	badges := ChatEventBadges{
		{SetID: BadgeModerator, ID: "1", Info: ""},
	}

	tests := []struct {
		name   string
		setIDs []string
		expect bool
	}{
		{"has one of moderator or broadcaster", []string{BadgeModerator, BadgeBroadcaster}, true},
		{"has none of VIP or subscriber", []string{BadgeVIP, BadgeSubscriber}, false},
		{"single match", []string{BadgeModerator}, true},
		{"empty check", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := badges.HasAnyBadge(tt.setIDs...); got != tt.expect {
				t.Errorf("HasAnyBadge(%v) = %v, want %v", tt.setIDs, got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_HasModeratorPrivileges(t *testing.T) {
	tests := []struct {
		name   string
		badges ChatEventBadges
		expect bool
	}{
		{
			name: "moderator badge",
			badges: ChatEventBadges{
				{SetID: BadgeModerator, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "lead moderator badge",
			badges: ChatEventBadges{
				{SetID: BadgeLeadModerator, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "lead moderator with subscriber",
			badges: ChatEventBadges{
				{SetID: BadgeLeadModerator, ID: "1", Info: ""},
				{SetID: BadgeSubscriber, ID: "12", Info: "12"},
			},
			expect: true,
		},
		{
			name: "only subscriber",
			badges: ChatEventBadges{
				{SetID: BadgeSubscriber, ID: "12", Info: "12"},
			},
			expect: false,
		},
		{
			name:   "no badges",
			badges: ChatEventBadges{},
			expect: false,
		},
		{
			name: "broadcaster only (not mod)",
			badges: ChatEventBadges{
				{SetID: BadgeBroadcaster, ID: "1", Info: ""},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.badges.HasModeratorPrivileges(); got != tt.expect {
				t.Errorf("HasModeratorPrivileges() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_HasBroadcasterPrivileges(t *testing.T) {
	tests := []struct {
		name   string
		badges ChatEventBadges
		expect bool
	}{
		{
			name: "broadcaster badge",
			badges: ChatEventBadges{
				{SetID: BadgeBroadcaster, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "moderator only",
			badges: ChatEventBadges{
				{SetID: BadgeModerator, ID: "1", Info: ""},
			},
			expect: false,
		},
		{
			name:   "no badges",
			badges: ChatEventBadges{},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.badges.HasBroadcasterPrivileges(); got != tt.expect {
				t.Errorf("HasBroadcasterPrivileges() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_HasVIPStatus(t *testing.T) {
	tests := []struct {
		name   string
		badges ChatEventBadges
		expect bool
	}{
		{
			name: "VIP badge",
			badges: ChatEventBadges{
				{SetID: BadgeVIP, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "moderator only",
			badges: ChatEventBadges{
				{SetID: BadgeModerator, ID: "1", Info: ""},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.badges.HasVIPStatus(); got != tt.expect {
				t.Errorf("HasVIPStatus() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_IsSubscriber(t *testing.T) {
	tests := []struct {
		name   string
		badges ChatEventBadges
		expect bool
	}{
		{
			name: "subscriber badge",
			badges: ChatEventBadges{
				{SetID: BadgeSubscriber, ID: "12", Info: "12"},
			},
			expect: true,
		},
		{
			name: "founder badge",
			badges: ChatEventBadges{
				{SetID: BadgeFounder, ID: "0", Info: ""},
			},
			expect: true,
		},
		{
			name: "neither subscriber nor founder",
			badges: ChatEventBadges{
				{SetID: BadgeModerator, ID: "1", Info: ""},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.badges.IsSubscriber(); got != tt.expect {
				t.Errorf("IsSubscriber() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_IsStaff(t *testing.T) {
	tests := []struct {
		name   string
		badges ChatEventBadges
		expect bool
	}{
		{
			name: "staff badge",
			badges: ChatEventBadges{
				{SetID: BadgeStaff, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "admin badge",
			badges: ChatEventBadges{
				{SetID: BadgeAdmin, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "global mod badge",
			badges: ChatEventBadges{
				{SetID: BadgeGlobalMod, ID: "1", Info: ""},
			},
			expect: true,
		},
		{
			name: "regular moderator",
			badges: ChatEventBadges{
				{SetID: BadgeModerator, ID: "1", Info: ""},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.badges.IsStaff(); got != tt.expect {
				t.Errorf("IsStaff() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestChatEventBadges_GetBadge(t *testing.T) {
	badges := ChatEventBadges{
		{SetID: BadgeModerator, ID: "1", Info: ""},
		{SetID: BadgeSubscriber, ID: "12", Info: "12"},
	}

	t.Run("found badge", func(t *testing.T) {
		badge := badges.GetBadge(BadgeSubscriber)
		if badge == nil {
			t.Fatal("expected to find subscriber badge")
		}
		if badge.ID != "12" {
			t.Errorf("expected badge ID '12', got %q", badge.ID)
		}
		if badge.Info != "12" {
			t.Errorf("expected badge Info '12', got %q", badge.Info)
		}
	})

	t.Run("badge not found", func(t *testing.T) {
		badge := badges.GetBadge(BadgeBroadcaster)
		if badge != nil {
			t.Errorf("expected nil, got %+v", badge)
		}
	})
}

func TestToChatEventBadges(t *testing.T) {
	original := []ChatEventBadge{
		{SetID: BadgeModerator, ID: "1", Info: ""},
		{SetID: BadgeSubscriber, ID: "12", Info: "12"},
	}

	converted := ToChatEventBadges(original)

	if len(converted) != len(original) {
		t.Errorf("expected length %d, got %d", len(original), len(converted))
	}

	// Verify helper methods work after conversion
	if !converted.HasModeratorPrivileges() {
		t.Error("expected HasModeratorPrivileges() to return true")
	}
}

func TestBadgeConstants(t *testing.T) {
	// Verify badge constants have expected values
	constants := map[string]string{
		"BadgeBroadcaster":   BadgeBroadcaster,
		"BadgeModerator":     BadgeModerator,
		"BadgeLeadModerator": BadgeLeadModerator,
		"BadgeVIP":           BadgeVIP,
		"BadgeSubscriber":    BadgeSubscriber,
		"BadgeFounder":       BadgeFounder,
	}

	expected := map[string]string{
		"BadgeBroadcaster":   "broadcaster",
		"BadgeModerator":     "moderator",
		"BadgeLeadModerator": "lead_moderator",
		"BadgeVIP":           "vip",
		"BadgeSubscriber":    "subscriber",
		"BadgeFounder":       "founder",
	}

	for name, got := range constants {
		if want := expected[name]; got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}
