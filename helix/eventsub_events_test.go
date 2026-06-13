package helix

import (
	"encoding/json"
	"testing"
)

// Tests for eventsub_events.go event types using official Twitch API documentation payloads.
// Reference: https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/

// TestEventSubEvents_NewlyAddedFields verifies fields added to match the
// official EventSub reference (shared-chat source fields, host_user_*,
// is_source_only, watch_streak, modiversary).
func TestEventSubEvents_NewlyAddedFields(t *testing.T) {
	t.Run("moderate", func(t *testing.T) {
		var e ChannelModerateEvent
		if err := json.Unmarshal([]byte(`{
			"broadcaster_user_id":"1","moderator_user_id":"2","action":"shared_chat_unban",
			"source_broadcaster_user_id":"9","source_broadcaster_user_login":"src","source_broadcaster_user_name":"Src",
			"shared_chat_unban":{"user_id":"5","user_login":"u","user_name":"U"}
		}`), &e); err != nil {
			t.Fatal(err)
		}
		if e.SourceBroadcasterUserID == nil || *e.SourceBroadcasterUserID != "9" {
			t.Errorf("source_broadcaster_user_id not decoded: %v", e.SourceBroadcasterUserID)
		}
		if e.SharedChatUnban == nil || e.SharedChatUnban.UserID != "5" {
			t.Errorf("shared_chat_unban not decoded: %+v", e.SharedChatUnban)
		}
	})

	t.Run("notification", func(t *testing.T) {
		var e ChannelChatNotificationEvent
		if err := json.Unmarshal([]byte(`{
			"notice_type":"watch_streak","is_source_only":true,
			"watch_streak":{"streak_count":3,"channel_points_awarded":350},
			"modiversary":{"months":12}
		}`), &e); err != nil {
			t.Fatal(err)
		}
		if e.WatchStreak == nil || e.WatchStreak.StreakCount != 3 || e.WatchStreak.ChannelPointsAwarded != 350 {
			t.Errorf("watch_streak not decoded: %+v", e.WatchStreak)
		}
		if e.Modiversary == nil || e.Modiversary.Months != 12 {
			t.Errorf("modiversary not decoded: %+v", e.Modiversary)
		}
		if e.IsSourceOnly == nil || !*e.IsSourceOnly {
			t.Errorf("is_source_only not decoded: %v", e.IsSourceOnly)
		}
	})

	t.Run("gueststar_host", func(t *testing.T) {
		var e ChannelGuestStarSessionEndEvent
		if err := json.Unmarshal([]byte(`{"session_id":"s","host_user_id":"7","host_user_name":"Host","host_user_login":"host"}`), &e); err != nil {
			t.Fatal(err)
		}
		if e.HostUserID != "7" || e.HostUserLogin != "host" || e.HostUserName != "Host" {
			t.Errorf("host_user_* not decoded: %+v", e)
		}
	})

	t.Run("chatmessage_source_only", func(t *testing.T) {
		var e ChannelChatMessageEvent
		if err := json.Unmarshal([]byte(`{"chatter_user_id":"1","message_id":"m","is_source_only":false}`), &e); err != nil {
			t.Fatal(err)
		}
		if e.IsSourceOnly == nil || *e.IsSourceOnly {
			t.Errorf("is_source_only not decoded: %v", e.IsSourceOnly)
		}
	})
}

