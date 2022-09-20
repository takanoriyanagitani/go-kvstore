package fscommon

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
	ks "github.com/takanoriyanagitani/go-kvstore/pkg/key/str"
)

const FilemodeDefault fs.FileMode = 0644

type BulkUpsertBuilderFs func(basedir string) kv.BulkUpsert

func BulkUpsertBuilderNew(vb ks.ValidateBucket) func(ks.ValidateId) BulkUpsertBuilderFs {
	return func(vi ks.ValidateId) BulkUpsertBuilderFs {
		return func(basedir string) kv.BulkUpsert {
			return func(ctx context.Context, bucket kv.Bucket, items kv.Iter[kv.BucketItem]) error {
				// TODO
				return nil
			}
		}
	}
}

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

type GetBasename func(fullpath string) (basename string)

var GetBasenameFs GetBasename = filepath.Base

type TimestampProvider func() time.Time
type FilemodeProvider func() fs.FileMode

var FilemodeProviderDefault FilemodeProvider = func() fs.FileMode { return FilemodeDefault }

type Items2writer func(context.Context, io.Writer, kv.Iter[kv.BucketItem]) error

type BulkUpsertFactory struct {
	iw Items2writer
	vb ks.ValidateBucket
	vi ks.ValidateId
}
