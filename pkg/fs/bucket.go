package fscommon

import (
	"fmt"
	"path/filepath"

	kv "github.com/takanoriyanagitani/go-kvstore"
	ks "github.com/takanoriyanagitani/go-kvstore/pkg/key/str"
)

type FsBucket struct {
	checkedBucket string
}

type FsBucketChecker func(kv.Bucket) bool

func (f FsBucketChecker) BuildConverter(dirname string) Bucket2filename {
	return func(b kv.Bucket) kv.Either[string, error] {
		var efb kv.Either[FsBucket, error] = FsBucketBuilderNew(f)(b)
		return kv.EitherMap(efb, func(fb FsBucket) string {
			return fb.ToFullpath(dirname)
		})
	}
}

var FsBucketCheckerSimple FsBucketChecker = func(b kv.Bucket) bool {
	return ks.ValidUtf8Str(b())
}

func FsBucketBuilderNew(bc FsBucketChecker) func(kv.Bucket) kv.Either[FsBucket, error] {
	return func(unchecked kv.Bucket) kv.Either[FsBucket, error] {
		var valid bool = bc(unchecked)
		var of kv.Option[FsBucket] = kv.OptionFromBool(valid, func() FsBucket {
			var checked string = unchecked()
			return FsBucket{
				checkedBucket: checked,
			}
		})
		return of.OkOrElse(func() error { return fmt.Errorf("Invalid bucket: %s", unchecked()) })
	}
}

func (f FsBucket) ToFullpath(dirname string) string {
	return filepath.Join(dirname, f.checkedBucket)
}

type Bucket2filename func(b kv.Bucket) kv.Either[string, error]

func Bucket2filenameBuilderNew(fbc FsBucketChecker) func(dirname string) Bucket2filename {
	return func(dirname string) Bucket2filename {
		return fbc.BuildConverter(dirname)
	}
}

var Bucket2filenameBuilderSimple func(dirname string) Bucket2filename = Bucket2filenameBuilderNew(FsBucketCheckerSimple)
