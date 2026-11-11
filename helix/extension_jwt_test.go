package helix

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewExtensionJWT(t *testing.T) {
	// Create a test secret (base64 encoded)
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))

	jwt, err := NewExtensionJWT("ext123", secret, "owner456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if jwt.ExtensionID() != "ext123" {
		t.Errorf("expected extension ID ext123, got %s", jwt.ExtensionID())
	}
	if jwt.OwnerID() != "owner456" {
		t.Errorf("expected owner ID owner456, got %s", jwt.OwnerID())
	}
}

func TestNewExtensionJWT_InvalidSecret(t *testing.T) {
	_, err := NewExtensionJWT("ext123", "not-base64!", "owner456")
	if err == nil {
		t.Error("expected error for invalid base64 secret")
	}
}

func TestExtensionJWT_CreateToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	claims := &ExtensionJWTClaims{
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: "owner456",
		Role:   ExtensionRoleExternal,
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 token parts, got %d", len(parts))
	}

	// Verify we can parse the token back
	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsed.UserID != "owner456" {
		t.Errorf("expected user ID owner456, got %s", parsed.UserID)
	}
	if parsed.Role != ExtensionRoleExternal {
		t.Errorf("expected role external, got %s", parsed.Role)
	}
}

func TestExtensionJWT_CreateEBSToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	token, err := jwt.CreateEBSToken(time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsed.Role != ExtensionRoleExternal {
		t.Errorf("expected role external, got %s", parsed.Role)
	}
	if parsed.UserID != "owner456" {
		t.Errorf("expected user ID owner456, got %s", parsed.UserID)
	}
}

func TestExtensionJWT_CreateBroadcasterToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	token, err := jwt.CreateBroadcasterToken("channel789", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsed.Role != ExtensionRoleBroadcaster {
		t.Errorf("expected role broadcaster, got %s", parsed.Role)
	}
	if parsed.ChannelID != "channel789" {
		t.Errorf("expected channel ID channel789, got %s", parsed.ChannelID)
	}
}

func TestExtensionJWT_CreatePubSubToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	listen := []string{"broadcast", "global"}
	send := []string{"broadcast"}

	token, err := jwt.CreatePubSubToken("channel789", listen, send, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if len(parsed.PubsubPermsListen) != 2 {
		t.Errorf("expected 2 listen permissions, got %d", len(parsed.PubsubPermsListen))
	}
	if len(parsed.PubsubPermsSend) != 1 {
		t.Errorf("expected 1 send permission, got %d", len(parsed.PubsubPermsSend))
	}
}

func TestParseExtensionJWT_InvalidFormat(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))

	tests := []struct {
		name  string
		token string
	}{
		{"no dots", "nodots"},
		{"one dot", "one.dot"},
		{"empty parts", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExtensionJWT(tt.token, secret)
			if err == nil {
				t.Error("expected error for invalid token format")
			}
		})
	}
}

func TestParseExtensionJWT_InvalidSignature(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("correct-secret"))
	wrongSecret := base64.StdEncoding.EncodeToString([]byte("wrong-secret"))

	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")
	token, _ := jwt.CreateEBSToken(time.Hour)

	// Try to parse with wrong secret
	_, err := ParseExtensionJWT(token, wrongSecret)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestParseExtensionJWT_Expired(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	// Create a token that's already expired
	claims := &ExtensionJWTClaims{
		Exp:    time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		UserID: "owner456",
		Role:   ExtensionRoleExternal,
	}

	token, _ := jwt.CreateToken(claims)

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expiration error, got: %v", err)
	}
}

func TestExtensionTokenProvider_GetToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// First call should generate a new token
	token1 := provider.GetToken()
	if token1 == nil {
		t.Fatal("expected token, got nil")
	}
	if token1.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	// Second call should return the same token (cached)
	token2 := provider.GetToken()
	if token2.AccessToken != token1.AccessToken {
		t.Error("expected same cached token")
	}
}

func TestExtensionTokenProvider_GetToken_Concurrent(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// Run concurrent GetToken calls to test thread safety
	const goroutines = 50
	done := make(chan *Token, goroutines)

	for range goroutines {
		go func() {
			token := provider.GetToken()
			done <- token
		}()
	}

	// Collect all tokens
	var tokens []*Token
	for range goroutines {
		token := <-done
		if token == nil {
			t.Error("expected token, got nil")
		}
		tokens = append(tokens, token)
	}

	// All tokens should be valid
	for _, token := range tokens {
		if token.AccessToken == "" {
			t.Error("expected non-empty access token")
		}
	}
}

