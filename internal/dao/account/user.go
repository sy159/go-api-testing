package account

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go-api-testing/internal/dal/model"
	"go-api-testing/internal/dal/repo"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type UserDao struct {
	repo *repo.Repository
}

type UserCondition struct {
	SearchValue string
	Limit       uint
	Offset      uint
}

func NewUserDao() *UserDao {
	return &UserDao{
		repo: repo.NewRepository(),
	}
}

// CreateUser 创建用户
func (u *UserDao) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	err := u.repo.Query().User.WithContext(ctx).Create(user)
	if err != nil {
		return nil, errors.WithMessage(err, "用户创建失败")
	}
	return user, nil
}

// UpdateUser 更新用户(UpdateColumns有零值问题，Updates多一个钩子函数调用，更新时间等，save会全量更新)
func (u *UserDao) UpdateUser(ctx context.Context, userId int32, user *model.User) (*model.User, error) {
	m := u.repo.Query().User
	_, err := m.WithContext(ctx).Where(m.ID.Eq(userId)).Select(m.Password, m.IsDelete, m.Description).UpdateColumns(user)
	//_, err := m.WithContext(ctx).Where(m.ID.Eq(userId)).UpdateSimple(m.IsDelete.Value(false)) // 0值处理方式
	if err != nil {
		return nil, errors.WithMessage(err, "用户更新失败")
	}
	return user, nil
}

// GetUserById 查询用户by id
func (u *UserDao) GetUserById(ctx context.Context, userId int32) (*model.User, error) {
	if userId <= 0 {
		return nil, errors.New("用户id不存在")
	}
	m := u.repo.Query().User
	user, err := m.WithContext(ctx).Where(m.ID.Eq(userId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("该用户不存在")
		}
		return nil, errors.WithMessage(err, "用户查询异常")
	}
	return user, nil
}

// GetUserByName 用户登录
func (u *UserDao) GetUserByName(ctx context.Context, username, pwd string) (*model.User, error) {
	if len(username) <= 0 || len(pwd) <= 0 {
		return nil, errors.New("用户不存在")
	}
	m := u.repo.Query().User
	user, err := m.WithContext(ctx).
		Where(m.IsDelete.Is(false)).
		Where(m.Username.Eq(username)).
		Where(m.Password.Eq(pwd)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("该用户不存在")
		}
		return nil, errors.WithMessage(err, "用户查询异常")
	}
	return user, nil
}

// IsNameDuplicate 校验是否有重复的用户名
func (u *UserDao) IsNameDuplicate(ctx context.Context, username string) (bool, error) {
	m := u.repo.Query().User
	_, err := m.WithContext(ctx).Where(m.Username.Eq(username)).Where(m.IsDelete.Is(false)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return true, errors.WithStack(err)
	}
	return true, nil
}

// ListUserByConditions 获取用户列表
func (u *UserDao) ListUserByConditions(ctx context.Context, condition *UserCondition) ([]*model.User, int64, error) {
	m := u.repo.Query().User
	userList, count, err := m.WithContext(ctx).
		Where(m.IsDelete.Is(false)).
		Scopes(
			func(dao gen.Dao) gen.Dao {
				searchValue := condition.SearchValue
				if len(searchValue) > 0 {
					dao = dao.Where(
						m.WithContext(ctx).Where(m.Username.Like(fmt.Sprintf("%%%s%%", searchValue))).
							Or(m.Description.Like(fmt.Sprintf("%%%s%%", searchValue))),
					)
				}
				return dao
			},
			//u.SearchValueScope(condition.SearchValue),
		).
		FindByPage(int(condition.Offset), int(condition.Limit))
	if err != nil {
		return nil, 0, errors.WithMessage(err, "用户列表查询异常")
	}
	return userList, count, nil
}

// SearchValueScope 用户名或者备注模糊匹配
func (u *UserDao) SearchValueScope(searchValue string) func(tx gen.Dao) gen.Dao {
	return func(tx gen.Dao) gen.Dao {
		if searchValue == "" {
			return tx
		}
		m := u.repo.Query().User
		return tx.Where(m.Username.Eq(searchValue))
	}
}
