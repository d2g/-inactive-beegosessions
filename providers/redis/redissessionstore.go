package redis

import (
	"github.com/beego/redigo/redis"
	"github.com/d2g/beegosessions"
	"sync"
)

type RedisSessionStore struct {
	c      redis.Conn
	sid    string
	lock   sync.RWMutex
	values map[interface{}]interface{}
}

func (rs *RedisSessionStore) Set(key, value interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values[key] = value
	return nil
}

func (rs *RedisSessionStore) Get(key interface{}) interface{} {
	rs.lock.RLock()
	defer rs.lock.RUnlock()
	if v, ok := rs.values[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (rs *RedisSessionStore) Delete(key interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	delete(rs.values, key)
	return nil
}

func (rs *RedisSessionStore) Flush() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values = make(map[interface{}]interface{})
	return nil
}

func (rs *RedisSessionStore) SessionID() string {
	return rs.sid
}

func (rs *RedisSessionStore) SessionRelease() {
	defer rs.c.Close()
	if len(rs.values) > 0 {
		b, err := beegosessions.EncodeGob(rs.values)
		if err != nil {
			return
		}
		rs.c.Do("SET", rs.sid, string(b))
	}
}
