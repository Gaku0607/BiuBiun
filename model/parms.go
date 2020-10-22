package model

type RegisterParms struct {
	UserName   string `json:"userName" form:"userName"  binding:"required"`
	Password   string `json:"password" form:"password"  binding:"required,alphanum"`
	RePassword string `json:"repassword" binding:"required,eqfield=Password"`
	City       string `json:"city" form:"city"  binding:"required"`
	Email      Email  `json:"email" form:"email" binding:"required,email"`
}

type LoginParms struct {
	UserName string `json:"userName" form:"userName" binding:"required"`
	PassWord string `json:"password" form:"password" binding:"required,alphanum"`
}

// ModifyShopParms
type MShopParms struct {
	ShopName string         `json:"shopName" `
	Address  string         `json:"address"`
	Phone    string         `json:"phone"`
	Status   *bool          `json:"status"`
	Category []ShopCategory `json:"category" binding:"max=3"`
}

// ModifyFoodParms
type MFoodParms struct {
	Status   uint8   `json:"status" form:"status" binding:"omitempty,oneof=0 1"`
	FoodName string  `json:"foodName" form:"food" binding:"required"`
	Price    float64 `json:"price" form:"price" binding:"omitempty,numeric"`
}

//Desc 0 Rating  1 RatingCount  2 Id
type QueryShopParms struct {
	Offset       uint64     `json:"offset" form:"offset" binding:"omitempty,numeric"`
	ShopCategory []Category `json:"shopCategory" form:"shopCategory" binding:"omitempty,SSP"`
	Status       uint8      `json:"status" form:"status" binding:"omitempty,oneof=0 1"`
	IsAsc        uint8      `json:"isAsa" form:"isAsc" binding:"omitempty,oneof=0 1"`
	Desc         uint8      `json:"desc" form:"desc" binging:"omitemty,oneof=0 1 2"`
}

type QueryFoodParms struct {
	Offset       uint64 `json:"offset" form:"offset" binding:"omitempty,numeric"`
	Status       uint8  `json:"status" form:"status" binding:"omitempty,oneof=0 1"`
	PriceIsAsc   uint8  `json:"isAsa" form:"isAsc" binding:"omitempty,oneof=0 1"`
	PriceDesc    uint8  `json:"desc" form:"desc" binging:"omitemty,oneof=0 1"`
	MaximumPrice uint64 `json:"max" form:"max" binding:"omitempty,numeric"`
	MinimumPrice uint64 `json:"min" form:"min" binding:"omitempty,numeric"`
}
