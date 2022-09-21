package kvstore

type Either[T, E any] interface {
	IsOk() bool
	IsNg() bool
	TryForEach(f func(T) E) E
	Ok() Option[T]
	UnwrapOrElse(func(E) T) T
	Left() Option[E]
	Right() Option[T]
	FlatMap(func(T) Either[T, E]) Either[T, E]
	Map(func(T) T) Either[T, E]
}

type eitherRight[T, E any] struct {
	t T
}

func EitherRight[T, E any](t T) Either[T, E] {
	return eitherRight[T, E]{t}
}

func (r eitherRight[T, E]) IsOk() bool                                  { return true }
func (r eitherRight[T, E]) IsNg() bool                                  { return false }
func (r eitherRight[T, E]) TryForEach(f func(T) E) E                    { return f(r.t) }
func (r eitherRight[T, E]) Ok() Option[T]                               { return OptionNew(r.t) }
func (r eitherRight[T, E]) UnwrapOrElse(_ func(E) T) T                  { return r.t }
func (r eitherRight[T, E]) Left() Option[E]                             { return OptionEmpty[E]() }
func (r eitherRight[T, E]) Right() Option[T]                            { return OptionNew(r.t) }
func (r eitherRight[T, E]) UnwrapOr(_ T) T                              { return r.t }
func (r eitherRight[T, E]) FlatMap(f func(T) Either[T, E]) Either[T, E] { return f(r.t) }
func (r eitherRight[T, E]) Map(f func(T) T) Either[T, E] {
	var mapd T = f(r.t)
	return EitherRight[T, E](mapd)
}

func EitherMap[T, U, E any](r Either[T, E], f func(T) U) Either[U, E] {
	return EitherFlatMap(r, func(t T) Either[U, E] {
		var u U = f(t)
		return EitherRight[U, E](u)
	})
}

func EitherFlatMap[T, U, E any](r Either[T, E], f func(T) Either[U, E]) Either[U, E] {
	switch lr := r.(type) {
	case eitherRight[T, E]:
		var t T = lr.t
		return f(t)
	case eitherLeft[T, E]:
		return EitherLeft[U, E](lr.e)
	}
	var e E
	return EitherLeft[U, E](e)
}

type eitherLeft[T, E any] struct {
	e E
}

func EitherLeft[T, E any](e E) Either[T, E] {
	return eitherLeft[T, E]{e}
}

func (l eitherLeft[T, E]) IsOk() bool                                  { return false }
func (l eitherLeft[T, E]) IsNg() bool                                  { return true }
func (l eitherLeft[T, E]) TryForEach(_ func(T) E) E                    { return l.e }
func (l eitherLeft[T, E]) Ok() Option[T]                               { return OptionEmpty[T]() }
func (l eitherLeft[T, E]) UnwrapOrElse(f func(E) T) T                  { return f(l.e) }
func (l eitherLeft[T, E]) Left() Option[E]                             { return OptionNew(l.e) }
func (l eitherLeft[T, E]) Right() Option[T]                            { return OptionEmpty[T]() }
func (l eitherLeft[T, E]) UnwrapOr(alt T) T                            { return alt }
func (l eitherLeft[T, E]) FlatMap(_ func(T) Either[T, E]) Either[T, E] { return l }
func (l eitherLeft[T, E]) Map(_ func(T) T) Either[T, E]                { return l }

func EitherOk[T any](t T) Either[T, error]     { return EitherRight[T, error](t) }
func EitherNg[T any](e error) Either[T, error] { return EitherLeft[T, error](e) }

func EitherNew[T any](t T, e error) Either[T, error] {
	if nil == e {
		return EitherRight[T, error](t)
	}
	return EitherLeft[T, error](e)
}

func eitherNg[T, E any](left eitherLeft[T, E], okf func(E) (ok bool)) Either[Option[T], E] {
	var e E = left.e
	var ok bool = okf(e)
	if ok {
		return EitherRight[Option[T], E](OptionEmpty[T]())
	}
	return EitherLeft[Option[T], E](e)
}

func EitherOkWhen[T, E any](e Either[T, E], okf func(E) (ok bool)) Either[Option[T], E] {
	switch lr := e.(type) {
	case eitherRight[T, E]:
		return EitherRight[Option[T], E](OptionNew(lr.t))
	case eitherLeft[T, E]:
		return eitherNg(lr, okf)
	}
	var err E
	return EitherLeft[Option[T], E](err)
}
