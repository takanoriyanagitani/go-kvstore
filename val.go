package kvstore

type Val struct {
	val []byte
}

func (v Val) Raw() []byte { return v.val }
