package helix

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The payloads below are the event examples from the Twitch EventSub
// reference, decoded as received.

func TestAutomodMessageUpdateEventV2_DocsExample(t *testing.T) {
	payload := `{
	  "broadcaster_user_id": "1337",
	  "broadcaster_user_login": "blah",
	  "broadcaster_user_name": "blahblah",
	  "moderator_user_id": "9001",
	  "moderator_user_login": "the_mod",
	  "moderator_user_name": "The_Mod",
	  "user_id": "4242",
	  "user_login": "baduser",
	  "user_name": "badbaduser",
	  "message_id": "bad-message-id",
	  "message": {
	    "text": "This is a bad message… pogchamp",
	    "fragments": [
	      {
	        "type": "text",
	        "text": "This is a bad message… ",
	        "cheermote": null,
	        "emote": null
	      },
	      {
	        "type": "cheermote",
	        "text": "pogchamp",
	        "cheermote": {
	          "prefix": "pogchamp",
	          "bits": 1000,
	          "tier": 1
	        },
	        "emote": null
	      }
	    ]
	  },
	  "reason": "automod",
	  "automod": null,
	  "blocked_term": {
	    "terms_found": [
	      {
	        "term_id": "123",
	        "owner_broadcaster_user_id": "1337",
	        "owner_broadcaster_user_login": "blah",
	        "owner_broadcaster_user_name": "blahblah",
	        "boundary": {
	          "start_pos": 0,
	          "end_pos": 30
	        }
	      }
	    ]
	  },
	  "status": "approved",
	  "held_at": "2022-12-02T15:00:00.00Z"
	}`

	var e AutomodMessageUpdateEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.ModeratorUserID != "9001" || e.UserID != "4242" || e.MessageID != "bad-message-id" {
		t.Errorf("unexpected ids %+v", e)
	}
	if !strings.HasPrefix(e.Message.Text, "This is a bad message") || len(e.Message.Fragments) != 2 {
		t.Errorf("unexpected message %+v", e.Message)
	}
	if c := e.Message.Fragments[1].Cheermote; c == nil || c.Prefix != "pogchamp" || c.Bits != 1000 {
		t.Errorf("unexpected cheermote %+v", c)
	}
	if !strings.EqualFold(e.Status, "approved") {
		t.Errorf("unexpected status %q", e.Status)
	}
	if !e.HeldAt.Valid || e.HeldAt.Year() != 2022 {
		t.Errorf("unexpected held_at %+v", e.HeldAt)
	}
	if e.Reason != "automod" || e.Automod != nil {
		t.Errorf("unexpected reason/automod %q %+v", e.Reason, e.Automod)
	}
	if e.BlockedTerm == nil || len(e.BlockedTerm.TermsFound) != 1 {
		t.Fatalf("unexpected blocked_term %+v", e.BlockedTerm)
	}
	term := e.BlockedTerm.TermsFound[0]
	if term.TermID != "123" || term.OwnerBroadcasterUserID != "1337" || term.OwnerBroadcasterUserLogin != "blah" || term.OwnerBroadcasterUserName != "blahblah" {
		t.Errorf("unexpected term %+v", term)
	}
	if term.Boundary.EndPos != 30 {
		t.Errorf("unexpected boundary %+v", term.Boundary)
	}
}

