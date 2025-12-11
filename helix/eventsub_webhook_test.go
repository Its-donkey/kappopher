package helix

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEventSubWebhookHandler_Verification(t *testing.T) {
	secret := "test-webhook-secret"
	challenge := "test-challenge-string"

	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithVerificationHandler(func(msg *EventSubWebhookMessage) bool {
			receivedMsg = msg
			return true
		}),
	)

	// Create verification payload
	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "webhook_callback_verification_pending",
		},
		Challenge: challenge,
	}
	body, _ := json.Marshal(payload)

	// Create request with proper headers
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	messageID := "msg-123"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)
	req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeStreamOnline)

	// Sign the message
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	// Execute request
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != challenge {
		t.Errorf("expected challenge %s, got %s", challenge, w.Body.String())
	}
	if receivedMsg == nil {
		t.Error("verification handler was not called")
	}
	if receivedMsg.Challenge != challenge {
		t.Errorf("expected challenge in message, got %s", receivedMsg.Challenge)
	}
}

func TestEventSubWebhookHandler_Notification(t *testing.T) {
	secret := "test-webhook-secret"

	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithNotificationHandler(func(msg *EventSubWebhookMessage) {
			receivedMsg = msg
		}),
	)

	// Create notification payload
	event := StreamOnlineEvent{
		ID: "stream123",
		EventSubBroadcaster: EventSubBroadcaster{
			BroadcasterUserID:    "12345",
			BroadcasterUserLogin: "testuser",
			BroadcasterUserName:  "TestUser",
		},
		Type:      "live",
		StartedAt: time.Now(),
	}
	eventJSON, _ := json.Marshal(event)

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "enabled",
		},
		Event: eventJSON,
	}
	body, _ := json.Marshal(payload)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	messageID := "msg-456"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)
	req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeStreamOnline)

	// Sign the message
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	// Execute request
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if receivedMsg == nil {
		t.Fatal("notification handler was not called")
	}
	if receivedMsg.SubscriptionType != EventSubTypeStreamOnline {
		t.Errorf("expected subscription type %s, got %s", EventSubTypeStreamOnline, receivedMsg.SubscriptionType)
	}

	// Parse the event
	parsedEvent, err := ParseEventSubEvent[StreamOnlineEvent](receivedMsg)
	if err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}
	if parsedEvent.ID != "stream123" {
		t.Errorf("expected event ID stream123, got %s", parsedEvent.ID)
	}
	if parsedEvent.BroadcasterUserID != "12345" {
		t.Errorf("expected broadcaster ID 12345, got %s", parsedEvent.BroadcasterUserID)
	}
}

func TestEventSubWebhookHandler_InvalidSignature(t *testing.T) {
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret("correct-secret"),
	)

	body := []byte(`{"subscription":{},"challenge":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)
	req.Header.Set(EventSubHeaderMessageSignature, "sha256=invalid")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_ExpiredTimestamp(t *testing.T) {
	secret := "test-secret"
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithMaxTimestampAge(5*time.Minute),
	)

	body := []byte(`{"subscription":{},"challenge":"test"}`)
	messageID := "msg-123"
	// Timestamp 10 minutes ago
	timestamp := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)

	// Sign with old timestamp
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for expired timestamp, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_Revocation(t *testing.T) {
	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithRevocationHandler(func(msg *EventSubWebhookMessage) {
			receivedMsg = msg
		}),
	)

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "authorization_revoked",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-789")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeRevocation)
	req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeStreamOnline)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if receivedMsg == nil {
		t.Error("revocation handler was not called")
	}
	if receivedMsg.Subscription.Status != "authorization_revoked" {
		t.Errorf("expected status authorization_revoked, got %s", receivedMsg.Subscription.Status)
	}
}

func TestVerifyEventSubSignature(t *testing.T) {
	secret := "my-secret"
	messageID := "msg-123"
	timestamp := "2024-01-01T00:00:00Z"
	body := []byte(`{"test": "data"}`)

	// Create valid signature
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	validSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		signature string
		expected  bool
	}{
		{"valid signature", validSig, true},
		{"invalid signature", "sha256=invalid", false},
		{"empty signature", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyEventSubSignature(secret, messageID, timestamp, body, tt.signature)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMessageDeduplicator(t *testing.T) {
	dedup := NewMessageDeduplicator(time.Minute, 100)

	// First occurrence should not be duplicate
	if dedup.IsDuplicate("msg-1") {
		t.Error("first occurrence should not be duplicate")
	}

	// Second occurrence should be duplicate
	if !dedup.IsDuplicate("msg-1") {
		t.Error("second occurrence should be duplicate")
	}

	// Different message should not be duplicate
	if dedup.IsDuplicate("msg-2") {
		t.Error("different message should not be duplicate")
	}

	// After clear, should not be duplicate
	dedup.Clear()
	if dedup.IsDuplicate("msg-1") {
		t.Error("after clear, should not be duplicate")
	}
}
