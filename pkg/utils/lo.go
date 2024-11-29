package utils

import "github.com/samber/lo"

func FilterNonNil[T any](item T, _ int) bool {
	return !lo.IsNil(item)
}

func MapTypeAssert[F any, T any](item F, _ int) T {
	val, _ := any(item).(T)
	return val
}

func TypeAssertFrom[F any, T any](items []F) []T {
	filters := lo.Map(items, MapTypeAssert[F, T])
	return lo.Filter(filters, FilterNonNil)
}
