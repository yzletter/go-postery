package client

import (
	"context"

	"github.com/yzletter/go-postery/api/proto/code/v1"
)

type CodeClient interface {
	Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error)
	Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error)
	Close()
}
