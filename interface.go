package invoker

type Invoker[T any, O any] interface {
	Invoke(T) ([]O, error)
}
