package controller

import (
	"fmt"

	midd "github.com/gaku/BiuBiun/middleware"
	m "github.com/gaku/BiuBiun/model"
	s "github.com/gaku/BiuBiun/server"
	t "github.com/gaku/BiuBiun/tool"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type MemberContorller struct {
	*gin.RouterGroup
}

func NewMerberContorller(g *gin.RouterGroup) *MemberContorller {
	return &MemberContorller{RouterGroup: g}
}
func (this MemberContorller) Router() {
	//一般遊客
	TouristGroup := this.RouterGroup.Group("/api")
	{ //登入
		TouristGroup.POST("/login", midd.MemberStatus(true), t.WrapperHandler(this.Login))
		//註冊
		TouristGroup.POST("/register", t.WrapperHandler(this.Register))
		//驗證
		TouristGroup.POST("/certification", t.WrapperHandler(this.CertificationIdAndCreate))
		//會員
		MemberGroup := TouristGroup.Group("/home")
		MemberGroup.Use(midd.MemberStatus(false))
		{ //Home
			MemberGroup.GET("/", t.WrapperHandler(this.Home))
			//更新會員頭像
			MemberGroup.POST("/upload", t.WrapperHandler(this.UploadAvatar))
			//升級賣家
			MemberGroup.POST("/applyion/seller", t.WrapperHandler(this.ApplyForSeller))
			//升級會員等級
			MemberGroup.POST("/applyion/vip")
		}
	}
}

//註冊
func (this MemberContorller) Register(c *gin.Context) (data interface{}, err error) {
	var mr m.RegisterParms
	if err := c.ShouldBindJSON(&mr); err != nil {
		return data, m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.WithStack(err))
	}
	//註冊信息相關驗證 成功寄出驗證信
	err = s.MS.RegisterVerification(&mr)
	return
}

//驗證以及創建用戶
func (this MemberContorller) CertificationIdAndCreate(c *gin.Context) (data interface{}, err error) {
	CertifiacationId := c.PostForm("ID")
	//驗證信確認與創建
	err = s.MS.VerificationIdAndCreate(CertifiacationId)
	return
}

//登入
func (this MemberContorller) Login(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	//確認是否為登入狀態
	if err == nil {
		//當為nil時 代表已登入
		if err = s.MS.IsLogin(MemberId); err == nil {
			return
		}
	}
	var mes m.LoginParms
	if err = c.ShouldBindJSON(&mes); err != nil {
		return
	}
	//校驗登錄信息
	return s.MS.VerificationLoginMsg(&mes)
}

//Home...
func (this MemberContorller) Home(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	return s.MS.SearchMemberInfo(MemberId)
}

//上傳Icon
func (this MemberContorller) UploadAvatar(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	file, err := c.FormFile("avatar")
	if err != nil {
		return data, m.NewAPIErr(m.ERROR_UPLOAD_FILE, errors.WithStack(err))
	}
	err = s.MS.SaveUploadFile(MemberId, file, c)
	return
}

//升級賣家
func (this MemberContorller) ApplyForSeller(c *gin.Context) (data interface{}, err error) {
	MemberId, err := GetMemberId(c)
	if err != nil {
		return
	}
	if IsSeller := c.GetBool("IsSeller"); IsSeller {
		return
	}
	return s.MS.VerificationIdentity(MemberId)
}
func GetMemberId(c *gin.Context) (MemberId uint, err error) {
	tempID, exists := c.Get("MemberId")
	if !exists {
		return MemberId, m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS, errors.New("MemberId cannt null"))
	}
	MemberId, ok := tempID.(uint)
	if !ok {
		err = m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS, errors.New(fmt.Sprintf("MemberId is not uint: %v", tempID)))
	}
	return
}
