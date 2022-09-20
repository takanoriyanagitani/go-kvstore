package fscommon

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
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

type FsBulkUpsert func(ctx context.Context, fullpath string, items kv.Iter[kv.BucketItem]) error

func FsBulkUpsertNew(chk GetCheckedFilepath) func(TempfilenameGenerator) func(Items2writer) FsBulkUpsert {
	return func(tg TempfilenameGenerator) func(Items2writer) FsBulkUpsert {
		return func(iw Items2writer) FsBulkUpsert {
			return func(ctx context.Context, fullpath string, items kv.Iter[kv.BucketItem]) error {
				var dirname string = filepath.Dir(fullpath)
				var tmpname string = tg(fullpath)
				iwGen := func(w io.Writer) error {
					var bw *bufio.Writer = bufio.NewWriter(w)
					e := iw(ctx, w, items)
					if nil != e {
						return e
					}
					return bw.Flush()
				}
				return kv.Error1st([]func() error{
					func() error {
						return write2tmp(chk, dirname, tmpname, iwGen)
					},
					func() error { return os.Rename(tmpname, fullpath) },
				})
			}
		}
	}
}

func BulkUpsertNew(fbu FsBulkUpsert) func(Bucket2filename) kv.BulkUpsert {
	return func(pg Bucket2filename) kv.BulkUpsert {
		return func(ctx context.Context, bucket kv.Bucket, items kv.Iter[kv.BucketItem]) error {
			var epath kv.Either[string, error] = pg(bucket)
			var ee kv.Either[error, error] = kv.EitherMap(epath, func(fullpath string) error {
				return fbu(ctx, fullpath, items)
			})
			return ee.UnwrapOrElse(kv.Identity[error])
		}
	}
}

type GetCheckedFilepath func(prefix string, unchecked string) string

func GetCheckedFilepathBuilderNew(alt string) GetCheckedFilepath {
	return func(prefix, unchecked string) string {
		var clean string = filepath.Clean(unchecked)
		var safe bool = strings.HasPrefix(prefix, clean)
		if safe {
			return clean
		}
		return alt
	}
}

func write2tmp(chk GetCheckedFilepath, dirname string, tmpname string, cb func(io.Writer) error) error {
	f, e := os.Create(chk(dirname, tmpname))
	if nil != e {
		return e
	}

	e = cb(f)
	if nil != e {
		_ = f.Close()
		return e
	}

	return f.Close()
}
