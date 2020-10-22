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
	SS *ShopServer = &ShopServer{}
)

type ShopServer struct {
}

//製作指定內容商店清單
func (this ShopServer) MekeShopList(Parms *m.QueryShopParms) (list []*m.Shop, Counts int, err error) {
	return dao.SDao.QueryConditionShop(Parms)
}

//取得商店內容
func (this ShopServer) GetShopInfo(ShopId int) (Shop m.Shop, err error) {
	return dao.SDao.SearchShopInfoByID(ShopId)
}

//取得所有商店內容（Seller）
func (this ShopServer) GetShopsInfo(MemberId uint) (shops []m.Shop, err error) {
	return dao.SDao.SearchShopsByMemberId(MemberId)
}

//增加商店
func (this ShopServer) AddShop(s *m.Shop, MemberId uint) (err error) {
	s.MemberID = MemberId
	return dao.SDao.CreateShop(s)
}

//上傳檔案
func (this ShopServer) SaveUploadFile(ShopId int, MemberId uint, File *multipart.FileHeader, c *gin.Context) (err error) {
	//查看有無此路徑
	ShopDir := path.Join(ShopSrcPath, strconv.Itoa(ShopId))
	//創建失敗時愈刪除路徑
	RemovePath := ""
	//File路徑
	FilePath := path.Join(ShopDir, strconv.FormatInt(time.Now().Unix(), 10)+File.Filename)
	RemovePath = FilePath
	exists, err := t.PathExists(ShopDir)
	if err != nil {
		return
	}
	if !exists {
		if err = os.Mkdir(ShopDir, os.ModePerm); err != nil {
			return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
		RemovePath = ShopDir
	}
	//儲存檔案
	if err := c.SaveUploadedFile(File, FilePath); err != nil {
		os.Remove(RemovePath)
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	Shop, err := dao.SDao.UpdateAvatar(ShopId, MemberId, FilePath[len(ShopSrcPath):])
	if err != nil {
		//如果儲存數據庫失敗 刪除更新檔
		os.RemoveAll(RemovePath)
		return
	}
	if Shop.IconPath != "" {
		//更新成功刪除上一個檔案
		os.Remove(ShopSrcPath + Shop.IconPath)
	}
	return
}

//修改商店
func (this ShopServer) ModifyShop(s *m.MShopParms, ShopId int, MemberId uint) (err error) {
	return dao.SDao.ModifyShopInfo(s, ShopId, MemberId)
}

//刪除相關所有文件
func (this ShopServer) DelRelatedFile(ShopId int, MemberId uint) (err error) {
	err = dao.SDao.DelShopInfo(ShopId, MemberId)
	if err != nil {
		return
	}
	//刪除相關文件
	os.RemoveAll(path.Join(ShopSrcPath, strconv.Itoa(ShopId)))
	return
}