func TestExtensionJWT_CreateToken_DefaultExpiration(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	// Create claims with Exp=0 (should use default 1 hour expiration)
	claims := &ExtensionJWTClaims{
		UserID: "user123",
		Role:   ExtensionRoleExternal,
		// Exp is 0 - should be set to default
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify that Exp was set
	if claims.Exp == 0 {
		t.Error("expected Exp to be set to default value")
	}
}

func TestExtensionTokenProvider_GetToken_DoubleCheck(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// Pre-populate with a valid token
	provider.mu.Lock()
	provider.token = &Token{
		AccessToken: "pre-existing-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	provider.mu.Unlock()

	// GetToken should return the cached token (hitting the fast path)
	token := provider.GetToken()
	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.AccessToken != "pre-existing-token" {
		t.Error("expected cached token to be returned")
	}
}

func TestExtensionTokenProvider_GetToken_ExpiredThenRefreshed(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// Pre-populate with an expired token
	provider.mu.Lock()
	provider.token = &Token{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-time.Hour), // Already expired
	}
	provider.mu.Unlock()

	// GetToken should generate a new token
	token := provider.GetToken()
	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.AccessToken == "expired-token" {
		t.Error("expected new token, got expired token")
	}
}

func TestNewExtensionClient(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	client := NewExtensionClient("client-id", jwt)
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.tokenProvider == nil {
		t.Error("expected token provider to be set")
	}
}

func TestWithExtensionJWT(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("client-id", authClient, WithExtensionJWT(jwt))

	if client.tokenProvider == nil {
		t.Error("expected token provider to be set")
	}
	if client.authClient != nil {
		t.Error("expected authClient to be nil when using extension JWT")
	}
}

func TestClient_SetExtensionJWT(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("client-id", authClient)

	// Initially should have authClient set
	if client.authClient == nil {
		t.Error("expected authClient to be set initially")
	}

	// After setting extension JWT, authClient should be nil
	client.SetExtensionJWT(jwt)
	if client.authClient != nil {
		t.Error("expected authClient to be nil after SetExtensionJWT")
	}
	if client.tokenProvider == nil {
		t.Error("expected tokenProvider to be set after SetExtensionJWT")
	}
}

func TestJWTClaimsStructure(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	claims := &ExtensionJWTClaims{
		Exp:               time.Now().Add(time.Hour).Unix(),
		UserID:            "user123",
		Role:              ExtensionRoleBroadcaster,
		ChannelID:         "channel456",
		OpaqueUserID:      "U123",
		IsUnlinked:        true,
		PubsubPermsListen: []string{"broadcast"},
		PubsubPermsSend:   []string{"broadcast", "whisper-user123"},
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode the claims part manually to verify structure
	parts := strings.Split(token, ".")
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])

	var decoded map[string]any
	if err := json.Unmarshal(claimsJSON, &decoded); err != nil {
		t.Fatalf("failed to decode claims: %v", err)
	}

	// Verify key fields are present
	if decoded["user_id"] != "user123" {
		t.Errorf("expected user_id user123, got %v", decoded["user_id"])
	}
	if decoded["role"] != string(ExtensionRoleBroadcaster) {
		t.Errorf("expected role broadcaster, got %v", decoded["role"])
	}
	if decoded["channel_id"] != "channel456" {
		t.Errorf("expected channel_id channel456, got %v", decoded["channel_id"])
	}
	if decoded["is_unlinked"] != true {
		t.Errorf("expected is_unlinked true, got %v", decoded["is_unlinked"])
	}
}

func TestParseExtensionJWT_InvalidClaimsBase64(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))

	// Create a token with invalid base64 in the claims part
	// Use valid header but invalid claims that will fail base64 decoding
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	invalidClaims := "!!!invalid-base64!!!"
	sig := base64.RawURLEncoding.EncodeToString([]byte("fake-sig"))

	token := header + "." + invalidClaims + "." + sig

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for invalid claims base64")
	}
}

func TestParseExtensionJWT_InvalidClaimsJSON(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))

	// Create a token with valid base64 but invalid JSON in claims
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	invalidJSON := base64.RawURLEncoding.EncodeToString([]byte(`{not valid json`))

	// Create a valid signature for these parts
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")
	message := header + "." + invalidJSON
	sig := base64.RawURLEncoding.EncodeToString(jwt.sign([]byte(message)))

	token := message + "." + sig

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for invalid claims JSON")
	}
	if !strings.Contains(err.Error(), "parsing claims") {
		t.Errorf("expected 'parsing claims' error, got: %v", err)
	}
}

func TestParseExtensionJWT_InvalidSecretBase64(t *testing.T) {
	_, err := ParseExtensionJWT("header.claims.sig", "!!!invalid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid secret base64")
	}
	if !strings.Contains(err.Error(), "decoding secret") {
		t.Errorf("expected 'decoding secret' error, got: %v", err)
	}
}

func TestParseExtensionJWT_InvalidSignatureBase64(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))

	// Create a token with invalid base64 in the signature part
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":9999999999,"user_id":"test","role":"external"}`))
	invalidSig := "!!!invalid-base64!!!"

	token := header + "." + claims + "." + invalidSig

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for invalid signature base64")
	}
	if !strings.Contains(err.Error(), "decoding signature") {
		t.Errorf("expected 'decoding signature' error, got: %v", err)
	}
}

// claimsOf decodes a token's claims segment without verifying it.
func claimsOf(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts", len(parts))
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decoding claims: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parsing claims: %v", err)
	}
	return m
}

