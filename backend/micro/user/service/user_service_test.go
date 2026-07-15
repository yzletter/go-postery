package service

import (
	"context"
	"testing"
	"time"

	"github.com/yzletter/go-postery/backend/micro/user/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUserRepository struct {
	updateID      int64
	updateFields  map[string]any
	updateCalled  bool
	updateErr     error
	profile       domain.Profile
	profileErr    error
	idsAfterTime  []int64
	idsAfterError error
}

func (repo *fakeUserRepository) GetProfile(ctx context.Context, uid int64) (domain.Profile, error) {
	return repo.profile, repo.profileErr
}

func (repo *fakeUserRepository) GetIDAfterTime(ctx context.Context, timeAfter time.Time) ([]int64, error) {
	return repo.idsAfterTime, repo.idsAfterError
}

func (repo *fakeUserRepository) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	repo.updateCalled = true
	repo.updateID = id
	repo.updateFields = updates
	return repo.updateErr
}

func TestUploadAvatarCallbackUpdatesProfile(t *testing.T) {
	repo := &fakeUserRepository{}
	svc := NewUserService(repo, nil, nil, nil, nil)
	object := "users/avatar/1/avatar-test.png"

	if err := svc.UploadAvatarCallback(context.Background(), 1, object); err != nil {
		t.Fatalf("UploadAvatarCallback() error = %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected UpdateProfile to be called")
	}
	if repo.updateID != 1 {
		t.Fatalf("UpdateProfile id = %d, want 1", repo.updateID)
	}
	if got := repo.updateFields["avatar"]; got != object {
		t.Fatalf("avatar update = %v, want %s", got, object)
	}
}

func TestUploadAvatarCallbackRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
		object string
	}{
		{
			name:   "invalid user id",
			userID: 0,
			object: "users/avatar/1/avatar-test.png",
		},
		{
			name:   "object user id mismatch",
			userID: 1,
			object: "users/avatar/2/avatar-test.png",
		},
		{
			name:   "empty object",
			userID: 1,
			object: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{}
			svc := NewUserService(repo, nil, nil, nil, nil)

			err := svc.UploadAvatarCallback(context.Background(), tt.userID, tt.object)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("UploadAvatarCallback() code = %s, want %s, err = %v", status.Code(err), codes.InvalidArgument, err)
			}
			if repo.updateCalled {
				t.Fatal("UpdateProfile should not be called for invalid input")
			}
		})
	}
}
