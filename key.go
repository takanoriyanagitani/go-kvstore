package kvstore

type Key struct {
	bucket Bucket
	id     []byte
}

func KeyNew(bucket Bucket, id []byte) Key {
	return Key{
		bucket,
		id,
	}
}

func (k Key) Bucket() Bucket { return k.bucket }
func (k Key) Id() []byte     { return k.id }

func (k Key) BucketString() string { return k.Bucket()() }
