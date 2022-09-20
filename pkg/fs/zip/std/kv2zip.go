package kvzip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
	kf "github.com/takanoriyanagitani/go-kvstore/pkg/fs"
	ks "github.com/takanoriyanagitani/go-kvstore/pkg/key/str"
)

var ZipEpoch time.Time = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

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

func Items2writerBuilderNew(i2z Items2zip) kf.Items2writer {
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

var Items2writerRawDefault kf.Items2writer = Items2writerBuilderNew(Items2zipRawDefault)
