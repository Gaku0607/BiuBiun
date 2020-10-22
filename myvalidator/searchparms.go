package myvalidator

import (
	"log"

	"github.com/gaku/BiuBiun/model"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	//註冊驗證方式 SearchShopParms ----> SSP
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("SSP", SearchShopParms)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func SearchShopParms(v validator.FieldLevel) bool {
	if data, ok := v.Field().Interface().([]model.Category); ok {
		if len(data) > 4 {
			return false
		}
		for _, v := range data {
			if v > 10 || v <= 0 {
				return false
			}
		}
	}
	return true
}
