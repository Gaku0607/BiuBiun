package server

import (
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gaku/BiuBiun/dao"
	m "github.com/gaku/BiuBiun/model"
	t "github.com/gaku/BiuBiun/tool"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	FS *FoodServer = &FoodServer{}
)

type FoodServer struct {
}

//獲取指定的食品清單
func (this FoodServer) GetFoodList(Parms *m.QueryFoodParms) (List []map[string]interface{}, Counts uint, err error) {
	return dao.FDao.QueryConditionFood(Parms)
}

//取得該商店食品清單
func (this FoodServer) GetShopFoodList(ShopId int) (Foods []m.Food, Counts uint, err error) {
	return dao.FDao.SearchFoodlistById(ShopId)
}

//增添食品
func (this FoodServer) AddFoods(ShopId int, MemberId uint, f *[]m.Food) (err error) {
	return dao.FDao.AddFoods(ShopId, MemberId, f)
}

//修改食品
func (this FoodServer) ModityFoods(ShopId int, MemberId uint, f *[]m.MFoodParms) (err error) {
	return dao.FDao.ModityFoods(ShopId, MemberId, f)
}

//刪除食品
func (this FoodServer) DeleteFoods(ShopId int, MemberId uint, f *[]m.MFoodParms) (err error) {
	avatarPath, err := dao.FDao.DeleteFoods(ShopId, MemberId, f)
	for _, v := range avatarPath {
		if v != "" {
			os.Remove(ShopSrcPath + v)
		}
	}
	return
}

//儲存上傳檔案
func (this FoodServer) SaveUploadFile(ShopId int, MemberId uint, FoodName string, File *multipart.FileHeader, c *gin.Context) (err error) {
	//查看有無此路徑
	ShopDir := path.Join(ShopSrcPath, strconv.Itoa(ShopId))
	exists, err := t.PathExists(ShopDir)
	if err != nil {
		return
	}
	if !exists {
		if err = os.Mkdir(ShopDir, os.ModePerm); err != nil {
			return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
	}
	FilePath := path.Join(ShopDir, strconv.FormatInt(time.Now().Unix(), 10)+File.Filename)
	//儲存
	if err := c.SaveUploadedFile(File, FilePath); err != nil {
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	Food, err := dao.FDao.UpdataAvatar(ShopId, MemberId, FoodName, FilePath[len(ShopSrcPath):])
	if err != nil {
		//如果儲存數據庫失敗 刪除更新檔
		os.Remove(FilePath)
		return
	}
	if Food.IconPath != "" {
		//更新成功刪除舊檔案
		os.Remove(ShopSrcPath + Food.IconPath)
	}
	return
}
