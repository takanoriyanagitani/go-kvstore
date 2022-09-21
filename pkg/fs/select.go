package fscommon

import (
	"context"
	"io"
	"io/fs"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

type FsSelect func(ctx context.Context, key FsKey) kv.Either[kv.Option[kv.Val], error]

type Archive2Bytes func(fs.File) kv.Either[Name2Bytes, error]

type Name2Rc func(name string) kv.Either[io.ReadCloser, error]

type Name2File func(name string) kv.Either[fs.File, error]

type Name2Bytes func(name string) kv.Either[[]byte, error]

type Rc2Bytes func(r io.ReadCloser) kv.Either[[]byte, error]

type File2Bytes func(f fs.File) kv.Either[[]byte, error]

type Ids func() kv.Either[kv.Iter[string], error]

var UnlimitedRc2Bytes Rc2Bytes = func(r io.ReadCloser) kv.Either[[]byte, error] {
	return kv.EitherNew(io.ReadAll(r))
}

var UnlimitedFile2Bytes File2Bytes = func(f fs.File) kv.Either[[]byte, error] {
	return UnlimitedRc2Bytes(f)
}

func Name2bytesNew(r2b Rc2Bytes) func(Name2Rc) Name2Bytes {
	return func(n2r Name2Rc) Name2Bytes {
		return kv.ComposeEither(
			n2r,
			r2b,
		)
	}
}

func Name2bytesNewFs(f2b File2Bytes) func(Name2File) Name2Bytes {
	return func(n2f Name2File) Name2Bytes {
		return kv.ComposeEither(
			n2f,
			f2b,
		)
	}
}
