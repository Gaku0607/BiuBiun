package middleware

import (
	"time"

	m "github.com/gaku/BiuBiun/model"
	t "github.com/gaku/BiuBiun/tool"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func MemberStatus(flag bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		tokenClaims, err := jwt.ParseWithClaims(auth, &m.Claims{},
			func(token *jwt.Token) (i interface{}, err error) {
				return t.JwtSecret, nil
			})
		if err != nil {
			if flag {
				c.Next()
				return
			}
			if result, ok := err.(*jwt.ValidationError); ok {
				//如果為過期錯誤
				if result.Errors == jwt.ValidationErrorExpired {
					//當錯誤為過期時 判斷
					if Claims, ok := tokenClaims.Claims.(*m.Claims); ok {
						//過期時間大餘3小時錯誤
						if time.Now().Unix() > Claims.ExpiresAt+10800 {
							t.Failed(c, m.NewAPIErr(m.ERROR_EXPIRED_JWT, errors.WithStack(err)).(*m.APIErr))
							c.Abort()
							return
						}
						//重新發放token
						token, err := t.PutOutJWT(Claims.MemberId, Claims.UserName, Claims.Level, Claims.IsSeller)
						if err != nil {
							t.Failed(c, err.(*m.APIErr))
							c.Abort()
							return
						}
						t.Success(c, token)
						c.Set("MemberId", Claims.MemberId)
						c.Set("IsSeller", Claims.IsSeller)
						c.Next()
						return
					}
					//jwt解析錯誤且不為過期錯誤
				} else {
					t.Failed(c, m.NewAPIErr(m.ERROR_FORMAT_JWT, errors.WithStack(err)).(*m.APIErr))
					c.Abort()
					return
				}
			}
		}
		if Claims, ok := tokenClaims.Claims.(*m.Claims); ok {
			c.Set("IsSeller", Claims.IsSeller)
			c.Set("MemberId", Claims.MemberId)
			c.Next()
			return
		}
	}
}
