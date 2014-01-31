package file

import (
	"github.com/d2g/beegosessions"
	"github.com/d2g/beegosessions/providers"
	"os"
	"sync"
)

type FileSessionStore struct {
	f      *os.File
	sid    string
	lock   sync.RWMutex
	values map[interface{}]interface{}
}

func (fs *FileSessionStore) Set(key, value interface{}) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values[key] = value
	return nil
}

func (fs *FileSessionStore) Get(key interface{}) interface{} {
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	if v, ok := fs.values[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (fs *FileSessionStore) Delete(key interface{}) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	delete(fs.values, key)
	return nil
}

func (fs *FileSessionStore) Flush() error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.values = make(map[interface{}]interface{})
	return nil
}

func (fs *FileSessionStore) SessionID() string {
	return fs.sid
}

func (fs *FileSessionStore) SessionRelease() {
	defer fs.f.Close()
	b, err := beegosessions.EncodeGob(fs.values)
	if err != nil {
		return
	}
	fs.f.Truncate(0)
	fs.f.Seek(0, 0)
	fs.f.Write(b)
}

func init() {
	providers.Register("file", filepder)
}
