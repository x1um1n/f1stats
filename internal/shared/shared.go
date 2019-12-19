package shared

import (
	"log"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/x1um1n/checkerr"

	"github.com/gomodule/redigo/redis"
)

/******************** Config handling stuff ***********************************/

// K is the global koanf instance
var K = koanf.New(".")

// LoadKoanf populates k with default values from configs/default.yml, then
// overrides/appends to those values with environment variables prefixed with KOANF_
func LoadKoanf() {
	log.Println("Reading default config")
	err := K.Load(file.Provider("config/default.yaml"), yaml.Parser())
	checkerr.CheckFatal(err, "Error reading default config file")

	log.Println("Checking environment for overrides")
	K.Load(env.Provider("KOANF_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "KOANF_"))
	}), nil)

	log.Printf("Using config for %s environment", K.String("environment"))
}

/******************** End Config handling stuff *******************************/

/******************** DB handling stuff ***************************************/

// newPool creates a redis connection pool
func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", K.String("redis_host")+":6379")
			checkerr.Check(err, "Error connecting to redis")
			return c, err
		},
	}
}

// ping tests connectivity for redis (PONG should be returned)
func ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	if err != nil {
		return err
	}

	log.Printf("PING Response = %s\n", s)
	return nil
}

// InitRedis initialises the redis connection pool
func InitRedis() *redis.Pool {
	//create redis connection pool
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	for i := 0; i < 10; i++ {
		err := ping(conn)
		if !checkerr.Check(err, "Error pinging redis..") {
			return pool
		}
		log.Printf("Attempt %d of 10, retrying in 5s\n", i)
		time.Sleep(5 * time.Second)
	}
	return nil
}

/******************** End DB handling stuff ***********************************/
