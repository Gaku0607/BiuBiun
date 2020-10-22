package initialization

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gaku/BiuBiun/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	Mysql *gorm.DB
)

func InitDB() (err error) {

	//Mysql conf
	dbPassword := os.Getenv("dbPassword")
	dbHost := os.Getenv("dbHost")
	dbName := os.Getenv("dbName")
	tempdbConnMaxage := os.Getenv("dbConnMaxage")
	tempdbMaxOpenConns := os.Getenv("dbMaxOpenConns")
	tempdbMaxIdConns := os.Getenv("dbMaxIdConns")
	tempGormLogMode := os.Getenv("GormLogMode")
	//轉為int型
	dbMaxOpenConns, _ := strconv.Atoi(tempdbMaxOpenConns)
	dbMaxIdConns, _ := strconv.Atoi(tempdbMaxIdConns)
	dbConnMaxage, _ := strconv.Atoi(tempdbConnMaxage)
	//轉為bool
	GormLogMode, err := strconv.ParseBool(tempGormLogMode)
	if err != nil {
		return
	}
	DBPath := fmt.Sprintf("root:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbPassword, dbHost, dbName)
	Mysql, err = gorm.Open("mysql", DBPath)
	if err != nil {
		return errors.New("mysql start failed err")
	}
	Mysql.SingularTable(true)  //默認s去除
	Mysql.LogMode(GormLogMode) //開啟日誌
	Mysql.AutoMigrate(&model.Member{}, &model.MemberLevel{}, &model.Email{},
		&model.Shop{}, &model.ShopCategory{}, &model.Food{})
	//設置連接持參數
	Mysql.DB().SetMaxOpenConns(dbMaxOpenConns)
	Mysql.DB().SetMaxIdleConns(dbMaxIdConns)
	Mysql.DB().SetConnMaxLifetime(time.Second * time.Duration(dbConnMaxage))
	return
}

func GetDB() *gorm.DB {
	return Mysql
}
