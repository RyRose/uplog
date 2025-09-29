package util

import (
	"net/url"
	"path"

	"golang.org/x/exp/constraints"
)

func UrlPathJoin(base string, parts ...string) string {
	p := []string{base}
	for _, part := range parts {
		p = append(p, url.PathEscape(part))
	}
	return path.Join(p...)
}

func Minimum[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Maximum[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Zero[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func DeZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
