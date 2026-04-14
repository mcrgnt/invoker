package invoker

import "errors"

type ClueFunc[T any, O any] func(T) ([]O, error)

type invoker[T any, O any] struct {
	clue ClueFunc[T, O]
}

func New[T any, O any](f ClueFunc[T, O]) Invoker[T, O] {
	return &invoker[T, O]{clue: f}
}

func (i *invoker[T, O]) Invoke(target T) ([]O, error) {
	if i.clue == nil {
		return nil, errors.New("clueFunc is not initialized")
	}
	return i.clue(target)
}
