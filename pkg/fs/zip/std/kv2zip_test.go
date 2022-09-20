package kvzip

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"testing"

	kv "github.com/takanoriyanagitani/go-kvstore"
	kf "github.com/takanoriyanagitani/go-kvstore/pkg/fs"
)

func checkBuilderNew[T any](comp func(a, b T) (same bool)) func(got, expected T) func(*testing.T) {
	return func(got, expected T) func(*testing.T) {
		return func(t *testing.T) {
			var same bool = comp(got, expected)
			if !same {
				t.Errorf("Unexpected value got.")
				t.Errorf("Expected: %v", expected)
				t.Fatalf("Got:      %v", got)
			}
		}
	}
}

func checker[T comparable](got, expected T) func(*testing.T) {
	return checkBuilderNew(
		func(a, b T) (same bool) {
			return a == b
		},
	)(got, expected)
}

var checkBytes func(got, expected []byte) func(*testing.T) = checkBuilderNew(func(a, b []byte) (same bool) {
	return 0 == bytes.Compare(a, b)
})

func TestAll(t *testing.T) {
	t.Parallel()

	t.Run("Items2writer", func(t *testing.T) {
		t.Parallel()

		t.Run("Items2writerRawDefault", func(t *testing.T) {
			t.Parallel()

			var i2w kf.Items2writer = Items2writerRawDefault

			t.Run("empty", func(t *testing.T) {
				t.Parallel()

				var buf bytes.Buffer
				var e error = i2w(context.Background(), &buf, kv.IterEmpty[kv.BucketItem]())
				t.Run("wrote without error", checker(nil == e, true))

				var rdr *bytes.Reader = bytes.NewReader(buf.Bytes())

				var ezr kv.Either[*zip.Reader, error] = kv.EitherNew(zip.NewReader(
					rdr,
					rdr.Size(),
				))
				t.Run("valid zip", checker(ezr.IsOk(), true))

				var zr *zip.Reader = ezr.Ok().Value()
				t.Run("zip empty", checker(len(zr.File), 0))
			})

			t.Run("single empty zip item", func(t *testing.T) {
				t.Parallel()

				var buf bytes.Buffer
				var e error = i2w(context.Background(), &buf, kv.IterFromArray([]kv.BucketItem{
					kv.BucketItemNew(
						kv.KeyNew(
							func() string { return "" },
							[]byte("filename-inside-zip.empty"),
						),
						kv.ValNew(nil),
					),
				}))
				t.Run("wrote without error", checker(nil == e, true))

				var rdr *bytes.Reader = bytes.NewReader(buf.Bytes())

				var ezr kv.Either[*zip.Reader, error] = kv.EitherNew(zip.NewReader(
					rdr,
					rdr.Size(),
				))
				t.Run("valid zip", checker(ezr.IsOk(), true))

				var zr *zip.Reader = ezr.Ok().Value()
				t.Run("zip empty", checker(len(zr.File), 1))

				var zf *zip.File = zr.File[0]

				var zh zip.FileHeader = zf.FileHeader
				t.Run("zip item name", checker(zh.Name, "filename-inside-zip.empty"))
				t.Run("compression method", checker(zh.Method, zip.Store))
				t.Run("zip item size", checker(zh.UncompressedSize64, 0))
			})

			t.Run("single non empty zip item", func(t *testing.T) {
				t.Parallel()

				var buf bytes.Buffer
				var e error = i2w(context.Background(), &buf, kv.IterFromArray([]kv.BucketItem{
					kv.BucketItemNew(
						kv.KeyNew(
							func() string { return "" },
							[]byte("filename-inside-zip.txt"),
						),
						kv.ValNew([]byte("hw")),
					),
				}))
				t.Run("wrote without error", checker(nil == e, true))

				var rdr *bytes.Reader = bytes.NewReader(buf.Bytes())

				var ezr kv.Either[*zip.Reader, error] = kv.EitherNew(zip.NewReader(
					rdr,
					rdr.Size(),
				))
				t.Run("valid zip", checker(ezr.IsOk(), true))

				var zr *zip.Reader = ezr.Ok().Value()
				t.Run("zip empty", checker(len(zr.File), 1))

				var zf *zip.File = zr.File[0]

				var zh zip.FileHeader = zf.FileHeader
				t.Run("zip item name", checker(zh.Name, "filename-inside-zip.txt"))
				t.Run("compression method", checker(zh.Method, zip.Store))
				t.Run("zip item size", checker(zh.UncompressedSize64, 2))
			})

			t.Run("many non empty zip items", func(t *testing.T) {
				t.Parallel()

				var buf bytes.Buffer
				var e error = i2w(context.Background(), &buf, kv.IterFromArray([]kv.BucketItem{
					kv.BucketItemNew(
						kv.KeyNew(
							func() string { return "" },
							[]byte("path/inside/zip/file1.txt"),
						),
						kv.ValNew([]byte("test 1")),
					),
					kv.BucketItemNew(
						kv.KeyNew(
							func() string { return "" },
							[]byte("path/inside/zip/file2.txt"),
						),
						kv.ValNew([]byte("test II")),
					),
				}))
				t.Run("wrote without error", checker(nil == e, true))

				var rdr *bytes.Reader = bytes.NewReader(buf.Bytes())

				var ezr kv.Either[*zip.Reader, error] = kv.EitherNew(zip.NewReader(
					rdr,
					rdr.Size(),
				))
				t.Run("valid zip", checker(ezr.IsOk(), true))

				var zr *zip.Reader = ezr.Ok().Value()
				t.Run("zip empty", checker(len(zr.File), 2))

				chk := func(zf *zip.File, name string, content []byte) func(*testing.T) {
					return func(t *testing.T) {
						var zh zip.FileHeader = zf.FileHeader
						t.Run("zip item name", checker(zh.Name, name))
						t.Run("compression method", checker(zh.Method, zip.Store))
						t.Run("zip item size", checker(zh.UncompressedSize64, uint64(len(content))))

						var ezf kv.Either[io.ReadCloser, error] = kv.EitherNew(zf.Open())
						t.Run("zip item open", checker(ezf.IsOk(), true))

						var rc io.ReadCloser = ezf.Ok().Value()
						defer rc.Close()

						var eb kv.Either[[]byte, error] = kv.EitherNew(io.ReadAll(rc))
						t.Run("zip content read", checker(eb.IsOk(), true))

						var ba []byte = eb.Ok().Value()
						t.Run("content len", checker(len(ba), len(content)))

						t.Run("content check", checkBytes(ba, content))
					}
				}

                t.Run("item 0", chk(zr.File[0], "path/inside/zip/file1.txt", []byte("test 1")))
                t.Run("item 1", chk(zr.File[1], "path/inside/zip/file2.txt", []byte("test II")))
			})
		})
	})
}
