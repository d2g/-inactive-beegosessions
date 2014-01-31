package mysql

//CREATE TABLE `session` (
//  `session_key` char(64) NOT NULL,
//  `session_data` blob,
//  `session_expiry` int(11) unsigned NOT NULL,
//  PRIMARY KEY (`session_key`)
//) ENGINE=MyISAM DEFAULT CHARSET=utf8;

import (
	"database/sql"
	"github.com/d2g/beegosessions"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type MysqlSessionStore struct {
	c      *sql.DB
	sid    string
	lock   sync.RWMutex
	values map[interface{}]interface{}
}

func (st *MysqlSessionStore) Set(key, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values[key] = value
	return nil
}

func (st *MysqlSessionStore) Get(key interface{}) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.values[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (st *MysqlSessionStore) Delete(key interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.values, key)
	return nil
}

func (st *MysqlSessionStore) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values = make(map[interface{}]interface{})
	return nil
}

func (st *MysqlSessionStore) SessionID() string {
	return st.sid
}

func (st *MysqlSessionStore) SessionRelease() {
	defer st.c.Close()
	if len(st.values) > 0 {
		b, err := beegosessions.EncodeGob(st.values)
		if err != nil {
			return
		}
		st.c.Exec("UPDATE session set `session_data`= ? where session_key=?", b, st.sid)
	}
}
