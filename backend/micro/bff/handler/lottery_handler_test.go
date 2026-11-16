package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	lottery_grpc "github.com/yzletter/go-postery/backend/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/bff/errno"
	"github.com/yzletter/go-postery/backend/utils/response"
)

type lotteryHandlerClient struct {
	grpcclient.LotteryClient
	drawErr        error
	resultErr      error
	payErr         error
	payCalls       int
	lastPayRequest *lottery_grpc.LotteryCommonRequest
}

func (client *lotteryHandlerClient) Lottery(ctx context.Context, request *lottery_grpc.UserID) (*lottery_grpc.LotteryResponse, error) {
	return nil, client.drawErr
}

func (client *lotteryHandlerClient) Result(ctx context.Context, request *lottery_grpc.UserID) (*lottery_grpc.Order, error) {
	return nil, client.resultErr
}

func (client *lotteryHandlerClient) Pay(ctx context.Context, request *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	client.payCalls++
	client.lastPayRequest = request
	return &lottery_grpc.EmptyResponse{}, client.payErr
}

func TestLotteryMapsEndedAndPaidOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name    string
		err     error
		wantErr *errno.Error
	}{
		{name: "ended order", err: errs.ErrNotFound, wantErr: errno.ErrLotteryEnded},
		{name: "paid order", err: errs.ErrAlreadyExits, wantErr: errno.ErrLotteryAlreadyPaid},
		{name: "database failure", err: errs.ErrInternal, wantErr: errno.ErrServerInternal},
		{name: "unavailable", err: errs.ErrUnavailable, wantErr: errno.ErrServiceUnavailable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/lottery/lucky", nil)
			ctx.Set(conf.UserIDInContext, int64(101))
			hdl := NewLotteryHandler(&lotteryHandlerClient{drawErr: tt.err}, nil)
			hdl.Lottery(ctx)
			var body response.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tt.wantErr.HTTPStatus || body.Code != tt.wantErr.Code || body.Msg != tt.wantErr.Msg {
				t.Fatalf("got HTTP=%d body=%+v, want %+v", recorder.Code, body, tt.wantErr)
			}
		})
	}
}

func TestLotteryResultPreservesQueryErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name    string
		err     error
		wantErr *errno.Error
	}{
		{name: "missing order", err: errs.ErrNotFound, wantErr: errno.ErrOrderNotFound},
		{name: "database failure", err: errs.ErrInternal, wantErr: errno.ErrServerInternal},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/lottery/result", nil)
			ctx.Set(conf.UserIDInContext, int64(101))
			hdl := NewLotteryHandler(&lotteryHandlerClient{resultErr: tt.err}, nil)
			hdl.Result(ctx)
			var body response.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tt.wantErr.HTTPStatus || body.Code != tt.wantErr.Code {
				t.Fatalf("got HTTP=%d body=%+v, want %+v", recorder.Code, body, tt.wantErr)
			}
		})
	}
}

func TestLotteryPayAllowsIdempotentSuccessAndPreservesIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &lotteryHandlerClient{}
	hdl := NewLotteryHandler(client, nil)
	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/lottery/pay", strings.NewReader(
			`{"user_id":"9007199254740993","gift_id":"170406720000000001","temp_order_id":"9007199254740995"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set(conf.UserIDInContext, int64(9007199254740993))
		hdl.Pay(ctx)
		if recorder.Code != http.StatusOK {
			t.Fatalf("pay returned HTTP %d: %s", recorder.Code, recorder.Body.String())
		}
	}
	if client.payCalls != 2 || client.lastPayRequest.UserID != 9007199254740993 ||
		client.lastPayRequest.GiftID != 170406720000000001 || client.lastPayRequest.TempOrderID != 9007199254740995 {
		t.Fatalf("unexpected payment request: calls=%d request=%+v", client.payCalls, client.lastPayRequest)
	}
}
