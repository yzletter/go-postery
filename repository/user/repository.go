package repository

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

// todo 代码优化

var (
	ErrUniqueKeyConflict = errors.New("唯一键冲突")
	ErrMySQLInternal     = errors.New("数据库内部错误")
)

// GormUserRepository 用 Gorm 实现 UserRepository
type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{
		db: db,
	}
}

// Create 创建一条 User 记录
func (repo *GormUserRepository) Create(name, password, ip string) (model.User, error) {
	// 将模型绑定到结构体
	user := model.User{
		Id:          snowflake.NextID(), // 用户 ID 雪花算法
		Name:        name,
		PassWord:    password,
		Status:      1,  // 用户状态为正常
		LastLoginIP: ip, // 用户登录 IP
		BirthDay:    nil,
	}

	// 到 MySQL 中创建新记录, 需要传指针
	err := repo.db.Create(&user).Error
	if err != nil {
		// 判断是否为 MySQL 错误
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				// Unique Key 冲突
				return model.User{}, ErrUniqueKeyConflict
			}
		}
		// 非主键冲突, 数据库出错
		slog.Error("MySQL Create Record Failed", "user", user, "error", err) // 记录日志, 方便后续人工定位问题所在
		return model.User{}, ErrMySQLInternal
	}

	// 返回 User
	return user, nil
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

func (repo *GormUserRepository) GetByID(uid int) (bool, model.User) {
	user := model.User{Id: uid}
	tx := repo.db.Select("*").First(&user) // 隐含的where条件是id, 注意：Find不会返回ErrRecordNotFound
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("go-postery GetBriefById : 查找用户失败", "uid", uid, "error", tx.Error)
		}
		return false, model.User{}
	}
	return true, user
}

func (repo *GormUserRepository) GetByName(name string) (model.User, error) {
	user := model.User{}
	tx := repo.db.Select("*").Where("name=?", name).First(&user) // 隐含的where条件是id, 注意：Find不会返回ErrRecordNotFound
	if tx.Error != nil {
		// 若错误不是记录未找到, 记录系统错误
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("go-postery GetUserByName : 查找用户失败", "uid", name, "error", tx.Error)
		}
		return model.User{}, errors.New("查找用户失败")
	}

	return user, nil
}
