package repository

import (
	"errors"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	repository "github.com/yzletter/go-postery/repository/user"
	"gorm.io/gorm"
)

type TagDBRepository struct {
	db *gorm.DB
}

func NewTagDBRepository(db *gorm.DB) *TagDBRepository {
	return &TagDBRepository{db: db}
}

func (repo *TagDBRepository) Create(name string, slug string) error {
	tag := model.Tag{
		Id:         snowflake.NextID(),
		Name:       name,
		Slug:       slug,
		DeleteTime: nil,
	}

	tx := repo.db.Create(&tag)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			// 唯一键冲突
			return repository.ErrUniqueKeyConflict
		}

		// 数据库内部错误
		slog.Error("MySQL Create Tag Failed", "error", tx.Error)
		return repository.ErrMySQLInternal
	}

	return nil
}

// FindTagsByPostID 根据 PostID 查找 Tags
func (repo *TagDBRepository) FindTagsByPostID(pid int) ([]string, error) {
	var res []string
	tx := repo.db.Model(&model.PostTag{}).Where("post_id = ?", pid).Pluck("name", &res)
	if tx.Error != nil {
		// 数据库内部错误
		slog.Error("MySQL Find Tag Failed", "error", tx.Error)
		return nil, repository.ErrMySQLInternal
	}
	return res, nil
}
