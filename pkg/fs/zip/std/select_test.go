package kvzip

import (
	"archive/zip"
	"bytes"
	"testing"

	kv "github.com/takanoriyanagitani/go-kvstore"
	kf "github.com/takanoriyanagitani/go-kvstore/pkg/fs"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	t.Run("UnlimitedName2BytesBuilderNew", func(t *testing.T) {
		t.Parallel()

		t.Run("Invalid zip", func(t *testing.T) {
			t.Parallel()

			var rdr *bytes.Reader = bytes.NewReader(nil)
			var era kv.Either[kf.ReaderAtSize, error] = kf.ReaderAtSizeNew(rdr, 0)
			t.Run("reader at size got", checker(era.IsOk(), true))

			var ra kf.ReaderAtSize = era.Ok().Value()

			var enb kv.Either[kf.Name2Bytes, error] = UnlimitedName2BytesBuilderNew(ra)
			t.Run("Must fail(invalid zip)", checker(enb.IsOk(), false))
		})

		t.Run("empty zip", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			var zw *zip.Writer = zip.NewWriter(&buf)
			var e error = zw.Close()
			t.Run("zip created", checker(nil == e, true))

			var rdr *bytes.Reader = bytes.NewReader(buf.Bytes())
			var era kv.Either[kf.ReaderAtSize, error] = kf.ReaderAtSizeNew(rdr, rdr.Size())
			t.Run("reader at size got", checker(era.IsOk(), true))

			var ra kf.ReaderAtSize = era.Ok().Value()

			var enb kv.Either[kf.Name2Bytes, error] = UnlimitedName2BytesBuilderNew(ra)
			t.Run("Name2Bytes got", checker(enb.IsOk(), true))

			var nb kf.Name2Bytes = enb.Ok().Value()

			var eb kv.Either[[]byte, error] = nb("path/to/non-exist/file")
			t.Run("Must fail(must not exist)", checker(eb.IsOk(), false))
		})
	})
}
