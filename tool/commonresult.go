package tool

import (
	"net/http"

	m "github.com/gaku/BiuBiun/model"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type HandlerFunc func(c *gin.Context) (interface{}, error)

func WrapperHandler(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c)
		if err != nil {
			if APIErr, ok := errors.Cause(err).(*m.APIErr); ok {
				Failed(c, APIErr)
				return
			}
		}
		Success(c, data)
	}
}

func Success(c *gin.Context, val interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": val,
	})
}
func Failed(c *gin.Context, APIErr *m.APIErr) {
	c.Set("APIErr", APIErr)
	c.JSON(APIErr.GetStatusCode(), gin.H{
		"code": APIErr.Code,
		"data": APIErr.GetErrMsg(),
	})
}
