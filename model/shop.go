package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Shop struct {
	MemberID    uint `json:"memberId" gorm:"not null"`
	gorm.Model  `json:"model"`
	ShopName    string         `json:"shopName" gorm:"type:varchar(20);not null" binding:"required"`
	Address     string         `json:"address" gorm:"type:varchar(100);not null;unique" binding:"required"`
	Phone       string         `json:"phone" gorm:"type:varchar(10);not null;unique"`
	Status      *bool          `json:"status" gorm:"type:bool;default:0"`
	IconPath    string         `json:"iconPath" gorm:"type:varchar(100)"`
	IsPremium   *bool          `josn:"isPremium" gorm:"type:bool;default:0"`
	Rating      float32        `json:"rating" gorm:"type:float;default:0"`
	RatingCount uint           `json:"ratingCount" gorm:"default:0"`
	Category    []ShopCategory `json:"category" binding:"max=3"`
	FoodList    []Food         `json:"foodlist"`
}

type Food struct {
	ShopID    uint   `json:"shopId" gorm:"primary_key;auto_increment:false"`
	FoodName  string `json:"foodName" gorm:"type:varchar(30);not null;primary_key" binding:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Status    *bool      `json:"status" gorm:"type:bool;default:0"`
	Price     float64    `json:"price" gorm:"type:double;not null"`
	IconPath  string     `json:"iconPath" gorm:"varchar(100)"`
}
type ShopCategory struct {
	ShopID   uint     `json:"-" gorm:"primary_key;auto_increment:false"`
	Category Category `json:"category" form:"category" binding:"omitempty,gt=1,lte=10" gorm:"primary_key;auto_increment:false"`
}

type Category uint8

const (
	Cafe Category = iota + 1
	Taiwanese
	Continental
	American
	Japanese
	Breakfast
	Yakiniku
	Steak
	HotPot
	Bar
)
