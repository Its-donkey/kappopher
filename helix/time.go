package helix

import (
	"bytes"
	"time"
)

// NullableTime is a time.Time that tolerates the Twitch API's representation of
// an absent timestamp. Optional timestamps such as next_ad_at, ended_at, and
// expires_at arrive as an empty string ("") or null rather than being omitted,
// neither of which a plain time.Time can unmarshal. NullableTime decodes those
// to an invalid (unset) value instead of erroring, and encodes an invalid value
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

// UnmarshalJSON accepts an RFC 3339 timestamp, an empty string, or null. The
// latter two yield an unset value (Valid == false) rather than an error.
func (t *NullableTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) || bytes.Equal(data, []byte(`""`)) {
		t.Time = time.Time{}
		t.Valid = false
		return nil
	}
	if err := t.Time.UnmarshalJSON(data); err != nil {
		return err
	}
	t.Valid = true
	return nil
}
