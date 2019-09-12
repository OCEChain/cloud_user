package redis

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/henrylee2cn/faygo"
	"os"
	"time"
	"user/config"
)

type Redis struct {
	conn redis.Conn
}

var redisCient *redis.Pool

var host string

func init() {
	host = config.GetConfig("redis", "host").String()
	MaxIdle, err := config.GetConfig("redis", "MaxIdle").Int()
	if err != nil {
		faygo.Info("获取配置出错")
		os.Exit(2)
	}
	MaxActive, err := config.GetConfig("redis", "MaxActive").Int()
	if err != nil {
		faygo.Info("获取配置出错")
		os.Exit(2)
	}
	redisCient = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     MaxIdle,
		MaxActive:   MaxActive,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			// 选择db
			c.Do("SELECT", "1")
			return c, nil
		},
	}
	redisCient.Get()
	redisCient.Dial()
	return
}

func NewRedis() (r *Redis) {
	r = new(Redis)
	r.conn = redisCient.Get()
	return r
}

func (r *Redis) Get(key string) (data string, err error) {
	data, err = redis.String(r.conn.Do("GET", key))
	defer r.conn.Close()
	//如果返回结果为空的话，那么错误置为nil，返回值为空
	if err == redis.ErrNil {
		err = nil
		data = ""
	}
	return
}

func (r *Redis) Set(key string, value interface{}, expire ...int) (reply interface{}, err error) {
	b, err := json.Marshal(value)
	if err != nil {
		return
	}
	if len(expire) == 1 {
		reply, err = r.conn.Do("SET", key, string(b), "EX", expire[0])
	} else {
		reply, err = r.conn.Do("SET", key, string(b))
	}
	defer r.conn.Close()
	return
}

func (r *Redis) Del(key string) (err error) {
	_, err = r.conn.Do("DEL", key)
	return
}
