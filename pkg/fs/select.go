package fscommon

import (
	"context"

	kv "github.com/takanoriyanagitani/go-kvstore"
)

type FsSelect func(ctx context.Context, key FsKey) kv.Either[kv.Option[kv.Val], error]