// TestChannelBitsUseEvent_OfficialPayload unmarshals the official channel.bits.use
// event payload and verifies the message object, fragments, and null power-ups.
func TestChannelBitsUseEvent_OfficialPayload(t *testing.T) {
	// Official Twitch example event object for channel.bits.use.
	payload := []byte(`{
		"user_id": "1234",
		"user_login": "cool_user",
		"user_name": "Cool_User",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cooler_user",
		"broadcaster_user_name": "Cooler_User",
		"bits": 2,
		"type": "cheer",
		"power_up": null,
		"custom_power_up": null,
		"message": {
			"text": "cheer1 hi cheer1",
			"fragments": [
				{"type": "cheermote", "text": "cheer1", "cheermote": {"prefix": "cheer", "bits": 1, "tier": 1}, "emote": null},
				{"type": "text", "text": " hi ", "cheermote": null, "emote": null},
				{"type": "cheermote", "text": "cheer1", "cheermote": {"prefix": "cheer", "bits": 1, "tier": 1}, "emote": null}
			]
		}
	}`)

	var event ChannelBitsUseEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if event.UserID != "1234" || event.BroadcasterUserID != "1337" {
		t.Errorf("embedded user/broadcaster not decoded: %+v", event)
	}
	if event.BitsUsed != 2 {
		t.Errorf("BitsUsed = %d, want 2", event.BitsUsed)
	}
	if event.Type != "cheer" {
		t.Errorf("Type = %q, want cheer", event.Type)
	}
	// power_up and custom_power_up are null -> nil.
	if event.PowerUp != nil {
		t.Errorf("PowerUp = %s, want nil", *event.PowerUp)
	}
	if event.CustomPowerUp != nil {
		t.Errorf("CustomPowerUp = %s, want nil", *event.CustomPowerUp)
	}
	if event.Message == nil {
		t.Fatal("Message is nil, want decoded object")
	}
	if event.Message.Text != "cheer1 hi cheer1" {
		t.Errorf("Message.Text = %q", event.Message.Text)
	}
	if len(event.Message.Fragments) != 3 {
		t.Fatalf("Fragments len = %d, want 3", len(event.Message.Fragments))
	}
	frag := event.Message.Fragments[0]
	if frag.Type != "cheermote" || frag.Cheermote == nil {
		t.Fatalf("fragment[0] = %+v, want cheermote", frag)
	}
	if frag.Cheermote.Prefix != "cheer" || frag.Cheermote.Bits != 1 || frag.Cheermote.Tier != 1 {
		t.Errorf("fragment[0].Cheermote = %+v", frag.Cheermote)
	}
}

func TestHypeTrainBeginEvent_V1ToV2Conversion(t *testing.T) {
	// Official Twitch v1 example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelhype_trainbegin
	// Modified is_golden_kappa_train to true for golden kappa test
	v1GoldenKappa := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{ "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
			{ "user_id": "456", "user_login": "kappa", "user_name": "Kappa", "type": "subscription", "total": 45 }
		],
		"last_contribution": { "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
		"level": 2,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_golden_kappa_train": true
	}`

	var event ChannelHypeTrainBeginEvent
	if err := json.Unmarshal([]byte(v1GoldenKappa), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.Type != HypeTrainTypeGoldenKappa {
		t.Errorf("expected Type=golden_kappa, got %s", event.Type)
	}
	if !event.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=true")
	}
	if event.ID != "1b0AsbInCHZW2SQFQkCzqN07Ib2" {
		t.Errorf("expected ID=1b0AsbInCHZW2SQFQkCzqN07Ib2, got %s", event.ID)
	}
	if event.Total != 137 {
		t.Errorf("expected Total=137, got %d", event.Total)
	}
	if len(event.TopContributions) != 2 {
		t.Errorf("expected 2 top contributions, got %d", len(event.TopContributions))
	}

	// Official Twitch v1 example (unmodified - is_golden_kappa_train: false)
	v1Regular := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{ "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
			{ "user_id": "456", "user_login": "kappa", "user_name": "Kappa", "type": "subscription", "total": 45 }
		],
		"last_contribution": { "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
		"level": 2,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_golden_kappa_train": false
	}`

	var event2 ChannelHypeTrainBeginEvent
	if err := json.Unmarshal([]byte(v1Regular), &event2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event2.Type != HypeTrainTypeRegular {
		t.Errorf("expected Type=regular, got %s", event2.Type)
	}
	if event2.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=false")
	}
}

