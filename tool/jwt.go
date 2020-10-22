package tool

import (
	"os"
	"strconv"
	"time"

	m "github.com/gaku/BiuBiun/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

var (
	JwtSecret []byte
)

func SetJwtSecret() (err error) {
	jwtSecret := os.Getenv("jwtSecret")
	if jwtSecret == "" {
		return errors.New("jwtSecret is null")
	}
	JwtSecret = []byte(jwtSecret)
	return
}
func PutOutJWT(MemberId uint, Username string, level m.Level, IsSeller *bool) (token string, err error) {
	now := time.Now()
	jwtId := Username + strconv.FormatInt(time.Now().Unix(), 10)
	claims := m.Claims{
		MemberId: MemberId,
		Level:    level,
		IsSeller: &(*IsSeller),
		StandardClaims: jwt.StandardClaims{
			Audience:  Username,
			ExpiresAt: now.Unix() + 10800,
			Id:        jwtId,
			IssuedAt:  now.Unix(),
			Issuer:    "BiuBiun",
			NotBefore: now.Add(time.Duration(10) * time.Second).Unix(),
			Subject:   Username,
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString(JwtSecret)
	if err != nil {
		return "", m.NewAPIErr(m.ERROR_SERVER_FAILD, errors.WithStack(err))
	}
	return
}
