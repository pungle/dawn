//Copyright (C) Mr.Pungle

package web

import (
	"encoding/json"
	"errors"
	"github.com/garyburd/redigo/redis"
	"time"
)

type redisDriver struct {
	pool *redis.Pool
}

func NewRedisSessionDriver(network string, addr string, maxIdle int,
	maxActive int, idleTimeout time.Duration) SessionDriver {

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			r, err := redis.String(c.Do("PING"))
			if r != "PONG" {
				return errors.New("redis not connect")
			}
			return err
		},
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
	}
	return &redisDriver{pool}
}

func (self *redisDriver) Get(key string) (interface{}, error) {
	conn := self.pool.Get()
	res, err := redis.Bytes(conn.Do("GET", key))
	conn.Close()
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (self *redisDriver) Set(key string, value interface{}, expire time.Duration) error {
	encodeData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	conn := self.pool.Get()
	_, err = conn.Do("SETEX", key, int(expire/time.Second), encodeData)
	conn.Close()
	return err

}
