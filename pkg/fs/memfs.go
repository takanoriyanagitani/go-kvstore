package fscommon

import (
	"io/fs"
)

type MemFs struct {
	raw map[string]fs.File
}

func (m MemFs) open(validname string) (fs.File, error) {
	f, ok := m.raw[validname]
	if ok {
		return f, nil
	}
	return nil, fs.ErrNotExist
}

func (m MemFs) Open(name string) (fs.File, error) {
	var valid bool = fs.ValidPath(name)
	if valid {
		return m.open(name)
	}
	return nil, fs.ErrInvalid
}

func (m MemFs) upsert(validname string, f fs.File) {
	m.raw[validname] = f
}

func (m MemFs) Upsert(name string, f fs.File) error {
	var valid bool = fs.ValidPath(name)
	if valid {
		m.upsert(name, f)
		return nil
	}
	return fs.ErrInvalid
}
