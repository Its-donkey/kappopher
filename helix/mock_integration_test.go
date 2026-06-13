package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMockAPI_ResubSubPlanDeserialization tests the sub_plan/sub_tier issue
// (twitchdev/issues#1039) using realistic channel.chat.notification payloads
// processed through kappopher's EventSub webhook handler.
func TestMockAPI_ResubSubPlanDeserialization(t *testing.T) {
	// Payload where Twitch sends "sub_plan" instead of "sub_tier" for resub.
	// This is the actual behavior reported in twitchdev/issues#1039.
	subPlanPayload := `{
		"subscription": {
			"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
			"status": "enabled",
			"type": "channel.chat.notification",
			"version": "1",
			"condition": {
				"broadcaster_user_id": "12826",
				"user_id": "141981764"
			},
			"transport": {
				"method": "websocket",
				"session_id": "AQoQILE98gtqShGmLD7AM6yJThAB"
			},
			"created_at": "2023-07-19T14:56:51.634234626Z",
			"cost": 0
		},
		"event": {
			"broadcaster_user_id": "12826",
			"broadcaster_user_login": "twitch",
			"broadcaster_user_name": "Twitch",
			"chatter_user_id": "1337",
			"chatter_user_login": "cool_resub_user",
			"chatter_user_name": "Cool_Resub_User",
			"chatter_is_anonymous": false,
			"color": "#FF4500",
			"badges": [{"set_id": "subscriber", "id": "24", "info": "24"}],
			"system_message": "Cool_Resub_User subscribed at Tier 1. They've subscribed for 24 months!",
			"message_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			"message": {"text": "Love this channel!", "fragments": []},
			"notice_type": "resub",
			"resub": {
				"cumulative_months": 24,
				"duration_months": 1,
				"streak_months": 24,
				"sub_plan": "1000",
				"is_prime": false,
				"is_gift": false,
				"gifter_is_anonymous": false
			}
		}
	}`

	// Payload using the documented "sub_tier" field name.
	subTierPayload := `{
		"subscription": {
			"id": "a2b3c487-261b-59fa-b265-1f31e8a5f2b5",
			"status": "enabled",
			"type": "channel.chat.notification",
			"version": "1",
			"condition": {
				"broadcaster_user_id": "12826",
				"user_id": "141981764"
			},
			"transport": {
				"method": "websocket",
				"session_id": "AQoQILE98gtqShGmLD7AM6yJThAB"
			},
			"created_at": "2023-07-19T14:56:51.634234626Z",
			"cost": 0
		},
		"event": {
			"broadcaster_user_id": "12826",
			"broadcaster_user_login": "twitch",
			"broadcaster_user_name": "Twitch",
			"chatter_user_id": "9001",
			"chatter_user_login": "tier3_user",
			"chatter_user_name": "Tier3_User",
			"chatter_is_anonymous": false,
			"color": "#00FF00",
			"badges": [{"set_id": "subscriber", "id": "36", "info": "36"}],
			"system_message": "Tier3_User subscribed at Tier 3. They've subscribed for 36 months!",
			"message_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
			"message": {"text": "Three years strong!", "fragments": []},
			"notice_type": "resub",
			"resub": {
				"cumulative_months": 36,
				"duration_months": 1,
				"streak_months": 36,
				"sub_tier": "3000",
				"is_prime": false,
				"is_gift": false,
				"gifter_is_anonymous": false
			}
		}
	}`

	// Payload for shared_chat_resub using "sub_plan" (also affected per twitchdev/issues#1039).
	sharedChatResubPayload := `{
		"subscription": {
			"id": "c3d4e587-362c-69gb-c376-2f42f9b6g3c6",
			"status": "enabled",
			"type": "channel.chat.notification",
			"version": "1",
			"condition": {
				"broadcaster_user_id": "12826",
				"user_id": "141981764"
			},
			"transport": {
				"method": "websocket",
				"session_id": "AQoQILE98gtqShGmLD7AM6yJThAB"
			},
			"created_at": "2023-07-19T14:56:51.634234626Z",
			"cost": 0
		},
		"event": {
			"broadcaster_user_id": "12826",
			"broadcaster_user_login": "twitch",
			"broadcaster_user_name": "Twitch",
			"chatter_user_id": "5555",
			"chatter_user_login": "shared_chat_resub_user",
			"chatter_user_name": "Shared_Chat_Resub_User",
			"chatter_is_anonymous": false,
			"color": "#9146FF",
			"badges": [{"set_id": "subscriber", "id": "12", "info": "12"}],
			"system_message": "Shared_Chat_Resub_User subscribed at Tier 2.",
			"message_id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
			"message": {"text": "Shared chat resub!", "fragments": []},
			"notice_type": "shared_chat_resub",
			"shared_chat_resub": {
				"cumulative_months": 12,
				"duration_months": 1,
				"streak_months": 12,
				"sub_plan": "2000",
				"is_prime": false,
				"is_gift": false,
				"gifter_is_anonymous": false
			},
			"source_broadcaster_user_id": "99999",
			"source_broadcaster_user_login": "source_channel",
			"source_broadcaster_user_name": "Source_Channel"
		}
	}`

	tests := []struct {
		name        string
		payload     string
		noticeType  string
		wantSubTier string
		wantMonths  int
		checkShared bool
	}{
		{
			name:        "resub with sub_plan (twitchdev/issues#1039)",
			payload:     subPlanPayload,
			noticeType:  "resub",
			wantSubTier: "1000",
			wantMonths:  24,
		},
		{
			name:        "resub with sub_tier (documented field name)",
			payload:     subTierPayload,
			noticeType:  "resub",
			wantSubTier: "3000",
			wantMonths:  36,
		},
		{
			name:        "shared_chat_resub with sub_plan (twitchdev/issues#1039)",
			payload:     sharedChatResubPayload,
			noticeType:  "shared_chat_resub",
			wantSubTier: "2000",
			wantMonths:  12,
			checkShared: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the full EventSub notification envelope.
			var envelope struct {
				Subscription EventSubSubscription `json:"subscription"`
				Event        json.RawMessage      `json:"event"`
			}
			if err := json.Unmarshal([]byte(tt.payload), &envelope); err != nil {
				t.Fatalf("failed to unmarshal envelope: %v", err)
			}

			if envelope.Subscription.Type != EventSubTypeChannelChatNotification {
				t.Fatalf("expected type=%s, got %s", EventSubTypeChannelChatNotification, envelope.Subscription.Type)
			}

			// Parse the event into the notification struct.
			var event ChannelChatNotificationEvent
			if err := json.Unmarshal(envelope.Event, &event); err != nil {
				t.Fatalf("failed to unmarshal event: %v", err)
			}

			if event.NoticeType != tt.noticeType {
				t.Errorf("expected notice_type=%s, got %s", tt.noticeType, event.NoticeType)
			}

			// Check the resub or shared_chat_resub field.
			var resub *ChatNotificationResub
			if tt.checkShared {
				resub = event.SharedChatResub
				if resub == nil {
					t.Fatal("expected shared_chat_resub to be non-nil")
				}
			} else {
				resub = event.Resub
				if resub == nil {
					t.Fatal("expected resub to be non-nil")
				}
			}

			if resub.SubTier != tt.wantSubTier {
				t.Errorf("SubTier: got %q, want %q (sub_plan/sub_tier compatibility)", resub.SubTier, tt.wantSubTier)
			}
			if resub.CumulativeMonths != tt.wantMonths {
				t.Errorf("CumulativeMonths: got %d, want %d", resub.CumulativeMonths, tt.wantMonths)
			}
		})
	}
}

