// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package runtime

import "github.com/apmckinlay/gsuneido/runtime/types"

// SuIter is a Value that wraps a runtime.Iter
// and provides the Suneido interator interface,
// returning itself when it reaches the end
type SuIter struct {
	CantConvert
	Iter
}

// Value interface --------------------------------------------------

var _ Value = (*SuIter)(nil)

func (SuIter) Call(*Thread, Value, *ArgSpec) Value {
	panic("can't call Iterator")
}

var IterMethods Methods

func (SuIter) Lookup(_ *Thread, method string) Callable {
	return IterMethods[method]
}

func (SuIter) Type() types.Type {
	return types.Iterator
}

func (it SuIter) String() string {
	return "/* iterator */"
}

func (it SuIter) Equal(other interface{}) bool {
	it2, ok := other.(SuIter)
	return ok && it2.Iter == it.Iter
}

func (SuIter) Get(*Thread, Value) Value {
	panic("iterator does not support get")
}

func (SuIter) Put(*Thread, Value, Value) {
	panic("iterator does not support put")
}

func (SuIter) RangeTo(int, int) Value {
	panic("iterator does not support range")
}

func (SuIter) RangeLen(int, int) Value {
	panic("iterator does not support range")
}

func (SuIter) Hash() uint32 {
	panic("iterator hash not implemented")
}

func (SuIter) Hash2() uint32 {
	panic("iterator hash not implemented")
}

func (SuIter) Compare(Value) int {
	panic("iterator compare not implemented")
}
