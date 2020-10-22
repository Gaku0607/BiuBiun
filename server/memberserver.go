package server

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gaku/BiuBiun/dao"

	"github.com/gaku/BiuBiun/initialization"
	m "github.com/gaku/BiuBiun/model"
	t "github.com/gaku/BiuBiun/tool"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	MS            *MemberServer
	ShopSrcPath   string
	MemberSrcPath string
)

type MemberServer struct {
	*redis.Pool
}

func NewMemberServer() *MemberServer {
	ShopSrcPath = os.Getenv("ShopFileDirPath")
	MemberSrcPath = os.Getenv("MemberFileDirPath")
	return &MemberServer{
		Pool: initialization.GetRedis(),
	}
}

//確認是否為登入狀態
func (this MemberServer) IsLogin(MemberId uint) (err error) {
	return dao.MDao.SearchMemberById(MemberId)
}

//註冊驗證
func (this MemberServer) RegisterVerification(mb *m.RegisterParms) (err error) {
	//驗證有無重複會員資料
	err = dao.MDao.CheckRegisterMsgInfo(mb)
	if err != nil {
		return
	}
	//寄驗證信 並且相對應的資料存入redis中
	ID := t.GetRandCertificationMath()
	//可能產生出新的ID必須保存
	NewID, err := this.SaveIDCashe(ID, mb)
	if err != nil {
		return
	}
	if err = t.SendMail("Hello!! this is a Certification Email", ID, mb.Email.Addrs); err != nil {
		//寄失敗記得刪除CertificationID
		this.DelIDCashe(NewID)
	}
	return
}

//驗證創建會員
func (this MemberServer) VerificationIdAndCreate(ID string) (err error) {
	if ID == "" {
		return m.NewAPIErr(m.ERROR_WRONG_FORMAT, errors.New("CertificationId cannt null"))
	}
	//校驗信息 成功後添加到Database
	UserInfo, err := this.VerificationCertificationId(ID)
	if err != nil {
		return
	}
	var mb m.Member
	passwordbcrypt, err := bcrypt.GenerateFromPassword([]byte(UserInfo.Password), 10)
	if err != nil {
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	mb.UserName = UserInfo.UserName
	mb.Password = string(passwordbcrypt)
	mb.City = UserInfo.City
	mb.Email.Addrs = UserInfo.Email.Addrs
	mb.MemberLevel.Level = 2
	err = dao.MDao.CreateMember(&mb)
	return
}

//驗證登入信息
func (this MemberServer) VerificationLoginMsg(mes *m.LoginParms) (token string, err error) {
	//Database校驗UserName
	mb, err := dao.MDao.GetMember(mes.UserName)
	if err != nil {
		return
	}
	//校驗PassWord
	err = bcrypt.CompareHashAndPassword([]byte(mb.Password), []byte(mes.PassWord))
	if err != nil {
		return "", m.NewAPIErr(m.ERROR_PWD_NOTEXISTS, err)
	}
	//發放JWT
	return t.PutOutJWT(mb.ID, mb.UserName, mb.MemberLevel.Level, &(*mb.IsSeller))
}

//儲存驗證碼至Redis
func (this MemberServer) SaveIDCashe(ID string, mb *m.RegisterParms) (NewID string, err error) {
	NewID = ID
	cxt, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
	Conn, err := this.Pool.GetContext(cxt)
	defer func() {
		cancel()
		Conn.Close()
	}()
	if err != nil {
		return "", m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	bytes, err := json.Marshal(mb)
	if err != nil {
		return "", m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	//檢測是否有相同ＩＤ 為0時代表沒有重複
	for {
		isExist, _ := redis.Int(Conn.Do("exists", NewID))
		if isExist == 0 {
			_, err = Conn.Do("setex", NewID, 600, bytes)
			if err != nil {
				return "", m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
			}
			return
		}
		NewID = t.GetRandCertificationMath()
	}
}

//刪除Redis中的驗證碼
func (this MemberServer) DelIDCashe(ID string) (err error) {
	cxt, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
	Conn, err := this.Pool.GetContext(cxt)
	defer func() {
		cancel()
		Conn.Close()
	}()
	if err != nil {
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	Conn.Do("del", ID)
	return
}

func (this MemberServer) VerificationCertificationId(ID string) (UserInfo m.RegisterParms, err error) {
	cxt, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
	Conn, err := this.Pool.GetContext(cxt)
	defer func() {
		cancel()
		Conn.Close()
	}()
	if err != nil {
		return UserInfo, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	tempUserInfo, err := Conn.Do("get", ID)
	if err != nil {
		return UserInfo, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	if tempUserInfo == nil {
		return UserInfo, m.NewAPIErr(m.ERROR_EXPIRED_CERTIFICATION_ID, errors.New("expored id"))
	}
	//刪除ＩＤ
	_, _ = Conn.Do("del", ID)
	err = json.Unmarshal(tempUserInfo.([]byte), &UserInfo)
	if err != nil {
		return UserInfo, m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	return
}

//儲存上傳檔案
func (this MemberServer) SaveUploadFile(MemberId uint, File *multipart.FileHeader, c *gin.Context) (err error) {
	FilePath := path.Join(MemberSrcPath, strconv.FormatInt(time.Now().Unix(), 10)+File.Filename)
	if err := c.SaveUploadedFile(File, FilePath); err != nil {
		return m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	mb, err := dao.MDao.UpdataAvatar(MemberId, FilePath[len(MemberSrcPath):])
	if err != nil {
		//如果儲存數據庫失敗 刪除更新檔
		os.Remove(FilePath)
		return
	}
	if mb.Avatar != "" {
		//更新成功刪除舊檔案
		os.Remove(MemberSrcPath + mb.Avatar)
	}
	return
}

//查詢會員信息
func (this MemberServer) SearchMemberInfo(MemberId uint) (mb m.Member, err error) {
	return dao.MDao.GetMemberInfo(MemberId)
}

//升級身份
func (this MemberServer) VerificationIdentity(MemberId uint) (token string, err error) {
	//成為賣家的的判斷......略
	mb, err := dao.MDao.ModifyIdentity(MemberId)
	if err != nil {
		return
	}
	//返回新的JWT
	return t.PutOutJWT(mb.ID, mb.UserName, mb.MemberLevel.Level, &(*mb.IsSeller))
}
