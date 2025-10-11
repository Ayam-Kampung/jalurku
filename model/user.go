package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model tabel pengguna
type User struct {
	ID        	uuid.UUID      	`gorm:"type:char(36);primaryKey"`
	Name      	string         	`gorm:"type:varchar(100);not null"`
	Email     	string         	`gorm:"type:varchar(100);unique;not null"`
	Password  	string         	`gorm:"type:varchar(255);not null"`
	Role      	string         	`gorm:"type:varchar(20);default:'user'"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 	`gorm:"index"`

	HasilAngket []HasilAngket 	`gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName overrides
func (User) TableName() string {
	return "users"
}