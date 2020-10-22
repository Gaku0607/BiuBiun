package initialization

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var (
	GinLogger *logrus.Logger
)

func InitLogger() (err error) {
	// log conf
	logMode := os.Getenv("logMode")
	FileTimeFormat := os.Getenv("FileTimeFormat")
	JsonTimeFormat := os.Getenv("JsonTimeFormat")
	tempLogFileMaxAgeTime := os.Getenv("FileMaxAgeTime")
	tempLogFileRotateionTime := os.Getenv("FileRotateionTime")
	GinDirPath := os.Getenv("GinDirPath")
	// log FileName
	GinFileName := os.Getenv("GinFileName")
	GinErrorlevelFileName := os.Getenv("GinErrorlevelFileName")
	//轉為int型
	logFileMaxAgeTime, _ := strconv.Atoi(tempLogFileMaxAgeTime)
	logFileRotateionTime, _ := strconv.Atoi(tempLogFileRotateionTime)
	err = GinLog(GinDirPath, GinFileName, GinErrorlevelFileName, FileTimeFormat, JsonTimeFormat, logMode, logFileMaxAgeTime, logFileRotateionTime)
	if err != nil {
		return
	}
	return
}
func GinLog(GinDirPath, FileName, BigErrFileName, FileTimeFormat, JsonTimeFormat, Mode string, MaxAgeTime, RotationTime int) (err error) {
	writer, err := rotatelogs.New(
		strings.Join([]string{GinDirPath, FileTimeFormat, " - ", FileName}, ""),
		rotatelogs.WithLinkName(GinDirPath+FileName), // 指向最新日志文件
		// WithMaxAge和WithRotationCount二者只能设置一個,
		// WithRotationCount设置文件清理前最多保存的個數.
		rotatelogs.WithMaxAge(time.Duration(MaxAgeTime)*time.Second),         // 文件最大保存時間
		rotatelogs.WithRotationTime(time.Duration(RotationTime)*time.Second), // 日志切割時間間隔
	)
	if err != nil {
		err = errors.New("config local systemfile system logger error")
		return
	}
	//Error級別Log
	werr, err := rotatelogs.New(
		strings.Join([]string{GinDirPath, FileTimeFormat, " - ", BigErrFileName}, ""),
		rotatelogs.WithLinkName(GinDirPath+BigErrFileName),
		rotatelogs.WithMaxAge(time.Duration(MaxAgeTime)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(RotationTime)*time.Second),
	)
	if err != nil {
		err = errors.New("config local systemfile system logger error")
		return
	}
	var level logrus.Level
	switch Mode {
	case logrus.DebugLevel.String():
		level = logrus.DebugLevel
	case logrus.InfoLevel.String():
		level = logrus.InfoLevel
	case logrus.WarnLevel.String():
		level = logrus.WarnLevel
	case logrus.ErrorLevel.String():
		level = logrus.ErrorLevel
	case logrus.PanicLevel.String():
		level = logrus.PanicLevel
	case logrus.FatalLevel.String():
		level = logrus.FatalLevel
	default:
		level = logrus.InfoLevel
	}
	GinLogger = logrus.New()
	GinLogger.SetLevel(level)
	errwriter := io.MultiWriter(writer, werr)
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: errwriter,
		logrus.FatalLevel: errwriter,
		logrus.PanicLevel: errwriter,
	}, &logrus.JSONFormatter{
		TimestampFormat: JsonTimeFormat,
	})
	//關閉控制台輸出
	stdout, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return
	}
	GinLogger.SetOutput(stdout)
	GinLogger.AddHook(lfsHook)
	return
}
func GetGinLogger() *logrus.Logger {
	return GinLogger
}
