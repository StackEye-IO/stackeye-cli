// Package testutil provides test fixtures and mock data for StackEye CLI tests.
// This file contains helper functions for creating test data.
package testutil

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// StringPtr returns a pointer to the given string.
// Useful for optional string fields in API requests/responses.
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int.
// Useful for optional int fields in API requests/responses.
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to the given int64.
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr returns a pointer to the given bool.
// Useful for optional bool fields in API requests/responses.
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to the given time.Time.
// Useful for optional timestamp fields.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// UUIDPtr returns a pointer to the given UUID.
func UUIDPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

// ParseTime parses a time string in RFC3339 format.
// Panics if the time string is invalid - use only in tests.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic("invalid time format: " + err.Error())
	}
	return t
}

// ParseUUID parses a UUID string.
// Panics if the UUID is invalid - use only in tests.
func ParseUUID(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		panic("invalid UUID: " + err.Error())
	}
	return u
}

// MustMarshalJSON marshals v to JSON bytes.
// Panics if marshaling fails - use only in tests.
func MustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic("failed to marshal JSON: " + err.Error())
	}
	return data
}

// MustMarshalJSONString marshals v to a JSON string.
// Panics if marshaling fails - use only in tests.
func MustMarshalJSONString(v interface{}) string {
	return string(MustMarshalJSON(v))
}

// MustMarshalJSONRaw marshals v to json.RawMessage.
// Panics if marshaling fails - use only in tests.
func MustMarshalJSONRaw(v interface{}) json.RawMessage {
	return json.RawMessage(MustMarshalJSON(v))
}

// NewRandomUUID generates a new random UUID.
// This is a convenience wrapper around uuid.New().
func NewRandomUUID() uuid.UUID {
	return uuid.New()
}

// RelativeTime returns a time relative to BaseTime.
// Positive durations are in the future, negative are in the past.
func RelativeTime(d time.Duration) time.Time {
	return BaseTime.Add(d)
}

// DaysAgo returns a time that is n days before BaseTime.
func DaysAgo(n int) time.Time {
	return BaseTime.Add(time.Duration(-n) * 24 * time.Hour)
}

// HoursAgo returns a time that is n hours before BaseTime.
func HoursAgo(n int) time.Time {
	return BaseTime.Add(time.Duration(-n) * time.Hour)
}

// MinutesAgo returns a time that is n minutes before BaseTime.
func MinutesAgo(n int) time.Time {
	return BaseTime.Add(time.Duration(-n) * time.Minute)
}

// DaysFromNow returns a time that is n days after BaseTime.
func DaysFromNow(n int) time.Time {
	return BaseTime.Add(time.Duration(n) * 24 * time.Hour)
}

// HoursFromNow returns a time that is n hours after BaseTime.
func HoursFromNow(n int) time.Time {
	return BaseTime.Add(time.Duration(n) * time.Hour)
}
