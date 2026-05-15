package util

import "context"

// Pointer returns a pointer to v. Useful for converting literals to *T.
func Pointer[T any](v T) *T {
	return &v
}

// ConvertArray maps each element of arr through callback and returns the results.
// The context is forwarded to the callback so response builders can lazily load
// relations when needed.
func ConvertArray[K any, T any](ctx context.Context, arr []K, callback func(ctx context.Context, k K) T) []T {
	nodes := []T{}
	for _, v := range arr {
		nodes = append(nodes, callback(ctx, v))
	}
	return nodes
}

// SliceValueToSlicePointer converts a slice of values to a slice of pointers.
// Each pointer points into the original slice, so mutations through the pointer
// are reflected in the original slice.
func SliceValueToSlicePointer[T any](sliceValue []T) []*T {
	slicePointer := make([]*T, len(sliceValue))
	for i := range sliceValue {
		slicePointer[i] = &sliceValue[i]
	}
	return slicePointer
}