func TestExtensionJWT_PubSubPermsAreNested(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	token, err := jwt.CreatePubSubToken("channel789", nil, []string{ExtensionPubSubTargetBroadcast}, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := claimsOf(t, token)
	perms, ok := m["pubsub_perms"].(map[string]any)
	if !ok {
		t.Fatalf("expected a pubsub_perms object, got claims %v", m)
	}
	send, _ := perms["send"].([]any)
	if len(send) != 1 || send[0] != "broadcast" {
		t.Errorf("unexpected pubsub_perms.send %v", perms["send"])
	}
	if _, ok := perms["listen"]; ok {
		t.Errorf("empty listen should be omitted, got %v", perms["listen"])
	}
	if _, ok := m["pubsub_perms_send"]; ok {
		t.Error("flat pubsub_perms_send must not be signed")
	}
	if m["role"] != "external" || m["user_id"] != "owner456" || m["channel_id"] != "channel789" {
		t.Errorf("unexpected claims %v", m)
	}
}

func TestExtensionJWT_DeprecatedFlatPermsAreSignedNested(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	claims := &ExtensionJWTClaims{
		UserID:            "owner456",
		Role:              ExtensionRoleExternal,
		ChannelID:         ExtensionChannelAll,
		PubsubPermsListen: []string{"broadcast"},
		PubsubPermsSend:   []string{ExtensionPubSubTargetGlobal},
	}
	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.PubsubPerms != nil {
		t.Error("CreateToken must not modify the caller's PubsubPerms")
	}

	perms, ok := claimsOf(t, token)["pubsub_perms"].(map[string]any)
	if !ok {
		t.Fatal("expected flat perms to be signed as a pubsub_perms object")
	}
	if send, _ := perms["send"].([]any); len(send) != 1 || send[0] != "global" {
		t.Errorf("unexpected pubsub_perms.send %v", perms["send"])
	}
}

func TestExtensionJWT_ParseViewerToken(t *testing.T) {
	// A viewer token as the Twitch extension frontend sends it to the EBS.
	secretBytes := []byte("test-secret-key-1234567890")
	secret := base64.StdEncoding.EncodeToString(secretBytes)
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	payload := fmt.Sprintf(`{"exp":%d,"opaque_user_id":"UG12X345T6J78","user_id":"123454321","channel_id":"98765","role":"viewer","is_unlinked":false,"pubsub_perms":{"listen":["broadcast","whisper-UG12X345T6J78","global"],"send":["broadcast","whisper-*"]}}`,
		time.Now().Add(time.Minute).Unix())
	message := base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + base64URLEncode([]byte(payload))
	token := message + "." + base64URLEncode(jwt.sign([]byte(message)))

	claims, err := jwt.ParseAuthorizationHeader("Bearer " + token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Role != ExtensionRoleViewer || claims.ChannelID != "98765" || claims.OpaqueUserID != "UG12X345T6J78" {
		t.Errorf("unexpected claims %+v", claims)
	}
	if claims.PubsubPerms == nil || len(claims.PubsubPerms.Listen) != 3 || claims.PubsubPerms.Send[1] != "whisper-*" {
		t.Fatalf("unexpected pubsub_perms %+v", claims.PubsubPerms)
	}
	if claims.PubsubPerms.Listen[1] != ExtensionPubSubWhisperTarget("UG12X345T6J78") {
		t.Errorf("unexpected whisper target %q", claims.PubsubPerms.Listen[1])
	}
	if len(claims.PubsubPermsListen) != 3 || len(claims.PubsubPermsSend) != 2 {
		t.Error("deprecated flat fields should be filled from pubsub_perms")
	}

	other, _ := NewExtensionJWT("ext123", base64.StdEncoding.EncodeToString([]byte("another-secret")), "owner456")
	if _, err := other.ParseToken(token); err == nil {
		t.Error("expected a token signed with another secret to be refused")
	}
}

func TestExtensionJWT_ParseAuthorizationHeader_Invalid(t *testing.T) {
	jwt, _ := NewExtensionJWT("ext123", base64.StdEncoding.EncodeToString([]byte("secret")), "owner456")
	for _, h := range []string{"", "Bearer", "Bearer   ", "Basic abc.def.ghi", "abc.def.ghi"} {
		if _, err := jwt.ParseAuthorizationHeader(h); err == nil {
			t.Errorf("expected an error for header %q", h)
		}
	}
}

func TestExtensionJWT_CreateExternalToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	token, err := jwt.CreateExternalToken("channel789", nil, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := claimsOf(t, token)
	if m["role"] != "external" || m["channel_id"] != "channel789" {
		t.Errorf("unexpected claims %v", m)
	}
	if _, ok := m["pubsub_perms"]; ok {
		t.Error("nil perms should be omitted")
	}
	exp, _ := m["exp"].(float64)
	if d := time.Until(time.Unix(int64(exp), 0)); d <= 0 || d > time.Minute {
		t.Errorf("exp should be seconds within a minute, got %v", d)
	}
}
