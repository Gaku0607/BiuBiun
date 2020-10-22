package initialization

import (
	"fmt"
	"log"

	t "github.com/gaku/BiuBiun/tool"

	"github.com/joho/godotenv"
)

//LoadSystem..
func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannt Open envFile Failed Err")
	}
	//初始化Logger
	err = InitLogger()
	if err != nil {
		log.Fatal(err.Error())
	}
	//初始化Redis
	InitRedis()
	//初始化DB
	err = InitDB()
	if err != nil {
		log.Fatal(err.Error())
	}
	//初始化FileDir
	err = InitFileDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	// 初始化JwtSecret
	err = t.SetJwtSecret()
	if err != nil {
		log.Fatal(err.Error())
	}
	//初始化Email
	err = t.SetEmailParameter()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