func TestAutomodMessageHoldEventV2_DocsExample(t *testing.T) {
	payload := `{
	  "broadcaster_user_id": "1337",
	  "broadcaster_user_login": "blah",
	  "broadcaster_user_name": "blahblah",
	  "user_id": "4242",
	  "user_login": "baduser",
	  "user_name": "badbaduser",
	  "message_id": "bad-message-id",
	  "message": {
	    "text": "This is a bad message… pogchamp",
	    "fragments": [
	      {
	        "type": "text",
	        "text": "This is a bad message… ",
	        "cheermote": null,
	        "emote": null
	      },
	      {
	        "type": "cheermote",
	        "text": "pogchamp",
	        "cheermote": {
	          "prefix": "pogchamp",
	          "bits": 1000,
	          "tier": 1
	        },
	        "emote": null
	      }
	    ]
	  },
	  "reason": "automod",
	  "automod": {
	    "category": "aggressive",
	    "level": 1,
	    "boundaries": [
	      {
	        "start_pos": 0,
	        "end_pos": 10
	      },
	      {
	        "start_pos": 20,
	        "end_pos": 30
	      }
	    ]
	  },
	  "blocked_term": null,
	  "held_at": "2022-12-02T15:00:00.00Z"
	}`

	var e AutomodMessageHoldEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Reason != "automod" || e.Automod == nil || e.Automod.Category != "aggressive" || len(e.Automod.Boundaries) != 2 {
		t.Errorf("unexpected automod %+v", e.Automod)
	}
	if e.BlockedTerm != nil {
		t.Errorf("expected nil blocked_term, got %+v", e.BlockedTerm)
	}
	if e.Message.Fragments[0].Type != "text" || e.HeldAt.Year() != 2022 {
		t.Errorf("unexpected message/held_at %+v %v", e.Message, e.HeldAt)
	}
}

func TestChannelCharityCampaignStartEvent_DocsExample(t *testing.T) {
	payload := `{
	  "id": "123-abc-456-def",
	  "broadcaster_id": "123456",
	  "broadcaster_name": "SunnySideUp",
	  "broadcaster_login": "sunnysideup",
	  "charity_name": "Example name",
	  "charity_description": "Example description",
	  "charity_logo": "https://abc.cloudfront.net/ppgf/1000/100.png",
	  "charity_website": "https://www.example.com",
	  "current_amount": {
	    "value": 0,
	    "decimal_places": 2,
	    "currency": "USD"
	  },
	  "target_amount": {
	    "value": 1500000,
	    "decimal_places": 2,
	    "currency": "USD"
	  },
	  "started_at": "2022-07-26T17:00:03.17106713Z"
	}`

	var e ChannelCharityCampaignStartEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.BroadcasterUserID != "123456" || e.BroadcasterUserLogin != "sunnysideup" || e.BroadcasterUserName != "SunnySideUp" {
		t.Errorf("broadcaster not decoded: %+v", e.EventSubBroadcaster)
	}
	if e.ID != "123-abc-456-def" || e.TargetAmount.Value != 1500000 || e.StartedAt.IsZero() {
		t.Errorf("unexpected event %+v", e)
	}

	// Progress shares the start event's type.
	var p ChannelCharityCampaignProgressEvent
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatalf("unmarshal progress: %v", err)
	}
	if p.BroadcasterUserID != "123456" {
		t.Errorf("progress broadcaster not decoded: %+v", p.EventSubBroadcaster)
	}
}

func TestChannelCharityCampaignStopEvent_Broadcaster(t *testing.T) {
	payload := `{"id":"123-abc-456-def","broadcaster_id":"123456","broadcaster_name":"SunnySideUp","broadcaster_login":"sunnysideup","charity_name":"Example name","current_amount":{"value":1450000,"decimal_places":2,"currency":"USD"},"target_amount":{"value":1500000,"decimal_places":2,"currency":"USD"},"stopped_at":"2022-07-26T22:00:03.17106713Z"}`

	var e ChannelCharityCampaignStopEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.BroadcasterUserID != "123456" || e.BroadcasterUserLogin != "sunnysideup" || e.BroadcasterUserName != "SunnySideUp" {
		t.Errorf("broadcaster not decoded: %+v", e.EventSubBroadcaster)
	}
	if e.CurrentAmount.Value != 1450000 || e.StoppedAt.IsZero() {
		t.Errorf("unexpected event %+v", e)
	}
}

