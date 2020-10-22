package controller

import (
	"fmt"
	"strconv"

	midd "github.com/gaku/BiuBiun/middleware"
	m "github.com/gaku/BiuBiun/model"
	s "github.com/gaku/BiuBiun/server"
	t "github.com/gaku/BiuBiun/tool"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
)

type ShopController struct {
	*gin.RouterGroup
}

func NewShopController(g *gin.RouterGroup) *ShopController {
	return &ShopController{
		RouterGroup: g,
	}
}

func (this *ShopController) Router() {
	//遊客
	TouristGroup := this.RouterGroup.Group("/api")
	{
		//獲取指定的商店清單
		TouristGroup.GET("/shoplist", t.WrapperHandler(this.ShowShoplist))
		//獲取指定的商店內容
		TouristGroup.GET("/shoplist/:shopid", t.WrapperHandler(this.GetShopInfo))

		//有賣家資格才可以訪問
		SellerGroup := TouristGroup.Group("/myshop")
		SellerGroup.Use(midd.MemberStatus(false), midd.IsSeller())
		{
			//獲取該賣家所有商店內容
			SellerGroup.GET("/", t.WrapperHandler(this.GetShopsInfo))
			//更新該商店頭像
			SellerGroup.PUT("/:shopid/uploadavatar", t.WrapperHandler(this.UploadAvatar))
			//避免路由相撞使用 :shopid  參數必為create
			SellerGroup.POST("/:shopid", t.WrapperHandler(this.Create))
			//修改指定的商店內容
			SellerGroup.PUT("/:shopid", t.WrapperHandler(this.Modity))
			//刪除指定商店
			SellerGroup.DELETE("/:shopid", t.WrapperHandler(this.Close))
		}
	}
}

//依據格式返回Shop清單
func (this ShopController) ShowShoplist(c *gin.Context) (data interface{}, err error) {
	var p m.QueryShopParms
	if err = c.ShouldBindWith(&p, binding.Form); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	list, Counts, err := s.SS.MekeShopList(&p)
	if err != nil {
		return
	}
	return map[string]interface{}{"shoplist": list, "counts": Counts}, err
}

//取得商店所有內容
func (this ShopController) GetShopInfo(c *gin.Context) (data interface{}, err error) {
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	return s.SS.GetShopInfo(ShopId)
}

//取得所有商店所有信息（Seller）
func (this ShopController) GetShopsInfo(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	return s.SS.GetShopsInfo(MemberId)
}

//建立商店
func (this ShopController) Create(c *gin.Context) (data interface{}, err error) {
	if Path := c.Param("shopid"); Path != "create" {
		return data, m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS, errors.New(fmt.Sprintf("URL Path Failed: %s", Path)))
	}
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	var shop m.Shop
	if err = c.ShouldBindJSON(&shop); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	if err = s.SS.AddShop(&shop, MemberId); err != nil {
		return
	}
	return &shop, err
}

//更新Icon
func (this ShopController) UploadAvatar(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	File, err := c.FormFile("avatar")
	if err != nil {
		return data, m.NewAPIErr(m.ERROR_UPLOAD_FILE, errors.WithStack(err))
	}
	err = s.SS.SaveUploadFile(ShopId, MemberId, File, c)
	return
}

//修改商店信息
func (this ShopController) Modity(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	var Shop m.MShopParms
	if err := c.ShouldBindJSON(&Shop); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	if err = s.SS.ModifyShop(&Shop, ShopId, MemberId); err != nil {
		return
	}
	return &Shop, err
}

//關閉商店
func (this ShopController) Close(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	ShopId, err := GetShopId(c)
	if err != nil {
		return
	}
	if err = s.SS.DelRelatedFile(ShopId, MemberId); err != nil {
		return
	}
	return
}

//取得Parms上的ShopId
func GetShopId(c *gin.Context) (ShopId int, err error) {
	tempShopId := c.Param("shopid")
	if ShopId, err = strconv.Atoi(tempShopId); err != nil {
		err = m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS, errors.WithStack(err))
	}
	return
}
