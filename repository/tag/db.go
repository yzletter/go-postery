package repository

import (
	"errors"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	userRepository "github.com/yzletter/go-postery/repository/user"
	"gorm.io/gorm"
)

type TagDBRepository struct {
	db *gorm.DB
}

func NewTagDBRepository(db *gorm.DB) *TagDBRepository {
	return &TagDBRepository{db: db}
}

func (repo *TagDBRepository) Create(name string, slug string) (int, error) {
	tag := model.Tag{
		Id:         int(snowflake.NextID()),
		Name:       name,
		Slug:       slug,
		DeleteTime: nil,
	}

	tx := repo.db.Create(&tag)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			// 唯一键冲突
			return 0, userRepository.ErrUniqueKeyConflict
		}

		// 数据库内部错误
		slog.Error("MySQL Create Tag Failed", "error", tx.Error)
		return 0, userRepository.ErrMySQLInternal
	}

	return tag.Id, nil
}

func (repo *TagDBRepository) Exist(name string) (int, error) {
	tag := model.Tag{}
	tx := repo.db.Select("id").Where("name = ?", name).First(&tag)
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("MySQL Find Tag Failed")
			return 0, userRepository.ErrMySQLInternal
		}
		return 0, gorm.ErrRecordNotFound
	}
	return tag.Id, nil
}

func (repo *TagDBRepository) Bind(pid, tid int) error {
	postTag := model.PostTag{
		Id:     int(snowflake.NextID()),
		PostId: pid,
		TagId:  tid,
	}
	tx := repo.db.Create(&postTag)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return userRepository.ErrUniqueKeyConflict
		}
		slog.Error("MySQL Create Post_Tag Failed", "error", tx.Error) // 记录日志, 方便后续人工定位问题所在
		return userRepository.ErrMySQLInternal
	}

	return nil
}

func (repo *TagDBRepository) DeleteBind(pid, tid int) error {
	tx := repo.db.Where("post_id = ? AND tag_id = ?", pid, tid).Delete(&model.PostTag{})
	if tx.RowsAffected == 0 {
		slog.Error("MySQL Delete Post_Tag Failed", "error", tx.Error) // 记录日志, 方便后续人工定位问题所在
		return errors.New("删除失败")
	}

	return nil
}

// FindTagsByPostID 根据 PostID 查找 Tags
func (repo *TagDBRepository) FindTagsByPostID(pid int) ([]string, error) {
	var names []string
	tx := repo.db.Table("post_tag pt").
		Joins("JOIN tag t ON t.id = pt.tag_id").
		Where("pt.post_id = ?", pid).
		Pluck("t.name", &names)

	if tx.Error != nil {
		// 数据库内部错误
		slog.Error("MySQL Find Tag Failed", "error", tx.Error)
		return nil, userRepository.ErrMySQLInternal
	}
	return names, nil
}
