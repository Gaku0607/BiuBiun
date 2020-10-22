package middleware

import (
	"fmt"

	m "github.com/gaku/BiuBiun/model"
	t "github.com/gaku/BiuBiun/tool"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func IsSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsSeller := c.GetBool("IsSeller"); !IsSeller {
			MemberId, _ := c.Get("MemberId")
			t.Failed(c, m.NewAPIErr(m.ERROR_INSUFFICIENT_PERMISSIONS,
				errors.New(fmt.Sprintf("%v is not Seller", MemberId))).(*m.APIErr))
			c.Abort()
			return
		}
		c.Next()
	}
}
