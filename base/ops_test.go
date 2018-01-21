package base

import (
	"testing"

	"github.com/apmckinlay/gsuneido/util/dnum"
	. "github.com/apmckinlay/gsuneido/util/hamcrest"
)

func TestCmp(t *testing.T) {
	vals := []Value{False, True, SuDnum{dnum.MinusInf}, SuInt(-1), SuInt(0), SuInt(+1), SuDnum{dnum.Inf},
		SuStr(""), SuStr("abc"), NewSuConcat().Add("foo"), SuStr("world")}
	for i := 1; i < len(vals); i++ {
		Assert(t).That(Cmp(vals[i], vals[i]), Equals(0))
		Assert(t).That(Cmp(vals[i-1], vals[i]), Equals(-1).Comment(vals[i-1], vals[i]))
		Assert(t).That(Cmp(vals[i], vals[i-1]), Equals(+1))
	}
}
