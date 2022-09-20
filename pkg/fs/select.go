package fscommon

import (
	"context"
	"io"
	"io/fs"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

type FsSelect func(ctx context.Context, key FsKey) kv.Either[kv.Option[kv.Val], error]

type File2Bytes func(f fs.File) kv.Either[[]byte, error]

var File2bytesSimple File2Bytes = func(f fs.File) kv.Either[[]byte, error] {
	return kv.EitherNew(io.ReadAll(f))
}
