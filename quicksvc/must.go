package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func must0(err error) {
	if err != nil {
		haltf("Error: %s", err)
	}
}

func halt(values ...any) {
	fmt.Fprint(os.Stderr, values...)
	fmt.Fprintln(os.Stderr)
	os.Exit(-1)
}

func haltf(format string, values ...any) {
	fmt.Fprintf(os.Stderr, format, values...)
	fmt.Fprintln(os.Stderr)
	os.Exit(-1)
}

func must[T any](value T, err error) T {
	must0(err)
	return value
}

type trial[T any] struct {
	value T
	err   error
}

func try[T any](value T, err error) trial[T] {
	return trial[T]{
		value: value,
		err:   err,
	}
}

func (t trial[T]) must() T {
	must0(t.err)
	return t.value
}

func (t trial[T]) withMessage(msg string) trial[T] {
	return trial[T]{
		value: t.value,
		err:   errors.WithMessage(t.err, msg),
	}
}

func (t trial[T]) replaceMessage(msg string) trial[T] {
	return trial[T]{
		value: t.value,
		err:   errors.New(msg),
	}
}

func (t trial[T]) replaceError(err error) trial[T] {
	if t.err == nil || err == nil {
		return t
	}
	return trial[T]{
		value: t.value,
		err:   err,
	}
}