func TestHypeTrainBeginEvent_V2ToV1Conversion(t *testing.T) {
	// Official Twitch v2 example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelhype_trainbegin
	// Modified type to golden_kappa for conversion test
	v2GoldenKappa := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{
				"user_id": "123",
				"user_login": "pogchamp",
				"user_name": "PogChamp",
				"type": "bits",
				"total": 50
			}
		],
		"shared_train_participants": null,
		"level": 1,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_shared_train": false,
		"type": "golden_kappa",
		"all_time_high_level": 4,
		"all_time_high_total": 2845
	}`

	var event ChannelHypeTrainBeginEvent
	if err := json.Unmarshal([]byte(v2GoldenKappa), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.Type != HypeTrainTypeGoldenKappa {
		t.Errorf("expected Type=golden_kappa, got %s", event.Type)
	}
	if !event.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=true (converted from v2)")
	}

	// Official Twitch v2 example (unmodified - type: regular)
	v2Regular := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{
				"user_id": "123",
				"user_login": "pogchamp",
				"user_name": "PogChamp",
				"type": "bits",
				"total": 50
			}
		],
		"shared_train_participants": null,
		"level": 1,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_shared_train": false,
		"type": "regular",
		"all_time_high_level": 4,
		"all_time_high_total": 2845
	}`

	var event2 ChannelHypeTrainBeginEvent
	if err := json.Unmarshal([]byte(v2Regular), &event2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event2.Type != HypeTrainTypeRegular {
		t.Errorf("expected Type=regular, got %s", event2.Type)
	}
	if event2.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=false")
	}
	if event2.AllTimeHighLevel != 4 {
		t.Errorf("expected AllTimeHighLevel=4, got %d", event2.AllTimeHighLevel)
	}
	if event2.AllTimeHighTotal != 2845 {
		t.Errorf("expected AllTimeHighTotal=2845, got %d", event2.AllTimeHighTotal)
	}

	// Test v2 payload with type=shared
	v2Shared := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{
				"user_id": "123",
				"user_login": "pogchamp",
				"user_name": "PogChamp",
				"type": "bits",
				"total": 50
			}
		],
		"shared_train_participants": [
			{"broadcaster_user_id": "111", "broadcaster_user_login": "user1", "broadcaster_user_name": "User1"},
			{"broadcaster_user_id": "222", "broadcaster_user_login": "user2", "broadcaster_user_name": "User2"}
		],
		"level": 1,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_shared_train": true,
		"type": "shared",
		"all_time_high_level": 4,
		"all_time_high_total": 2845
	}`

	var event3 ChannelHypeTrainBeginEvent
	if err := json.Unmarshal([]byte(v2Shared), &event3); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event3.Type != HypeTrainTypeShared {
		t.Errorf("expected Type=shared, got %s", event3.Type)
	}
	if !event3.IsSharedTrain {
		t.Error("expected IsSharedTrain=true")
	}
	if len(event3.SharedTrainParticipants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(event3.SharedTrainParticipants))
	}
	if event3.SharedTrainParticipants[0].BroadcasterID != "111" || event3.SharedTrainParticipants[0].BroadcasterLogin != "user1" {
		t.Errorf("participant broadcaster_user_* not decoded: %+v", event3.SharedTrainParticipants[0])
	}
	if event3.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=false for shared train")
	}
}

func TestHypeTrainEndEvent_V1ToV2Conversion(t *testing.T) {
	// Official Twitch v1 example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelhype_trainend
	// Modified is_golden_kappa_train to true for golden kappa test
	v1GoldenKappa := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"level": 2,
		"total": 137,
		"top_contributions": [
			{ "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
			{ "user_id": "456", "user_login": "kappa", "user_name": "Kappa", "type": "subscription", "total": 45 }
		],
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"ended_at": "2020-07-15T17:16:11.17106713Z",
		"cooldown_ends_at": "2020-07-15T18:16:11.17106713Z",
		"is_golden_kappa_train": true
	}`

	var event ChannelHypeTrainEndEvent
	if err := json.Unmarshal([]byte(v1GoldenKappa), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.Type != HypeTrainTypeGoldenKappa {
		t.Errorf("expected Type=golden_kappa, got %s", event.Type)
	}
	if !event.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=true")
	}
	if event.ID != "1b0AsbInCHZW2SQFQkCzqN07Ib2" {
		t.Errorf("expected ID=1b0AsbInCHZW2SQFQkCzqN07Ib2, got %s", event.ID)
	}
	if event.Total != 137 {
		t.Errorf("expected Total=137, got %d", event.Total)
	}
	if len(event.TopContributions) != 2 {
		t.Errorf("expected 2 top contributions, got %d", len(event.TopContributions))
	}
}

