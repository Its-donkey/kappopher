package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetCharityCampaign(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/charity/campaigns" {
			t.Errorf("expected /charity/campaigns, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id 12345, got %s", broadcasterID)
		}

		resp := Response[CharityCampaign]{
			Data: []CharityCampaign{
				{
					ID:                 "campaign123",
					BroadcasterID:      "12345",
					BroadcasterLogin:   "testuser",
					BroadcasterName:    "TestUser",
					CharityName:        "Test Charity",
					CharityDescription: "A test charity",
					CharityLogo:        "https://example.com/logo.png",
					CharityWebsite:     "https://example.com",
					CurrentAmount: CharityAmount{
						Value:         1000,
						DecimalPlaces: 2,
						Currency:      "USD",
					},
					TargetAmount: CharityAmount{
						Value:         5000,
						DecimalPlaces: 2,
						Currency:      "USD",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCharityCampaign(context.Background(), &GetCharityCampaignParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(resp.Data))
	}
	if resp.Data[0].CharityName != "Test Charity" {
		t.Errorf("expected charity name 'Test Charity', got %s", resp.Data[0].CharityName)
	}
}

func TestClient_GetCharityCampaign_Pagination(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		first := r.URL.Query().Get("first")
		if first != "1" {
			t.Errorf("expected first=1, got %s", first)
		}
		after := r.URL.Query().Get("after")
		if after != "abc123" {
			t.Errorf("expected after=abc123, got %s", after)
		}

		resp := Response[CharityCampaign]{
			Data: []CharityCampaign{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetCharityCampaign(context.Background(), &GetCharityCampaignParams{
		BroadcasterID:    "12345",
		PaginationParams: &PaginationParams{First: 1, After: "abc123"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetCharityCampaignDonations(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/charity/donations" {
			t.Errorf("expected /charity/donations, got %s", r.URL.Path)
		}

		resp := Response[CharityDonation]{
			Data: []CharityDonation{
				{
					ID:         "donation123",
					CampaignID: "campaign123",
					UserID:     "67890",
					UserLogin:  "donor1",
					UserName:   "Donor1",
					Amount: CharityAmount{
						Value:         100,
						DecimalPlaces: 2,
						Currency:      "USD",
					},
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCharityCampaignDonations(context.Background(), &GetCharityCampaignDonationsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 donation, got %d", len(resp.Data))
	}
	if resp.Data[0].UserLogin != "donor1" {
		t.Errorf("expected user_login 'donor1', got %s", resp.Data[0].UserLogin)
	}
}

func TestClient_GetCharityCampaign_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetCharityCampaign(context.Background(), &GetCharityCampaignParams{
		BroadcasterID: "12345",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetCharityCampaign_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[CharityCampaign]{
			Data: []CharityCampaign{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCharityCampaign(context.Background(), &GetCharityCampaignParams{
		BroadcasterID: "12345",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 campaigns, got %d", len(resp.Data))
	}
}

func TestClient_GetCharityCampaignDonations_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetCharityCampaignDonations(context.Background(), &GetCharityCampaignDonationsParams{
		BroadcasterID: "12345",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
