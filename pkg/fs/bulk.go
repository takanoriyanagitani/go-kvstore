package fscommon

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

type FsBulkUpsert func(ctx context.Context, fullpath string, items kv.Iter[kv.BucketItem]) error

type FsBulkUpsertFactory struct {
	i2w Items2writer
	chk GetCheckedFilepath
	tfg TempfilenameGenerator
}

func FsBulkUpsertFactoryNew(i2w Items2writer, chk GetCheckedFilepath, tfg TempfilenameGenerator) FsBulkUpsertFactory {
	return FsBulkUpsertFactory{
		i2w,
		chk,
		tfg,
	}
}

// FsBulkUpsertFactoryDefault creates a partially-populated FsBulkUpsertFactory.
func FsBulkUpsertFactoryDefault() FsBulkUpsertFactory {
	return FsBulkUpsertFactoryNew(
		nil,
		GetCheckedFilepathDefault,
		TempfilenameGeneratorSimpleDefault,
	)
}

func (f FsBulkUpsertFactory) Build() kv.Either[FsBulkUpsert, error] {
	var valid bool = kv.IterAll(
		kv.IterFromArray([]bool{
			nil != f.i2w,
			nil != f.chk,
			nil != f.tfg,
		}),
		kv.Identity[bool],
	)
	var o kv.Option[FsBulkUpsert] = kv.OptionFromBool(valid, func() FsBulkUpsert {
		return func(ctx context.Context, fullpath string, items kv.Iter[kv.BucketItem]) error {
			var dirname string = filepath.Dir(fullpath)
			var tmpname string = f.tfg(fullpath)
			iwGen := func(w io.Writer) error {
				var bw *bufio.Writer = bufio.NewWriter(w)
				e := f.i2w(ctx, w, items)
				if nil != e {
					return e
				}
				return bw.Flush()
			}
			return kv.Error1st([]func() error{
				func() error {
					return write2tmp(f.chk, dirname, tmpname, iwGen)
				},
				func() error { return os.Rename(tmpname, fullpath) },
			})
		}
	})
	return o.OkOrElse(func() error { return fmt.Errorf("Invalid argument") })
}

func (f FsBulkUpsertFactory) BuildWithConverter(i2w Items2writer) kv.Either[FsBulkUpsert, error] {
    return f.WithConverter(i2w).Build()
}

func (f FsBulkUpsertFactory) WithConverter(i2w Items2writer) FsBulkUpsertFactory {
	f.i2w = i2w
	return f
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
