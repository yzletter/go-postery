package service

import (
	"context"
	"testing"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	infraRocketMQ "github.com/yzletter/go-postery/backend/infra/mq/rocketmq"
	"github.com/yzletter/go-postery/backend/micro/lottery/dto"
	"github.com/yzletter/go-postery/backend/micro/lottery/model"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryStateGiftRepository struct {
	repository.GiftRepository
	hasStock    bool
	detailCalls int
}

func (repo *lotteryStateGiftRepository) GetCacheInventory(ctx context.Context) ([]*model.Gift, error) {
	if !repo.hasStock {
		return nil, nil
	}
	return []*model.Gift{{ID: 1, Count: 1}}, nil
}

func (repo *lotteryStateGiftRepository) Lottery(ctx context.Context, requestID, userID, giftID int64) (int64, int64, error) {
	return 1001, 1, nil
}

func (repo *lotteryStateGiftRepository) Reuse(ctx context.Context, requestID, userID int64) (int64, int64, error) {
	return 1001, 1, nil
}

func (repo *lotteryStateGiftRepository) GetByID(ctx context.Context, gid int64) (*model.Gift, error) {
	repo.detailCalls++
	return &model.Gift{ID: gid, Name: "gift"}, nil
}

type lotteryStateOrderRepository struct {
	repository.OrderRepository
	order    *model.Order
	getErr   error
	payErr   error
	payCalls int
	payID    int64
}

func (repo *lotteryStateOrderRepository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	return repo.order, nil
}

func (repo *lotteryStateOrderRepository) GetOrder(ctx context.Context, uid int64) (*model.Order, error) {
	return repo.order, repo.getErr
}

func (repo *lotteryStateOrderRepository) Pay(ctx context.Context, orderID int64) error {
	repo.payCalls++
	repo.payID = orderID
	return repo.payErr
}

type lotteryStateProducer struct {
	rmq_client.Producer
	calls int
}

func (producer *lotteryStateProducer) Send(ctx context.Context, message *rmq_client.Message) ([]*rmq_client.SendReceipt, error) {
	producer.calls++
	return nil, nil
}

func TestLotteryChecksSavedOrderState(t *testing.T) {
	now := time.Now()
	for _, path := range []struct {
		name     string
		hasStock bool
	}{
		{name: "inventory available", hasStock: true},
		{name: "reuse after inventory exhausted"},
	} {
		t.Run(path.name, func(t *testing.T) {
			for _, tt := range []struct {
				name     string
				state    int
				expireAt time.Time
				wantCode codes.Code
			}{
				{name: "pending", state: model.OrderStatusPending, expireAt: now.Add(time.Hour), wantCode: codes.OK},
				{name: "cancelled", state: model.OrderStatusCancelled, expireAt: now.Add(time.Hour), wantCode: codes.NotFound},
				{name: "expired", state: model.OrderStatusExpired, expireAt: now.Add(-time.Hour), wantCode: codes.NotFound},
				{name: "pending past deadline", state: model.OrderStatusPending, expireAt: now.Add(-time.Hour), wantCode: codes.NotFound},
				{name: "paid", state: model.OrderStatusPaid, expireAt: now.Add(time.Hour), wantCode: codes.AlreadyExists},
			} {
				t.Run(tt.name, func(t *testing.T) {
					giftRepo := &lotteryStateGiftRepository{hasStock: path.hasStock}
					orderRepo := &lotteryStateOrderRepository{order: &model.Order{
						ID: 1001, UserID: 101, GiftID: 1, Status: tt.state, ExpireAt: tt.expireAt,
					}}
					producer := &lotteryStateProducer{}
					svc := NewLotteryService(orderRepo, giftRepo, &infraRocketMQ.RocketMQ{RocketProducer: producer}, &orderRetryIDGenerator{id: 2000})
					result, err := svc.Lottery(context.Background(), 101, 1002)
					if status.Code(err) != tt.wantCode {
						t.Fatalf("error = %v, want code %v", err, tt.wantCode)
					}
					if tt.wantCode != codes.OK {
						if result != nil || producer.calls != 0 || giftRepo.detailCalls != 0 {
							t.Fatalf("ended order continued lottery flow: result=%+v messages=%d details=%d", result, producer.calls, giftRepo.detailCalls)
						}
						return
					}
					if result == nil || !result.Success || result.Description != dto.DescriptionLotterySuccess || result.OrderID != 1001 ||
						producer.calls != 1 || giftRepo.detailCalls != 1 {
						t.Fatalf("valid pending order was not returned: result=%+v messages=%d details=%d", result, producer.calls, giftRepo.detailCalls)
					}
				})
			}
		})
	}
}

func TestPayReusesPaidOrderAndChecksOwnership(t *testing.T) {
	paidOrder := &model.Order{
		ID: 1001, UserID: 101, GiftID: 1,
		Status: model.OrderStatusPaid, ExpireAt: time.Now().Add(-time.Hour),
	}
	pendingOrder := &model.Order{
		ID: 1001, UserID: 101, GiftID: 1,
		Status: model.OrderStatusPending, ExpireAt: time.Now().Add(time.Hour),
	}
	for _, tt := range []struct {
		name         string
		order        *model.Order
		getErr       error
		payErr       error
		wantCode     codes.Code
		wantPayCalls int
	}{
		{name: "paid order retry", order: paidOrder, wantCode: codes.OK},
		{name: "wrong user", order: &model.Order{ID: 1001, UserID: 102, GiftID: 1, Status: model.OrderStatusPaid}, wantCode: codes.NotFound},
		{name: "wrong order", order: &model.Order{ID: 1002, UserID: 101, GiftID: 1, Status: model.OrderStatusPaid}, wantCode: codes.NotFound},
		{name: "wrong gift", order: &model.Order{ID: 1001, UserID: 101, GiftID: 2, Status: model.OrderStatusPaid}, wantCode: codes.NotFound},
		{name: "nil order", wantCode: codes.NotFound},
		{name: "missing order", getErr: repository.ErrRecordNotFound, wantCode: codes.NotFound},
		{name: "query failure", getErr: repository.ErrServerInternal, wantCode: codes.Internal},
		{name: "pending order", order: pendingOrder, wantCode: codes.OK, wantPayCalls: 1},
		{name: "payment rejected", order: pendingOrder, payErr: repository.ErrRecordNotFound, wantCode: codes.NotFound, wantPayCalls: 1},
		{name: "payment failure", order: pendingOrder, payErr: repository.ErrServerInternal, wantCode: codes.Internal, wantPayCalls: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &lotteryStateOrderRepository{order: tt.order, getErr: tt.getErr, payErr: tt.payErr}
			svc := NewLotteryService(repo, nil, nil, nil)
			err := svc.Pay(context.Background(), 101, 1001, 1)
			if status.Code(err) != tt.wantCode || repo.payCalls != tt.wantPayCalls {
				t.Fatalf("got err=%v payCalls=%d, want code=%v payCalls=%d", err, repo.payCalls, tt.wantCode, tt.wantPayCalls)
			}
			if repo.payCalls != 0 && repo.payID != 1001 {
				t.Fatalf("paid order ID = %d, want 1001", repo.payID)
			}
		})
	}
}