func TestHypeTrainEndEvent_V2ToV1Conversion(t *testing.T) {
	// Official Twitch v2 example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelhype_trainend
	// Modified type to golden_kappa for conversion test
	v2GoldenKappa := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"top_contributions": [
			{
				"user_id": "123",
				"user_login": "pogchamp",
				"user_name": "PogChamp",
				"type": "bits",
				"total": 50
			}
		],
		"shared_train_participants": null,
		"level": 1,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"ended_at": "2020-07-15T17:16:11.17106713Z",
		"cooldown_ends_at": "2020-07-16T17:16:11.17106713Z",
		"is_shared_train": false,
		"type": "golden_kappa"
	}`

	var event ChannelHypeTrainEndEvent
	if err := json.Unmarshal([]byte(v2GoldenKappa), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.Type != HypeTrainTypeGoldenKappa {
		t.Errorf("expected Type=golden_kappa, got %s", event.Type)
	}
	if !event.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=true (converted from v2)")
	}

	// Official Twitch v2 example (unmodified - type: regular)
	v2Regular := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"top_contributions": [
			{
				"user_id": "123",
				"user_login": "pogchamp",
				"user_name": "PogChamp",
				"type": "bits",
				"total": 50
			}
		],
		"shared_train_participants": null,
		"level": 1,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"ended_at": "2020-07-15T17:16:11.17106713Z",
		"cooldown_ends_at": "2020-07-16T17:16:11.17106713Z",
		"is_shared_train": false,
		"type": "regular"
	}`

	var event2 ChannelHypeTrainEndEvent
	if err := json.Unmarshal([]byte(v2Regular), &event2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event2.Type != HypeTrainTypeRegular {
		t.Errorf("expected Type=regular, got %s", event2.Type)
	}
	if event2.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=false")
	}
}

func TestHypeTrainProgressEvent_Conversion(t *testing.T) {
	// Official Twitch v1 example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelhype_trainprogress
	// Modified is_golden_kappa_train to true for conversion test
	v1Payload := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{ "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
			{ "user_id": "456", "user_login": "kappa", "user_name": "Kappa", "type": "subscription", "total": 45 }
		],
		"last_contribution": { "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
		"level": 2,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_golden_kappa_train": true
	}`

	var event ChannelHypeTrainProgressEvent
	if err := json.Unmarshal([]byte(v1Payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.Type != HypeTrainTypeGoldenKappa {
		t.Errorf("expected Type=golden_kappa, got %s", event.Type)
	}
	if event.ID != "1b0AsbInCHZW2SQFQkCzqN07Ib2" {
		t.Errorf("expected ID=1b0AsbInCHZW2SQFQkCzqN07Ib2, got %s", event.ID)
	}
	if event.Total != 137 {
		t.Errorf("expected Total=137, got %d", event.Total)
	}
	if event.Progress != 137 {
		t.Errorf("expected Progress=137, got %d", event.Progress)
	}
	if event.Goal != 500 {
		t.Errorf("expected Goal=500, got %d", event.Goal)
	}

	// Official Twitch v2 example (unmodified - type: regular)
	v2Payload := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"total": 137,
		"progress": 137,
		"goal": 500,
		"top_contributions": [
			{
				"user_id": "123",
				"user_login": "pogchamp",
				"user_name": "PogChamp",
				"type": "bits",
				"total": 50
			}
		],
		"shared_train_participants": null,
		"level": 1,
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"expires_at": "2020-07-15T17:16:11.17106713Z",
		"is_shared_train": false,
		"type": "regular"
	}`

	var event2 ChannelHypeTrainProgressEvent
	if err := json.Unmarshal([]byte(v2Payload), &event2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event2.Type != HypeTrainTypeRegular {
		t.Errorf("expected Type=regular, got %s", event2.Type)
	}
	if event2.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=false")
	}
}

