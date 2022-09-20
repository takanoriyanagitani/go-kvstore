package strk

import (
	"fmt"
	"unicode/utf8"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

type Id2str func(k kv.Key) kv.Either[string, error]
type Bucket2str func(k kv.Key) kv.Either[string, error]

type ValidateId func(k kv.Key) (valid bool)
type ValidateBucket func(k kv.Key) (valid bool)

type ValidateBytes func([]byte) (valid bool)
type ValidateString func(string) (valid bool)

var ValidUtf8 ValidateBytes = utf8.Valid
var ValidUtf8Str ValidateString = utf8.ValidString

func Id2strNew(vi ValidateId) Id2str {
	return func(k kv.Key) kv.Either[string, error] {
		var valid bool = vi(k)
		var o kv.Option[string] = kv.OptionFromBool(valid, func() string {
			return string(k.Id())
		})
		return o.OkOrElse(func() error { return fmt.Errorf("Invalid id") })
	}
}

func ValidateIdNew(vb ValidateBytes) ValidateId {
	return func(k kv.Key) (valid bool) {
		return vb(k.Id())
	}
}

func ValidateIdLowerBound(bucket int, id int) ValidateId {
	return func(k kv.Key) (valid bool) {
		var b bool = bucket <= len(k.BucketString())
		var i bool = id <= len(k.Id())
		return b && i
	}
}

func (v ValidateId) Concat(other ValidateId) ValidateId {
	return func(k kv.Key) (valid bool) {
		return v(k) && other(k)
	}
}

func ValidateBucketNew(vs ValidateString) ValidateBucket {
	return func(k kv.Key) (valid bool) {
		var bs string = k.BucketString()
		return vs(bs)
	}
}

var ValidateIdUtf8 ValidateId = ValidateIdNew(ValidUtf8)
var ValidateBucketUtf8 ValidateBucket = ValidateBucketNew(ValidUtf8Str)
