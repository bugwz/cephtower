package store

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:64;not null"`
	DisplayName  string `gorm:"size:96;not null"`
	Email        string `gorm:"uniqueIndex;size:128"`
	Role         string `gorm:"size:24;not null;index"`
	Permissions  string `gorm:"type:text;not null"`
	PasswordHash string `gorm:"type:text;not null"`
	Enabled      bool   `gorm:"not null;default:true"`
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "user"
}

type PasswordResetCode struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	CodeHash  string    `gorm:"type:text;not null"`
	Used      bool      `gorm:"not null;default:false"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PasswordResetCode) TableName() string {
	return "password_reset_code"
}

type UserSession struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"uniqueIndex;size:96;not null"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserSession) TableName() string {
	return "user_session"
}

func (d *Database) FindUserByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return user, err
}

func (d *Database) FindUserByAccount(ctx context.Context, account string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("username = ? OR email = ?", account, account).First(&user).Error
	return user, err
}

func (d *Database) FindUserByID(ctx context.Context, id uint) (User, error) {
	var user User
	err := d.db.WithContext(ctx).First(&user, id).Error
	return user, err
}

func (d *Database) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := d.db.WithContext(ctx).Order("id asc").Find(&users).Error
	return users, err
}

func (d *Database) CreateUser(ctx context.Context, user *User) error {
	return d.db.WithContext(ctx).Create(user).Error
}

func (d *Database) UpdateUser(ctx context.Context, user *User, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(user).Updates(updates).Error
}

func (d *Database) DeleteUserSessions(ctx context.Context, userID uint) error {
	return d.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserSession{}).Error
}

func (d *Database) CreateSessionAndTouchUser(ctx context.Context, session *UserSession, user *User, now time.Time) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}
		return tx.Model(user).Update("last_login_at", now).Error
	})
}

func (d *Database) UserForSession(ctx context.Context, token string, now time.Time) (User, error) {
	var session UserSession
	err := d.db.WithContext(ctx).Preload("User").Where("token = ? AND expires_at > ?", token, now).First(&session).Error
	return session.User, err
}

func (d *Database) ReplacePasswordReset(ctx context.Context, reset *PasswordResetCode) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&PasswordResetCode{}).Where("user_id = ? AND used = ?", reset.UserID, false).Update("used", true).Error; err != nil {
			return err
		}
		return tx.Create(reset).Error
	})
}

func (d *Database) FindValidPasswordReset(ctx context.Context, userID uint, now time.Time) (PasswordResetCode, error) {
	var reset PasswordResetCode
	err := d.db.WithContext(ctx).Where("user_id = ? AND used = ? AND expires_at > ?", userID, false, now).Order("id desc").First(&reset).Error
	return reset, err
}

func (d *Database) CompletePasswordReset(ctx context.Context, user *User, reset *PasswordResetCode, passwordHash string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(user).Updates(map[string]any{"password_hash": passwordHash, "enabled": true}).Error; err != nil {
			return err
		}
		if err := tx.Model(reset).Update("used", true).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", user.ID).Delete(&UserSession{}).Error
	})
}

func (d *Database) HasUsers(ctx context.Context) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&User{}).Count(&count).Error
	return count > 0, err
}
