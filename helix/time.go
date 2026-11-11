package helix

import (
	"bytes"
	"strconv"
	"time"
)

// NullableTime is a time.Time that tolerates the Twitch API's representation of
// an absent timestamp. Optional timestamps such as next_ad_at, ended_at, and
// expires_at arrive as an empty string ("") or null rather than being omitted,
// neither of which a plain time.Time can unmarshal. Some (the ad schedule's
// next_ad_at, last_ad_at and snooze_refresh_at) are Unix timestamps in seconds,
// with 0 meaning unset. NullableTime decodes all of these, the absent forms to
// an invalid (unset) value instead of erroring, and encodes an invalid value
// back to null.
//
// The embedded time.Time promotes the usual accessors (Format, IsZero, Unix,
// ...), so a NullableTime field is a near drop-in for time.Time. Use Valid to
// distinguish a real timestamp from an absent one.
type NullableTime struct {
	time.Time
	Valid bool // Valid is true when Time was populated from a non-empty value.
}

// NewNullableTime returns a valid NullableTime wrapping t.
func NewNullableTime(t time.Time) NullableTime {
	return NullableTime{Time: t, Valid: true}
}

// MarshalJSON encodes an unset value as null and otherwise delegates to
// time.Time's RFC 3339 encoding.
func (t NullableTime) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return t.Time.MarshalJSON()
}

// UnmarshalJSON accepts an RFC 3339 timestamp, a Unix timestamp in seconds, an
// empty string, or null. An empty string, null or 0 yields an unset value
// (Valid == false) rather than an error.
func (t *NullableTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) || bytes.Equal(data, []byte(`""`)) {
		t.Time = time.Time{}
		t.Valid = false
		return nil
	}
	if len(data) > 0 && data[0] != '"' {
		secs, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return err
		}
		if secs == 0 {
			t.Time = time.Time{}
			t.Valid = false
			return nil
		}
		t.Time = time.Unix(secs, 0).UTC()
		t.Valid = true
		return nil
	}
	if err := t.Time.UnmarshalJSON(data); err != nil {
		return err
	}
	t.Valid = true
	return nil
}
