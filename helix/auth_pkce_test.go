package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// RFC 7636 Appendix B test vector, which Twitch's PKCE docs also use.
const (
	rfcCodeVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	rfcCodeChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
)

func TestNewPKCEFromVerifier(t *testing.T) {
	pkce, err := NewPKCEFromVerifier(rfcCodeVerifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkce.Verifier != rfcCodeVerifier {
		t.Errorf("expected verifier %s, got %s", rfcCodeVerifier, pkce.Verifier)
	}
	if pkce.Challenge != rfcCodeChallenge {
		t.Errorf("expected challenge %s, got %s", rfcCodeChallenge, pkce.Challenge)
	}
}

func TestNewPKCEFromVerifier_Invalid(t *testing.T) {
	tests := map[string]string{
		"empty":        "",
		"too short":    strings.Repeat("a", 42),
		"too long":     strings.Repeat("a", 129),
		"bad char":     strings.Repeat("a", 42) + "+",
		"space":        strings.Repeat("a", 42) + " ",
		"padding":      strings.Repeat("a", 42) + "=",
		"non-ascii":    strings.Repeat("a", 41) + "é",
		"slash":        strings.Repeat("a", 42) + "/",
		"percent-code": strings.Repeat("a", 40) + "%7E",
	}
	for name, verifier := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPKCEFromVerifier(verifier); err != ErrInvalidCodeVerifier {
				t.Errorf("expected ErrInvalidCodeVerifier, got %v", err)
			}
		})
	}

	for _, verifier := range []string{strings.Repeat("a", 43), strings.Repeat("Z9-._~", 21) + "ab"} {
		if _, err := NewPKCEFromVerifier(verifier); err != nil {
			t.Errorf("verifier %q (len %d): unexpected error: %v", verifier, len(verifier), err)
		}
	}
}

func TestNewPKCE(t *testing.T) {
	a, err := NewPKCE()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := NewPKCE()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Verifier == b.Verifier {
		t.Error("expected a fresh verifier each time")
	}
	if len(a.Verifier) != 43 || len(a.Challenge) != 43 {
		t.Errorf("expected 43-character verifier and challenge, got %d and %d", len(a.Verifier), len(a.Challenge))
	}
	again, err := NewPKCEFromVerifier(a.Verifier)
	if err != nil || again.Challenge != a.Challenge {
		t.Errorf("challenge does not match its verifier: %v", err)
	}
}

func TestAuthClient_GetPKCEAuthURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:    "hof5gwx0su6owfnys0nyan9c87zr6t",
		RedirectURI: "http://localhost:3000",
		Scopes:      []string{ScopeChannelManagePolls, ScopeChannelReadPolls},
		State:       "c3ab8aa609ea11e793ae92361f002671",
		ForceVerify: true,
	})

	authURL, err := client.GetPKCEAuthURL(&PKCE{Verifier: rfcCodeVerifier, Challenge: rfcCodeChallenge})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(authURL, AuthorizeEndpoint+"?") {
		t.Fatalf("expected URL on %s, got %s", AuthorizeEndpoint, authURL)
	}
	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing URL: %v", err)
	}
	q := u.Query()
	want := map[string]string{
		"response_type":         "code",
		"client_id":             "hof5gwx0su6owfnys0nyan9c87zr6t",
		"redirect_uri":          "http://localhost:3000",
		"scope":                 "channel:manage:polls channel:read:polls",
		"state":                 "c3ab8aa609ea11e793ae92361f002671",
		"force_verify":          "true",
		"code_challenge":        rfcCodeChallenge,
		"code_challenge_method": "S256",
	}
	for key, value := range want {
		if got := q.Get(key); got != value {
			t.Errorf("%s: expected %q, got %q", key, value, got)
		}
	}
	if q.Has("code_verifier") {
		t.Error("the verifier must never be on the authorization URL")
	}
}

func TestAuthClient_GetPKCEAuthURL_Errors(t *testing.T) {
	pkce := &PKCE{Verifier: rfcCodeVerifier, Challenge: rfcCodeChallenge}
	tests := []struct {
		name        string
		config      AuthConfig
		pkce        *PKCE
		expectError error
	}{
		{"nil pkce", AuthConfig{ClientID: "client", RedirectURI: "http://localhost"}, nil, ErrMissingCodeChallenge},
		{"empty challenge", AuthConfig{ClientID: "client", RedirectURI: "http://localhost"}, &PKCE{}, ErrMissingCodeChallenge},
		{"missing client ID", AuthConfig{RedirectURI: "http://localhost"}, pkce, ErrMissingClientID},
		{"missing redirect URI", AuthConfig{ClientID: "client"}, pkce, ErrMissingRedirectURI},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuthClient(tt.config).GetPKCEAuthURL(tt.pkce)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}

	// A public client can start this flow.
	public := NewAuthClient(AuthConfig{ClientID: "client", RedirectURI: "http://localhost", ClientType: PublicClient})
	if _, err := public.GetPKCEAuthURL(pkce); err != nil {
		t.Errorf("public client: unexpected error: %v", err)
	}
}

