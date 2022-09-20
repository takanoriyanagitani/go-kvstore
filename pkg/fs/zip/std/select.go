package kvzip

import (
	"archive/zip"
	"io/fs"

	kv "github.com/takanoriyanagitani/go-kvstore"
	kf "github.com/takanoriyanagitani/go-kvstore/pkg/fs"
)

func name2fileBuilder(zr *zip.Reader) kf.Name2File {
	return func(name string) kv.Either[fs.File, error] {
		return kv.EitherNew(zr.Open(name))
	}
}

func name2BytesBuilderNew(f2b kf.File2Bytes) func(zr *zip.Reader) kf.Name2Bytes {
	return func(zr *zip.Reader) kf.Name2Bytes {
		var n2f kf.Name2File = name2fileBuilder(zr)
		return kf.Name2bytesNewFs(f2b)(n2f)
	}
}

func ras2zipReader(ras kf.ReaderAtSize) kv.Either[*zip.Reader, error] {
	return kv.EitherNew(zip.NewReader(ras.ReaderAt, ras.Size))
}

func Name2BytesBuilderNew(ras kf.ReaderAtSize) func(kf.File2Bytes) kv.Either[kf.Name2Bytes, error] {
	return func(f2b kf.File2Bytes) kv.Either[kf.Name2Bytes, error] {
		var ezr kv.Either[*zip.Reader, error] = ras2zipReader(ras)
		var z2n func(zr *zip.Reader) kf.Name2Bytes = name2BytesBuilderNew(f2b)
		return kv.EitherMap(ezr, z2n)
	}
}

func UnlimitedName2BytesBuilderNew(ras kf.ReaderAtSize) kv.Either[kf.Name2Bytes, error] {
	return Name2BytesBuilderNew(ras)(kf.UnlimitedFile2Bytes)
}
