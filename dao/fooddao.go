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
	FDao *FoodDao
)

const (
	MAX_FOOD_COUNTS = 10
)

type FoodDao struct {
	*gorm.DB
}

func NewFoodDao() *FoodDao {
	return &FoodDao{DB: initialization.GetDB()}
}
func (this FoodDao) QueryConditionFood(Parms *m.QueryFoodParms) (list []map[string]interface{}, Count uint, err error) {
	var sql bytes.Buffer
	//查詢需要的參數
	sql.WriteString("select f.food_name,f.price,f.status,s.shop_name,s.id from food as f join shop as s on f.shop_id=s.id ")
	sql.WriteString("where s.status=1 and s.deleted_at is null ")
	//Status不為默認值時只搜索 Status為1的值
	if Parms.Status != 0 {
		sql.WriteString("and f.status=1")
	}
	//價格最大值設置
	if Parms.MaximumPrice != 0 {
		sql.WriteString(strings.Join([]string{" and f.price <=", strconv.FormatUint(Parms.MaximumPrice, 10)}, " "))
	}
	//價格最小值設置
	if Parms.MinimumPrice != 0 {
		sql.WriteString(strings.Join([]string{" and f.price >=", strconv.FormatUint(Parms.MinimumPrice, 10)}, " "))
	}
	//排序已商店優良度決定
	sql.WriteString(" Order by s.is_premium Desc ")
	//價格排序
	if Parms.PriceDesc != 0 {
		sql.WriteString(",f.price ")
		//大至小 or 小至大
		if Parms.PriceIsAsc == 0 {
			sql.WriteString("Desc ")
		} else {
			sql.WriteString("Asc ")
		}
	}
	//最大值為10
	sql.WriteString(strings.Join([]string{"limit 10 offset ", strconv.FormatUint(Parms.Offset, 10)}, ""))
	fmt.Println(sql.String())
	Rows, err := this.Raw(sql.String()).Rows()
	if err != nil {
		return list, Count, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	Columns, err := Rows.Columns()
	if err != nil {
		return list, Count, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	defer Rows.Close()
	//這裡使用map[string]interface{} 也可以用對應的struck
	for Rows.Next() {
		columns := make([]interface{}, len(Columns))
		columnPointers := make([]interface{}, len(Columns))
		for index, _ := range columns {
			columnPointers[index] = &columns[index]
		}
		if err = Rows.Scan(columnPointers...); err != nil {
			return list, Count, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
		var m map[string]interface{} = make(map[string]interface{})
		for index, ColName := range Columns {
			val := columnPointers[index].(*interface{})
			m[ColName] = string((*val).([]byte))
		}
		list = append(list, m)
		Count++
	}
	return
}

//查詢商店食品總數
func (this FoodDao) QueryShopFoodCounts(ShopId int) (Counts int) {
	return this.Model(&m.Shop{
		Model: gorm.Model{
			ID: uint(ShopId),
		},
	}).Association("FoodList").Count()
}

//查詢有無重複值
func (this FoodDao) CheackRepectFoodInfo(ShopId int, f *[]m.Food) (err error) {
	var Food m.Food
	for _, v := range *f {
		if !this.Unscoped().Select("shop_id,food_name").Scopes(SearchFoodByName(ShopId, v.FoodName)).First(&Food).RecordNotFound() {
			return m.NewAPIErr(m.ERROR_FOOD_EXISTS,
				errors.New(fmt.Sprintf("food already exist FoodName: %s", v.FoodName)))
		}
	}
	return
}

//確認該賣家是否擁有此商店
func (this FoodDao) IsShopExists(ShopId int, MemberId uint) (err error) {
	if this.Scopes(SearchShop(ShopId, MemberId)).Select("id").First(&m.Shop{}).RecordNotFound() {
		return m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS,
			errors.New(fmt.Sprintf("MemberId: %d operations on the shop: %d  ", MemberId, ShopId)))
	}
	return
}

//添增食品
func (this FoodDao) AddFoods(ShopId int, MemberId uint, f *[]m.Food) (err error) {
	if err = this.IsShopExists(ShopId, MemberId); err != nil {
		return
	}
	Counts := this.QueryShopFoodCounts(ShopId)
	if Counts+len(*f) > MAX_FOOD_COUNTS {
		return m.NewAPIErr(m.ERROR_OUT_OF_FOOD_RANGE,
			errors.New(fmt.Sprintf("over range MaxFoodCounts: %d", MAX_FOOD_COUNTS)))
	}
	if err = this.CheackRepectFoodInfo(ShopId, f); err != nil {
		return
	}
	tx := this.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(r.(error)))
		}
	}()
	if err := tx.Model(
		&m.Shop{
			Model: gorm.Model{
				ID: uint(ShopId),
			},
		}).Association("FoodList").Append(f).Error; err != nil {
		tx.Rollback()
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	tx.Commit()
	return
}

