package kvstore

type Key struct {
	bucket Bucket
	id     []byte
}

func (k Key) Bucket() Bucket { return k.bucket }
func (k Key) Id() []byte     { return k.id }

func (k Key) BucketString() string { return k.Bucket()() }
