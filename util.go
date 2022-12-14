package kvstore

func Compose[T, U, V any](f func(T) U, g func(U) V) func(T) V {
	return func(t T) V {
		var u U = f(t)
		return g(u)
	}
}

func ComposeEither[T, U, V, E any](f func(T) Either[U, E], g func(U) Either[V, E]) func(T) Either[V, E] {
	return func(t T) Either[V, E] {
		var eu Either[U, E] = f(t)
		return EitherFlatMap(eu, g)
	}
}

func CoalesceError(e1 error, e2 error) error {
	if nil == e1 {
		return e2
	}
	return e1
}

func Identity[T any](t T) T { return t }

func MapGet[T comparable, U any](m map[T]U, t T) Option[U] {
	u, found := m[t]
	if found {
		return OptionNew(u)
	}
	return OptionEmpty[U]()
}

func Curry[T, U, V any](f func(T, U) V) func(T) func(U) V {
	return func(t T) func(U) V {
		return func(u U) V {
			return f(t, u)
		}
	}
}

func Error1st(ef []func() error) error {
	return IterReduce(IterFromArray(ef), nil, func(e error, f func() error) error {
		if nil == e {
			return f()
		}
		return e
	})
}

func Bool2ef(ok bool, ng func() error) func() error {
	if ok {
		return func() error { return nil }
	}
	return ng
}

func Bool2error(ok bool, ng func() error) error {
	return Bool2ef(ok, ng)()
}
