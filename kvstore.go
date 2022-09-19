package kvstore

import (
	"context"
)

type Create func(ctx context.Context, bucket Bucket) error
type Select func(ctx context.Context, key Key) Either[Option[Val], error]
type Upsert func(ctx context.Context, item BucketItem) error
type Insert func(ctx context.Context, item BucketItem) error
type Delete func(ctx context.Context, key Key) error

// BulkUpsert upserts items into single bucket.
type BulkUpsert func(ctx context.Context, bucket Bucket, items Iter[BucketItem]) error

func NonAtomicUpsertBuilder(sel Select) func(Delete) func(Insert) Upsert {
	return func(del Delete) func(Insert) Upsert {
		return func(ins Insert) Upsert {
			return func(ctx context.Context, item BucketItem) error {
				var eov Either[Option[Val], error] = sel(ctx, item.Key())
				return eov.TryForEach(func(ov Option[Val]) error {
					var k Option[Key] = OptionMap(ov, func(_ Val) Key { return item.Key() })
					var oe Option[error] = OptionMap(k, func(k Key) error {
						return del(ctx, k)
					})
					var e error = oe.UnwrapOr(nil)
					var ei Either[BucketItem, error] = EitherNew(item, e)
					var ee Either[error, error] = EitherMap(ei, func(_ BucketItem) error {
						return ins(ctx, item)
					})
					return ee.UnwrapOrElse(Identity[error])
				})
			}
		}
	}
}
