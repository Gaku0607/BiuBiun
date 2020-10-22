package main

import (
	"net/http"
	"os"

	"github.com/gaku/BiuBiun/controller"
	"github.com/gaku/BiuBiun/dao"
	midd "github.com/gaku/BiuBiun/middleware"
	_ "github.com/gaku/BiuBiun/myvalidator"
	"github.com/gaku/BiuBiun/server"

	"github.com/gaku/BiuBiun/initialization"
	"github.com/gin-gonic/gin"
)

func main() {
	//初始化會員庫
	dao.MDao = dao.NewMemberDao()
	//初始化會員伺服器
	server.MS = server.NewMemberServer()
	//初始化商店庫
	dao.SDao = dao.NewShopDao()
	//初始化食物庫
	dao.FDao = dao.NewFoodDao()
	//gin conf
	ginHost := os.Getenv("ginHost")
	ginPost := os.Getenv("ginPost")
	ginMode := os.Getenv("ginMode")
	gin.SetMode(ginMode)
	g := gin.Default()
	//添加Cose路由
	g.Use(midd.GinLogger(), midd.Cors())
	//v1 !!!
	v1 := g.Group("/v1")
	//初始化會員系統
	mc := controller.NewMerberContorller(v1)
	//初始化商店系統
	sc := controller.NewShopController(v1)
	//初始化商店系統
	fc := controller.NewFoodController(v1)
	//會員路由
	sc.Router()
	//商店路由
	mc.Router()
	//食品路由
	fc.Router()
	//設定無效路由
	g.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, ":((")
	})
	err := g.Run(ginHost + ginPost)
	if err != nil {
		panic(err.Error())
	}
	defer initialization.GetDB().Close()
	defer initialization.GetRedis().Close()
}
