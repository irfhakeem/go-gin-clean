package unit_test

import (
	"context"
	"testing"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/domain/policy"
	"go-gin-clean/internal/domain/vo"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"
	"go-gin-clean/test/mock"

	"github.com/google/uuid"
)

func TestUserUseCase_DeleteUser_DeletesAvatar(t *testing.T) {
	logger.Init("test")

	userRepo := mock.NewMockUserRepository()
	refreshRepo := mock.NewMockRefreshTokenRepository()
	storage := mock.NewMockStorage()
	cache := mock.NewMockCache()
	txMgr := mock.NewMockTransactionManager()
	pol := policy.NewUserPolicy()

	adminID := uuid.New()
	adminActor := policy.Actor{ID: adminID, Role: vo.RoleSuperAdmin}

	targetID := uuid.New()
	targetUser, _ := entity.NewUser("Target", "target@example.com", "hash")
	targetUser.ID = targetID
	targetUser.Avatar = "https://res.cloudinary.com/demo/image/upload/v123/users/" + targetID.String() + "/avatar/pic.jpg"
	userRepo.Users[targetID.String()] = targetUser
	storage.Files[targetUser.Avatar] = []byte("image_data")

	uc := usecase.NewUserUseCase(
		pol,
		txMgr,
		userRepo,
		refreshRepo,
		nil,
		mock.NewMockHasher(),
		storage,
		cache,
		&config.ServerConfig{},
	)

	err := uc.DeleteUser(context.Background(), adminActor, targetID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(storage.DeletedFiles) != 1 {
		t.Fatalf("expected 1 deleted file from storage, got %d", len(storage.DeletedFiles))
	}
	if storage.DeletedFiles[0] != targetUser.Avatar {
		t.Fatalf("expected %s deleted from storage, got %s", targetUser.Avatar, storage.DeletedFiles[0])
	}
}
