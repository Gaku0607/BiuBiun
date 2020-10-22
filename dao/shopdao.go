package dao

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/gaku/BiuBiun/initialization"
	m "github.com/gaku/BiuBiun/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	SDao *ShopDao
)

const (
	MAX_SHOP_COUNTS int = 3
)

type ShopDao struct {
	*gorm.DB
}

func NewShopDao() *ShopDao {
	return &ShopDao{DB: initialization.GetDB()}
}

//分頁條件查詢商家列表 最多10筆
func (this ShopDao) QueryConditionShop(Parms *m.QueryShopParms) (Shops []*m.Shop, Count int, err error) {
	var sql bytes.Buffer
	//要呈現的商家資訊
	sql.WriteString("select s.id,s.shop_name,s.member_id,s.address,s.phone," +
		"s.icon_path,s.is_premium,s.rating,s.rating_count,s.status from shop as s ")
	//有類型條件的話 遞歸拼接join語法
	if len(Parms.ShopCategory) != 0 {
		sqlstr := this.SearchShopByCategorys(1, &Parms.ShopCategory,
			fmt.Sprintf("select a.shop_id from shop_category as a where a.category=%d", Parms.ShopCategory[0]))
		sql.WriteString(sqlstr)
	}
	sql.WriteString("where s.deleted_at is null ")
	//Status不為默認值時 將搜索方式改為 status=1
	if Parms.Status != 0 {
		sql.WriteString("and s.status=1 ")
	}
	//排序方式
	switch Parms.Desc {
	case 0:
		sql.WriteString(" Order by s.is_premium Desc, s.rating ")
	case 1:
		sql.WriteString(" Order by s.is_premium Desc, s.rating_count ")
	case 2:
		sql.WriteString(" Order by s.is_premium Desc, s.id ")
	}
	if Parms.IsAsc == 0 {
		sql.WriteString("Desc ")
	} else {
		sql.WriteString("Asc ")
	}
	sql.WriteString(strings.Join([]string{" limit 10 offset ", strconv.FormatUint(Parms.Offset, 10)}, ""))
	rows, err := this.Raw(sql.String()).Rows()
	if err != nil {
		return Shops, Count, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	defer rows.Close()
	for rows.Next() {
		Shop := &m.Shop{}
		err = this.ScanRows(rows, Shop)
		if err != nil {
			return Shops, Count, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
		err = this.GetShopCategory(Shop)
		if err != nil {
			return Shops, Count, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
		Shops = append(Shops, Shop)
	}
	Count = len(Shops)
	return
}

//取得商家類型
func (this ShopDao) GetShopCategory(s *m.Shop) (err error) {
	if err := this.Model(&s).Select("category").Related(&s.Category).
		Error; err != nil && err != gorm.ErrRecordNotFound {
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	return
}

//查詢複數類型商店
func (this ShopDao) SearchShopByCategorys(index int, arr *[]m.Category, sqlstr string) string {
	if index == len(*arr) {
		newsql := strings.Join([]string{"join (", sqlstr, ") as a on s.id=a.shop_id "}, "")
		return newsql
	}
	newsql := strings.Join([]string{"select a.shop_id from(", sqlstr, ")as a ",
		"join shop_category as s on a.shop_id=s.shop_id where s.category=", strconv.FormatUint(uint64((*arr)[index]), 10)}, "")
	return this.SearchShopByCategorys(index+1, arr, newsql)
}

//取得商家所有信息
func (this ShopDao) SearchShopInfoByID(ShopId int) (Shop m.Shop, err error) {
	if this.Where("id = ?", ShopId).Preload("Category").First(&Shop).
		RecordNotFound() {
		return Shop, m.NewAPIErr(m.ERROR_NOSUCH_DATA, errors.New(fmt.Sprintf("not found ShopId :%d", ShopId)))
	}
	return
}

//取得會員的所有商家信息
func (this ShopDao) SearchShopsByMemberId(MemberId uint) (Shops []m.Shop, err error) {
	if err := this.Model(&m.Shop{}).Where("member_id = ?", MemberId).Preload("Category").Scan(&Shops).Error; err != nil {
		return Shops, m.NewAPIErr(m.ERROR_NOSUCH_DATA, errors.New(fmt.Sprintf("not found MemberId :%d", MemberId)))
	}
	return
}

//創建商家
func (this ShopDao) CreateShop(s *m.Shop) (err error) {
	Counts := this.QuerySellerShopCounts(s.MemberID)
	if Counts >= MAX_SHOP_COUNTS {
		return m.NewAPIErr(m.ERROR_OUT_OF_SHOP_RANGE, errors.New(fmt.Sprintf("over range MaxShopCount : %d", MAX_SHOP_COUNTS)))
	}
	if err = this.CheackRepectShopInfo(s.Address, s.Phone); err != nil {
		return
	}
	if err := this.Model(&m.Member{
		Model: gorm.Model{
			ID: s.MemberID,
		},
	}).Association("Shops").Append(s).Error; err != nil && err != gorm.ErrRecordNotFound {
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	return
}

//查詢賣家商店數
func (this ShopDao) QuerySellerShopCounts(MemberId uint) (Counts int) {
	return this.Model(&m.Member{
		Model: gorm.Model{
			ID: MemberId,
		},
	}).Association("Shops").Count()
}

//查詢有無重複內容 地址 電話等等
func (this ShopDao) CheackRepectShopInfo(Address, Phone string) (err error) {
	var Shop m.Shop
	//查無正確返回
	if !this.Unscoped().Select("address, phone").Where("address = ? or phone = ?", Address, Phone).First(&Shop).RecordNotFound() {
		if Shop.Address == Address {
			return m.NewAPIErr(m.ERROR_ADDRESS_EXISTS, errors.New("address already exist"))
		} else {
			return m.NewAPIErr(m.ERROR_PHONE_EXISTS, errors.New("phone already exist"))
		}
	}
	return
}

//確認該賣家是否擁有此商店
func (this ShopDao) IsShopExists(ShopId int, MemberId uint) (err error) {
	if this.Scopes(SearchShop(ShopId, MemberId)).Select("id").First(&m.Shop{}).RecordNotFound() {
		return m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS,
			errors.New(fmt.Sprintf("MemberId: %d operations on the shopId: %d  ", MemberId, ShopId)))
	}
	return
}

//修改商店內容
func (this ShopDao) ModifyShopInfo(s *m.MShopParms, ShopId int, MemberId uint) (err error) {
	if err = this.IsShopExists(ShopId, MemberId); err != nil {
		return
	}
	if err = this.CheackRepectShopInfo(s.Address, s.Phone); err != nil {
		return
	}
	tx := this.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(r.(error)))
		}
	}()
	if len(s.Category) != 0 {
		//刪除關聯在更新
		if err := tx.Where("shop_id = ?", ShopId).Delete(&m.ShopCategory{}).Error; err != nil {
			tx.Rollback()
			return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
	}
	if result := tx.Scopes(SearchShop(ShopId, MemberId)).Omit("Category").Updates(s); result.RowsAffected == 0 {
		tx.Rollback()
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.New(fmt.Sprintf("update failed ShopId: %d", ShopId)))
	}
	if err := tx.Model(&m.Shop{
		Model: gorm.Model{
			ID: uint(ShopId),
		},
	}).Association("Category").Append(s.Category).Error; err != nil {
		tx.Rollback()
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	tx.Commit()
	return
}

//更新商店 商標
func (this ShopDao) UpdateAvatar(ShopId int, MemberId uint, AvatarPath string) (Shop m.Shop, err error) {
	if this.Scopes(SearchShop(ShopId, MemberId)).Select("icon_path").First(&Shop).
		RecordNotFound() {
		return Shop, m.NewAPIErr(m.ERROR_NOSUCH_DATA, errors.New(fmt.Sprintf("not found ShopId : %d", ShopId)))
	}
	if result := this.Scopes(SearchShop(ShopId, MemberId)).UpdateColumn("icon_path", AvatarPath); result.RowsAffected == 0 {
		return Shop, m.NewAPIErr(m.ERROR_SERVER_FAILD,
			errors.New(fmt.Sprintf("update ShopAvatar Failed ShopId : %d", ShopId)))
	}
	return
}

//刪除商店所有內容
func (this ShopDao) DelShopInfo(ShopId int, MemberId uint) (err error) {
	var Shop m.Shop
	//查詢 商店相關內容
	if this.Scopes(SearchShop(ShopId, MemberId)).Preload("FoodList").First(&Shop).RecordNotFound() {
		return m.NewAPIErr(m.ERROR_NOSUCH_DATA, errors.New(fmt.Sprintf("not found ShopId : %d", ShopId)))
	}
	//開啟事務
	tx := this.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(r.(error)))
		}
	}()
	//當存在著商品列表一併刪除
	if len(Shop.FoodList) != 0 {
		//刪除 關聯食品資料
		if err = tx.Where("shop_id = ?", ShopId).Delete(&m.Food{}).Error; err != nil {
			tx.Rollback()
			return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
	}
	//刪除店家資訊
	if err = tx.Where("shop_id = ?", ShopId).Delete(&m.ShopCategory{}).Error; err != nil {
		tx.Rollback()
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	//刪除店家
	if err = tx.Scopes(SearchShop(ShopId, MemberId)).Delete(&m.Shop{}).Error; err != nil {
		tx.Rollback()
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	tx.Commit()
	return
}

//查詢店家
func SearchShop(ShopId, MemberId interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Model(&m.Shop{}).Where("member_id = ? and id = ?", MemberId, ShopId)
	}
}