func TestDropEntitlementGrant_DocsExample(t *testing.T) {
	// Webhook notification body from the drop.entitlement.grant reference:
	// batched, with an "events" array and no "event".
	body := `{
	  "subscription": {
	    "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
	    "type": "drop.entitlement.grant",
	    "version": "1",
	    "status": "enabled",
	    "cost": 0,
	    "condition": {"organization_id": "9001", "category_id": "9002", "campaign_id": "9003"},
	    "transport": {"method": "webhook", "callback": "https://example.com/webhooks/callback"},
	    "created_at": "2019-11-16T10:11:12.634234626Z",
	    "is_batching_enabled": true
	  },
	  "events": [
	    {
	      "id": "bf7c8577-e3e3-4881-a78a-e9446641d45d",
	      "data": {
	        "organization_id": "9001",
	        "category_id": "9002",
	        "category_name": "Fortnite",
	        "campaign_id": "9003",
	        "user_id": "1234",
	        "user_name": "Cool_User",
	        "user_login": "cool_user",
	        "entitlement_id": "fb78259e-fb81-4d1b-8333-34a06ffc24c0",
	        "benefit_id": "74c52265-e214-48a6-91b9-23b6014e8041",
	        "created_at": "2019-01-28T04:17:53.325Z"
	      }
	    },
	    {
	      "id": "bf7c8577-e3e3-4881-a78a-e9446641d45c",
	      "data": {
	        "organization_id": "9001",
	        "category_id": "9002",
	        "category_name": "Fortnite",
	        "campaign_id": "9003",
	        "user_id": "12345",
	        "user_name": "Cooler_User",
	        "user_login": "cooler_user",
	        "entitlement_id": "fb78259e-fb81-4d1b-8333-34a06ffc24c0",
	        "benefit_id": "74c52265-e214-48a6-91b9-23b6014e8041",
	        "created_at": "2019-01-28T04:17:53.325Z"
	      }
	    }
	  ]
	}`

	check := func(t *testing.T, events []DropEntitlementGrantEvent) {
		t.Helper()
		if len(events) != 2 {
			t.Fatalf("got %d events, want 2", len(events))
		}
		e := events[1]
		if e.ID != "bf7c8577-e3e3-4881-a78a-e9446641d45c" {
			t.Errorf("unexpected id %q", e.ID)
		}
		d := e.Data
		if d.OrganizationID != "9001" || d.CategoryID != "9002" || d.CategoryName != "Fortnite" || d.CampaignID != "9003" {
			t.Errorf("unexpected entitlement %+v", d)
		}
		if d.UserID != "12345" || d.UserLogin != "cooler_user" || d.UserName != "Cooler_User" {
			t.Errorf("unexpected user %+v", d)
		}
		if d.EntitlementID != "fb78259e-fb81-4d1b-8333-34a06ffc24c0" || d.BenefitID != "74c52265-e214-48a6-91b9-23b6014e8041" || d.CreatedAt.Year() != 2019 {
			t.Errorf("unexpected entitlement %+v", d)
		}
	}

	t.Run("webhook", func(t *testing.T) {
		var got *EventSubWebhookMessage
		h := NewEventSubWebhookHandler(WithNotificationHandler(func(m *EventSubWebhookMessage) { got = m }))
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, "msg-1")
		req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
		req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)
		req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeDropEntitlementGrant)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent || got == nil {
			t.Fatalf("status %d, message %v", w.Code, got)
		}
		if len(got.Event) != 0 {
			t.Errorf("expected no event, got %s", got.Event)
		}
		if !got.Subscription.IsBatchingEnabled {
			t.Error("expected is_batching_enabled")
		}
		events, err := ParseEventSubEvents[DropEntitlementGrantEvent](got.Events)
		if err != nil {
			t.Fatalf("ParseEventSubEvents: %v", err)
		}
		check(t, events)
	})

	t.Run("websocket", func(t *testing.T) {
		var raw json.RawMessage
		c := NewEventSubWebSocketClient(WithWSNotificationHandler(func(_ *EventSubSubscription, ev json.RawMessage) { raw = ev }))
		c.handleNotification(WebSocketMessage{Payload: json.RawMessage(body)})
		events, err := ParseEventSubEvents[DropEntitlementGrantEvent](raw)
		if err != nil {
			t.Fatalf("ParseEventSubEvents: %v", err)
		}
		check(t, events)
	})
}

