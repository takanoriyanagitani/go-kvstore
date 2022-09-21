package fscommon

import (
	"io/fs"
	"os"
)

type RealFs struct{}

func (r RealFs) Open(name string) (fs.File, error) {
	return os.Open(name)
}
