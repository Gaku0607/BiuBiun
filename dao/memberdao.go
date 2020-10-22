package dao

import (
	"fmt"
	"strings"

	"github.com/gaku/BiuBiun/initialization"
	m "github.com/gaku/BiuBiun/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	MDao *MemberDao
)

type MemberDao struct {
	*gorm.DB
}

func NewMemberDao() *MemberDao {
	return &MemberDao{
		DB: initialization.GetDB(),
	}
}

//ID查詢會員(查看登錄狀態)
func (this MemberDao) SearchMemberById(MemberId uint) (err error) {
	var Counts uint
	if this.Scopes(SearchByMemberId(MemberId)).Count(&Counts); Counts == 0 {
		return m.NewAPIErr(m.ERROR_USER_NOTEXISTS,
			errors.New(fmt.Sprintf("not found MemberId: %d", MemberId)))
	}
	return
}

//登入用
func (this MemberDao) GetMember(MemberName string) (mb m.Member, err error) {
	mb, err = this.SearchMemberByName(MemberName)
	if err != nil {
		return
	}
	return mb, this.GetMemberLevel(&mb)
}

//名稱查詢會員 (用於處理 登入 以及 註冊確認重複)
func (this MemberDao) SearchMemberByName(MemberName string) (mb m.Member, err error) {
	if this.Where("user_name = ?", MemberName).First(&mb).RecordNotFound() {
		return mb, m.NewAPIErr(m.ERROR_USER_NOTEXISTS,
			errors.New(fmt.Sprintf("not found MemberName: %s", MemberName)))
	}
	return
}

//取得會員等級紅利信息
func (this MemberDao) GetMemberLevel(mb *m.Member) (err error) {
	if this.Model(&m.MemberLevel{}).Scopes(MemberLevelInfo()).Where("member_id = ?", mb.ID).First(&mb.MemberLevel).RecordNotFound() {
		return m.NewAPIErr(m.ERROR_NOSUCH_DATA,
			errors.New(fmt.Sprintf("not found MemberId: %d", mb.ID)))
	}
	return
}

//取得會員相關信息 個資 SHOP SHOPPINGCAR
func (this MemberDao) GetMemberInfo(MemberId uint) (mb m.Member, err error) {
	if this.Where("id = ?", MemberId).Preload("Shops").Preload("Shops.Category").Preload("MemberLevel", MemberLevelInfo()).
		Preload("Email").First(&mb).RecordNotFound() {
		return mb, m.NewAPIErr(m.ERROR_NOSUCH_DATA,
			errors.New(fmt.Sprintf("not found MemberId: %d", MemberId)))
	}
	mb.Password = ""
	return
}

//驗證註冊內容有無重複
func (this MemberDao) CheckRegisterMsgInfo(mb *m.RegisterParms) (err error) {
	if _, err = this.SearchMemberByName(mb.UserName); err != nil {
		if this.Where("addrs = ?", mb.Email.Addrs).First(&m.Email{}).RecordNotFound() {
			return nil
		}
		return m.NewAPIErr(m.ERROR_EMAIL_EXISTS,
			errors.New(fmt.Sprintf("Email already exist %s", mb.Email.Addrs)))
	}
	//當err為nil時 代表已有用戶使用
	if err == nil {
		return m.NewAPIErr(m.ERROR_USER_EXISTS,
			errors.New(fmt.Sprintf("UserName already exist %s", mb.UserName)))
	}
	return
}

//建立用戶
func (this MemberDao) CreateMember(mb *m.Member) (err error) {
	if err := this.Create(mb).Error; err != nil {
		if strings.HasPrefix(fmt.Sprintf("%s", err), "Error 1062: Duplicate entry") {
			//建立失敗時 確認 失敗原因
			var newmb m.RegisterParms
			newmb.UserName = mb.UserName
			newmb.Email.Addrs = mb.Email.Addrs
			return this.CheckRegisterMsgInfo(&newmb)
		}
		err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	return
}

//上傳用戶頭貼
func (this MemberDao) UpdataAvatar(MemberId uint, FilePath string) (mb m.Member, err error) {
	//查出是否有被替換資料有的話查出以便刪除
	if this.Select("avatar").Where("id = ?", MemberId).First(&mb).RecordNotFound() {
		return mb, m.NewAPIErr(m.ERROR_USER_NOTEXISTS,
			errors.New(fmt.Sprintf("not found MemberId: %d", MemberId)))
	}
	if Counts := this.Scopes(SearchByMemberId(MemberId)).UpdateColumn("avatar", FilePath).RowsAffected; Counts == 0 {
		return mb, m.NewAPIErr(m.ERROR_SERVER_FAILD,
			errors.New(fmt.Sprintf("update MemberAvatar failed MemberId: %d", MemberId)))
	}
	return
}

//更新身份為賣家
func (this MemberDao) ModifyIdentity(MemberId uint) (mb m.Member, err error) {
	tx := this.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(r.(error)))
		}
	}()
	if Counts := tx.Scopes(SearchByMemberId(MemberId)).UpdateColumn("is_seller", true).RowsAffected; Counts != 1 {
		tx.Rollback()
		return mb, m.NewAPIErr(m.ERROR_NOSUCH_DATA,
			errors.New(fmt.Sprintf("not found MemberId: %d", MemberId)))
	}
	if tx.Scopes(SearchByMemberId(MemberId)).Preload("MemberLevel", MemberLevelInfo()).First(&mb).RecordNotFound() {
		tx.Rollback()
		return mb, m.NewAPIErr(m.ERROR_SERVER_FAILD,
			errors.New(fmt.Sprintf("not found MemberId: %d", MemberId)))
	}
	tx.Commit()
	return
}

//獲取會員等級相關內容
func MemberLevelInfo() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select("level,expires_at,point")
	}
}

//ID查詢會員
func SearchByMemberId(MemberId uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Model(&m.Member{}).Where("id = ?", MemberId)
	}
}
