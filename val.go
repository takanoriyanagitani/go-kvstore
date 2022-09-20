package kvstore

type Val struct {
	val []byte
}

func ValNew(val []byte) Val {
	return Val{val}
}

func (v Val) Raw() []byte { return v.val }
