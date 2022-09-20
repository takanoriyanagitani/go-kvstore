package fscommon

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

const FilemodeDefault fs.FileMode = 0644

type GetBasename func(fullpath string) (basename string)

var GetBasenameFs GetBasename = filepath.Base

type TimestampProvider func() time.Time
type FilemodeProvider func() fs.FileMode

var FilemodeProviderDefault FilemodeProvider = func() fs.FileMode { return FilemodeDefault }

type Items2writer func(context.Context, io.Writer, kv.Iter[kv.BucketItem]) error

type GetCheckedFilepath func(prefix string, unchecked string) string

func GetCheckedFilepathBuilderNew(alt string) GetCheckedFilepath {
	return func(prefix, unchecked string) string {
		var clean string = filepath.Clean(unchecked)
		var safe bool = strings.HasPrefix(prefix, clean)
		if safe {
			return clean
		}
		return alt
	}
}

// GetCheckedFilepathDefault creates GetCheckedFilepath using empty path(which should be invalid).
var GetCheckedFilepathDefault GetCheckedFilepath = GetCheckedFilepathBuilderNew("")

func write2tmp(chk GetCheckedFilepath, dirname string, tmpname string, cb func(io.Writer) error) error {
	f, e := os.Create(chk(dirname, tmpname))
	if nil != e {
		return e
	}

	e = cb(f)
	if nil != e {
		_ = f.Close()
		return e
	}

	return f.Close()
}
