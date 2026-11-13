package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository/dao"
)

type orderQueryDAO struct {
	dao.OrderDAO
	err error
}

func (dao *orderQueryDAO) Get(ctx context.Context, uid int64) (*model.Order, error) {
	return nil, dao.err
}

func TestGetOrderPreservesErrorType(t *testing.T) {
	for _, tt := range []struct {
		name    string
		daoErr  error
		wantErr error
	}{
		{name: "missing order", daoErr: dao.ErrRecordNotFound, wantErr: ErrRecordNotFound},
		{name: "database failure", daoErr: dao.ErrServerInternal, wantErr: ErrServerInternal},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewOrderRepository(&orderQueryDAO{err: tt.daoErr})
			_, err := repo.GetOrder(context.Background(), 101)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetOrder error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
