package repository

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/model"
	"gorm.io/gorm"
)

// todo 代码优化

// GormUserRepository 用 Gorm 实现 UserRepository
type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{
		db: db,
	}
}

func (repo *GormUserRepository) Create(name, password string) (int, error) {
	// 将模型绑定到结构体
	user := model.User{
		Name:     name,
		PassWord: password,
	}

	// 到 MySQL 中创建新记录
	err := repo.db.Create(&user).Error // 需要传指针
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) { // 判断是否为 MySQL 错误
			if mysqlErr.Number == 1062 { // Unique Key 冲突
				return 0, fmt.Errorf("用户[%s]已存在", name)
			}
		}
		// 记录日志, 方便后续人工定位问题所在
		slog.Error("go-postery RegisterUser : 用户注册失败", "name", name, "error", err)
		return 0, fmt.Errorf("用户注册失败，请稍后重试")
	}

	// 返回 Id
	return user.Id, nil
}

func (repo *GormUserRepository) Delete(uid int) (bool, error) {
	// 将模型绑定到结构体
	user := model.User{
		Id: uid,
	}

	// 删除记录
	tx := repo.db.Delete(&user)
	if tx.Error != nil {
		// 系统层面错误
		slog.Error("go-postery LogOffUser : 用户注销失败", "uid", uid, "error", tx.Error)
		return false, fmt.Errorf("用户注销失败，请稍后重试")
	} else if tx.RowsAffected == 0 {
		// 业务层面错误
		return false, fmt.Errorf("用户注销失败, uid %d 不存在", uid)
	}

	return true, nil
}

func (repo *GormUserRepository) UpdatePassword(uid int, oldPass, newPass string) error {
	tx := repo.db.Model(&model.User{}).Where("id=? and password=?", uid, oldPass).Update("password", newPass)
	if tx.Error != nil {
		// 系统错误
		slog.Error("go-postery UpdatePassword : 密码更改失败", "uid", uid, "error", tx.Error)
		return fmt.Errorf("更改用户密码失败, 请稍后再试")
	} else if tx.RowsAffected == 0 {
		// 业务错误
		return fmt.Errorf("用户 id 或旧密码错误")
	}

	return nil
}

func (repo *GormUserRepository) GetByID(uid int) *model.User {
	user := model.User{Id: uid}
	tx := repo.db.Select("*").First(&user) // 隐含的where条件是id, 注意：Find不会返回ErrRecordNotFound
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("go-postery GetById : 查找用户失败", "uid", uid, "error", tx.Error)
		}
		return nil
	}

	return &user
}

func (repo *GormUserRepository) GetByName(name string) *model.User {
	user := model.User{}
	tx := repo.db.Select("*").Where("name=?", name).First(&user) // 隐含的where条件是id, 注意：Find不会返回ErrRecordNotFound
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("go-postery GetUserByName : 查找用户失败", "uid", name, "error", tx.Error)
		}
		return nil
	}

	return &user
}
