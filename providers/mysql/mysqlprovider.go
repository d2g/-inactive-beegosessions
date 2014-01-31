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
	"github.com/d2g/beegosessions/providers"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var mysqlpder = &MysqlProvider{}

type MysqlProvider struct {
	maxlifetime int64
	savePath    string
}

func (mp *MysqlProvider) connectInit() *sql.DB {
	db, e := sql.Open("mysql", mp.savePath)
	if e != nil {
		return nil
	}
	return db
}

func (mp *MysqlProvider) SessionInit(maxlifetime int64, savePath string) error {
	mp.maxlifetime = maxlifetime
	mp.savePath = savePath
	return nil
}

func (mp *MysqlProvider) SessionRead(sid string) (providers.SessionStore, error) {
	c := mp.connectInit()
	row := c.QueryRow("select session_data from session where session_key=?", sid)
	var sessiondata []byte
	err := row.Scan(&sessiondata)
	if err == sql.ErrNoRows {
		c.Exec("insert into session(`session_key`,`session_data`,`session_expiry`) values(?,?,?)", sid, "", time.Now().Unix())
	}
	var kv map[interface{}]interface{}
	if len(sessiondata) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = beegosessions.DecodeGob(sessiondata)
		if err != nil {
			return nil, err
		}
	}
	rs := &MysqlSessionStore{c: c, sid: sid, values: kv}
	return rs, nil
}

func (mp *MysqlProvider) SessionExist(sid string) bool {
	c := mp.connectInit()
	row := c.QueryRow("select session_data from session where session_key=?", sid)
	var sessiondata []byte
	err := row.Scan(&sessiondata)
	if err == sql.ErrNoRows {
		return false
	} else {
		return true
	}
}

func (mp *MysqlProvider) SessionRegenerate(oldsid, sid string) (providers.SessionStore, error) {
	c := mp.connectInit()
	row := c.QueryRow("select session_data from session where session_key=?", oldsid)
	var sessiondata []byte
	err := row.Scan(&sessiondata)
	if err == sql.ErrNoRows {
		c.Exec("insert into session(`session_key`,`session_data`,`session_expiry`) values(?,?,?)", oldsid, "", time.Now().Unix())
	}
	c.Exec("update session set `session_key`=? where session_key=?", sid, oldsid)
	var kv map[interface{}]interface{}
	if len(sessiondata) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = beegosessions.DecodeGob(sessiondata)
		if err != nil {
			return nil, err
		}
	}
	rs := &MysqlSessionStore{c: c, sid: sid, values: kv}
	return rs, nil
}

func (mp *MysqlProvider) SessionDestroy(sid string) error {
	c := mp.connectInit()
	c.Exec("DELETE FROM session where session_key=?", sid)
	c.Close()
	return nil
}

func (mp *MysqlProvider) SessionGC() {
	c := mp.connectInit()
	c.Exec("DELETE from session where session_expiry < ?", time.Now().Unix()-mp.maxlifetime)
	c.Close()
	return
}

func (mp *MysqlProvider) SessionAll() int {
	c := mp.connectInit()
	defer c.Close()
	var total int
	err := c.QueryRow("SELECT count(*) as num from session").Scan(&total)
	if err != nil {
		return 0
	}
	return total
}

func init() {
	providers.Register("mysql", mysqlpder)
}