func TestParseEventSubEvents_EmptyAndInvalid(t *testing.T) {
	events, err := ParseEventSubEvents[DropEntitlementGrantEvent](nil)
	if err != nil || events != nil {
		t.Errorf("empty: got %v, %v", events, err)
	}
	if _, err := ParseEventSubEvents[DropEntitlementGrantEvent](json.RawMessage(`{"id":"x"}`)); err == nil {
		t.Error("expected error for a non-array")
	}
}

func TestChannelPointsAutomaticRewardRedemptionAddEventV2_DocsExample(t *testing.T) {
	payload := `{
	  "broadcaster_user_id": "12826",
	  "broadcaster_user_name": "Twitch",
	  "broadcaster_user_login": "twitch",
	  "user_id": "141981764",
	  "user_name": "TwitchDev",
	  "user_login": "twitchdev",
	  "id": "f024099a-e0fe-4339-9a0a-a706fb59f353",
	  "reward": {
	    "type": "send_highlighted_message",
	    "channel_points": 100,
	    "emote": null
	  },
	  "message": {
	    "text": "Hello world! VoHiYo",
	    "fragments": [
	      {"type": "text", "text": "Hello world! ", "emote": null},
	      {"type": "emote", "text": "VoHiYo", "emote": {"id": "81274"}}
	    ]
	  },
	  "redeemed_at": "2024-08-12T21:14:34.260398045Z"
	}`

	var e ChannelPointsAutomaticRewardRedemptionAddEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Reward.Type != AutomaticRewardTypeSendHighlightedMessage || e.Reward.ChannelPoints != 100 || e.Reward.Emote != nil {
		t.Errorf("unexpected reward %+v", e.Reward)
	}
	if e.Message.Text != "Hello world! VoHiYo" || len(e.Message.Fragments) != 2 {
		t.Fatalf("unexpected message %+v", e.Message)
	}
	if f := e.Message.Fragments[1]; f.Type != "emote" || f.Emote == nil || f.Emote.ID != "81274" {
		t.Errorf("unexpected fragment %+v", f)
	}
	if e.Message.Fragments[0].Emote != nil {
		t.Errorf("expected nil emote on text fragment")
	}
	if e.ID != "f024099a-e0fe-4339-9a0a-a706fb59f353" || e.UserID != "141981764" || e.RedeemedAt.Year() != 2024 {
		t.Errorf("unexpected event %+v", e)
	}

	// The reward's emote and a null message.
	payload = `{"id":"x","reward":{"type":"random_sub_emote_unlock","channel_points":50,"emote":{"id":"301","name":"Kappa"}},"message":null,"redeemed_at":"2024-08-12T21:14:34Z"}`
	e = ChannelPointsAutomaticRewardRedemptionAddEvent{}
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Reward.Emote == nil || e.Reward.Emote.ID != "301" || e.Reward.Emote.Name != "Kappa" {
		t.Errorf("unexpected reward emote %+v", e.Reward.Emote)
	}
	if e.Message.Text != "" || e.Message.Fragments != nil {
		t.Errorf("expected zero message, got %+v", e.Message)
	}
}

func TestChannelPointsAutomaticRewardRedemptionAddEventV1_DocsExample(t *testing.T) {
	payload := `{"broadcaster_user_id": "12826", "broadcaster_user_name": "Twitch", "broadcaster_user_login": "twitch", "user_id": "141981764", "user_name": "TwitchDev", "user_login": "twitchdev", "id": "f024099a-e0fe-4339-9a0a-a706fb59f353", "reward": {"type": "send_highlighted_message", "cost": 100, "unlocked_emote": null}, "message": {"text": "Hello world! VoHiYo", "emotes": [{"id": "81274", "begin": 13, "end": 18}]}, "user_input": "Hello world! VoHiYo ", "redeemed_at": "2024-02-23T21:14:34.260398045Z"}`

	var e ChannelPointsAutomaticRewardRedemptionAddEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Reward.Cost != 100 || len(e.Message.Emotes) != 1 || e.Message.Emotes[0].Begin != 13 || e.UserInput == "" {
		t.Errorf("v1 fields not decoded: %+v", e)
	}
}

