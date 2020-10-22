package model

import (
	"fmt"
	"net/http"
)

// 1000...1999 為登入註冊錯誤
const (
	ERROR_USER_NOTEXISTS           uint = 1001 + iota //用戶不存在
	ERROR_PWD_NOTEXISTS                               //密碼不存在
	ERROR_USER_EXISTS                                 //用戶已存在
	ERROR_PWD_EXISTS                                  //密碼已存在
	ERROR_EXPIRED_CERTIFICATION_ID                    // 驗證碼過期
	ERROR_EMAIL_EXISTS                                //帳號已存在
)

//2001...2999 為用戶錯誤
const (
	ERROR_INSUFFICIENT_PERMISSIONS uint = 2001 + iota //權限不足
	ERROR_FORMAT_JWT                                  //Jwt 格式錯誤
	ERROR_EXPIRED_JWT                                 //Jwt 過期
)

//3001...4000 為操作Shop錯誤
const (
	ERROR_OUT_OF_SHOP_RANGE uint = 3001 + iota //超出Shop總數範圍
	ERROR_OUT_OF_FOOD_RANGE                    //超出Food總數範圍
	ERROR_ADDRESS_EXISTS                       //地址重複
	ERROR_PHONE_EXISTS                         //手機號碼重複
	ERROR_FOOD_EXISTS                          //商品品名重複
)

//4001...5000...通用錯誤
const (
	ERROR_WRONG_FORMAT uint = 4001 + iota //格式錯誤
	ERROR_NOSUCH_DATA                     //查無資料
	ERROR_UPLOAD_FILE                     //上傳失敗
)

//5001系統錯誤
const (
	ERROR_SERVER_FAILD uint = 5001 + iota //伺服器故障
)

var errMsg = map[uint]string{
	//登入 註冊 錯誤
	ERROR_USER_NOTEXISTS:           "User not exist",
	ERROR_PWD_NOTEXISTS:            "Password  not exist",
	ERROR_USER_EXISTS:              "User already exist",
	ERROR_PWD_EXISTS:               "Password already exist",
	ERROR_EMAIL_EXISTS:             "Email already exist",
	ERROR_EXPIRED_CERTIFICATION_ID: "Expired CertificationID Time",
	//用戶 錯誤
	ERROR_INSUFFICIENT_PERMISSIONS: "Insufficient Permissions",
	ERROR_EXPIRED_JWT:              "Expired JWT Time",
	ERROR_FORMAT_JWT:               "JWT format failed",
	//為操作Shop錯誤
	ERROR_OUT_OF_SHOP_RANGE: "Out of range for ShopMaxCount",
	ERROR_OUT_OF_FOOD_RANGE: "Out of range for FoodMaxCount",
	ERROR_FOOD_EXISTS:       "Food already exist",
	ERROR_ADDRESS_EXISTS:    "Shop address already exist",
	ERROR_PHONE_EXISTS:      "Phone already exist",
	//通用錯誤
	ERROR_WRONG_FORMAT: "wrong format",
	ERROR_UPLOAD_FILE:  "upload file failed",
	ERROR_NOSUCH_DATA:  "No Such Data",
	//伺服器錯誤
	ERROR_SERVER_FAILD: "Server error",
}

type APIErr struct {
	Context string
	Stack   string
	Code    uint
}

func NewAPIErr(Code uint, err error) error {
	return &APIErr{
		Context: err.Error(),
		Stack:   fmt.Sprintf("%+v", err),
		Code:    Code,
	}
}
func (this APIErr) Error() string {
	return fmt.Sprintf("Stack:%s \n", this.Stack)
}
func (this APIErr) GetStatusCode() int {
	switch this.Code {
	case ERROR_INSUFFICIENT_PERMISSIONS, ERROR_EXPIRED_JWT, ERROR_FORMAT_JWT:
		return http.StatusForbidden //403
	case ERROR_SERVER_FAILD:
		return http.StatusInternalServerError //500
	case ERROR_WRONG_FORMAT:
		return http.StatusBadRequest //400
	}
	return http.StatusOK //200
}
func (this APIErr) GetErrMsg() string {
	return errMsg[this.Code]
}
