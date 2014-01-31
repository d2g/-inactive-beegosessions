package file

import (
	"os"
)

type activeSession struct {
	total int
}

func (self *activeSession) visit(paths string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f.IsDir() {
		return nil
	}
	self.total = self.total + 1
	return nil
}
