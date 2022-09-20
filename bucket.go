package kvstore

type Bucket func() string

type BucketItem struct {
	key Key
	val Val
}

func (b BucketItem) Key() Key { return b.key }
func (b BucketItem) Val() Val { return b.val }

func BucketItemNew(key Key, val Val) BucketItem {
	return BucketItem{
		key,
		val,
	}
}
