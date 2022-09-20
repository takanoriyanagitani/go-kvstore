package kvzip

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
	kf "github.com/takanoriyanagitani/go-kvstore/pkg/fs"
)

type Bytes2zip func(b []byte, zw *zip.Writer) error

type Bytes2zipBuilder func(name string) Bytes2zip

var TimestampProviderZipEpoch kf.TimestampProvider = func() time.Time { return ZipEpoch }

type Bytes2zipBuilderFactory struct {
	fmp kf.FilemodeProvider
	tsp kf.TimestampProvider
	gbn kf.GetBasename
}

func (f Bytes2zipBuilderFactory) ZipRaw() Bytes2zipBuilder {
	return func(fullpath string) Bytes2zip {
		return func(b []byte, zw *zip.Writer) error {
			var mode fs.FileMode = f.fmp()
			var timestamp time.Time = f.tsp()

			var basename string = f.gbn(fullpath)
			var mf kf.MemFile = kf.MemFileNew(basename, b, mode, timestamp)
			var zi ZipItemInfo = ZipItemInfoNew(fullpath, mf)
			var efh kv.Either[*zip.FileHeader, error] = zi.ToHeader()
			var ew kv.Either[io.Writer, error] = kv.EitherFlatMap(
				efh,
				func(fh *zip.FileHeader) kv.Either[io.Writer, error] {
					return kv.EitherNew(zw.CreateHeader(fh))
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

func Bytes2zipBuilderFactoryNew(fmp kf.FilemodeProvider, tsp kf.TimestampProvider, gbn kf.GetBasename) kv.Either[Bytes2zipBuilderFactory, error] {
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
	kf.FilemodeProviderDefault,
	TimestampProviderZipEpoch,
	kf.GetBasenameFs,
).Ok().Must()

var Bytes2zipBuilderRawDefault Bytes2zipBuilder = Bytes2zipBuilderFactoryDefault.ZipRaw()
