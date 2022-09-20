package kvzip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
	ks "github.com/takanoriyanagitani/go-kvstore/pkg/key/str"
)

var ZipEpoch time.Time = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

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

type ZipItemInfo struct {
	fullpath string
	fsinfo   fs.FileInfo
}

func ZipItemInfoNew(fullpath string, fsinfo fs.FileInfo) ZipItemInfo {
	return ZipItemInfo{
		fullpath,
		fsinfo,
	}
}

func (z ZipItemInfo) ToHeader() kv.Either[*zip.FileHeader, error] {
	var efh kv.Either[*zip.FileHeader, error] = kv.EitherNew(zip.FileInfoHeader(z.fsinfo))
	return efh.Map(func(fh *zip.FileHeader) *zip.FileHeader {
		fh.Name = z.fullpath
		return fh
	})
}

type Bytes2zip func(b []byte, zw *zip.Writer) error

type Bytes2zipBuilder func(name string) Bytes2zip

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

func (m MemFile) ToZipItemInfo(fullpath string) ZipItemInfo {
	return ZipItemInfoNew(fullpath, m)
}

type GetBasename func(fullpath string) (basename string)

var GetBasenameFs GetBasename = filepath.Base

type TimestampProvider func() time.Time
type FilemodeProvider func() fs.FileMode

var TimestampProviderZipEpoch TimestampProvider = func() time.Time { return ZipEpoch }
var FilemodeProviderDefault FilemodeProvider = func() fs.FileMode { return FilemodeDefault }

type Bytes2zipBuilderFactory struct {
	fmp FilemodeProvider
	tsp TimestampProvider
	gbn GetBasename
}

func (f Bytes2zipBuilderFactory) ZipRaw() Bytes2zipBuilder {
	return func(fullpath string) Bytes2zip {
		return func(b []byte, zw *zip.Writer) error {
			var mode fs.FileMode = f.fmp()
			var timestamp time.Time = f.tsp()

			var basename string = f.gbn(fullpath)
			var mf MemFile = MemFileNew(basename, b, mode, timestamp)
			var zi ZipItemInfo = mf.ToZipItemInfo(fullpath)
			var efh kv.Either[*zip.FileHeader, error] = zi.ToHeader()
			var ew kv.Either[io.Writer, error] = kv.EitherFlatMap(
				efh,
				func(fh *zip.FileHeader) kv.Either[io.Writer, error] {
					return kv.EitherNew(zw.CreateRaw(fh))
				},
			)
			var ei kv.Either[int, error] = kv.EitherFlatMap(
				ew,
				func(w io.Writer) kv.Either[int, error] {
					return kv.EitherNew(w.Write(b))
				},
			)
			var oe kv.Option[error] = ei.Left()
			return oe.UnwrapOr(nil)
		}
	}
}

func Bytes2zipBuilderFactoryNew(fmp FilemodeProvider, tsp TimestampProvider, gbn GetBasename) kv.Either[Bytes2zipBuilderFactory, error] {
	var ok bool = kv.IterReduce(
		kv.IterFromArray([]bool{
			nil != fmp,
			nil != tsp,
			nil != gbn,
		}),
		true,
		func(state bool, b bool) bool { return state && b },
	)
	var of kv.Option[Bytes2zipBuilderFactory] = kv.OptionFromBool(ok, func() Bytes2zipBuilderFactory {
		return Bytes2zipBuilderFactory{
			fmp,
			tsp,
			gbn,
		}
	})
	return of.OkOrElse(func() error { return fmt.Errorf("Invalid arguments") })
}

var Bytes2zipBuilderFactoryDefault Bytes2zipBuilderFactory = Bytes2zipBuilderFactoryNew(
	FilemodeProviderDefault,
	TimestampProviderZipEpoch,
	GetBasenameFs,
).Ok().Must()

var Bytes2zipBuilderRawDefault Bytes2zipBuilder = Bytes2zipBuilderFactoryDefault.ZipRaw()

type Item2zip func(*zip.Writer) func(kv.BucketItem) error

func Item2zipBuilderNew(b2zb Bytes2zipBuilder) Item2zip {
	return func(zw *zip.Writer) func(kv.BucketItem) error {
		return func(item kv.BucketItem) error {
			var k kv.Key = item.Key()
			var v kv.Val = item.Val()

			var id []byte = k.Id()
			var dt []byte = v.Raw()

			var vi ks.ValidateId = ks.ValidateIdUtf8

			return kv.Error1st([]func() error{
				kv.Bool2ef(vi(k), func() error { return fmt.Errorf("Invalid id") }),
				func() error {
					var name string = string(id)
					var b2z Bytes2zip = b2zb(name)
					return b2z(dt, zw)
				},
			})
		}
	}
}

var Item2zipRawDefault Item2zip = Item2zipBuilderNew(Bytes2zipBuilderRawDefault)

type Items2zip func(context.Context, *zip.Writer, kv.Iter[kv.BucketItem]) error

func Items2zipBuilderNewDefault(i2z Item2zip) Items2zip {
	return func(_ctx context.Context, zw *zip.Writer, items kv.Iter[kv.BucketItem]) error {
		var iz func(kv.BucketItem) error = i2z(zw)
		return items.TryForEach(iz)
	}
}

var Items2zipRawDefault Items2zip = Items2zipBuilderNewDefault(Item2zipRawDefault)

type Items2writer func(context.Context, io.Writer, kv.Iter[kv.BucketItem]) error

func Items2writerBuilderNew(i2z Items2zip) Items2writer {
	return func(ctx context.Context, w io.Writer, items kv.Iter[kv.BucketItem]) error {
		var zw *zip.Writer = zip.NewWriter(w)
		e := i2z(ctx, zw, items)
		if nil != e {
			_ = zw.Close()
			return e
		}
		e = zw.Flush()
		if nil != e {
			_ = zw.Close()
			return e
		}
		return zw.Close()
	}
}

var Items2writerRawDefault Items2writer = Items2writerBuilderNew(Items2zipRawDefault)
