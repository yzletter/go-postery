package dao

import (
	"errors"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type GormTagDAO struct {
	db *gorm.DB
}

func NewTagDAO(db *gorm.DB) TagDAO {
	return &GormTagDAO{db: db}
}

func (dao *GormTagDAO) Create(name string, slug string) (int, error) {
	tag := model.Tag{
		Id:         int(snowflake.NextID()),
		Name:       name,
		Slug:       slug,
		DeleteTime: nil,
	}

	tx := dao.db.Create(&tag)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			// 唯一键冲突
			return 0, ErrUniqueKeyConflict
		}

		// 数据库内部错误
		slog.Error("MySQL Create Tag Failed", "error", tx.Error)
		return 0, ErrInternal
	}

	return tag.Id, nil
}

func (dao *GormTagDAO) Exist(name string) (int, error) {
	tag := model.Tag{}
	tx := dao.db.Select("id").Where("name = ?", name).First(&tag)
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("MySQL Find Tag Failed")
			return 0, ErrInternal
		}
		return 0, gorm.ErrRecordNotFound
	}
	return tag.Id, nil
}

func (dao *GormTagDAO) Bind(pid, tid int) error {
	postTag := model.PostTag{
		Id:     int(snowflake.NextID()),
		PostId: pid,
		TagId:  tid,
	}
	tx := dao.db.Create(&postTag)
	if tx.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tx.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUniqueKeyConflict
		}
		slog.Error("MySQL Create Post_Tag Failed", "error", tx.Error) // 记录日志, 方便后续人工定位问题所在
		return ErrInternal
	}

	return nil
}

func (dao *GormTagDAO) DeleteBind(pid, tid int) error {
	tx := dao.db.Where("post_id = ? AND tag_id = ?", pid, tid).Delete(&model.PostTag{})
	if tx.RowsAffected == 0 {
		slog.Error("MySQL Delete Post_Tag Failed", "error", tx.Error) // 记录日志, 方便后续人工定位问题所在
		return errors.New("删除失败")
	}

	return nil
}

// FindTagsByPostID 根据 PostID 查找 Tags
func (dao *GormTagDAO) FindTagsByPostID(pid int) ([]string, error) {
	var names []string
	tx := dao.db.Table("post_tag pt").
		Joins("JOIN tag t ON t.id = pt.tag_id").
		Where("pt.post_id = ?", pid).
		Pluck("t.name", &names)

	if tx.Error != nil {
		// 数据库内部错误
		slog.Error("MySQL Find Tag Failed", "error", tx.Error)
		return nil, ErrInternal
	}
	return names, nil
}
