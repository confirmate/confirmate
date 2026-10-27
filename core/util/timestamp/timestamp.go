package timestamp

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Now returns a new timestamp for the current time, truncating to microsecond precision.
func Now() *timestamppb.Timestamp {
	return New(time.Now().Truncate(time.Microsecond))
}

// New returns a new timestamp from the given time, truncating to microsecond precision.
func New(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t.Truncate(time.Microsecond))
}
