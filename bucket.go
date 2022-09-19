package kvstore

type Bucket func() string

type BucketItem struct {
	key Key
	val Val
}

func (b BucketItem) Key() Key { return b.key }
