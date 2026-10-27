package utils

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// RPCTimeToGoTime *timestamppb.Timestamp 转 *time.Time
func RPCTimeToGoTime(t *timestamppb.Timestamp) time.Time {
	if t != nil {
		return t.AsTime()
	}
	return time.Time{}
}

// GoTimeToRPCTime *time.Time 转 *timestamppb.Timestamp
func GoTimeToRPCTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// FormatTimePtr *time.Time 转 RFC3339 格式的 string
func FormatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return FormatTime(*t)
}

// FormatTime time.Time 转 RFC3339 格式的 string
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
