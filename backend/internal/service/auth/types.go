package auth

import (
	"errors"
	"time"

	"cephtower/backend/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidResetCode   = errors.New("验证码无效或已过期")
	ErrEmailRequired      = errors.New("当前账号未绑定邮箱，请联系管理员重设密码")
)

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      store.User
}

type CreateUserInput struct {
	Username    string
	DisplayName string
	Email       string
	Role        string
	Permissions []string
	Password    string
	Enabled     *bool
}

type UpdateUserInput struct {
	DisplayName *string
	Email       *string
	Role        *string
	Permissions []string
	Password    *string
	Enabled     *bool
}
