package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/rs/xid"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/auth/conf"
	model2 "github.com/yzletter/go-postery/auth/model"
	code_conf "github.com/yzletter/go-postery/code/conf"
	code_model "github.com/yzletter/go-postery/code/model"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/service/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/repository"
)

type authService struct {
	authRepo   repository.AuthRepository
	jwtManager ports.JwtManager
	passHasher ports.PasswordHasher
	idGen      ports.IDGenerator

	codeConn *grpc.ClientConn
}
