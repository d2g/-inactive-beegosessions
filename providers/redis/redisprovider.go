package redis

import (
	"github.com/beego/redigo/redis"
	"github.com/d2g/beegosessions"
	"github.com/d2g/beegosessions/providers"
	"strconv"
	"strings"
)

var redispder = &RedisProvider{}

var MAX_POOL_SIZE = 100

var redisPool chan redis.Conn

type RedisProvider struct {
	maxlifetime int64
	savePath    string
	poolsize    int
	password    string
	poollist    *redis.Pool
}

//savepath like redisserveraddr,poolsize,password
//127.0.0.1:6379,100,astaxie
func (rp *RedisProvider) SessionInit(maxlifetime int64, savePath string) error {
	rp.maxlifetime = maxlifetime
	configs := strings.Split(savePath, ",")
	if len(configs) > 0 {
		rp.savePath = configs[0]
	}
	if len(configs) > 1 {
		poolsize, err := strconv.Atoi(configs[1])
		if err != nil || poolsize <= 0 {
			rp.poolsize = MAX_POOL_SIZE
		} else {
			rp.poolsize = poolsize
		}
	} else {
		rp.poolsize = MAX_POOL_SIZE
	}
	if len(configs) > 2 {
		rp.password = configs[2]
	}
	rp.poollist = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", rp.savePath)
		if err != nil {
			return nil, err
		}
		if rp.password != "" {
			if _, err := c.Do("AUTH", rp.password); err != nil {
				c.Close()
				return nil, err
			}
		}
		return c, err
	}, rp.poolsize)
	return nil
}

func (rp *RedisProvider) SessionRead(sid string) (providers.SessionStore, error) {
	c := rp.poollist.Get()
	if existed, err := redis.Int(c.Do("EXISTS", sid)); err != nil || existed == 0 {
		c.Do("SET", sid)
	}
	c.Do("EXPIRE", sid, rp.maxlifetime)
	kvs, err := redis.String(c.Do("GET", sid))
	var kv map[interface{}]interface{}
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = beegosessions.DecodeGob([]byte(kvs))
		if err != nil {
			return nil, err
		}
	}
	rs := &RedisSessionStore{c: c, sid: sid, values: kv}
	return rs, nil
}

func (rp *RedisProvider) SessionExist(sid string) bool {
	c := rp.poollist.Get()
	if existed, err := redis.Int(c.Do("EXISTS", sid)); err != nil || existed == 0 {
		return false
	} else {
		return true
	}
}

func (rp *RedisProvider) SessionRegenerate(oldsid, sid string) (providers.SessionStore, error) {
	c := rp.poollist.Get()
	if existed, err := redis.Int(c.Do("EXISTS", oldsid)); err != nil || existed == 0 {
		c.Do("SET", oldsid)
	}
	c.Do("RENAME", oldsid, sid)
	c.Do("EXPIRE", sid, rp.maxlifetime)
	kvs, err := redis.String(c.Do("GET", sid))
	var kv map[interface{}]interface{}
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = beegosessions.DecodeGob([]byte(kvs))
		if err != nil {
			return nil, err
		}
	}
	rs := &RedisSessionStore{c: c, sid: sid, values: kv}
	return rs, nil
}

func (rp *RedisProvider) SessionDestroy(sid string) error {
	c := rp.poollist.Get()
	c.Do("DEL", sid)
	return nil
}

func (rp *RedisProvider) SessionGC() {
	return
}

//@todo
func (rp *RedisProvider) SessionAll() int {

	return 0
}

func init() {
	providers.Register("redis", redispder)
}