// TestMockAPI_ClientAgainstMockServer tests that kappopher's Client can
// communicate with a Twitch-like mock API server.
func TestMockAPI_ClientAgainstMockServer(t *testing.T) {
	// Spin up a local mock that mimics Twitch Helix responses.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify kappopher sends the right headers.
		clientID := r.Header.Get("Client-Id")
		auth := r.Header.Get("Authorization")
		if clientID == "" {
			t.Error("missing Client-Id header")
		}
		if auth == "" {
			t.Error("missing Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/users":
			_, _ = w.Write([]byte(`{"data":[{"id":"11138301","login":"testuser","display_name":"TestUser","type":"","broadcaster_type":"partner","description":"","created_at":"2026-02-18T07:00:03Z","profile_image_url":"","offline_image_url":"","view_count":0}]}`))
		case "/streams":
			_, _ = w.Write([]byte(`{"data":[{"id":"123456","user_id":"11138301","user_login":"testuser","user_name":"TestUser","game_id":"12345","game_name":"TestGame","type":"live","title":"Test Stream","viewer_count":100,"started_at":"2026-02-18T07:00:03Z","language":"en","thumbnail_url":"","is_mature":false}],"pagination":{}}`))
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":"Not Found","status":404,"message":"endpoint not found"}`))
		}
	}))
	defer mockServer.Close()

	auth := NewAuthClient(AuthConfig{ClientID: "mock-client-id"})
	auth.SetToken(&Token{AccessToken: "mock-token"})
	client := NewClient("mock-client-id", auth)
	client.baseURL = mockServer.URL

	ctx := context.Background()

	// Test GetUsers.
	users, err := client.GetUsers(ctx, &GetUsersParams{IDs: []string{"11138301"}})
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}
	if len(users.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users.Data))
	}
	if users.Data[0].Login != "testuser" {
		t.Errorf("expected login=testuser, got %s", users.Data[0].Login)
	}

	// Test GetStreams.
	streams, err := client.GetStreams(ctx, &GetStreamsParams{UserIDs: []string{"11138301"}})
	if err != nil {
		t.Fatalf("GetStreams failed: %v", err)
	}
	if len(streams.Data) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(streams.Data))
	}
	if streams.Data[0].ViewerCount != 100 {
		t.Errorf("expected viewer_count=100, got %d", streams.Data[0].ViewerCount)
	}

	t.Log("kappopher Client successfully communicates with mock API")
}
