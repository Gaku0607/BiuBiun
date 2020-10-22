package middleware

import (
	"fmt"
	"os"
	"time"

	logger "github.com/gaku/BiuBiun/initialization"
	"github.com/gaku/BiuBiun/model"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GinLogger() gin.HandlerFunc {
	log := logger.GetGinLogger()
	return func(c *gin.Context) {
		StartTime := time.Now()
		c.Next()
		EndTime := time.Now()
		//花費時間 毫秒
		LatencyTime := EndTime.Sub(StartTime)
		SpendTime := fmt.Sprintf("%f ms", float64(LatencyTime)/1000000.0)
		//請求IP
		ClientIP := c.ClientIP()
		//請求主機
		HostName, err := os.Hostname()
		if err != nil {
			HostName = "unknown"
		}
		//狀態
		StatusCode := c.Writer.Status()
		//請求方法
		ReqMethod := c.Request.Method
		//ＵＲＬ
		ReqURL := c.Request.RequestURI
		//請求所使用的瀏覽器
		UserAgent := c.Request.UserAgent()
		//請求的數據大小
		DataSize := c.Writer.Size()
		entry := log.WithFields(logrus.Fields{
			"host_name":    HostName,
			"status_code":  StatusCode,
			"latency_time": SpendTime,
			"client_ip":    ClientIP,
			"req_method":   ReqMethod,
			"req_uri":      ReqURL,
			"user_agent":   UserAgent,
			"data_size":    DataSize,
		})
		//取得Err具體內容
		val, exists := c.Get("APIErr")
		EorrorMsg := ""
		if exists {
			EorrorMsg = val.(*model.APIErr).Error()
		}
		//Error級別 紀錄並且 發送信件給管理員
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		}
		if StatusCode >= 500 {
			entry.Error(EorrorMsg)
		} else if StatusCode >= 400 {
			entry.Warn(EorrorMsg)
		} else {
			entry.Info(EorrorMsg)
		}
	}
}
