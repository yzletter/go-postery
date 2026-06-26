package handler

import (
	"context"

	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	"github.com/yzletter/go-postery/backend/grpc/manager"
)

type InteractiveHandler struct {
	interClient manager.InteractiveClient
}

func NewInteractiveHandler(interClient manager.InteractiveClient) *InteractiveHandler {
	return &InteractiveHandler{
		interClient: interClient,
	}
}

func (hdl *InteractiveHandler) GetFollowers(ctx context.Context, userID int64) (int64, error) {
	resp, err := hdl.interClient.GetFollowers(ctx, &interactive_grpc.ListFollowRequest{UserID: userID, PageNo: 1, PageSize: 1})
	if err != nil {
		return 0, err
	}
	return int64(resp.Count), nil
}

func (hdl *InteractiveHandler) GetFollowees(ctx context.Context, userID int64) (int64, error) {
	resp, err := hdl.interClient.GetFollowees(ctx, &interactive_grpc.ListFollowRequest{UserID: userID, PageNo: 1, PageSize: 1})
	if err != nil {
		return 0, err
	}
	return int64(resp.Count), nil
}