func TestAuthClient_ExchangeCodeWithPKCE_Errors(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		code        string
		verifier    string
		expectError error
	}{
		{
			name:        "missing client ID",
			config:      AuthConfig{ClientSecret: "secret", RedirectURI: "http://localhost"},
			code:        "code",
			verifier:    rfcCodeVerifier,
			expectError: ErrMissingClientID,
		},
		{
			name:        "confidential client missing secret",
			config:      AuthConfig{ClientID: "client", RedirectURI: "http://localhost"},
			code:        "code",
			verifier:    rfcCodeVerifier,
			expectError: ErrMissingClientSecret,
		},
		{
			name:        "missing redirect URI",
			config:      AuthConfig{ClientID: "client", ClientType: PublicClient},
			code:        "code",
			verifier:    rfcCodeVerifier,
			expectError: ErrMissingRedirectURI,
		},
		{
			name:        "missing code",
			config:      AuthConfig{ClientID: "client", ClientType: PublicClient, RedirectURI: "http://localhost"},
			verifier:    rfcCodeVerifier,
			expectError: ErrMissingCode,
		},
		{
			name:        "missing verifier",
			config:      AuthConfig{ClientID: "client", ClientType: PublicClient, RedirectURI: "http://localhost"},
			code:        "code",
			expectError: ErrInvalidCodeVerifier,
		},
		{
			name:        "invalid verifier",
			config:      AuthConfig{ClientID: "client", ClientType: PublicClient, RedirectURI: "http://localhost"},
			code:        "code",
			verifier:    "too-short",
			expectError: ErrInvalidCodeVerifier,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuthClient(tt.config).ExchangeCodeWithPKCE(context.Background(), tt.code, tt.verifier)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

// pkceTokenServer answers a PKCE code exchange with Twitch's documented sample
// response, after checking the form it was sent.
func pkceTokenServer(t *testing.T, wantSecret string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		want := map[string]string{
			"client_id":     "hof5gwx0su6owfnys0yan9c87zr6t",
			"code":          "gulfwdmys5lsm6qyz4xiz9q32l10",
			"code_verifier": rfcCodeVerifier,
			"grant_type":    "authorization_code",
			"redirect_uri":  "http://localhost:3000",
		}
		for key, value := range want {
			if got := r.Form.Get(key); got != value {
				t.Errorf("%s: expected %q, got %q", key, value, got)
			}
		}
		if wantSecret == "" && r.Form.Has("client_secret") {
			t.Error("expected client_secret to be missing")
		}
		if wantSecret != "" && r.Form.Get("client_secret") != wantSecret {
			t.Errorf("client_secret: expected %q, got %q", wantSecret, r.Form.Get("client_secret"))
		}
		_ = json.NewEncoder(w).Encode(Token{
			AccessToken:  "rfx2uswqe8l4g1mkagrvg5tv0ks3",
			RefreshToken: "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01",
			TokenType:    "bearer",
			ExpiresIn:    14124,
			Scope:        []string{"channel:moderate", "chat:edit", "chat:read"},
		})
	}))
}

func TestAuthClient_ExchangeCodeWithPKCE_PublicClient(t *testing.T) {
	server := pkceTokenServer(t, "")
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "hof5gwx0su6owfnys0yan9c87zr6t",
		ClientSecret: "ignored", // a public client never sends it, even if one is set
		ClientType:   PublicClient,
		RedirectURI:  "http://localhost:3000",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	token, err := client.ExchangeCodeWithPKCE(context.Background(), "gulfwdmys5lsm6qyz4xiz9q32l10", rfcCodeVerifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Errorf("expected rfx2uswqe8l4g1mkagrvg5tv0ks3, got %s", token.AccessToken)
	}
	if token.RefreshToken != "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01" {
		t.Errorf("unexpected refresh token %s", token.RefreshToken)
	}
	if token.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}
	if client.GetToken() != token {
		t.Error("expected the client to store the new token")
	}
}

func TestAuthClient_ExchangeCodeWithPKCE_ConfidentialClient(t *testing.T) {
	server := pkceTokenServer(t, "secret")
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "hof5gwx0su6owfnys0yan9c87zr6t",
		ClientSecret: "secret",
		RedirectURI:  "http://localhost:3000",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	if _, err := client.ExchangeCodeWithPKCE(context.Background(), "gulfwdmys5lsm6qyz4xiz9q32l10", rfcCodeVerifier); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthClient_ExchangeCodeWithPKCE_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(AuthErrorResponse{Status: 400, Message: "Invalid authorization code"})
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client", ClientType: PublicClient, RedirectURI: "http://localhost:3000"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	_, err := client.ExchangeCodeWithPKCE(context.Background(), "bad-code", rfcCodeVerifier)
	if err == nil || !strings.Contains(err.Error(), "Invalid authorization code") {
		t.Errorf("expected the Twitch error message, got %v", err)
	}
	if client.GetToken() != nil {
		t.Error("a failed exchange must not store a token")
	}
}
