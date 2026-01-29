package utils

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func RPCTimeToGoTime(t *timestamppb.Timestamp) time.Time {
	if t != nil {
		return t.AsTime()
	}
	return time.Time{}
}

func GoTimeToRPCTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}
