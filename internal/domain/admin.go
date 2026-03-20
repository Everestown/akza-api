package domain

import "time"

type Admin struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Name         string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

func (Admin) TableName() string { return "admins" }
