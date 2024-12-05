package account

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-api-testing/internal/dal/model"
	"go-api-testing/internal/dao/account"
	"go-api-testing/utils/auth/jwt"
	"go-api-testing/utils/cryption"
	e "go-api-testing/utils/error"
	"go-api-testing/utils/logger"
	"go-api-testing/utils/response"
	"go.uber.org/zap"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserListRequest struct {
	Search string `form:"search"`
	Limit  uint   `form:"limit"`
	Offset uint   `form:"offset"`
}

type UserRequest struct {
	Id          int32  `json:"id"`
	Username    string `json:"username" binding:"required,min=3,max=20"`
	Password    string `json:"password" binding:"required,min=4"`
	IsDelete    bool   `json:"is_delete"`
	Description string `json:"description"`
}

type UserResponse struct {
	Id          int32  `json:"id"`
	UserName    string `json:"username"`
	Description string `json:"description"`
	CreateTime  string `json:"create_time"`
}

func RefreshToken(c *gin.Context) {
	var refreshTokenRequest RefreshTokenRequest
	if err := c.ShouldBindJSON(&refreshTokenRequest); err != nil {
		response.Fail(c, e.HttpBadRequest.GetErrCode(), err.Error())
		return
	}
	accessToken, refreshToken, err := jwt.RefreshTokens(refreshTokenRequest.RefreshToken)
	if err != nil {
		response.Fail(c, e.HttpInternalServerError.GetErrCode(), err.Error())
		return
	}
	response.Success(c, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

func Login(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		response.Fail(c, e.HttpBadRequest.GetErrCode(), err.Error())
		return
	}
	userDao := account.NewUserDao()
	user, err := userDao.GetUserByName(c, loginReq.Username, cryption.Sha256(loginReq.Password))
	if err != nil {
		response.FailByError(c, e.UserLoginError)
		return
	}
	if user == nil {
		response.FailByError(c, e.UserLoginError)
		return
	}
	accessToken, refreshToken, err := jwt.GenerateTokens(uint(user.ID), user.Username, "test")
	if err != nil {
		response.Fail(c, e.HttpInternalServerError.GetErrCode(), err.Error())
	}
	response.Success(c, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

// GetUserList 用户信息
func GetUserList(c *gin.Context) {
	var userListReq UserListRequest
	if err := c.BindQuery(&userListReq); err != nil {
		response.Fail(c, e.HttpBadRequest.GetErrCode(), err.Error())
		return
	}
	userDao := account.NewUserDao()
	// limit 默认为10
	if userListReq.Limit == 0 {
		userListReq.Limit = 10
	}
	userList, count, err := userDao.ListUserByConditions(c, &account.UserCondition{
		SearchValue: userListReq.Search,
		Limit:       userListReq.Limit,
		Offset:      userListReq.Offset,
	})
	if err != nil {
		response.FailByError(c, e.HttpInternalServerError)
		return
	}
	data := make([]UserResponse, 0, len(userList))
	for _, user := range userList {
		data = append(data, UserResponse{
			Id:          user.ID,
			UserName:    user.Username,
			Description: *user.Description,
			CreateTime:  user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, gin.H{"total": count, "list": data})
}

func CreateUser(c *gin.Context) {
	var userReq UserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		response.Fail(c, e.HttpBadRequest.GetErrCode(), err.Error())
		return
	}
	userDao := account.NewUserDao()

	// 校验用户名是否已经存在
	isDuplicate, err := userDao.IsNameDuplicate(c, userReq.Username)
	if err != nil {
		response.Fail(c, e.HttpInternalServerError.GetErrCode(), err.Error())
		return
	}
	if isDuplicate {
		response.Fail(c, e.UserNameError.GetErrCode(), e.UserNameError.GetErrMsg())
		return
	}

	// 创建用户
	user, err := userDao.CreateUser(c, &model.User{
		Username:    userReq.Username,
		Password:    cryption.Sha256(userReq.Password), // 密码加密(还可以加盐，增加密码难度)
		IsDelete:    false,
		Description: &userReq.Description,
	})
	if err != nil {
		response.Fail(c, e.UserCreateError.GetErrCode(), err.Error())
		return
	}
	operatorId, _ := c.Get("userId")
	operatorName, _ := c.Get("username")
	logger.Info(fmt.Sprintf(
		"用户(%d-%s)添加用户(%d-%s)成功", operatorId, operatorName, user.ID, user.Username),
		zap.String("type", "creat_user"),
		zap.Any("operator_id", operatorId),
		zap.Any("operator_name", operatorName),
		zap.Any("operator_name", operatorName),
		zap.Int32("user_id", user.ID),
		zap.String("description", userReq.Description),
	)
	response.Success(c, gin.H{"user_id": user.ID})
}

func DelUser(c *gin.Context) {
	var userReq UserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		response.Fail(c, e.HttpBadRequest.GetErrCode(), err.Error())
	}
	userDao := account.NewUserDao()
	_, err := userDao.UpdateUser(c, userReq.Id, &model.User{
		IsDelete: true,
	})
	if err != nil {
		response.Fail(c, e.HttpInternalServerError.GetErrCode(), err.Error())
	}
	operatorId, _ := c.Get("userId")
	operatorName, _ := c.Get("username")
	logger.Info(fmt.Sprintf(
		"用户(%d-%s)删除用户(%d)成功", operatorId, operatorName, userReq.Id),
		zap.String("type", "del_user"),
		zap.Any("operator_id", operatorId),
		zap.Any("operator_name", operatorName),
		zap.Int32("user_id", userReq.Id),
	)
	response.Success(c, gin.H{"user_id": userReq.Id})
}

func UpdateUser(c *gin.Context) {
	var userReq UserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		response.Fail(c, e.HttpBadRequest.GetErrCode(), err.Error())
	}
	userDao := account.NewUserDao()
	_, err := userDao.UpdateUser(c, userReq.Id, &model.User{
		Password:    cryption.Sha256(userReq.Password), // 密码加密(还可以加盐，增加密码难度),
		Description: &userReq.Description,
	})
	if err != nil {
		response.Fail(c, e.HttpInternalServerError.GetErrCode(), err.Error())
	}
	operatorId, _ := c.Get("userId")
	operatorName, _ := c.Get("username")
	logger.Info(fmt.Sprintf(
		"用户(%d-%s)更新用户(%d)成功", operatorId, operatorName, userReq.Id),
		zap.String("type", "update_user"),
		zap.Any("operator_id", operatorId),
		zap.Any("operator_name", operatorName),
		zap.Int32("user_id", userReq.Id),
	)
	response.Success(c, gin.H{"user_id": userReq.Id})
}
