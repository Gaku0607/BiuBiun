package model

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

// JWT
type Claims struct {
	UserName string
	MemberId uint
	Level    Level
	IsSeller *bool
	jwt.StandardClaims
}
type Member struct {
	gorm.Model  `json:"-"`
	UserName    string      `json:"userName" form:"userName" gorm:"type:varchar(20);not null;unique"`
	Password    string      `json:"password" form:"password" gorm:"type:varchar(255);not null;unique" `
	Avatar      string      `json:"avatar" form:"avatar" gorm:"type:varchar(255)"`
	City        string      `json:"city" form:"city" gorm:"type:varchar(10)"`
	Balanc      float64     `json:"balanc" form:"balanc" gorm:"default:0"`
	MemberLevel MemberLevel `json:"memberLevel" `
	Email       Email       `json:"email" form:"email"`
	IsSeller    *bool       `json:"isSeller" gorm:"default:0"`
	Shops       []Shop      `json:"shop"`
}

type Email struct {
	MemberID  uint      `json:"memberId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Addrs     string    `json:"addrs" gorm:"type:varchar(255);primary_key;not null;unique" binding:"email"`
}

type MemberLevel struct {
	MemberID  uint      `json:"memberid"`
	UpdatedAt time.Time `json:"updatedat"`
	ExpiresAt time.Time `json:"expiresat" gorm:"default:0"`
	Level     Level     `json:"level" form:"level" gorm:"default:2" binding:"omitempty,oneof=0 1 2"`
	Point     uint32    `json:"point" gorm:"default:0"`
}

const (
	Viplevel Level = iota
	Goldlevel
	Memberlevel
)

type Level uint32
