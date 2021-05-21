// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MI[]string license found in the LICENSE file.

package setord

import "github.com/apmckinlay/gsuneido/util/ints"

// this provides set operations on [][]string.
// Operations do not modify their inputs
// but may return an input unmodified, rather than a copy.
// Where applicable, the order of the values is maintained.

// Contains returns true if x contains s
func Contains(x [][]string, s []string) bool {
	// no point checking for sorted because that will do a full scan
	for _, xs := range x {
		if eq(xs, s) {
			return true
		}
	}
	return false
}

// Copy returns a copy of the set
func Copy(x [][]string) [][]string {
	z := make([][]string, len(x))
	copy(z, x)
	return z
}

// AddUnique appends s unless it is already in the set
//
// WARNING: it does append so usage must be: x = AddUnique(x, y)
func AddUnique(x [][]string, s []string) [][]string {
	if Contains(x, s) {
		return x
	}
	return append(x, s)
}

// Equal returns true if x and y contain the same set of strings.
//
// WARNING: requires that x does not contain duplicates
func Equal(x, y [][]string) bool {
	if len(x) != len(y) {
		return false
	}
	if Same(x, y) {
		return true
	}
outer:
	for _, xs := range x {
		for _, ys := range y {
			if eq(xs, ys) {
				continue outer
			}
		}
		return false // xs wasn't found in y
	}
	return true
}

func Same(x, y [][]string) bool {
	return len(x) > 0 && len(y) > 0 && len(x) == len(y) && &x[0] == &y[0]
}

// Union returns a combined list.
// If either input contains duplicates, so will the output.
//
// WARNING: If either x or y is empty, it returns the *original* of the other.
func Union(x, y [][]string) [][]string {
	if len(x) == 0 {
		return y[:len(y):len(y)] // so append won't share
	}
	if len(y) == 0 {
		return x[:len(x):len(x)] // so append won't share
	}
	z := make([][]string, 0, len(x)+len(y))
	z = append(z, x...)
outer:
	for _, ys := range y {
		for _, xs := range x {
			if eq(xs, ys) {
				continue outer
			}
		}
		z = append(z, ys)
	}
	return z
}

// Difference returns the elements of x that are not in y.
//
// WARNING: If y is empty, it returns the *original* x.
//
// WARNING: duplicates in the inputs may give duplicates in the result
func Difference(x, y [][]string) [][]string {
	if len(x) == 0 {
		return [][]string{}
	}
	if len(y) == 0 {
		return x[:len(x):len(x)] // so append won't share
	}
	if Same(x, y) {
		return [][]string{}
	}
	z := make([][]string, 0, len(x))
outer:
	for _, xs := range x {
		for _, ys := range y {
			if eq(xs, ys) {
				continue outer
			}
		}
		z = append(z, xs)
	}
	return z

}

// Intersect returns a list of the strings common to the inputs,
// the result is in the same order as the first argument (x).
//
// WARNING: If x and y are the same list, it returns the *original*.
//
// WARNING: duplicates in the inputs may give duplicates in the result
func Intersect(x, y [][]string) [][]string {
	if Same(x, y) {
		return x[:len(x):len(x)] // so append won't share
	}
	if len(x) == 0 || len(y) == 0 {
		return [][]string{}
	}
	z := make([][]string, 0, ints.Min(len(x), len(y))/2) // ???
outer:
	for _, xs := range x {
		for _, ys := range y {
			if eq(xs, ys) {
				z = append(z, xs)
				continue outer
			}
		}
	}
	return z
}

// Subset returns true is y is a subset of x
// i.e. x contains all of y
func Subset(x, y [][]string) bool {
outer:
	for _, ys := range y {
		for _, xs := range x {
			if eq(xs, ys) {
				continue outer
			}
		}
		return false // ys wasn't found in x
	}
	return true
}

// Disjoint returns true if x and y have no elements in common.
// i.e. Intersect(x, y) is empty
func Disjoint(x, y [][]string) bool {
	for _, xs := range x {
		for _, ys := range y {
			if eq(xs, ys) {
				return false
			}
		}
	}
	return true
}
