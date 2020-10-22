package controller

import (
	midd "github.com/gaku/BiuBiun/middleware"
	m "github.com/gaku/BiuBiun/model"
	s "github.com/gaku/BiuBiun/server"

	t "github.com/gaku/BiuBiun/tool"

	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

type FoodController struct {
	*gin.RouterGroup
}

func NewFoodController(g *gin.RouterGroup) *FoodController {
	return &FoodController{RouterGroup: g}
}
func (this FoodController) Router() {
	TouristGroup := this.RouterGroup.Group("/api")
	{ //獲取該商店所有食品
		TouristGroup.GET("/shoplist/:shopid/foodlist", t.WrapperHandler(this.QueryShopFoodList))
		//獲取指定的食品清單
		TouristGroup.GET("/foodlist", t.WrapperHandler(this.QueryFoodList))
		//會員 尚未完成
		MemberGroup := TouristGroup.Group("/foodlist/:shopid/:foodname")
		MemberGroup.Use(midd.MemberStatus(false))
		{
			MemberGroup.POST("/addlist")
			MemberGroup.POST("/buy")
		}
		//賣家
		SellerGroup := TouristGroup.Group("/myshop/:shopid/foodlist")
		SellerGroup.Use(midd.MemberStatus(false), midd.IsSeller())
		{
			//獲取該商店所有食品
			SellerGroup.GET("/", t.WrapperHandler(this.QueryShopFoodList))
			//刪除該商店食品
			SellerGroup.DELETE("/", t.WrapperHandler(this.Delete))
			//添加該商店食品
			SellerGroup.POST("/", t.WrapperHandler(this.Add))
			//修改該商店食品
			SellerGroup.PUT("/", t.WrapperHandler(this.Modity))
			//修改該商店食品圖片
			SellerGroup.PUT("/uploadavatar", t.WrapperHandler(this.UploadAvatar))
		}
	}
}

//取得指定的食品清單
func (this FoodController) QueryFoodList(c *gin.Context) (data interface{}, err error) {
	var Parms m.QueryFoodParms
	if err := c.ShouldBindWith(&Parms, binding.Form); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	list, Counts, err := s.FS.GetFoodList(&Parms)
	if err != nil {
		return
	}
	return map[string]interface{}{"count": Counts, "list": &list}, err
}

//展示指定店家的食品清單
func (this FoodController) QueryShopFoodList(c *gin.Context) (data interface{}, err error) {
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	list, Counts, err := s.FS.GetShopFoodList(ShopId)
	if err != nil {
		return
	}
	return map[string]interface{}{"count": Counts, "list": &list}, err
}

//添加食品
func (this FoodController) Add(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	var Food []m.Food
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	if err := c.ShouldBindJSON(&Food); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	if err = s.FS.AddFoods(ShopId, MemberId, &Food); err != nil {
		return
	}
	return &Food, err
}

//修改食品
func (this FoodController) Modity(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	var Food []m.MFoodParms
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	if err = c.ShouldBindJSON(&Food); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	if err = s.FS.ModityFoods(ShopId, MemberId, &Food); err != nil {
		return
	}
	return &Food, err
}

//刪除食品
func (this FoodController) Delete(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	var Food []m.MFoodParms
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	if err = c.ShouldBindJSON(&Food); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	if err = s.FS.DeleteFoods(ShopId, MemberId, &Food); err != nil {
		return
	}
	return &Food, err
}

//上傳Avatar
func (this FoodController) UploadAvatar(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	FoodName := c.PostForm("foodname")
	File, err := c.FormFile("avatar")
	if err != nil {
		return data, m.NewAPIErr(m.ERROR_UPLOAD_FILE, errors.WithStack(err))
	}
	err = s.FS.SaveUploadFile(ShopId, MemberId, FoodName, File, c)
	return
}
