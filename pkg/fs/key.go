package fscommon

import (
	"path/filepath"
)

type FsKey struct {
	dirname string
	bucket  string // dirname/bucket = path to the archive(tar, zip, ...)
	id      string // filename inside the archive
}

func (f FsKey) ToFullpath() string { return filepath.Join(f.dirname, f.bucket) }

func (f FsKey) ToItemname() string { return f.id }