func TestChannelSharedChatBeginEvent_DocsExample(t *testing.T) {
	payload := `{
	  "session_id": "2b64a92a-dbb8-424e-b1c3-304423ba1b6f",
	  "broadcaster_user_id": "1971641",
	  "broadcaster_user_login": "streamer",
	  "broadcaster_user_name": "streamer",
	  "host_broadcaster_user_id": "1971641",
	  "host_broadcaster_user_login": "streamer",
	  "host_broadcaster_user_name": "streamer",
	  "participants": [
	    {"broadcaster_user_id": "1971641", "broadcaster_user_name": "streamer", "broadcaster_user_login": "streamer"},
	    {"broadcaster_user_id": "112233", "broadcaster_user_name": "streamer33", "broadcaster_user_login": "streamer33"}
	  ]
	}`

	var e ChannelSharedChatBeginEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(e.Participants) != 2 {
		t.Fatalf("got %d participants", len(e.Participants))
	}
	p := e.Participants[1]
	if p.BroadcasterUserID != "112233" || p.BroadcasterUserLogin != "streamer33" || p.BroadcasterUserName != "streamer33" {
		t.Errorf("unexpected participant %+v", p)
	}
	if p.BroadcasterID != "112233" {
		t.Errorf("deprecated BroadcasterID not filled: %+v", p)
	}

	// Update shares the type.
	var u ChannelSharedChatUpdateEvent
	if err := json.Unmarshal([]byte(payload), &u); err != nil || u.Participants[0].BroadcasterUserID != "1971641" {
		t.Errorf("update: %v %+v", err, u.Participants)
	}
}

func TestStreamOfflineEvent_DocsExample(t *testing.T) {
	payload := `{"id": "9001", "broadcaster_user_id": "1337", "broadcaster_user_login": "cool_user", "broadcaster_user_name": "Cool_User"}`

	var e StreamOfflineEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.ID != "9001" || e.BroadcasterUserID != "1337" {
		t.Errorf("unexpected event %+v", e)
	}
}

func TestChannelPredictionLockEvent_DocsExample(t *testing.T) {
	// The docs example sends the last top predictor's user_id as the number
	// 12345; Twitch sends IDs as strings, so it is quoted here.
	payload := `{
	  "id": "1243456",
	  "broadcaster_user_id": "1337",
	  "broadcaster_user_login": "cool_user",
	  "broadcaster_user_name": "Cool_User",
	  "title": "Aren't shoes just really hard socks?",
	  "outcomes": [
	    {
	      "id": "1243456",
	      "title": "Yeah!",
	      "color": "blue",
	      "users": 10,
	      "channel_points": 15000,
	      "top_predictors": [
	        {"user_name": "Cool_User", "user_login": "cool_user", "user_id": "1234", "channel_points_won": null, "channel_points_used": 500},
	        {"user_name": "Coolest_User", "user_login": "coolest_user", "user_id": "1236", "channel_points_won": null, "channel_points_used": 200}
	      ]
	    },
	    {
	      "id": "2243456",
	      "title": "No!",
	      "color": "pink",
	      "top_predictors": [
	        {"user_name": "Cooler_User", "user_login": "cooler_user", "user_id": "12345", "channel_points_won": null, "channel_points_used": 5000}
	      ]
	    }
	  ],
	  "started_at": "2020-07-15T17:16:03.17106713Z",
	  "locked_at": "2020-07-15T17:21:03.17106713Z"
	}`

	var e ChannelPredictionLockEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := time.Date(2020, 7, 15, 17, 21, 3, 171067130, time.UTC)
	if !e.LockedAt.Equal(want) {
		t.Errorf("LockedAt = %v, want %v", e.LockedAt, want)
	}
	if e.ID != "1243456" || e.BroadcasterUserID != "1337" || len(e.Outcomes) != 2 || e.StartedAt.IsZero() {
		t.Errorf("begin fields not decoded: %+v", e)
	}
	if !e.LocksAt.IsZero() {
		t.Errorf("LocksAt should be zero, got %v", e.LocksAt)
	}
	if e.Outcomes[1].TopPredictors[0].UserID != "12345" {
		t.Errorf("unexpected predictor %+v", e.Outcomes[1].TopPredictors[0])
	}
}

