package fscommon

import (
	"io/fs"
	"time"
)

type MemFile struct {
	basename string
	data     []byte
	mode     fs.FileMode
	modified time.Time
}

func MemFileNew(basename string, data []byte, mode fs.FileMode, modified time.Time) MemFile {
	return MemFile{
		basename,
		data,
		mode,
		modified,
	}
}

func (m MemFile) Name() string       { return m.basename }
func (m MemFile) Size() int64        { return int64(len(m.data)) }
func (m MemFile) Mode() fs.FileMode  { return m.mode }
func (m MemFile) ModTime() time.Time { return m.modified }
func (m MemFile) IsDir() bool        { return false }
func (m MemFile) Sys() any           { return nil }
