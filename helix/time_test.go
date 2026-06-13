package helix

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNullableTime_Unmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
		wantTime  time.Time
		wantErr   bool
	}{
		{name: "RFC3339", input: `"2022-03-15T02:00:28Z"`, wantValid: true, wantTime: time.Date(2022, 3, 15, 2, 0, 28, 0, time.UTC)},
		{name: "empty string", input: `""`, wantValid: false},
		{name: "null", input: `null`, wantValid: false},
		{name: "invalid", input: `"not-a-time"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nt NullableTime
			err := json.Unmarshal([]byte(tt.input), &nt)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if nt.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", nt.Valid, tt.wantValid)
			}
			if tt.wantValid && !nt.Equal(tt.wantTime) {
				t.Errorf("Time = %v, want %v", nt.Time, tt.wantTime)
			}
			if !tt.wantValid && !nt.IsZero() {
				t.Errorf("Time = %v, want zero", nt.Time)
			}
		})
	}
}

func TestNullableTime_Marshal(t *testing.T) {
	tests := []struct {
		name  string
		value NullableTime
		want  string
	}{
		{name: "valid", value: NewNullableTime(time.Date(2022, 3, 15, 2, 0, 28, 0, time.UTC)), want: `"2022-03-15T02:00:28Z"`},
		{name: "invalid", value: NullableTime{}, want: `null`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(b) != tt.want {
				t.Errorf("Marshal = %s, want %s", b, tt.want)
			}
		})
	}
}

// TestNullableTime_EmptyStringInStruct verifies the field decodes Twitch's
// empty-string timestamp without failing the whole unmarshal.
func TestNullableTime_EmptyStringInStruct(t *testing.T) {
	var p Poll
	if err := json.Unmarshal([]byte(`{"id":"123","status":"ACTIVE","ended_at":""}`), &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.EndedAt.Valid {
		t.Errorf("EndedAt.Valid = true, want false for empty string")
	}
}