func TestResultDistinguishesMissingOrderAndQueryFailure(t *testing.T) {
	for _, tt := range []struct {
		name     string
		getErr   error
		wantCode codes.Code
	}{
		{name: "existing order", wantCode: codes.OK},
		{name: "missing order", getErr: repository.ErrRecordNotFound, wantCode: codes.NotFound},
		{name: "database failure", getErr: repository.ErrServerInternal, wantCode: codes.Internal},
	} {
		t.Run(tt.name, func(t *testing.T) {
			savedOrder := &model.Order{ID: 1001, UserID: 101, GiftID: 1}
			orderRepo := &lotteryStateOrderRepository{order: savedOrder, getErr: tt.getErr}
			giftRepo := &lotteryStateGiftRepository{}
			svc := NewLotteryService(orderRepo, giftRepo, nil, nil)
			order, gift, err := svc.Result(context.Background(), 101)
			if status.Code(err) != tt.wantCode {
				t.Fatalf("Result error = %v, want code %v", err, tt.wantCode)
			}
			if tt.wantCode != codes.OK {
				if order != nil || gift != nil || giftRepo.detailCalls != 0 {
					t.Fatalf("query failure continued result lookup: order=%v gift=%v details=%d", order, gift, giftRepo.detailCalls)
				}
				return
			}
			if order != savedOrder || gift == nil || gift.ID != savedOrder.GiftID || giftRepo.detailCalls != 1 {
				t.Fatalf("unexpected result: order=%v gift=%v details=%d", order, gift, giftRepo.detailCalls)
			}
		})
	}
}