func TestChannelBitsUseEvent_PowerUps(t *testing.T) {
	payload := `{
	  "user_id": "1234", "user_login": "cool_user", "user_name": "Cool_User",
	  "broadcaster_user_id": "1337", "broadcaster_user_login": "cooler_user", "broadcaster_user_name": "Cooler_User",
	  "bits": 50,
	  "type": "power_up",
	  "power_up": {"type": "gigantify_an_emote", "emote": {"id": "301", "name": "Kappa"}, "message_effect_id": null},
	  "custom_power_up": null,
	  "message": {"text": "Kappa", "fragments": [{"type": "emote", "text": "Kappa", "cheermote": null, "emote": {"id": "301", "emote_set_id": "0"}}]}
	}`

	var e ChannelBitsUseEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Type != BitsUseTypePowerUp || e.PowerUp == nil || e.CustomPowerUp != nil {
		t.Fatalf("unexpected power-ups %+v %+v", e.PowerUp, e.CustomPowerUp)
	}
	if e.PowerUp.Type != PowerUpTypeGigantifyAnEmote || e.PowerUp.Emote == nil || e.PowerUp.Emote.Name != "Kappa" || e.PowerUp.MessageEffectID != nil {
		t.Errorf("unexpected power_up %+v", e.PowerUp)
	}

	payload = `{"bits": 100, "type": "power_up", "power_up": {"type": "message_effect", "emote": null, "message_effect_id": "cosmic-abyss"}, "custom_power_up": null}`
	e = ChannelBitsUseEvent{}
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.PowerUp == nil || e.PowerUp.Type != PowerUpTypeMessageEffect || e.PowerUp.MessageEffectID == nil || *e.PowerUp.MessageEffectID != "cosmic-abyss" || e.PowerUp.Emote != nil {
		t.Errorf("unexpected power_up %+v", e.PowerUp)
	}

	payload = `{"bits": 200, "type": "custom_power_up", "power_up": null, "custom_power_up": {"title": "Hydrate", "reward_id": "abc-123"}}`
	e = ChannelBitsUseEvent{}
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Type != BitsUseTypeCustomPowerUp || e.PowerUp != nil || e.CustomPowerUp == nil || e.CustomPowerUp.Title != "Hydrate" || e.CustomPowerUp.RewardID != "abc-123" {
		t.Errorf("unexpected custom_power_up %+v", e.CustomPowerUp)
	}
}

func TestChannelSubscriptionGiftEvent_Anonymous(t *testing.T) {
	// cumulative_total is null for anonymous gifts.
	payload := `{"user_id": null, "user_login": null, "user_name": null, "broadcaster_user_id": "1337", "broadcaster_user_login": "cooler_user", "broadcaster_user_name": "Cooler_User", "total": 2, "tier": "1000", "cumulative_total": null, "is_anonymous": true}`

	var e ChannelSubscriptionGiftEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !e.IsAnonymous || e.CumulativeTotal != nil || e.Total != 2 {
		t.Errorf("unexpected event %+v", e)
	}

	// A gifter who shares a count of zero is distinguishable from null.
	payload = `{"total": 1, "tier": "1000", "cumulative_total": 0, "is_anonymous": false}`
	e = ChannelSubscriptionGiftEvent{}
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.CumulativeTotal == nil || *e.CumulativeTotal != 0 {
		t.Errorf("expected cumulative_total 0, got %v", e.CumulativeTotal)
	}
}

