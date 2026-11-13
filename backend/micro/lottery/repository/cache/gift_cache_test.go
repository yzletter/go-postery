package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

type lotteryRecordClient struct {
	redis.UniversalClient
	keys       []string
	nextCursor uint64
	scanErr    error
	hashes     map[string]map[string]string
	hashErrors map[string]error
	cursor     uint64
	match      string
	count      int64
}

func (client *lotteryRecordClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	client.cursor, client.match, client.count = cursor, match, count
	return redis.NewScanCmdResult(client.keys, client.nextCursor, client.scanErr)
}

func (client *lotteryRecordClient) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return redis.NewMapStringStringResult(client.hashes[key], client.hashErrors[key])
}

func TestScanLotteryRecords(t *testing.T) {
	client := &lotteryRecordClient{
		keys: []string{
			"lottery:record:9007199254740993",
			"lottery:record:102",
			"lottery:record:103",
			"lottery:record:104",
			"lottery:record:105",
			"lottery:record:106",
			"lottery:record:107",
		},
		nextCursor: 37,
		hashes: map[string]map[string]string{
			"lottery:record:9007199254740993": {"request_id": "9007199254740995", "user_id": "9007199254740993", "gift_id": "170406720000000001"},
			"lottery:record:103":              {"request_id": "abc", "user_id": "103", "gift_id": "1"},
			"lottery:record:104":              {"request_id": "1", "user_id": "104", "gift_id": "0"},
			"lottery:record:105":              {"request_id": "1", "user_id": "106", "gift_id": "1"},
			"lottery:record:107":              {"request_id": "2", "user_id": "107", "gift_id": "3"},
		},
		hashErrors: map[string]error{"lottery:record:106": errors.New("read failed")},
	}
	cache := NewGiftCache(client)
	records, cursor, err := cache.ScanLotteryRecords(context.Background(), 12, 100)
	if err != nil {
		t.Fatal(err)
	}
	if cursor != 37 || client.cursor != 12 || client.match != "lottery:record:*" || client.count != 100 {
		t.Fatalf("unexpected scan: cursor=%d client=%+v", cursor, client)
	}
	if len(records) != 2 {
		t.Fatalf("records count = %d, want 2", len(records))
	}
	if records[0].RequestID != 9007199254740995 || records[0].UserID != 9007199254740993 || records[0].GiftID != 170406720000000001 {
		t.Fatalf("large IDs changed: %+v", records[0])
	}
	if records[1].UserID != 107 || records[1].GiftID != 3 {
		t.Fatalf("valid record after read failure was lost: %+v", records[1])
	}
}

func TestScanLotteryRecordsKeepsCursorOnFailure(t *testing.T) {
	wantErr := errors.New("scan failed")
	cache := NewGiftCache(&lotteryRecordClient{scanErr: wantErr})
	records, cursor, err := cache.ScanLotteryRecords(context.Background(), 37, 100)
	if !errors.Is(err, wantErr) || cursor != 37 || len(records) != 0 {
		t.Fatalf("got records=%v cursor=%d err=%v", records, cursor, err)
	}
}

func TestScanLotteryRecordsAdvancesOnEmptyPage(t *testing.T) {
	cache := NewGiftCache(&lotteryRecordClient{nextCursor: 37})
	records, cursor, err := cache.ScanLotteryRecords(context.Background(), 0, 100)
	if err != nil || cursor != 37 || len(records) != 0 {
		t.Fatalf("got records=%v cursor=%d err=%v", records, cursor, err)
	}
}
