package invoker

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mcrgnt/extractor"
)

type Builder interface {
	Build() (any, error)
}

type mockBuilder struct {
	data string
	err  error
}

func (m mockBuilder) Build() (any, error) {
	return m.data, m.err
}

func TestInvoker_Invoke(t *testing.T) {
	builderClue := func(b Builder) ([]any, error) {
		res, err := b.Build()
		if err != nil {
			return nil, err
		}
		return []any{res}, nil
	}

	t.Run("success", func(t *testing.T) {
		inv := New(builderClue)
		target := mockBuilder{data: "test-data"}

		got, err := inv.Invoke(target)

		if err != nil {
			t.Errorf("got error %v", err)
		}
		if !reflect.DeepEqual(got, []any{"test-data"}) {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("error", func(t *testing.T) {
		inv := New(builderClue)
		target := mockBuilder{err: errors.New("fail")}

		_, err := inv.Invoke(target)

		if err == nil || err.Error() != "fail" {
			t.Errorf("expected error 'fail', got %v", err)
		}
	})

	t.Run("nil_function", func(t *testing.T) {
		inv := New[Builder, any](nil)
		_, err := inv.Invoke(mockBuilder{})

		if err == nil || err.Error() != "clueFunc is not initialized" {
			t.Errorf("expected initialization error, got %v", err)
		}
	})
}

type stringBuilder struct{ val string }

func (s stringBuilder) Build() (any, error) {
	return s.val, nil
}

type ComplexTarget struct {
	Name   string
	BuildA stringBuilder
	BuildB stringBuilder
}

func TestInvoker_WithExtractor(t *testing.T) {
	extractorClue := func(target ComplexTarget) ([]any, error) {
		builders := extractor.New[Builder](target).Extract()

		var results []any
		for _, b := range builders {
			res, err := b.Build()
			if err == nil {
				results = append(results, res)
			}
		}
		return results, nil
	}

	t.Run("success", func(t *testing.T) {
		inv := New(extractorClue)
		target := ComplexTarget{
			BuildA: stringBuilder{"data-a"},
			BuildB: stringBuilder{"data-b"},
		}

		got, err := inv.Invoke(target)

		if err != nil {
			t.Fatalf("error: %v", err)
		}

		expected := []any{"data-a", "data-b"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})
}

func TestInvoker_WithFuncOutcome(t *testing.T) {
	type OutcomeFunc func() (string, bool, error)

	clue := func(target string) ([]OutcomeFunc, error) {
		f := func() (string, bool, error) {
			if target == "" {
				return "", false, errors.New("empty target")
			}
			return "processed:" + target, true, nil
		}
		return []OutcomeFunc{f}, nil
	}

	t.Run("execute outcome function", func(t *testing.T) {
		inv := New(clue)

		// Invoke возвращает слайс функций
		outcomes, err := inv.Invoke("go")
		if err != nil {
			t.Fatalf("invoke failed: %v", err)
		}

		if len(outcomes) != 1 {
			t.Fatalf("expected 1 function, got %d", len(outcomes))
		}

		val, ok, err := outcomes[0]()

		if err != nil {
			t.Errorf("outcome func returned error: %v", err)
		}
		if !ok || val != "processed:go" {
			t.Errorf("unexpected results: %v, %v", val, ok)
		}
	})
}
