package tool

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

var (
	email    string
	password string
	host     string
	port     int
)

func SetEmailParameter() (err error) {
	email = os.Getenv("Email")
	password = os.Getenv("EmailPassword")
	host = os.Getenv("Host")
	tempPort := os.Getenv("Port")
	if email == "" || password == "" || host == "" || tempPort == "" {
		return errors.New("Email Parameter cannt null")
	}
	port, err = strconv.Atoi(tempPort)
	return
}
func SendMail(subject string, body string, mailTo ...string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", strings.Join([]string{"BiuBiun", email}, ""))
	m.SetHeader("To", mailTo...)    //傳送可給多個使用者
	m.SetHeader("Subject", subject) //設定郵件主題
	m.SetBody("text/html", body)    //設定郵件正文

	d := gomail.NewDialer(host, port, email, password)
	err := d.DialAndSend(m)
	return errors.WithStack(err)
}

func GetRandCertificationMath() (ID string) {
	rand.Seed(time.Now().UnixNano())
	letters := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	b := make([]byte, 7)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
