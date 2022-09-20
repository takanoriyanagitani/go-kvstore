package fscommon

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
	ks "github.com/takanoriyanagitani/go-kvstore/pkg/key/str"
)

const FilemodeDefault fs.FileMode = 0644

type BulkUpsertBuilderFs func(basedir string) kv.BulkUpsert

type GetBasename func(fullpath string) (basename string)

var GetBasenameFs GetBasename = filepath.Base

type TimestampProvider func() time.Time
type FilemodeProvider func() fs.FileMode

var FilemodeProviderDefault FilemodeProvider = func() fs.FileMode { return FilemodeDefault }

type Items2writer func(context.Context, io.Writer, kv.Iter[kv.BucketItem]) error

type Bucket2filename func(b kv.Bucket) kv.Either[string, error]

type TempfilenameGenerator func(fullpath string) string

func TempfilenameGeneratorBuilderSimpleNew(suffix string) TempfilenameGenerator {
	return func(fullpath string) string {
		return fullpath + suffix
	}
}

var TempfilenameGeneratorSimpleDefault TempfilenameGenerator = TempfilenameGeneratorBuilderSimpleNew(".tmp")

type BulkUpsertFactory struct {
	iw Items2writer
	vb ks.ValidateBucket
	vi ks.ValidateId
	bf Bucket2filename
	tg TempfilenameGenerator
}

func fileCommit(fullpath, tmp string, cb func(io.Writer) error) error {
	f, e := os.Create(tmp)
	if nil != e {
		return e
	}
	e = cb(f)
	if nil != e {
		_ = f.Close()
		return e
	}
	e = f.Close()
	if nil != e {
		_ = os.Remove(tmp)
		return e
	}
	e = os.Rename(tmp, fullpath)
	if nil != e {
		_ = os.Remove(tmp)
		return e
	}
	return nil
}

func (b BulkUpsertFactory) Build() kv.BulkUpsert {
	return func(ctx context.Context, bucket kv.Bucket, items kv.Iter[kv.BucketItem]) error {
		var epath kv.Either[string, error] = b.bf(bucket)
		return epath.TryForEach(func(fullpath string) error {
			var tmpname string = b.tg(fullpath)
			return fileCommit(fullpath, tmpname, func(w io.Writer) error {
				var bw *bufio.Writer = bufio.NewWriter(w)
				e := b.iw(ctx, bw, items)
				if nil != e {
					return e
				}
				return bw.Flush()
			})
		})
	}
}
