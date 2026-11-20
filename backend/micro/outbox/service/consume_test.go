package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestHandleMessagesKeepsOffsetWhenFailedEventCannotBeSaved(t *testing.T) {
	db := newTestDB(t)
	saveErr := errors.New("mysql unavailable")
	var saveSQL string
	if err := db.Callback().Create().After("gorm:create").Register("test:fail_save", func(tx *gorm.DB) {
		saveSQL = tx.Statement.SQL.String()
		_ = tx.AddError(saveErr)
	}); err != nil {
		t.Fatal(err)
	}

	consumer := NewConsumer(conf.KafkaConfig{}, db, "events", "test")
	var handled []int64
	err := consumer.handleMessages(context.Background(), func(_ context.Context, messages ...kafka.Message) error {
		if len(messages) > 1 {
			return ErrInvalidMessage
		}
		handled = append(handled, messages[0].Offset)
		return ErrInvalidMessage
	}, []kafka.Message{{Topic: "events", Offset: 10}, {Topic: "events", Offset: 11}})
	if !errors.Is(err, saveErr) || !reflect.DeepEqual(handled, []int64{10}) {
		t.Fatalf("handleMessages() = %v, handled offsets = %v", err, handled)
	}
	if !strings.Contains(saveSQL, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("failed event insert must tolerate redelivery: %s", saveSQL)
	}
}

func TestReplayUpdatesOriginalFailedEvent(t *testing.T) {
	for _, tc := range []struct {
		name       string
		handleErr  error
		wantStatus string
	}{
		{name: "success", wantStatus: model.FailedEventStatusResolved},
		{name: "failure", handleErr: errors.New("business still unavailable"), wantStatus: model.FailedEventStatusManualRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t)
			var inserts int
			var finalStatus, finalError string
			var finalSQL string
			if err := db.Callback().Create().After("gorm:create").Register("test:count_insert", func(tx *gorm.DB) {
				inserts++
			}); err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Update().After("gorm:update").Register("test:replay_update", func(tx *gorm.DB) {
				updates := tx.Statement.Dest.(map[string]any)
				if strings.Contains(tx.Statement.SQL.String(), "WHERE id = ?") {
					finalStatus = updates["status"].(string)
					finalError = updates["last_error"].(string)
					finalSQL = tx.Statement.SQL.String()
				}
				tx.RowsAffected = 1
			}); err != nil {
				t.Fatal(err)
			}
			if err := db.Callback().Query().After("gorm:query").Register("test:load_event", func(tx *gorm.DB) {
				*tx.Statement.Dest.(*model.FailedEvent) = model.FailedEvent{
					ID: 12, Topic: "events", Partition: 1, Offset: 23,
					MessageValue: []byte("original payload"), Headers: `[{"Key":"test","Value":"ZGF0YQ=="}]`,
				}
			}); err != nil {
				t.Fatal(err)
			}

			consumer := NewConsumer(conf.KafkaConfig{}, db, "events", "test")
			var handled int
			err := consumer.replayOne(context.Background(), func(_ context.Context, messages ...kafka.Message) error {
				handled++
				msg := messages[0]
				if msg.Topic != "events" || msg.Partition != 1 || msg.Offset != 23 || string(msg.Value) != "original payload" ||
					len(msg.Headers) != 1 || string(msg.Headers[0].Value) != "data" {
					t.Fatalf("replay changed original message: %#v", msg)
				}
				return tc.handleErr
			})
			if err != nil || handled != 1 || inserts != 0 || finalStatus != tc.wantStatus {
				t.Fatalf("replayOne() = %v, handled = %d, inserts = %d, status = %q", err, handled, inserts, finalStatus)
			}
			if !strings.Contains(finalSQL, "lock_owner = ? AND locked_until > NOW()") {
				t.Fatalf("replay result must be written by the current lease owner: %s", finalSQL)
			}
			if tc.handleErr != nil && finalError != tc.handleErr.Error() {
				t.Fatalf("replay error = %q, want %q", finalError, tc.handleErr)
			}
		})
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user@tcp(127.0.0.1:3306)/test",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	return db
}
