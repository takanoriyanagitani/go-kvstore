package kvzip

import (
	"archive/zip"
	"io/fs"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

func name2fileBuilder(zr *zip.Reader) func(name string) kv.Either[fs.File, error] {
	return func(name string) kv.Either[fs.File, error] {
		return kv.EitherNew(zr.Open(name))
	}
}
