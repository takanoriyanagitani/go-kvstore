package fscommon

import (
	"io/fs"
	"os"
)

type RealFs struct {
	fo fsOpen
}

func realFsNew(fo fsOpen) RealFs {
	return RealFs{fo}
}

func RealFsNew() RealFs {
	return realFsNew(os.Open)
}

type fsOpen func(name string) (*os.File, error)

// Open tries to open named file.
// Its users responsibility to check the name.
func (r RealFs) Open(name string) (fs.File, error) {
	return r.fo(name)
}
