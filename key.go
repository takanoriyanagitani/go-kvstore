package kvstore

type Key struct {
	bucket Bucket
	id     []byte
}
