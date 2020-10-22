package initialization

import (
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	Redis *redis.Pool
)

func InitRedis() {

	// redis conf
	redisPassword := os.Getenv("redisPassword")
	redisHost := os.Getenv("redisHost")
	tempRedisMaxOpenConns := os.Getenv("redisMaxOpenConns")
	tempRedisMaxIdConns := os.Getenv("redisMaxIdConns")
	tempRedisConnMaxAge := os.Getenv("redisConnMaxAge")
	//轉為int型
	redisMaxOpenConns, _ := strconv.Atoi(tempRedisMaxOpenConns)
	redisMaxIdConns, _ := strconv.Atoi(tempRedisMaxIdConns)
	redisConnMaxage, _ := strconv.Atoi(tempRedisConnMaxAge)

	options := redis.DialPassword(redisPassword)
	Redis = &redis.Pool{
		MaxIdle:     redisMaxIdConns,
		MaxActive:   redisMaxOpenConns,
		IdleTimeout: time.Duration(redisConnMaxage),
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisHost, options)
			return conn, err
		},
	}
}

func GetRedis() *redis.Pool {
	return Redis
}