//修改食品內容
func (this FoodDao) ModityFoods(ShopId int, MemberId uint, f *[]m.MFoodParms) (err error) {
	if err = this.IsShopExists(ShopId, MemberId); err != nil {
		return
	}
	tx := this.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(r.(error)))
		}
	}()
	for _, food := range *f {
		if result := tx.Scopes(SearchFoodByName(ShopId, food.FoodName)).
			Updates(food); result.Error != nil || result.RowsAffected == 0 {
			tx.Rollback()
			return m.NewAPIErr(m.ERROR_SERVER_FAILD,
				errors.New(fmt.Sprintf("ShopId :%d update FoodName: %s failed", ShopId, food.FoodName)))
		}
	}
	tx.Commit()
	return
}

//依據ShopID查詢食品列表
func (this FoodDao) SearchFoodlistById(ShopId int) (Foods []m.Food, Counts uint, err error) {
	//Scan不會觸發RecordNotFoundErr
	if err := this.Model(&m.Food{}).Where("shop_id = ?", ShopId).Scan(&Foods).Count(&Counts).Error; err != nil {
		return Foods, Counts, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	return
}

//刪除食品
func (this FoodDao) DeleteFoods(ShopId int, MemberId uint, f *[]m.MFoodParms) (avatarPaths []string, err error) {
	if err = this.IsShopExists(ShopId, MemberId); err != nil {
		return
	}
	var avatarPath []string
	for _, food := range *f {
		if err := this.Scopes(SearchFoodByName(ShopId, food.FoodName)).Pluck("icon_path", &avatarPath).
			Error; err != nil {
			return avatarPath, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
		avatarPaths = append(avatarPaths, avatarPath...)
	}
	tx := this.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(r.(error)))
		}
	}()
	for _, food := range *f {
		if err := tx.Scopes(SearchFoodByName(ShopId, food.FoodName)).Delete(&m.Food{}).Error; err != nil {
			tx.Rollback()
			return avatarPath, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
		}
	}
	tx.Commit()
	return
}

//上傳Icon
func (this FoodDao) UpdataAvatar(ShopId int, MemberId uint, FoodName, FilePath string) (Food m.Food, err error) {
	if err = this.IsShopExists(ShopId, MemberId); err != nil {
		return
	}
	if this.Select("icon_path").Scopes(SearchFoodByName(ShopId, FoodName)).First(&Food).RecordNotFound() {
		return Food, m.NewAPIErr(m.ERROR_NOSUCH_DATA,
			errors.New(fmt.Sprintf("FoodName:%s that does not exists in the Shop:%d ", FoodName, ShopId)))
	}
	if Counts := this.Scopes(SearchFoodByName(ShopId, FoodName)).UpdateColumn("icon_path", FilePath).RowsAffected; Counts == 0 {
		return Food, m.NewAPIErr(m.ERROR_SERVER_FAILD,
			errors.New(fmt.Sprintf("ShopId :%d update FoodAvatar failed FoodName: %s", ShopId, FoodName)))
	}
	return
}

func SearchFoodByName(ShopId int, FoodName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Model(&m.Food{}).Where("shop_id = ? and food_name = ?", ShopId, FoodName)
	}
}
