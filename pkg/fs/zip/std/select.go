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

func Name2BytesBuilderNew(f2b kf.File2Bytes) func(kf.ReaderAtSize) kv.Either[kf.Name2Bytes, error] {
	return func(ras kf.ReaderAtSize) kv.Either[kf.Name2Bytes, error] {
		var ezr kv.Either[*zip.Reader, error] = ras2zipReader(ras)
		var z2n func(zr *zip.Reader) kf.Name2Bytes = name2BytesBuilderNew(f2b)
		return kv.EitherMap(ezr, z2n)
	}
}

func UnlimitedName2BytesBuilderNew(ras kf.ReaderAtSize) kv.Either[kf.Name2Bytes, error] {
	return Name2BytesBuilderNew(kf.UnlimitedFile2Bytes)(ras)
}

func file2id(f *zip.File) (id string) {
	var fh zip.FileHeader = f.FileHeader
	return fh.Name
}

func reader2names(r *zip.Reader) (names kv.Iter[string]) {
	var files kv.Iter[*zip.File] = kv.IterFromArray(r.File)
	return kv.IterMap(files, file2id)
}

func ras2names(ras kf.ReaderAtSize) kv.Either[kv.Iter[string], error] {
	var ez kv.Either[*zip.Reader, error] = ras2zipReader(ras)
	return kv.EitherMap(ez, reader2names)
}

func IdsBuilderNew(ras kf.ReaderAtSize) kf.Ids {
	return func() kv.Either[kv.Iter[string], error] {
		return ras2names(ras)
	}
}

func Archive2bytesNew(f2b kf.File2Bytes) kf.Archive2Bytes {
	return kv.ComposeEither(
		kf.File2ReaderAtSizeBuilderNew(f2b),
		Name2BytesBuilderNew(f2b),
	)
}

var UnlimitedArchive2bytes kf.Archive2Bytes = Archive2bytesNew(kf.UnlimitedFile2Bytes)
