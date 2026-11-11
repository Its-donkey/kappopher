package helix

import (
	"encoding/json"
	"strings"
	"testing"
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
