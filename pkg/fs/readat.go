package fscommon

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

type ReaderAtSize struct {
	io.ReaderAt
	Size int64
}

func ReaderAtSizeNew(ra io.ReaderAt, Size int64) kv.Either[ReaderAtSize, error] {
	var ora kv.Option[io.ReaderAt] = kv.OptionFromBool(nil != ra, func() io.ReaderAt {
		return ra
	})
	var oras kv.Option[ReaderAtSize] = kv.OptionMap(ora, func(ra io.ReaderAt) ReaderAtSize {
		return ReaderAtSize{
			ReaderAt: ra,
			Size:     Size,
		}
	})
	return oras.OkOrElse(func() error { return fmt.Errorf("Invalid arguments") })
}

type File2ReaderAtSize func(f fs.File) kv.Either[ReaderAtSize, error]

func File2ReaderAtSizeBuilderNew(f2b File2Bytes) File2ReaderAtSize {
	return func(f fs.File) kv.Either[ReaderAtSize, error] {
		var eb kv.Either[[]byte, error] = f2b(f)
		var er kv.Either[*bytes.Reader, error] = kv.EitherMap(eb, func(ba []byte) *bytes.Reader {
			return bytes.NewReader(ba)
		})
		return kv.EitherFlatMap(er, func(r *bytes.Reader) kv.Either[ReaderAtSize, error] {
			return ReaderAtSizeNew(r, r.Size())
		})
	}
}

var UnlimitedFile2ReaderAtSize File2ReaderAtSize = File2ReaderAtSizeBuilderNew(
	UnlimitedFile2Bytes,
)