func TestHypeTrainBeginEvent_InvalidJSON(t *testing.T) {
	var event ChannelHypeTrainBeginEvent
	// Use JSON with wrong type for a field to trigger error inside UnmarshalJSON
	err := json.Unmarshal([]byte(`{"level": "not a number"}`), &event)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHypeTrainEndEvent_InvalidJSON(t *testing.T) {
	var event ChannelHypeTrainEndEvent
	// Use JSON with wrong type for a field to trigger error inside UnmarshalJSON
	err := json.Unmarshal([]byte(`{"level": "not a number"}`), &event)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHypeTrainEndEvent_V1RegularTrain(t *testing.T) {
	// Official Twitch v1 example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelhype_trainend
	// (unmodified - is_golden_kappa_train: false)
	v1Regular := `{
		"id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"level": 2,
		"total": 137,
		"top_contributions": [
			{ "user_id": "123", "user_login": "pogchamp", "user_name": "PogChamp", "type": "bits", "total": 50 },
			{ "user_id": "456", "user_login": "kappa", "user_name": "Kappa", "type": "subscription", "total": 45 }
		],
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"ended_at": "2020-07-15T17:16:11.17106713Z",
		"cooldown_ends_at": "2020-07-15T18:16:11.17106713Z",
		"is_golden_kappa_train": false
	}`

	var event ChannelHypeTrainEndEvent
	if err := json.Unmarshal([]byte(v1Regular), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.Type != HypeTrainTypeRegular {
		t.Errorf("expected Type=regular, got %s", event.Type)
	}
	if event.IsGoldenKappaTrain {
		t.Error("expected IsGoldenKappaTrain=false")
	}
	if event.ID != "1b0AsbInCHZW2SQFQkCzqN07Ib2" {
		t.Errorf("expected ID=1b0AsbInCHZW2SQFQkCzqN07Ib2, got %s", event.ID)
	}
	if event.Level != 2 {
		t.Errorf("expected Level=2, got %d", event.Level)
	}
}

// Tests for other EventSub event types using official Twitch API documentation payloads.

func TestChannelFollowEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelfollow
	payload := `{
		"user_id": "1234",
		"user_login": "cool_user",
		"user_name": "Cool_User",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cooler_user",
		"broadcaster_user_name": "Cooler_User",
		"followed_at": "2020-07-15T18:16:11.17106713Z"
	}`

	var event ChannelFollowEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.UserID != "1234" {
		t.Errorf("expected UserID=1234, got %s", event.UserID)
	}
	if event.UserLogin != "cool_user" {
		t.Errorf("expected UserLogin=cool_user, got %s", event.UserLogin)
	}
	if event.BroadcasterUserID != "1337" {
		t.Errorf("expected BroadcasterUserID=1337, got %s", event.BroadcasterUserID)
	}
	if event.FollowedAt.IsZero() {
		t.Error("expected FollowedAt to be set")
	}
}

func TestChannelSubscribeEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelsubscribe
	payload := `{
		"user_id": "1234",
		"user_login": "cool_user",
		"user_name": "Cool_User",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cooler_user",
		"broadcaster_user_name": "Cooler_User",
		"tier": "1000",
		"is_gift": false
	}`

	var event ChannelSubscribeEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.UserID != "1234" {
		t.Errorf("expected UserID=1234, got %s", event.UserID)
	}
	if event.Tier != "1000" {
		t.Errorf("expected Tier=1000, got %s", event.Tier)
	}
	if event.IsGift {
		t.Error("expected IsGift=false")
	}
}

func TestChannelCheerEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelcheer
	payload := `{
		"is_anonymous": false,
		"user_id": "1234",
		"user_login": "cool_user",
		"user_name": "Cool_User",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cooler_user",
		"broadcaster_user_name": "Cooler_User",
		"message": "pogchamp",
		"bits": 1000
	}`

	var event ChannelCheerEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.IsAnonymous {
		t.Error("expected IsAnonymous=false")
	}
	if event.UserID != "1234" {
		t.Errorf("expected UserID=1234, got %s", event.UserID)
	}
	if event.Message != "pogchamp" {
		t.Errorf("expected Message=pogchamp, got %s", event.Message)
	}
	if event.Bits != 1000 {
		t.Errorf("expected Bits=1000, got %d", event.Bits)
	}
}

func TestChannelRaidEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelraid
	payload := `{
		"from_broadcaster_user_id": "1234",
		"from_broadcaster_user_login": "cool_user",
		"from_broadcaster_user_name": "Cool_User",
		"to_broadcaster_user_id": "1337",
		"to_broadcaster_user_login": "cooler_user",
		"to_broadcaster_user_name": "Cooler_User",
		"viewers": 9001
	}`

	var event ChannelRaidEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.FromBroadcasterUserID != "1234" {
		t.Errorf("expected FromBroadcasterUserID=1234, got %s", event.FromBroadcasterUserID)
	}
	if event.ToBroadcasterUserID != "1337" {
		t.Errorf("expected ToBroadcasterUserID=1337, got %s", event.ToBroadcasterUserID)
	}
	if event.Viewers != 9001 {
		t.Errorf("expected Viewers=9001, got %d", event.Viewers)
	}
}

func TestStreamOnlineEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#streamonline
	payload := `{
		"id": "9001",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"type": "live",
		"started_at": "2020-10-11T10:11:12.123Z"
	}`

	var event StreamOnlineEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.ID != "9001" {
		t.Errorf("expected ID=9001, got %s", event.ID)
	}
	if event.BroadcasterUserID != "1337" {
		t.Errorf("expected BroadcasterUserID=1337, got %s", event.BroadcasterUserID)
	}
	if event.Type != "live" {
		t.Errorf("expected Type=live, got %s", event.Type)
	}
	if event.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
}

func TestStreamOfflineEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#streamoffline
	payload := `{
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User"
	}`

	var event StreamOfflineEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.BroadcasterUserID != "1337" {
		t.Errorf("expected BroadcasterUserID=1337, got %s", event.BroadcasterUserID)
	}
	if event.BroadcasterUserLogin != "cool_user" {
		t.Errorf("expected BroadcasterUserLogin=cool_user, got %s", event.BroadcasterUserLogin)
	}
}

func TestChannelUpdateEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelupdate
	payload := `{
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"title": "Best Stream Ever",
		"language": "en",
		"category_id": "21779",
		"category_name": "League of Legends",
		"content_classification_labels": ["MatureGame"]
	}`

	var event ChannelUpdateEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.BroadcasterUserID != "1337" {
		t.Errorf("expected BroadcasterUserID=1337, got %s", event.BroadcasterUserID)
	}
	if event.Title != "Best Stream Ever" {
		t.Errorf("expected Title=Best Stream Ever, got %s", event.Title)
	}
	if event.CategoryID != "21779" {
		t.Errorf("expected CategoryID=21779, got %s", event.CategoryID)
	}
	if event.CategoryName != "League of Legends" {
		t.Errorf("expected CategoryName=League of Legends, got %s", event.CategoryName)
	}
}

func TestChannelBanEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelban
	payload := `{
		"user_id": "1234",
		"user_login": "cool_user",
		"user_name": "Cool_User",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cooler_user",
		"broadcaster_user_name": "Cooler_User",
		"moderator_user_id": "1339",
		"moderator_user_login": "mod_user",
		"moderator_user_name": "Mod_User",
		"reason": "Offensive language",
		"banned_at": "2020-07-15T18:16:11.17106713Z",
		"ends_at": "2020-07-15T18:26:11.17106713Z",
		"is_permanent": false
	}`

	var event ChannelBanEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.UserID != "1234" {
		t.Errorf("expected UserID=1234, got %s", event.UserID)
	}
	if event.ModeratorUserID != "1339" {
		t.Errorf("expected ModeratorUserID=1339, got %s", event.ModeratorUserID)
	}
	if event.Reason != "Offensive language" {
		t.Errorf("expected Reason=Offensive language, got %s", event.Reason)
	}
	if event.IsPermanent {
		t.Error("expected IsPermanent=false")
	}
}

func TestChannelPointsRedemptionAddEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelchannel_points_custom_reward_redemptionadd
	payload := `{
		"id": "17fa2df1-ad76-4804-bfa5-a40ef63efe63",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"user_id": "9001",
		"user_login": "cooler_user",
		"user_name": "Cooler_User",
		"user_input": "pogchamp",
		"status": "unfulfilled",
		"reward": {
			"id": "92af127c-7326-4483-a52b-b0da0be61c01",
			"title": "title",
			"cost": 100,
			"prompt": "reward prompt"
		},
		"redeemed_at": "2020-07-15T17:16:03.17106713Z"
	}`

	var event ChannelPointsRedemptionAddEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.ID != "17fa2df1-ad76-4804-bfa5-a40ef63efe63" {
		t.Errorf("expected ID=17fa2df1-ad76-4804-bfa5-a40ef63efe63, got %s", event.ID)
	}
	if event.UserInput != "pogchamp" {
		t.Errorf("expected UserInput=pogchamp, got %s", event.UserInput)
	}
	if event.Status != "unfulfilled" {
		t.Errorf("expected Status=unfulfilled, got %s", event.Status)
	}
	if event.Reward.Title != "title" {
		t.Errorf("expected Reward.Title=title, got %s", event.Reward.Title)
	}
	if event.Reward.Cost != 100 {
		t.Errorf("expected Reward.Cost=100, got %d", event.Reward.Cost)
	}
}

func TestChannelSubscriptionGiftEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelsubscriptiongift
	payload := `{
		"user_id": "1234",
		"user_login": "cool_user",
		"user_name": "Cool_User",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cooler_user",
		"broadcaster_user_name": "Cooler_User",
		"total": 2,
		"tier": "1000",
		"cumulative_total": 284,
		"is_anonymous": false
	}`

	var event ChannelSubscriptionGiftEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.UserID != "1234" {
		t.Errorf("expected UserID=1234, got %s", event.UserID)
	}
	if event.Total != 2 {
		t.Errorf("expected Total=2, got %d", event.Total)
	}
	if event.Tier != "1000" {
		t.Errorf("expected Tier=1000, got %s", event.Tier)
	}
	if event.CumulativeTotal != 284 {
		t.Errorf("expected CumulativeTotal=284, got %d", event.CumulativeTotal)
	}
}

func TestChannelPollBeginEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelpollbegin
	payload := `{
		"id": "1243456",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"title": "Aren't we great?",
		"choices": [
			{"id": "123", "title": "Yeah!"},
			{"id": "124", "title": "Definitely!"},
			{"id": "125", "title": "100%!"}
		],
		"bits_voting": {
			"is_enabled": true,
			"amount_per_vote": 10
		},
		"channel_points_voting": {
			"is_enabled": true,
			"amount_per_vote": 10
		},
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"ends_at": "2020-07-15T17:16:08.17106713Z"
	}`

	var event ChannelPollBeginEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.ID != "1243456" {
		t.Errorf("expected ID=1243456, got %s", event.ID)
	}
	if event.Title != "Aren't we great?" {
		t.Errorf("expected Title=Aren't we great?, got %s", event.Title)
	}
	if len(event.Choices) != 3 {
		t.Errorf("expected 3 choices, got %d", len(event.Choices))
	}
	if event.BitsVoting.IsEnabled != true {
		t.Error("expected BitsVoting.IsEnabled=true")
	}
}

func TestChannelPredictionBeginEvent(t *testing.T) {
	// Official Twitch example from https://dev.twitch.tv/docs/eventsub/eventsub-subscription-types/#channelpredictionbegin
	payload := `{
		"id": "1243456",
		"broadcaster_user_id": "1337",
		"broadcaster_user_login": "cool_user",
		"broadcaster_user_name": "Cool_User",
		"title": "Will we win?",
		"outcomes": [
			{"id": "123", "title": "Yes", "color": "blue"},
			{"id": "456", "title": "No", "color": "pink"}
		],
		"started_at": "2020-07-15T17:16:03.17106713Z",
		"locks_at": "2020-07-15T17:21:03.17106713Z"
	}`

	var event ChannelPredictionBeginEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if event.ID != "1243456" {
		t.Errorf("expected ID=1243456, got %s", event.ID)
	}
	if event.Title != "Will we win?" {
		t.Errorf("expected Title=Will we win?, got %s", event.Title)
	}
	if len(event.Outcomes) != 2 {
		t.Errorf("expected 2 outcomes, got %d", len(event.Outcomes))
	}
	if event.Outcomes[0].Color != "blue" {
		t.Errorf("expected first outcome color=blue, got %s", event.Outcomes[0].Color)
	}
}

// TestChannelCustomPowerUpRedemptionAddEvent verifies the
// channel.custom_power_up_redemption.add event payload decodes.
func TestChannelCustomPowerUpRedemptionAddEvent(t *testing.T) {
	payload := []byte(`{
		"id": "redemption-1",
		"broadcaster_user_id": "1337", "broadcaster_user_login": "cool", "broadcaster_user_name": "Cool",
		"user_id": "9001", "user_login": "viewer", "user_name": "Viewer",
		"user_input": "go go go",
		"status": "unfulfilled",
		"custom_power_up": {"id": "pu-1", "title": "Hype", "bits": 100, "prompt": "Bring the hype"},
		"redeemed_at": "2026-05-19T12:00:00Z"
	}`)

	var e ChannelCustomPowerUpRedemptionAddEvent
	if err := json.Unmarshal(payload, &e); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if e.ID != "redemption-1" || e.BroadcasterUserID != "1337" || e.UserID != "9001" {
		t.Errorf("core fields not decoded: %+v", e)
	}
	if e.Status != "unfulfilled" || e.UserInput != "go go go" {
		t.Errorf("status/input not decoded: %+v", e)
	}
	if e.CustomPowerUp.ID != "pu-1" || e.CustomPowerUp.Bits != 100 || e.CustomPowerUp.Title != "Hype" {
		t.Errorf("custom_power_up not decoded: %+v", e.CustomPowerUp)
	}
	if e.RedeemedAt.IsZero() {
		t.Error("redeemed_at should be parsed")
	}
}
