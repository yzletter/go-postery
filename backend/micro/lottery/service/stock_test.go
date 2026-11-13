package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository"
)

type orderRetryGiftRepository struct {
	repository.GiftRepository
	scan func(context.Context, uint64, int64) ([]*model.LotteryRecord, uint64, error)
}

func (repo *orderRetryGiftRepository) ScanLotteryRecords(ctx context.Context, cursor uint64, count int64) ([]*model.LotteryRecord, uint64, error) {
	return repo.scan(ctx, cursor, count)
}

type orderRetryOrderRepository struct {
	repository.OrderRepository
	create func(context.Context, *model.Order) (*model.Order, error)
}

func (repo *orderRetryOrderRepository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	return repo.create(ctx, order)
}

type orderRetryIDGenerator struct {
	id int64
}

func (gen *orderRetryIDGenerator) NextID() int64 {
	gen.id++
	return gen.id
}

func (gen *orderRetryIDGenerator) NextIDUint64() uint64 {
	return uint64(gen.NextID())
}

func TestRetryLotteryOrdersRecoversAndReusesExistingOrders(t *testing.T) {
	existing := model.Order{
		ID:                  700,
		UserID:              101,
		GiftID:              1,
		ExpireAt:            time.Now().Add(-time.Hour),
		Status:              model.OrderStatusPaid,
		StockRollbackStatus: model.StockRollbackStatusUnneeded,
	}
	orders := map[int64]*model.Order{101: &existing}
	attempts := make(map[int64]int)
	var retryContexts []context.Context
	giftRepo := &orderRetryGiftRepository{
		scan: func(ctx context.Context, cursor uint64, count int64) ([]*model.LotteryRecord, uint64, error) {
			return []*model.LotteryRecord{
				{RequestID: 9007199254740993, UserID: 101, GiftID: 1},
				{RequestID: 9007199254740994, UserID: 102, GiftID: 2},
				{RequestID: 9007199254740995, UserID: 103, GiftID: 3},
			}, 0, nil
		},
	}
	orderRepo := &orderRetryOrderRepository{
		create: func(ctx context.Context, order *model.Order) (*model.Order, error) {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > lotteryOrderRetryTimeout {
				t.Fatal("order retry has no bounded timeout")
			}
			retryContexts = append(retryContexts, ctx)
			attempts[order.UserID]++
			if saved, ok := orders[order.UserID]; ok {
				return saved, nil
			}
			if order.UserID == 103 && attempts[103] == 1 {
				return nil, repository.ErrServerInternal
			}
			saved := *order
			orders[order.UserID] = &saved
			return &saved, nil
		},
	}
	svc := &lotteryService{giftRepo: giftRepo, orderRepo: orderRepo, idGen: &orderRetryIDGenerator{id: 1000}}
	before := time.Now()
	if cursor := svc.retryLotteryOrders(context.Background(), 0); cursor != 0 {
		t.Fatalf("cursor = %d, want 0", cursor)
	}
	if orders[103] != nil || orders[102] == nil {
		t.Fatalf("unexpected first retry result: %+v", orders)
	}
	created := *orders[102]
	if created.ID != 1002 || created.GiftID != 2 || created.Count != 1 ||
		created.Status != model.OrderStatusPending || created.StockRollbackStatus != model.StockRollbackStatusPending {
		t.Fatalf("unexpected recovered order: %+v", created)
	}
	if created.ExpireAt.Before(before.Add(conf.RocketLotteryPayDelay*time.Second)) ||
		created.ExpireAt.After(time.Now().Add(conf.RocketLotteryPayDelay*time.Second)) {
		t.Fatalf("unexpected payment deadline: %v", created.ExpireAt)
	}

	svc.retryLotteryOrders(context.Background(), 0)
	if len(orders) != 3 || orders[103] == nil || attempts[103] != 2 {
		t.Fatalf("failed order was not recovered: orders=%+v attempts=%v", orders, attempts)
	}
	if *orders[101] != existing || *orders[102] != created {
		t.Fatal("repeated scanning changed an existing order or extended its payment deadline")
	}
	for _, ctx := range retryContexts {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatal("retry context was not released")
		}
	}
}

func TestRetryLotteryOrdersPagination(t *testing.T) {
	wantErr := errors.New("scan failed")
	for _, tt := range []struct {
		name       string
		scanErr    error
		nextCursor uint64
		wantCursor uint64
	}{
		{name: "empty page", nextCursor: 37, wantCursor: 37},
		{name: "scan failure", scanErr: wantErr, wantCursor: 12},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc := &lotteryService{giftRepo: &orderRetryGiftRepository{
				scan: func(ctx context.Context, cursor uint64, count int64) ([]*model.LotteryRecord, uint64, error) {
					if cursor != 12 || count != lotteryRecordScanCount {
						t.Fatalf("unexpected scan arguments: cursor=%d count=%d", cursor, count)
					}
					return nil, tt.nextCursor, tt.scanErr
				},
			}}
			if cursor := svc.retryLotteryOrders(context.Background(), 12); cursor != tt.wantCursor {
				t.Fatalf("cursor = %d, want %d", cursor, tt.wantCursor)
			}
		})
	}
}

func TestLotteryOrderRetryScannerStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		(&lotteryService{}).StartLotteryOrderRetryScanner(ctx)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scanner did not stop after cancellation")
	}
}