func TestChannelSubscriptionMessageEvent_DocsExample(t *testing.T) {
	payload := `{
	  "user_id": "1234",
	  "user_login": "cool_user",
	  "user_name": "Cool_User",
	  "broadcaster_user_id": "1337",
	  "broadcaster_user_login": "cooler_user",
	  "broadcaster_user_name": "Cooler_User",
	  "tier": "1000",
	  "message": {
	    "text": "Love the stream! FevziGG",
	    "emotes": [{"begin": 23, "end": 30, "id": "302976485"}]
	  },
	  "cumulative_months": 15,
	  "streak_months": 1,
	  "duration_months": 6
	}`

	var e ChannelSubscriptionMessageEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.StreakMonths == nil || *e.StreakMonths != 1 || e.CumulativeMonths != 15 || e.DurationMonths != 6 {
		t.Errorf("unexpected event %+v", e)
	}

	// streak_months is null when the user doesn't share it.
	if err := json.Unmarshal([]byte(`{"cumulative_months": 15, "streak_months": null, "duration_months": 1}`), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.StreakMonths != nil {
		t.Errorf("expected nil StreakMonths, got %v", *e.StreakMonths)
	}
}

func TestChannelChatNotificationEvent_DocsExample(t *testing.T) {
	payload := `{"broadcaster_user_id": "1971641", "broadcaster_user_login": "streamer", "broadcaster_user_name": "streamer", "chatter_user_id": "49912639", "chatter_user_login": "viewer23", "chatter_user_name": "viewer23", "chatter_is_anonymous": false, "color": "", "badges": [], "system_message": "viewer23 subscribed at Tier 1. They've subscribed for 10 months!", "message_id": "d62235c8-47ff-a4f4--84e8-5a29a65a9c03", "message": {"text": "", "fragments": []}, "notice_type": "resub", "sub": null, "resub": {"cumulative_months": 10, "duration_months": 0, "streak_months": null, "sub_plan": "1000", "is_gift": false, "gifter_is_anonymous": null, "gifter_user_id": null, "gifter_user_name": null, "gifter_user_login": null}, "sub_gift": null, "community_sub_gift": null, "gift_paid_upgrade": null, "prime_paid_upgrade": null, "pay_it_forward": null, "raid": null, "unraid": null, "announcement": null, "bits_badge_tier": null, "charity_donation": null, "watch_streak": null, "modiversary": null, "shared_chat_sub": null, "shared_chat_resub": null, "shared_chat_sub_gift": null, "shared_chat_community_sub_gift": null, "shared_chat_gift_paid_upgrade": null, "shared_chat_prime_paid_upgrade": null, "shared_chat_pay_it_forward": null, "shared_chat_raid": null, "shared_chat_announcement": null, "shared_chat_modiversary": null, "source_broadcaster_user_id": null, "source_broadcaster_user_login": null, "source_broadcaster_user_name": null, "source_message_id": null, "source_badges": null, "is_source_only": null}`

	var e ChannelChatNotificationEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.NoticeType != ChatNoticeTypeResub || e.Resub == nil {
		t.Fatalf("unexpected notice %q %+v", e.NoticeType, e.Resub)
	}
	if e.Resub.StreakMonths != nil {
		t.Errorf("expected nil streak_months, got %d", *e.Resub.StreakMonths)
	}
	if e.Resub.CumulativeMonths != 10 || e.Resub.SubTier != "1000" {
		t.Errorf("unexpected resub %+v", e.Resub)
	}
	if e.GiftedDropsSummary != nil || e.SharedChatGiftedDropsSummary != nil {
		t.Error("expected nil gifted drops summaries")
	}
}

func TestChannelChatNotificationEvent_GiftedDropsSummary(t *testing.T) {
	payload := `{"notice_type": "gifted_drops_summary", "gifted_drops_summary": {"recipient_count": 25}, "shared_chat_gifted_drops_summary": null}`
	var e ChannelChatNotificationEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.NoticeType != ChatNoticeTypeGiftedDropsSummary || e.GiftedDropsSummary == nil || e.GiftedDropsSummary.RecipientCount != 25 {
		t.Errorf("unexpected summary %+v", e.GiftedDropsSummary)
	}

	payload = `{"notice_type": "shared_chat_gifted_drops_summary", "gifted_drops_summary": null, "shared_chat_gifted_drops_summary": {"recipient_count": 3}}`
	e = ChannelChatNotificationEvent{}
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.NoticeType != ChatNoticeTypeSharedChatGiftedDropsSummary || e.SharedChatGiftedDropsSummary == nil || e.SharedChatGiftedDropsSummary.RecipientCount != 3 {
		t.Errorf("unexpected shared summary %+v", e.SharedChatGiftedDropsSummary)
	}
}

func TestChannelChatSettingsUpdateEvent_DocsExample(t *testing.T) {
	payload := `{"broadcaster_user_id": "1337", "broadcaster_user_login": "cool_user", "broadcaster_user_name": "Cool_User", "emote_mode": true, "follower_mode": false, "follower_mode_duration_minutes": null, "slow_mode": true, "slow_mode_wait_time_seconds": 10, "subscriber_mode": false, "unique_chat_mode": false}`

	var e ChannelChatSettingsUpdateEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !e.SlowMode || e.SlowModeWaitTimeSeconds == nil || *e.SlowModeWaitTimeSeconds != 10 || e.FollowerModeDurationMinutes != nil {
		t.Errorf("unexpected settings %+v", e)
	}

	// slow_mode_wait_time_seconds is null when slow mode is off.
	e = ChannelChatSettingsUpdateEvent{}
	if err := json.Unmarshal([]byte(`{"slow_mode": false, "slow_mode_wait_time_seconds": null}`), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.SlowModeWaitTimeSeconds != nil {
		t.Errorf("expected nil wait time, got %d", *e.SlowModeWaitTimeSeconds)
	}
}

func TestChatEventFragment_Gif(t *testing.T) {
	payload := `{"text": "hi ", "fragments": [
	  {"type": "text", "text": "hi ", "cheermote": null, "emote": null, "mention": null, "gif": null},
	  {"type": "gif", "text": "", "cheermote": null, "emote": null, "mention": null, "gif": {"id": "gif-1", "url": "https://media.example/gif-1.gif"}},
	  {"type": "gif", "text": "", "gif": {"gif_id": "gif-2", "url": "https://media.example/gif-2.gif"}}
	]}`

	var m ChatEventMessage
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.Fragments[0].Gif != nil {
		t.Error("expected nil gif on text fragment")
	}
	if g := m.Fragments[1].Gif; g == nil || g.ID != "gif-1" || g.URL != "https://media.example/gif-1.gif" {
		t.Errorf("unexpected gif %+v", g)
	}
	if g := m.Fragments[2].Gif; g == nil || g.ID != "gif-2" {
		t.Errorf("gif_id spelling not accepted: %+v", g)
	}
	if err := json.Unmarshal([]byte(`{"gif": {"id": 5}}`), &ChatEventFragment{}); err == nil {
		t.Error("expected error for invalid gif")
	}
}

func TestChannelChatUserMessageHoldEvent_DocsExample(t *testing.T) {
	payload := `{"broadcaster_user_id": "123", "broadcaster_user_login": "bob", "broadcaster_user_name": "Bob", "user_id": "456", "user_login": "tom", "user_name": "Tommy", "message_id": "789", "message": {"text": "hey world", "fragments": [{"type": "emote", "text": "hey world", "cheermote": null, "emote": {"id": "foo", "emote_set_id": "7"}}, {"type": "cheermote", "text": "bye world", "cheermote": {"prefix": "prefix", "bits": 100, "tier": 1}, "emote": null}, {"type": "text", "text": "surprise", "cheermote": null, "emote": null}]}}`

	var e ChannelChatUserMessageHoldEvent
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(e.Message.Fragments) != 3 || e.Message.Fragments[0].Emote == nil || e.Message.Fragments[1].Cheermote == nil {
		t.Errorf("unexpected message %+v", e.Message)
	}

	// mention and gif fragments.
	payload = `{"message": {"text": "@bob x", "fragments": [
	  {"type": "mention", "text": "@bob", "mention": {"user_id": "123", "user_login": "bob", "user_name": "Bob"}},
	  {"type": "gif", "text": "", "gif": {"id": "g", "url": "https://media.example/g.gif"}}
	]}}`
	e = ChannelChatUserMessageHoldEvent{}
	if err := json.Unmarshal([]byte(payload), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m := e.Message.Fragments[0].Mention; m == nil || m.UserLogin != "bob" {
		t.Errorf("unexpected mention %+v", m)
	}
	if g := e.Message.Fragments[1].Gif; g == nil || g.ID != "g" {
		t.Errorf("unexpected gif %+v", g)
	}
}
